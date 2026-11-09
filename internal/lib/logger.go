package lib

import (
	"log/slog"
	"os"
)

func LoggerNew(level slog.Level) *slog.Logger {

	var opts slog.HandlerOptions
	if level%4 != 0 || level < -4 || level > 8 {
		opts = slog.HandlerOptions{
			AddSource: false,
			Level:     level,
		}
	} else {
		opts = slog.HandlerOptions{
			AddSource: false,
			Level:     level.Level(),
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	slog.SetDefault(logger)

	return logger
}
