package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

func GetFxLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: log}
}
