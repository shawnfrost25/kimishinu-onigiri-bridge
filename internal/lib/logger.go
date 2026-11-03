package lib

import (
	"log/slog"
	"os"
)

func LoggerNew(level slog.Level) {

	var opts slog.HandlerOptions
	if level%4 != 0 || level < -4 || level > 8 {
		opts = slog.HandlerOptions{
			AddSource: false,
			Level:     level,
		}
	} else {
		opts = slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelInfo,
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts))
	logger.With(slog.String("", ""))
	slog.SetDefault(logger)
}
