package logger

import (
	"context"
	"log"
	"os"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
	// WARNING. Тут такто код надеется на наличие ошибки в fields[1], хоть я так и делаю всегда, тем не менее...
	if len(fields) != 0 && domain.CheckIfCriticalError(fields[1].(error)) {
		zl.logger.Errorw(msg, append(fields, zap.Stack("stack"))...)
		return
	}
	zl.logger.Errorw(msg, fields...)
}

func (zl *zapLogger) Panic(msg string, fields ...any) {
	zl.logger.Panicw(msg, append(fields, zap.Stack("stack"))...)
}

func (zl *zapLogger) Sync() {
	zl.logger.Sync()
}

func (zl *zapLogger) Log(ctx context.Context, level string, msg string, fields ...any) {
	zl.logger.With(fields...).Log(stringToZapLevel(level), msg)
}

func MustInit(level, logFile, encoding string, dev bool) Logger {
	ll := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1, // Megabyte
		MaxBackups: 20,
		MaxAge:     90, // Days
		Compress:   false,
	}

	var enc zapcore.Encoder
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	switch encoding {
	case "json":
		enc = zapcore.NewJSONEncoder(encCfg)
	case "console":
		enc = zapcore.NewConsoleEncoder(encCfg)
	}

	zapCore := zapcore.NewCore(
		enc,
		zap.CombineWriteSyncers(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(ll),
		),
		zap.NewAtomicLevelAt(stringToZapLevel(level)),
	)

	opts := []zap.Option{
		zap.AddCallerSkip(1), zap.AddCaller(),
	}

	if dev {
		opts = append(opts, zap.Development())
	}

	zlogger := zap.New(zapCore, opts...)

	return &zapLogger{logger: zlogger.Sugar()}
}

func stringToZapLevel(s string) zapcore.Level {
	switch s {
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
		log.Printf("unknown log level: %s. setting to ErrorLevel", s)
		return zap.ErrorLevel
	}
}
