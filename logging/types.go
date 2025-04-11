package logging

import (
	"errors"

	"github.com/rs/zerolog"
)

type LogLevel string

const (
	Debug = "debug"
	Info  = "info"
	Warn  = "warn"
	Error = "error"
	Fatal = "fatal"
	Panic = "panic"
)

var ErrInvalidLogLevel = errors.New("invalid log level")

func (l LogLevel) validate() error {
	if l == Debug || l == Info || l == Warn ||
		l == Error || l == Fatal || l == Panic {
		return nil
	}

	return ErrInvalidLogLevel
}

func NewLevel(level string) (LogLevel, error) {
	lvl := LogLevel(level)
	err := lvl.validate()
	if err != nil {
		return "", err
	}

	return lvl, nil
}

func (l LogLevel) ToZerolog() (zerolog.Level, error) {
	switch l {
	case Debug:
		return zerolog.DebugLevel, nil
	case Info:
		return zerolog.InfoLevel, nil
	case Warn:
		return zerolog.WarnLevel, nil
	case Error:
		return zerolog.ErrorLevel, nil
	case Fatal:
		return zerolog.FatalLevel, nil
	case Panic:
		return zerolog.PanicLevel, nil
	default:
		return -1, ErrInvalidLogLevel
	}
}
