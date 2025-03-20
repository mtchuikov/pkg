package logtools

import "errors"

var (
	ErrInvalidLogLevel     = errors.New("invalid log level")
	ErrFailedToOpenLogFile = errors.New("failed to open log file")
)
