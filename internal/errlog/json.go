package errlog

import (
	"context"
	"encoding/json"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Json(ctx context.Context, subject string, err error) {
	synErr, isSyntaxError := errors.AsType[*json.SyntaxError](err)
	unmTypeErr, isUnmTypeErr := errors.AsType[*json.UnmarshalTypeError](err)
	invUnmErr, isInvUnmErr := errors.AsType[*json.InvalidUnmarshalError](err)
	marErr, isMarhsalErr := errors.AsType[*json.MarshalerError](err)
	unsupTypeErr, isUnsupTypeErr := errors.AsType[*json.UnsupportedTypeError](err)
	unsupValErr, isUnsupValErr := errors.AsType[*json.UnsupportedValueError](err)

	logger := GetLogger(ctx)
	span := trace.SpanFromContext(ctx)

	var code string
	var failed bool
	switch {
	case isSyntaxError:
		logger.WarnContext(ctx, "invalid json payload",
			"err", err,
			"offset", synErr.Offset,
			"code", subject+"_payload_syntax_error",
		)
	case isUnmTypeErr:
		logger.WarnContext(ctx, "invalid json type",
			"err", err,
			"field", unmTypeErr.Field,
			"given_type", unmTypeErr.Value,
			"expected_type", unmTypeErr.Type.String(),
			"offset", unmTypeErr.Offset,
			"code", subject+"_payload_type_error",
		)
	case isInvUnmErr:
		code = subject + "_invalid_unmarshal_target"
		failed = true
		logger.ErrorContext(ctx, "invalid json unmarshal target",
			"err", err,
			"target_type", invUnmErr.Type.String(),
			"code", subject+"_invalid_unmarshal_target",
		)
	case isMarhsalErr:
		code = subject + "_marshal_failed"
		failed = true
		logger.ErrorContext(ctx, "json marshaler failed",
			"err", err,
			"target_type", marErr.Type.String(),
			"code", subject+"_marshal_failed",
		)
	case isUnsupTypeErr:
		code = subject + "_unsupported_type"
		failed = true
		logger.ErrorContext(ctx, "unsupported json type",
			"err", err,
			"target_type", unsupTypeErr.Type.String(),
			"code", subject+"_unsupported_type",
		)
	case isUnsupValErr:
		code = subject + "_unsupported_value"
		failed = true
		logger.ErrorContext(ctx, "unsupported json value",
			"err", err,
			"value", unsupValErr.Value.String(),
			"description", unsupValErr.Str,
			"code", subject+"_unsupported_value",
		)
	default:
		code = subject + "_failed"
		failed = true
		logger.ErrorContext(ctx, "json operation failed",
			"err", err,
			"code", subject+"_failed",
		)
	}

	if failed {
		span.RecordError(err)
		span.SetStatus(codes.Error, code)
	}
}
