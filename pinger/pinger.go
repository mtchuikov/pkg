package pinger

import (
	"context"
	"sync"
	"time"

	"github.com/mtchuikov/pkg/backoff"
	"github.com/mtchuikov/pkg/chsubscription"
	"github.com/rs/zerolog"
)

const (
	DefaultPingInterval    = 3 * time.Second
	DefaultBackoffInterval = 2 * time.Second
	DefaultBackoffMax      = 6 * time.Second
)

type pinger interface {
	Ping(ctx context.Context) error
}

type Pinger struct {
	log       zerolog.Logger
	mu        sync.Mutex
	pinger    pinger
	ticker    *time.Ticker
	backoff   *backoff.Backoff
	err       error
	chsub     *chsubscription.ChSubscription[error]
	closeOnce sync.Once
}

func New(log zerolog.Logger, pinger pinger, opts ...Option) *Pinger {
	p := &Pinger{
		log:       log,
		mu:        sync.Mutex{},
		pinger:    pinger,
		ticker:    time.NewTicker(DefaultPingInterval),
		err:       nil,
		chsub:     chsubscription.New[error](),
		closeOnce: sync.Once{},
	}

	backoff := backoff.New()
	backoff.Interval = DefaultBackoffInterval
	backoff.Max = DefaultBackoffMax

	p.backoff = backoff

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Pinger) Ping(ctx context.Context, timeout time.Duration) {
	for {
		select {
		case <-ctx.Done():
			p.log.Debug().
				Msg("context cancelled, stopping ping loop")
			return
		case <-p.ticker.C:
			p.log.Debug().
				Msg("ticker ticked, starting ping")

			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)

			p.mu.Lock()
			err := p.pinger.Ping(timeoutCtx)
			p.mu.Unlock()

			cancel()

			if err != nil {
				p.log.Debug().
					Err(err).
					Msg("failed to ping")

				p.err = err
				p.chsub.Notify(ctx, err)

				delay := p.backoff.Next()
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
					continue
				}
			}

			p.log.Debug().
				Msg("ping successful")
			p.backoff.Reset()
		}
	}
}

func (p *Pinger) ChangePinger(pinger pinger) {
	p.mu.Lock()
	p.pinger = pinger
	p.mu.Unlock()
}

func (p *Pinger) Error() error {
	return p.err
}

func (p *Pinger) Subscribe() <-chan error {
	return p.chsub.Subscribe(3)
}

func (p *Pinger) Unsubscribe(item <-chan error) {
	p.chsub.Unsubscribe(item)
}

func (p *Pinger) Close(ctx context.Context) error {
	closeFn := func() {
		p.ticker.Stop()
		p.chsub.Close()
	}

	p.closeOnce.Do(closeFn)

	return nil
}
