package errlog

import (
	"kimishinu/internal/failure"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"
	"strings"
)

func InternalRes(w http.ResponseWriter, trace string) {
	lib.Pretty(w, http.StatusInternalServerError, lib.Error{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An internal error occurred",
		TraceID: trace,
	})
}

func Internal(logger slog.Logger, w http.ResponseWriter, subject string, kind failure.KindType, err error, trace string) {
	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.InternalEmptyValueError:
		logger.Error(
			subjectMsg+" is missing",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_missing",
		)
		InternalRes(w, trace)
	case failure.GenerationError:
		logger.Error("failed to generate "+subjectMsg,
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_generation_failed",
		)
		InternalRes(w, trace)
	}
}
