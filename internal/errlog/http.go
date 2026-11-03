package errlog

import (
	"kimishinu/internal/failure"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"
	"strings"
)

func Http(logger slog.Logger, w http.ResponseWriter, subject string, kind failure.KindType, err error, trace string) {
	respond := func(status int, errorCode string, message string) {
		lib.Pretty(w, status, lib.Error{
			Code:    errorCode,
			Message: message,
			TraceID: trace,
		})
	}

	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.EmptyValueError:
		logger.Warn(subjectMsg+" is missing",
			"err", err,
			"status", http.StatusBadRequest,
			"code", subject+"_missing",
		)
		respond(http.StatusBadRequest, "BAD_REQUEST", "A required field is missing")
	case failure.MalformedError:
		logger.Warn("resource is malformed",
			"err", err,
			"status", http.StatusBadRequest,
			"code", subject+"_malformed",
		)
		respond(http.StatusBadRequest, "BAD_REQUEST", "The provided data is malformed")
	case failure.InvalidError:
		logger.Warn("request validation failed",
			"err", err,
			"status", http.StatusBadRequest,
			"code", subject+"_invalid",
		)
		respond(http.StatusBadRequest, "BAD_REQUEST", "The given resource does not comply with the requirements")
	case failure.UnauthorizedError:
		logger.Warn("unauthorized "+subjectMsg,
			"err", err,
			"status", http.StatusUnauthorized,
			"code", subject+"_unauthorized",
		)
		respond(http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized action")
	case failure.ForbiddenError:
		logger.Warn("not enough permissions for "+subjectMsg,
			"err", err,
			"status", http.StatusForbidden,
			"code", subject+"_forbidden",
		)
		respond(http.StatusForbidden, "FORBIDDEN", "Restricted action")
	case failure.ConflictError:
		logger.Warn("resource conflict",
			"err", err,
			"status", http.StatusConflict,
			"code", subject+"_already_exists",
		)
		respond(http.StatusConflict, "CONFLICT", "A resource with the provided identifier already exists")
	case failure.NotFoundError:
		logger.Warn(subjectMsg+" not found",
			"err", err,
			"status", http.StatusNotFound,
			"code", subject+"_not_found",
		)
		respond(http.StatusNotFound, "NOT_FOUND", "Nothing matches the provided resource")
	case failure.RateLimitError:
		logger.Warn(subjectMsg+" is rate limited",
			"err", err,
			"status", http.StatusTooManyRequests,
			"code", subject+"_rate_limited",
		)
		respond(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Too many requests, try again later")
	}
}
