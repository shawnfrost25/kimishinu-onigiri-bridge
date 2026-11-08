package errlog

import (
	"kimishinu/internal/failure"
	"log/slog"
	"strings"
)

func General(logger slog.Logger, subject string, kind failure.KindType, err error) {
	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.GeneralError:
		logger.Warn(subjectMsg+" request failed",
			"err", err,
			"code", subject+"_failed",
		)
	case failure.InternalGeneralError:
		logger.Error(subjectMsg+" ended as unexpected internal error",
			"err", err,
			"code", subject+"_failed",
		)
	}
}
