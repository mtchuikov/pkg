package closer

type Option func(*Closer)

func WithMaxConcurrent(max int) Option {
	return func(c *Closer) {
		c.maxConcurrent = max
	}
}
