package errlog

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func Database(logger slog.Logger, subject string, err error) {

	pgErr, isPgErr := errors.AsType[*pgconn.PgError](err)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		logger.Error("database operation timed out",
			"err", err,
			"code", "database_timeout",
		)
	case errors.Is(err, pgconn.ErrConnClosed):
		logger.Error("database connection closed",
			"err", err,
			"code", "database_connection_closed",
		)
	case errors.Is(err, context.Canceled):
		logger.Warn("database operation canceled",
			"err", err,
			"code", "database_operation_canceled",
		)
	case errors.Is(err, pgx.ErrNoRows):
		logger.Warn("no rows found",
			"err", err,
			"code", subject+"_not_found",
		)
	case isPgErr:
		switch pgErr.SQLState() {
		case "23505":
			logger.Warn("insert conflict",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("constraint", pgErr.ConstraintName),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("code", subject+"_already_exists"),
				slog.Any("err", err),
			)
		case "23503":
			logger.Warn("resource relationship conflict",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("constraint", pgErr.ConstraintName),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("code", subject+"_foreign_key_violation"),
				slog.Any("err", err),
			)
		// Not Null violation
		case "23502":
			logger.Warn("required field missing",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("column", pgErr.ColumnName),
				slog.String("code", subject+"_missing"),
				slog.Any("err", err),
			)
		case "22P02":
			logger.Warn("invalid value",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("code", subject+"_invalid"),
				slog.Any("err", err),
			)
		case "40P01":
			logger.Error("database deadlock detected",
				slog.String("sqlstate", "40P01"),
				slog.String("code", subject+"_deadlock"),
				slog.Any("err", err),
			)
		case "40001":
			logger.Error("database serialization failure",
				slog.String("sqlstate", "40001"),
				slog.String("code", subject+"_serialization_failure"),
				slog.Any("err", err),
			)
		case "25P02":
			logger.Error("database transaction already failed",
				slog.String("sqlstate", "25P02"),
				slog.String("code", subject+"_transaction_failed"),
				slog.Any("err", err),
			)
		case "53300":
			logger.Error("database connection limit reached",
				slog.String("sqlstate", "53300"),
				slog.String("code", "database_connections_exhausted"),
				slog.Any("err", err),
			)
		case "57014":
			logger.Warn("database query canceled",
				slog.String("sqlstate", "57014"),
				slog.String("code", subject+"_query_canceled"),
				slog.Any("err", err),
			)
		default:
			logger.Error("unexpected database error",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("routine", pgErr.Routine),
				slog.String("code", subject+"_failed"),
				slog.Any("err", err),
			)
		}
	default:
		logger.Error("unexpected database error",
			"err", err,
			"code", subject+"_failed",
		)
	}
}
