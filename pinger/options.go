package pinger

import "time"

type Option func(*Pinger)

func WithPingInterval(interval time.Duration) Option {
	return func(p *Pinger) {
		p.ticker = time.NewTicker(interval)
	}
}

func WithBackoffInterval(interval time.Duration) Option {
	return func(p *Pinger) {
		p.backoff.Interval = interval
	}
}

func WithBackoffMax(interval time.Duration) Option {
	return func(p *Pinger) {
		p.backoff.Max = interval
	}
}
