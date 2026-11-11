package observability

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func TracerSetup(ctx context.Context) (*sdktrace.TracerProvider, error) {
	resources, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName("kimishinu-poll"),
		semconv.ServiceVersion("1.0.0"),
	))
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("alloy:4318"), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}
	processor := sdktrace.NewBatchSpanProcessor(exporter)
	return sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(processor), sdktrace.WithResource(resources)), nil
}

func LoggerSetup(ctx context.Context) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpoint("alloy:4318"), otlploghttp.WithInsecure())
	if err != nil {
		return nil, err
	}
	procesor := log.NewBatchProcessor(exporter)
	return log.NewLoggerProvider(log.WithProcessor(procesor)), nil
}
