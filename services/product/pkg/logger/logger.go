package logger

import (
	"context"
)

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Panic(msg string, fields ...any)

	Sync()
	Log(ctx context.Context, level string, msg string, fields ...any)
}

const (
	LevelDebug string = "debug"
	LevelInfo  string = "info"
	LevelWarn  string = "warn"
	LevelError string = "error"
	LevelPanic string = "panic"
)
