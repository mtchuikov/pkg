package randtools

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringGenerator_StringLength(t *testing.T) {
	generator := NewStringGenerator()

	const length = 10
	str := generator.Generate(length)
	strLen := len(str)

	errMsg := fmt.Sprintf("expected string length to be %v, got %v",
		length, strLen)
	require.Equal(t, length, strLen, errMsg)
}

func TestStringGenerator_Alphabet(t *testing.T) {
	const alphabet = "abc"
	generator := NewStringGenerator(WithAlphabet(alphabet))

	str := generator.Generate(100)
	for _, char := range str {
		symbol := string(char)

		errMsg := fmt.Sprintf("expected %v chat to be in alphabet %s", symbol, alphabet)
		require.Contains(t, alphabet, symbol, errMsg)
	}
}

func TestStringGenerator_Concurrency(t *testing.T) {
	generator := NewStringGenerator()

	var strs sync.Map

	var wg sync.WaitGroup
	queue := make(chan struct{}, 10)
	defer close(queue)

	for range 100 {
		wg.Add(1)
		queue <- struct{}{}

		go func() {
			defer func() {
				wg.Done()
				<-queue
			}()

			str := generator.Generate(10)
			_, exists := strs.Load(str)

			errMsg := fmt.Sprintf("expected unique string, but got duplicate: %s", str)
			require.False(t, exists, errMsg)

			strs.Store(str, struct{}{})
		}()
	}

	wg.Wait()
}
