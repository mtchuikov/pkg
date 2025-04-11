package strgen

import (
	"math/rand/v2"
	"testing"
)

func BenchmarkSimpleRand(b *testing.B) {
	runeAlphabet := []rune(DefaultAlphabet)

	for range 10000 {
		b := make([]rune, 32)
		for i := range b {
			b[i] = runeAlphabet[rand.IntN(DefaultAlphabetLen)]
		}

		_ = string(b)
	}
}

func BenchmarkGenerator(b *testing.B) {
	generage := New()

	for range 10000 {
		_ = generage.Generate(32)
	}
}
