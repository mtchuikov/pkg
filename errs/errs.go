package errs

import (
	"fmt"
)

type Error struct {
	Msg string
	Err error
}

func (e *Error) StringWrap() string {
	return fmt.Sprintf("%v: %v", e.Msg, e.Err)
}

func (e *Error) String() string {
	return e.Err.Error()
}
