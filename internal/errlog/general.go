package errlog

import (
	"kimishinu/internal/failure"
	"kimishinu/internal/lib"
	"net/http"
	"strings"

	"log/slog"
)

func General(logger slog.Logger, w http.ResponseWriter, subject string, kind failure.KindType, err error, trace string) {
	respond := func(status int, errorCode string, message string) {
		lib.Pretty(w, status, lib.Error{
			Code:    errorCode,
			Message: message,
			TraceID: trace,
		})
	}

	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.GeneralError:
		logger.Warn(subjectMsg+" request failed",
			"err", err,
			"status", http.StatusBadRequest,
			"code", subject+"_failed",
		)
		respond(http.StatusBadRequest, "BAD_REQUEST", "The request could not be processed")
	case failure.InternalGeneralError:
		logger.Error(subjectMsg+" ended as unexpected internal error",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_failed",
		)
		InternalRes(w, trace)
	}
}
