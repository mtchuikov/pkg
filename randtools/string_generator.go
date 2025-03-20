package randtools

import (
	cryptorand "crypto/rand"
	"math"
	"math/rand/v2"
	"sync"
	"unsafe"
)

const (
	DefaultAlphabet    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	DefaultAlphabetLen = 62
)

type StringGenerator struct {
	mu              sync.Mutex
	source          *rand.Rand
	alphabet        string
	alphabetLen     int
	letterIndexBits int
	letterIndexMask uint64
	letterIndexMax  int
}

type StringGeneratorOption func(*StringGenerator)

func WithAlphabet(alphabet string) StringGeneratorOption {
	return func(g *StringGenerator) {
		g.alphabet = alphabet
		g.alphabetLen = len(g.alphabet)
	}
}

// NewStringGenerator creates and returns a *StringGenerator
// object designed for generating random strings.
// The generator's configuration can be customized using
// optional parameters. If no alphabet is specified, a
// default alphabet with 62 characters (uppercase and lowercase
// Latin letters, and digits) is used.
//
// Note: if you want to use a custom alphabet, ensure it has
// sufficient length to provide variety in generated strings.
func NewStringGenerator(opts ...StringGeneratorOption) *StringGenerator {
	var seed [32]byte
	cryptorand.Read(seed[:])

	chaCha8 := rand.NewChaCha8(seed)
	source := rand.New(chaCha8)

	g := &StringGenerator{
		mu:     sync.Mutex{},
		source: source,
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.alphabet == "" {
		g.alphabet = DefaultAlphabet
		g.alphabetLen = DefaultAlphabetLen
	}

	// the smallest k such that 2^k >= alphabetLen
	letterIndexBits := int(math.Ceil(math.Log2(float64(g.alphabetLen))))

	// 2^k - 1
	letterIndexMask := uint64(1<<letterIndexBits - 1)

	// number of indexes that can be extracted from a 64-bit
	// number
	letterIndexMax := 64 / letterIndexBits

	g.letterIndexBits = letterIndexBits
	g.letterIndexMask = letterIndexMask
	g.letterIndexMax = letterIndexMax

	return g
}

// GenerateString produces a random string of a specified length
// using characters from a provided alphabet. It utilizes a
// source of randomness to select characters. This function is
// safe for concurrent use.
//
// The principle of operation of this function was proposed by
// icza on StackOverflow https://stackoverflow.com/a/31832326
func (g *StringGenerator) Generate(stringLen int) string {
	result := make([]byte, stringLen)

	g.mu.Lock()
	cache := g.source.Uint64()
	g.mu.Unlock()

	remain := g.letterIndexMax

	i := stringLen - 1
	for i >= 0 {
		if remain == 0 {
			g.mu.Lock()
			cache = g.source.Uint64()
			g.mu.Unlock()

			remain = g.letterIndexMax
		}

		idx := int(cache & g.letterIndexMask)
		if idx < g.alphabetLen {
			result[i] = g.alphabet[idx]
			i--
		}

		cache >>= g.letterIndexBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&result))
}
