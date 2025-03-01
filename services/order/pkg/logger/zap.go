package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BaseLogger interface {
	WithOptions(opts ...zap.Option) *zap.Logger
	With(fields ...zap.Field) *zap.Logger

	Debug(msg string, field ...zap.Field)
	Info(msg string, field ...zap.Field)
	Warn(msg string, field ...zap.Field)
	Error(msg string, field ...zap.Field)
	Panic(msg string, field ...zap.Field)
}

var ZapEntry = "inventory-zap"

func logLevel(lvl string) zapcore.Level {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	default:
		return zap.ErrorLevel
	}
}

func NewZapLogger(logLvl string, outputPaths, errOutputPaths []string) BaseLogger {
	zapConfig := zap.NewProductionConfig()

	encoder := zap.NewProductionEncoderConfig()

	zapConfig.EncoderConfig = encoder
	zapConfig.Level = zap.NewAtomicLevelAt(logLevel(logLvl))
	zapConfig.Development = true                       // TODO убрать хардкод
	zapConfig.Encoding = "json"                        // TODO хардкод
	zapConfig.InitialFields = map[string]interface{}{} // TODO
	zapConfig.OutputPaths = outputPaths
	zapConfig.ErrorOutputPaths = errOutputPaths
	logger, err := zapConfig.Build()

	if err != nil {
		panic(err)
	}

	return logger
}
