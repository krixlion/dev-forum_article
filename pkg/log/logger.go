package log

import (
	"go.uber.org/zap"
)

type Logger interface {
	Log(msg string, keyvals ...interface{})
}

// Log implements Logger
type log struct {
	logger *zap.SugaredLogger
}

// NewLogger returns an error on hardware error.
func NewLogger() (Logger, error) {
	logger, err := zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := logger.Sugar()
	defer sugar.Sync()

	return log{
		logger: logger.Sugar(),
	}, err
}

func (log log) Log(msg string, keyvals ...interface{}) {
	log.logger.Infow(msg, keyvals...)
}
