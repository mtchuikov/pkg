package strgen

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

type Generator struct {
	mu     sync.Mutex
	source *rand.Rand

	alphabet    string
	alphabetLen int

	letterIdxBits int
	letterIdxMask uint64
	letterIdxMax  int
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
func New(opts ...Option) *Generator {
	var seed [32]byte
	cryptorand.Read(seed[:])

	chaCha8 := rand.NewChaCha8(seed)
	source := rand.New(chaCha8)

	g := &Generator{
		mu:          sync.Mutex{},
		source:      source,
		alphabet:    DefaultAlphabet,
		alphabetLen: DefaultAlphabetLen,
	}

	for _, opt := range opts {
		opt(g)
	}

	// the smallest k such that 2^k >= alphabetLen
	letterIndexBits := int(math.Ceil(math.Log2(float64(g.alphabetLen))))

	// 2^k - 1
	letterIndexMask := uint64(1<<letterIndexBits - 1)

	// number of indexes that can be extracted from a 64-bit
	// number
	letterIndexMax := 64 / letterIndexBits

	g.letterIdxBits = letterIndexBits
	g.letterIdxMask = letterIndexMask
	g.letterIdxMax = letterIndexMax

	return g
}

// GenerateString produces a random string of a specified length
// using characters from a provided alphabet. It utilizes a
// source of randomness to select characters. This function is
// safe for concurrent use.
//
// The principle of operation of this function was proposed by
// icza on StackOverflow https://stackoverflow.com/a/31832326
func (g *Generator) Generate(stringLen int) string {
	result := make([]byte, stringLen)

	g.mu.Lock()
	cache := g.source.Uint64()
	g.mu.Unlock()

	remain := g.letterIdxMax

	i := stringLen - 1
	for i >= 0 {
		if remain == 0 {
			g.mu.Lock()
			cache = g.source.Uint64()
			g.mu.Unlock()

			remain = g.letterIdxMax
		}

		idx := int(cache & g.letterIdxMask)
		if idx < g.alphabetLen {
			result[i] = g.alphabet[idx]
			i--
		}

		cache >>= g.letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&result))
}
