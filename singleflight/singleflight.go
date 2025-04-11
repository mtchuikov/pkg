package singleflight

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var errGoexit = errors.New("runtime.Goexit was called")

type panicError struct {
	value any
	stack []byte
}

func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.value, p.stack)
}

func newPanicError(v any) error {
	stack := debug.Stack()

	line := bytes.IndexByte(stack[:], '\n')
	if line >= 0 {
		stack = stack[line+1:]
	}

	return &panicError{
		value: v,
		stack: stack,
	}
}

type Result[V any] struct {
	Val    V
	Err    error
	Shared bool
}

type call[V any] struct {
	wg sync.WaitGroup

	val V
	err error

	dups  int
	chans []chan<- Result[V]

	cancel     context.CancelFunc
	ctxWaiters atomic.Int64
}

type Group[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]*call[V]
}

func NewGroup[K comparable, V any]() *Group[K, V] {
	return &Group[K, V]{
		mu: sync.Mutex{},
		m:  make(map[K]*call[V], 0),
	}
}

func (g *Group[K, V]) waitCtx(ctx context.Context, c *call[V], result <-chan Result[V], output chan<- Result[V]) {
	var res Result[V]
	select {
	case <-ctx.Done():
	case res = <-result:
	}

	if c.ctxWaiters.Add(-1) == 0 {
		c.cancel()
		c.wg.Wait()
	}

	err := ctx.Err()
	if err != nil {
		res = Result[V]{Err: err}
	}

	output <- res
}

func (g *Group[K, V]) doCall(c *call[V], key K, fn func() (V, error)) {
	normalReturn := false
	recovered := false
	defer func() {
		if !normalReturn && !recovered {
			c.err = errGoexit
		}

		g.mu.Lock()
		defer g.mu.Unlock()

		c.wg.Done()
		if g.m[key] == c {
			delete(g.m, key)
		}

		e, ok := c.err.(*panicError)
		if ok {
			if len(c.chans) > 0 {
				go panic(e)
				select {}
			} else {
				panic(e)
			}
		} else if c.err == errGoexit {
		} else {
			for _, ch := range c.chans {
				ch <- Result[V]{
					Val:    c.val,
					Err:    c.err,
					Shared: c.dups > 0,
				}
			}
		}
	}()

	func() {
		defer func() {
			if !normalReturn {
				recovered := recover()
				if recovered != nil {
					c.err = newPanicError(recovered)
				}
			}
		}()

		c.val, c.err = fn()
		normalReturn = true
	}()

	if !normalReturn {
		recovered = true
	}
}

func (g *Group[K, V]) Do(key K, fn func() (V, error)) (v V, err error, shared bool) {
	g.mu.Lock()

	if c, ok := g.m[key]; ok {
		c.dups++
		g.mu.Unlock()
		c.wg.Wait()

		if e, ok := c.err.(*panicError); ok {
			panic(e)
		} else if c.err == errGoexit {
			runtime.Goexit()
		}

		return c.val, c.err, true
	}
	c := new(call[V])
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	g.doCall(c, key, fn)
	return c.val, c.err, c.dups > 0
}

func (g *Group[K, V]) DoChan(key K, fn func() (V, error)) <-chan Result[V] {
	ch := make(chan Result[V], 1)
	g.mu.Lock()

	if c, ok := g.m[key]; ok {
		c.dups++
		c.chans = append(c.chans, ch)

		g.mu.Unlock()

		return ch
	}

	c := &call[V]{chans: []chan<- Result[V]{ch}}
	c.wg.Add(1)
	g.m[key] = c

	g.mu.Unlock()

	go g.doCall(c, key, fn)

	return ch
}

func (g *Group[K, V]) DoChanContext(ctx context.Context, key K, fn func(context.Context) (V, error)) <-chan Result[V] {
	ch := make(chan Result[V], 1)
	g.mu.Lock()

	c, ok := g.m[key]
	if ok {
		c.dups++
		c.ctxWaiters.Add(1)
		c.chans = append(c.chans, ch)
		g.mu.Unlock()

	} else {
		callCtx, callCancel := context.WithCancel(context.Background())
		c = &call[V]{
			chans:  []chan<- Result[V]{ch},
			cancel: callCancel,
		}

		c.wg.Add(1)
		c.ctxWaiters.Add(1)
		g.m[key] = c

		g.mu.Unlock()

		go g.doCall(c, key, func() (V, error) {
			return fn(callCtx)
		})
	}

	final := make(chan Result[V], 1)
	go g.waitCtx(ctx, c, ch, final)

	return final
}

func (g *Group[K, V]) Forget(key K) {
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}
