package backoff

import (
	"math/rand/v2"
	"time"
)

const (
	DefaultInterval    = 500 * time.Millisecond
	DefaultMaxInterval = 15 * time.Second
	DefaultMultiplier  = 1.5
	DefaultRandFactor  = 0.5
)

type Backoff struct {
	Interval   time.Duration
	Max        time.Duration
	Current    time.Duration
	Multiplier float64
	RandFactor float64
}

func New() *Backoff {
	backoff := &Backoff{
		Interval:   DefaultInterval,
		Max:        DefaultMaxInterval,
		Current:    0,
		Multiplier: DefaultMultiplier,
		RandFactor: DefaultRandFactor,
	}

	return backoff
}

func (b *Backoff) randTime(random float64) time.Duration {
	if b.RandFactor == 0 {
		return b.Current
	}

	current := float64(b.Current)
	delta := b.RandFactor * float64(b.Current)

	min := current - delta
	max := current + delta

	return time.Duration(max + (random * (max - min + 1)))
}

func (b *Backoff) incrementCurrent() {
	current := float64(b.Current)
	if current >= float64(b.Max)/b.Multiplier {
		b.Current = b.Max
		return
	}

	b.Interval = time.Duration(current * b.Multiplier)
}

func (b *Backoff) Next() time.Duration {
	if b.Current == 0 {
		b.Current = b.Interval
	}

	next := b.randTime(rand.Float64())
	b.incrementCurrent()

	return next
}

func (b *Backoff) Reset() {
	b.Current = b.Interval
}
