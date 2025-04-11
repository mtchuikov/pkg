package interceptors

import "errors"

var ErrMaxOpenStreamsReached = errors.New("maximim open streams reached")
