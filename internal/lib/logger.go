package lib

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func LoggerNew(serviceName string) *slog.Logger {
	// Forces to take all the LevelTypes and also points to the OTel API
	logger := otelslog.NewLogger(serviceName)
	slog.SetDefault(logger)
	return logger
}
