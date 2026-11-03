package errlog

import (
	"context"
	"errors"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func Redis(logger slog.Logger, w http.ResponseWriter, subject string, err error, trace string) {
	respond := func(status int, errorCode string, message string) {
		lib.Pretty(w, status, lib.Error{
			Code:    errorCode,
			Message: message,
			TraceID: trace,
		})
	}

	red, isRedisError := errors.AsType[redis.Error](err)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		logger.Error("redis operation timed out",
			"err", err,
			"status", http.StatusGatewayTimeout,
			"code", "redis_timeout",
		)
		respond(http.StatusGatewayTimeout, "TIMEOUT", "The operation timed out")
	case errors.Is(err, context.Canceled):
		logger.Warn("redis operation canceled",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", "redis_operation_canceled",
		)
		InternalRes(w, trace)
	case redis.IsAuthError(err):
		logger.Error("failed redis authentication",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", "redis_authentication_failed",
		)
		InternalRes(w, trace)
	case errors.Is(err, redis.Nil):
		logger.Warn("expired or nonexistent redis entry",
			"err", err,
			"status", http.StatusNotFound,
			"code", subject+"_not_found",
		)
		respond(http.StatusNotFound, "NOT_FOUND", "Nothing matches the provided resource")
	case isRedisError:
		logger.Error("redis server rejected command: "+red.Error(),
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_failed",
		)
		InternalRes(w, trace)
	default:
		logger.Error("unexpected redis error",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_failed",
		)
		InternalRes(w, trace)
	}
}
