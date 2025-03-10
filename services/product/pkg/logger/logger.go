package logger

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Panic(msg string, fields ...any)

	Log(ctx context.Context, level Level, msg string, fields ...any)
	With(fields ...any) Logger
	Named(name string) Logger
}

type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic = 4
)

type zapLogger struct {
	logger *zap.SugaredLogger
}

func (zl *zapLogger) Debug(msg string, fields ...any) {
	zl.logger.Debugw(msg, fields...)
}

func (zl *zapLogger) Info(msg string, fields ...any) {
	zl.logger.Infow(msg, fields...)
}

func (zl *zapLogger) Warn(msg string, fields ...any) {
	zl.logger.Warnw(msg, fields...)
}

func (zl *zapLogger) Error(msg string, fields ...any) {
	// WARNING. Тут код надеется на наличие ошибки в fields[1], хоть я так и делаю всегда, тем не менее...
	if domain.CheckIfCriticalError(fields[1].(error)) {
		zl.logger.Errorw(msg, append(fields, zap.Stack("stack"))...)
		return
	}
	zl.logger.Errorw(msg, fields...)
}

func (zl *zapLogger) Panic(msg string, fields ...any) {
	zl.logger.Panicw(msg, append(fields, zap.Stack("stack"))...)
}

func (zl *zapLogger) Log(ctx context.Context, level Level, msg string, fields ...any) {
	zl.logger.With(fields...).Log(levelToZapLevel(level), msg)
}

func (zl *zapLogger) With(fields ...any) Logger {
	return &zapLogger{logger: zl.logger.With(fields)}
}

func (zl *zapLogger) Named(name string) Logger {
	return &zapLogger{logger: zl.logger.Named(name)}
}

func NewZapLogger(level string, options ...ZapOption) Logger {
	zapConfig := zap.NewProductionConfig()

	zapLevel := zapcore.Level(levelToZapLevel(stringToLevel(level)))
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	for _, opt := range options {
		opt(&zapConfig)
	}

	zapConfig.EncoderConfig.StacktraceKey = ""

	logger, err := zapConfig.Build(
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.AddCallerSkip(1),
		zap.AddCaller(),
	)
	if err != nil {
		panic(err)
	}

	return &zapLogger{logger: logger.Sugar()}
}

type ZapOption func(*zap.Config)

func WithDevelopmentMode() ZapOption {
	return func(cfg *zap.Config) {
		cfg.Development = true
	}
}

func WithEncoding(encoding string) ZapOption {
	return func(cfg *zap.Config) {
		cfg.Encoding = encoding
	}
}

func WithOutputPaths(outputPaths []string) ZapOption {
	return func(cfg *zap.Config) {
		cfg.OutputPaths = outputPaths
	}
}

func WithErrorOutputPaths(errorOutputPaths []string) ZapOption {
	return func(cfg *zap.Config) {
		cfg.ErrorOutputPaths = errorOutputPaths
	}
}

func WithFileOutput(path string) ZapOption {
	return func(cfg *zap.Config) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("failed to create dir: %v", err)
			return
		}

		cfg.OutputPaths = append(cfg.OutputPaths, path)
	}
}

func WithFileErrorsOutput(path string) ZapOption {
	return func(cfg *zap.Config) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("failed to create dir: %v", err)
			return
		}

		cfg.ErrorOutputPaths = append(cfg.ErrorOutputPaths, path)
	}
}

func stringToLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "panic":
		return LevelPanic
	default:
		return LevelError
	}
}

func levelToZapLevel(level Level) zapcore.Level {
	switch level {
	case LevelDebug:
		return zap.DebugLevel
	case LevelInfo:
		return zap.InfoLevel
	case LevelWarn:
		return zap.WarnLevel
	case LevelError:
		return zap.ErrorLevel
	case LevelPanic:
		return zap.PanicLevel
	default:
		return zap.ErrorLevel
	}
}
