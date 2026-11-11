package errlog

import (
	"context"
	"kimishinu/internal/failure"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Internal(ctx context.Context, subject string, kind failure.KindType, err error) {
	logger := GetLogger(ctx)
	subjectMsg := strings.ReplaceAll(subject, "_", " ")
	span := trace.SpanFromContext(ctx)

	var code string
	var failed bool
	switch kind {
	case failure.InternalEmptyValueError:
		code = subject + "_missing"
		failed = true
		logger.ErrorContext(ctx,
			subjectMsg+" is missing",
			"err", err,
			"code", code,
		)
	case failure.InternalGeneralError:
		code = subject + "_failed"
		failed = true
		logger.ErrorContext(ctx,
			subjectMsg+" ended as unexpected internal error",
			"err", err,
			"code", subject+"_failed",
		)
	case failure.GenerationError:
		code = subject + "_generation_failed"
		failed = true
		logger.ErrorContext(ctx, "failed to generate "+subjectMsg,
			"err", err,
			"code", subject+"_generation_failed",
		)
	}

	if failed {
		span.RecordError(err)
		span.SetStatus(codes.Error, code)
	}
}
