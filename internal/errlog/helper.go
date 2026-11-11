package errlog

import (
	"context"
	"log/slog"
)

type Logger struct{}

func SetLogger(ctx context.Context, log *slog.Logger) context.Context {
	// The key to fetch it will be "Logger{}"
	return context.WithValue(ctx, Logger{}, log)
}

func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(Logger{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
