package observability

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
)

func LoggerSetup(ctx context.Context) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}
	procesor := log.NewBatchProcessor(exporter)
	return log.NewLoggerProvider(log.WithProcessor(procesor)), nil
}
