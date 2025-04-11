package strgen

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testGeneratorSuite struct {
	suite.Suite

	length    int
	generator *Generator
}

func TestGenerator(t *testing.T) {
	suite.Run(t, new(testGeneratorSuite))
}

func (s *testGeneratorSuite) SetupTest() {
	s.length = 10
	s.generator = New()
}

func (s *testGeneratorSuite) TestGenerate() {
	str := s.generator.Generate(s.length)

	errMsg := "expected and generated length doesn't match"
	s.Require().Len(str, s.length, errMsg)
}

func (s *testGeneratorSuite) TestGenerate_CustomAlphabet() {
	const alphabet = "abc"
	s.generator = New(WithAlphabet(alphabet))

	str := s.generator.Generate(s.length)
	for _, char := range str {
		symbol := string(char)

		errMsg := "generated string contains unexpected symbol"
		s.Require().Contains(alphabet, symbol, errMsg)
	}
}

func (s *testGeneratorSuite) TestGenerator_Concurrent() {
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

			str := s.generator.Generate(10)
			_, exists := strs.Load(str)

			errMsg := "generated string must be unique"
			s.Require().False(exists, errMsg)

			strs.Store(str, struct{}{})
		}()
	}

	wg.Wait()
}
