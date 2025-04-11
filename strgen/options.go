package strgen

type Option func(*Generator)

func WithAlphabet(alphabet string) Option {
	return func(g *Generator) {
		g.alphabet = alphabet
		g.alphabetLen = len(g.alphabet)
	}
}
