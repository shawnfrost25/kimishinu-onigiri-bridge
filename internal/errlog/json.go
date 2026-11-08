package errlog

import (
	"encoding/json"
	"errors"
	"log/slog"
)

func Json(logger slog.Logger, subject string, err error) {
	synErr, isSyntaxError := errors.AsType[*json.SyntaxError](err)
	unmTypeErr, isUnmTypeErr := errors.AsType[*json.UnmarshalTypeError](err)
	invUnmErr, isInvUnmErr := errors.AsType[*json.InvalidUnmarshalError](err)
	marErr, isMarhsalErr := errors.AsType[*json.MarshalerError](err)
	unsupTypeErr, isUnsupTypeErr := errors.AsType[*json.UnsupportedTypeError](err)
	unsupValErr, isUnsupValErr := errors.AsType[*json.UnsupportedValueError](err)

	switch {
	case isSyntaxError:
		logger.Warn("invalid json payload",
			"err", err,
			"offset", synErr.Offset,
			"code", subject+"_payload_syntax_error",
		)
	case isUnmTypeErr:
		logger.Warn("invalid json type",
			"err", err,
			"field", unmTypeErr.Field,
			"given_type", unmTypeErr.Value,
			"expected_type", unmTypeErr.Type.String(),
			"offset", unmTypeErr.Offset,
			"code", subject+"_payload_type_error",
		)
	case isInvUnmErr:
		logger.Error("invalid json unmarshal target",
			"err", err,
			"target_type", invUnmErr.Type.String(),
			"code", subject+"_invalid_unmarshal_target",
		)
	case isMarhsalErr:
		logger.Error("json marshaler failed",
			"err", err,
			"target_type", marErr.Type.String(),
			"code", subject+"_marshal_failed",
		)
	case isUnsupTypeErr:
		logger.Error("unsupported json type",
			"err", err,
			"target_type", unsupTypeErr.Type.String(),
			"code", subject+"_unsupported_type",
		)
	case isUnsupValErr:
		logger.Error("unsupported json value",
			"err", err,
			"value", unsupValErr.Value.String(),
			"description", unsupValErr.Str,
			"code", subject+"_unsupported_value",
		)
	default:
		logger.Error("json operation failed",
			"err", err,
			"code", subject+"_failed",
		)
	}
}
