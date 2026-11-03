package errlog

import (
	"context"
	"errors"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func Database(logger slog.Logger, w http.ResponseWriter, subject string, trace string, err error) {
	respond := func(status int, errorCode string, message string) {
		lib.Pretty(w, status, lib.Error{
			Code:    errorCode,
			Message: message,
			TraceID: trace,
		})
	}

	pgErr, isPgErr := errors.AsType[*pgconn.PgError](err)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		logger.Error("database operation timed out",
			"err", err,
			"status", http.StatusGatewayTimeout,
			"code", "database_timeout",
		)
		respond(http.StatusGatewayTimeout, "TIMEOUT", "The operation timed out")
	case errors.Is(err, pgconn.ErrConnClosed):
		logger.Error("database connection closed",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", "database_connection_closed",
		)
		InternalRes(w, trace)
	case errors.Is(err, context.Canceled):
		logger.Warn("database operation canceled",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", "database_operation_canceled",
		)
		InternalRes(w, trace)
	case errors.Is(err, pgx.ErrNoRows):
		logger.Warn("no rows found",
			"err", err,
			"status", http.StatusNotFound,
			"code", subject+"_not_found",
		)
		respond(http.StatusNotFound, "NOT_FOUND", "Nothing matches the provided resource")
	case isPgErr:
		switch pgErr.SQLState() {
		case "23505":
			logger.Warn("insert conflict",
				slog.Int("status", http.StatusConflict),
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("constraint", pgErr.ConstraintName),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("code", subject+"_already_exists"),
				slog.Any("err", err),
			)
			respond(http.StatusConflict, "CONFLICT", "A resource with the provided identifier already exists")
		case "23503":
			logger.Warn("resource relationship conflict",
				slog.Int("status", http.StatusConflict),
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("constraint", pgErr.ConstraintName),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("code", subject+"_foreign_key_violation"),
				slog.Any("err", err),
			)
			respond(http.StatusConflict, "CONFLICT", "The operation conflicts with an existing resource relationship")
		// Not Null violation
		case "23502":
			logger.Warn("required field missing",
				slog.Int("status", http.StatusBadRequest),
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("column", pgErr.ColumnName),
				slog.String("code", subject+"_missing"),
				slog.Any("err", err),
			)
			respond(http.StatusBadRequest, "BAD_REQUEST", "A required field is missing")
		case "22P02":
			logger.Warn("invalid value",
				slog.Int("status", http.StatusBadRequest),
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("code", subject+"_invalid"),
				slog.Any("err", err),
			)
			respond(http.StatusBadRequest, "BAD_REQUEST", "A provided value is invalid")
		case "40P01":
			logger.Error("database deadlock detected",
				slog.Int("status", http.StatusServiceUnavailable),
				slog.String("sqlstate", "40P01"),
				slog.String("code", subject+"_deadlock"),
				slog.Any("err", err),
			)
			respond(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "The operation could not be completed temporarily")
		case "40001":
			logger.Error("database serialization failure",
				slog.Int("status", http.StatusServiceUnavailable),
				slog.String("sqlstate", "40001"),
				slog.String("code", subject+"_serialization_failure"),
				slog.Any("err", err),
			)
			respond(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "The operation could not be completed temporarily")
		case "25P02":
			logger.Error("database transaction already failed",
				slog.Int("status", http.StatusInternalServerError),
				slog.String("sqlstate", "25P02"),
				slog.String("code", subject+"_transaction_failed"),
				slog.Any("err", err),
			)
			InternalRes(w, trace)
		case "53300":
			logger.Error("database connection limit reached",
				slog.Int("status", http.StatusServiceUnavailable),
				slog.String("sqlstate", "53300"),
				slog.String("code", "database_connections_exhausted"),
				slog.Any("err", err),
			)
			respond(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "The service is temporarily unavailable")
		case "57014":
			logger.Warn("database query canceled",
				slog.Int("status", http.StatusInternalServerError),
				slog.String("sqlstate", "57014"),
				slog.String("code", subject+"_query_canceled"),
				slog.Any("err", err),
			)
			InternalRes(w, trace)
		default:
			logger.Error("unexpected database error",
				slog.Int("status", http.StatusInternalServerError),
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("routine", pgErr.Routine),
				slog.String("code", subject+"_failed"),
				slog.Any("err", err),
			)
			InternalRes(w, trace)
		}
	default:
		logger.Error("unexpected database error",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", subject+"_failed",
		)
		InternalRes(w, trace)
	}
}
