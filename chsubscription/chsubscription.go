package chsubscription

import (
	"context"
	"slices"
	"sync"
)

type ChSubscription[ChT any] struct {
	mu    sync.Mutex
	items []chan ChT
}

func New[ChT any]() *ChSubscription[ChT] {
	return &ChSubscription[ChT]{
		mu:    sync.Mutex{},
		items: make([]chan ChT, 0, 3),
	}
}

func (s *ChSubscription[ChT]) Subscribe(bufSize int) <-chan ChT {
	ch := make(chan ChT, bufSize)

	s.mu.Lock()
	s.items = append(s.items, ch)
	s.mu.Unlock()

	return ch
}

func (s *ChSubscription[ChT]) Notify(ctx context.Context, val ChT) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.items {
		select {
		case <-ctx.Done():
			return
		case item <- val:
		default:
		}
	}
}

func (s *ChSubscription[ChT]) Unsubscribe(item <-chan ChT) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, i := range s.items {
		if i == item {
			close(i)
			s.items = slices.Delete(s.items, idx, idx+1)
			break
		}
	}
}

func (s *ChSubscription[ChT]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, i := range s.items {
		close(i)
	}

	s.items = nil
}
