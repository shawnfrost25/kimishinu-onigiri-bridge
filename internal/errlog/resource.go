package errlog

import (
	"context"
	"kimishinu/internal/failure"
	"strings"
)

func Resource(ctx context.Context, subject string, kind failure.KindType, err error) {
	logger := GetLogger(ctx)
	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.EmptyValueError:
		logger.WarnContext(ctx, subjectMsg+" is missing",
			"err", err,
			"code", subject+"_missing",
		)
	case failure.MalformedError:
		logger.WarnContext(ctx, "resource is malformed",
			"err", err,
			"code", subject+"_malformed",
		)
	case failure.InvalidError:
		logger.WarnContext(ctx, "request validation failed",
			"err", err,
			"code", subject+"_invalid",
		)
	case failure.UnauthorizedError:
		logger.WarnContext(ctx, "unauthorized "+subjectMsg,
			"err", err,
			"code", subject+"_unauthorized",
		)
	case failure.ForbiddenError:
		logger.WarnContext(ctx, "not enough permissions for "+subjectMsg,
			"err", err,
			"code", subject+"_forbidden",
		)
	case failure.ConflictError:
		logger.WarnContext(ctx, "resource conflict",
			"err", err,
			"code", subject+"_already_exists",
		)
	case failure.NotFoundError:
		logger.WarnContext(ctx, subjectMsg+" not found",
			"err", err,
			"code", subject+"_not_found",
		)
	case failure.RateLimitError:
		logger.WarnContext(ctx, subjectMsg+" is rate limited",
			"err", err,
			"code", subject+"_rate_limited",
		)
	}
}
