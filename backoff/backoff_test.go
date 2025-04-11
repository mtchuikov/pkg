package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type testBackoff struct {
	suite.Suite
	backoff *Backoff
}

func TestBackoffSuite(t *testing.T) {
	suite.Run(t, new(testBackoff))
}

func (s *testBackoff) SetupTest() {
	s.backoff = New()
	s.backoff.RandFactor = 0
}

func (s *testBackoff) TestMaxOverflow() {
	s.backoff.Multiplier = 3

	dur := time.Duration(s.backoff.Multiplier)
	s.backoff.Current = s.backoff.Max / dur

	s.backoff.incrementCurrent()

	errMsg := "'current' should be to equal 'max' when exceeding threshold"
	s.Require().Equal(s.backoff.Max, s.backoff.Current, errMsg)
}

func (s *testBackoff) TestNext() {
	next := s.backoff.Next()

	errMsg := "next must return 'current' value when 'rand factor' == 0"
	s.Require().Equal(s.backoff.Current, next, errMsg)
}

func (s *testBackoff) TestReset() {
	s.backoff.Next()
	s.backoff.Reset()

	errMsg := "reset the 'current' value to 'initial interval'"
	s.Require().Equal(s.backoff.Interval, s.backoff.Current, errMsg)
}
