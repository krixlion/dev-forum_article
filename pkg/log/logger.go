package log

import (
	"os"

	"github.com/go-kit/log"
)

var logger log.Logger = MakeLogger()

func MakeLogger() log.Logger {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return logger
}

func PrintLn(keyvals ...interface{}) error {
	err := logger.Log(keyvals...)
	return err
}
