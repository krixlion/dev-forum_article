package log

import (
	"go.uber.org/zap"
)

// MakeZapLogger returns an error on hardware error.
func MakeZapLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := logger.Sugar()
	defer sugar.Sync()
	return sugar, err
}

func PrintLn(msg string, keyvals ...interface{}) {
	logger, _ := MakeZapLogger()
	logger.Infow(msg, keyvals...)
}
