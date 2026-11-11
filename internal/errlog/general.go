package errlog

import (
	"context"
	"kimishinu/internal/failure"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func General(ctx context.Context, subject string, kind failure.KindType, err error) {
	span := trace.SpanFromContext(ctx)
	subjectMsg := strings.ReplaceAll(subject, "_", " ")
	logger := GetLogger(ctx)
	switch kind {
	case failure.GeneralError:
		logger.WarnContext(ctx,
			subjectMsg+" request failed",
			"err", err,
			"code", subject+"_failed",
		)
	case failure.InternalGeneralError:
		span.RecordError(err)
		span.SetStatus(codes.Error, subjectMsg+" ended as unexpected internal error")
		logger.ErrorContext(ctx,
			subjectMsg+" ended as unexpected internal error",
			"err", err,
			"code", subject+"_failed",
		)
	}
}
