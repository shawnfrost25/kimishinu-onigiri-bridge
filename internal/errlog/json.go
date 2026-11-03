package errlog

import (
	"encoding/json"
	"errors"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"
)

func Json(logger slog.Logger, w http.ResponseWriter, subject string, trace string, err error) {

	respond := func(status int, errorCode string, message string) {
		lib.Pretty(w, status, lib.Error{
			Code:    errorCode,
			Message: message,
			TraceID: trace,
		})
	}

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
			"status", http.StatusBadRequest,
			"offset", synErr.Offset,
			"code", subject+"_payload_syntax_error",
		)
		respond(http.StatusBadRequest, "BAD_REQUEST", "The payload contains invalid JSON syntax")
	case isUnmTypeErr:
		logger.Warn("invalid json type",
			"err", err,
			"status", http.StatusBadRequest,
			"field", unmTypeErr.Field,
			"given_type", unmTypeErr.Value,
			"expected_type", unmTypeErr.Type.String(),
			"offset", unmTypeErr.Offset,
			"code", subject+"_payload_type_error",
		)
		respond(http.StatusBadRequest, "BAD_REQUEST", "The payload contains a value with an invalid type")
	case isInvUnmErr:
		logger.Error("invalid json unmarshal target",
			"err", err,
			"status", http.StatusInternalServerError,
			"target_type", invUnmErr.Type.String(),
			"code", subject+"_invalid_unmarshal_target",
		)
		InternalRes(w, trace)
	case isMarhsalErr:
		logger.Error("json marshaler failed",
			"err", err,
			"status", http.StatusInternalServerError,
			"target_type", marErr.Type.String(),
			"code", subject+"_marshal_failed",
		)
		InternalRes(w, trace)
	case isUnsupTypeErr:
		logger.Error("unsupported json type",
			"err", err,
			"status", http.StatusInternalServerError,
			"target_type", unsupTypeErr.Type.String(),
			"code", subject+"_unsupported_type",
		)
		InternalRes(w, trace)
	case isUnsupValErr:
		logger.Error("unsupported json value",
			"err", err,
			"status", http.StatusInternalServerError,
			"value", unsupValErr.Value.String(),
			"description", unsupValErr.Str,
			"code", subject+"_unsupported_value",
		)
		InternalRes(w, trace)
	default:
		logger.Error("json operation failed",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_failed",
		)
		InternalRes(w, trace)
	}
}
