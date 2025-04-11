package logging

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

func NewZerolog(appName string, output ...io.Writer) zerolog.Logger {
	zerolog.LevelFieldName = LevelFieldName
	zerolog.ErrorFieldName = ErrorFieldName
	zerolog.MessageFieldName = MessageFieldName
	zerolog.TimeFieldFormat = TimeFieldFormat

	return zerolog.New(os.Stdout).With().
		Str("app", appName).Timestamp().
		Logger()
}
