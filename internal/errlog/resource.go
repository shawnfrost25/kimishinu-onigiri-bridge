package errlog

import (
	"kimishinu/internal/failure"
	"log/slog"
	"strings"
)

func Resource(logger slog.Logger, subject string, kind failure.KindType, err error) {
	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.EmptyValueError:
		logger.Warn(subjectMsg+" is missing",
			"err", err,
			"code", subject+"_missing",
		)
	case failure.MalformedError:
		logger.Warn("resource is malformed",
			"err", err,
			"code", subject+"_malformed",
		)
	case failure.InvalidError:
		logger.Warn("request validation failed",
			"err", err,
			"code", subject+"_invalid",
		)
	case failure.UnauthorizedError:
		logger.Warn("unauthorized "+subjectMsg,
			"err", err,
			"code", subject+"_unauthorized",
		)
	case failure.ForbiddenError:
		logger.Warn("not enough permissions for "+subjectMsg,
			"err", err,
			"code", subject+"_forbidden",
		)
	case failure.ConflictError:
		logger.Warn("resource conflict",
			"err", err,
			"code", subject+"_already_exists",
		)
	case failure.NotFoundError:
		logger.Warn(subjectMsg+" not found",
			"err", err,
			"code", subject+"_not_found",
		)
	case failure.RateLimitError:
		logger.Warn(subjectMsg+" is rate limited",
			"err", err,
			"code", subject+"_rate_limited",
		)
	}
}
