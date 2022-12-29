package logging

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

var global log

func init() {
	l, err := NewLogger()
	if err != nil {
		panic(err)
	}

	global = l.(log)
}

func Log(msg string, keyvals ...interface{}) {
	global.l.Infow(msg, keyvals...)
}

type Logger interface {
	Log(ctx context.Context, msg string, keyvals ...interface{})
}

// Log implements Logger
type log struct {
	l *otelzap.SugaredLogger
}

// NewLogger returns an error on hardware error.
func NewLogger() (Logger, error) {
	logger, err := zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(2))
	otelLogger := otelzap.New(logger)
	sugar := otelLogger.Sugar()
	defer sugar.Sync()

	return log{
		l: sugar,
	}, err
}

func (log log) Log(ctx context.Context, msg string, keyvals ...interface{}) {
	log.l.InfowContext(ctx, msg, keyvals...)
}
