package chsubscription

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testChSubscription struct {
	suite.Suite
	ctx     context.Context
	val     int
	bufSize int
	chsub   *ChSubscription[int]
}

func TestChSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(testChSubscription))
}

func (s *testChSubscription) SetupTest() {
	s.ctx = context.Background()
	s.val = 41
	s.bufSize = 3
	s.chsub = New[int]()
}

func (s *testChSubscription) TearDownSuite() {
	s.chsub.Close()
}

func (s *testChSubscription) TestSubscribe() {
	ch := s.chsub.Subscribe(s.bufSize)

	errMsg := "channel must be non-nil "
	s.Require().NotNil(ch, errMsg)
}

func (s *testChSubscription) TestNotify() {
	ch1 := s.chsub.Subscribe(s.bufSize)
	ch2 := s.chsub.Subscribe(s.bufSize)

	s.chsub.Notify(s.ctx, s.val)

	testCh := func(sub <-chan int) {
		select {
		case val, closed := <-sub:
			errMsg := "channel must to be open"
			s.Require().True(closed, errMsg)

			errMsg = "must receive %s through channel, got %d"
			s.Require().Equalf(s.val, val, errMsg, s.val, val)
		default:
			errMsg := "channel must contain value"
			s.Require().FailNow(errMsg)
		}
	}

	testCh(ch1)
	testCh(ch2)
}

func (s *testChSubscription) TestUnsubscribe() {
	ch := s.chsub.Subscribe(0)

	s.chsub.Unsubscribe(ch)

	errMsg := "items must to be empty after unsubscribe"
	s.Require().Len(s.chsub.items, 0, errMsg)

	_, closed := <-ch
	errMsg = "channel must to be closed"
	s.Require().False(closed)
}

func (s *testChSubscription) TestClose() {
	ch1 := s.chsub.Subscribe(s.bufSize)
	ch2 := s.chsub.Subscribe(s.bufSize)

	s.chsub.Close()

	_, closed := <-ch1
	errMsg := "channel ch1 must to be closed"
	s.Require().False(closed, errMsg)

	_, closed = <-ch2
	errMsg = "channel ch2 must to be closed"
	s.Require().False(closed, errMsg)
}
