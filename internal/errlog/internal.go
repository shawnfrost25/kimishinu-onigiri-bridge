package errlog

import (
	"kimishinu/internal/failure"
	"log/slog"
	"strings"
)

func Internal(logger slog.Logger, subject string, kind failure.KindType, err error) {
	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.InternalEmptyValueError:
		logger.Error(
			subjectMsg+" is missing",
			"err", err,
			"code", subject+"_missing",
		)
	case failure.InternalGeneralError:
		logger.Error(
			subjectMsg+" ended as unexpected internal error",
			"err", err,
			"code", subject+"_failed",
		)
	case failure.GenerationError:
		logger.Error("failed to generate "+subjectMsg,
			"err", err,
			"code", subject+"_generation_failed",
		)
	}
}
