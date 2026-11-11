package errlog

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Database(ctx context.Context, subject string, err error) {
	logger := GetLogger(ctx)
	span := trace.SpanFromContext(ctx)
	pgErr, isPgErr := errors.AsType[*pgconn.PgError](err)

	var code string
	var failed bool
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		failed = true
		code = "database_timeout"
		logger.ErrorContext(ctx, "database operation timed out",
			"err", err,
			"code", code,
		)
	case errors.Is(err, pgconn.ErrConnClosed):
		failed = true
		code = "database_connection_closed"
		logger.ErrorContext(ctx, "database connection closed",
			"err", err,
			"code", code,
		)
	case errors.Is(err, context.Canceled):
		code = "database_operation_canceled"
		logger.WarnContext(ctx, "database operation canceled",
			"err", err,
			"code", code,
		)
	case errors.Is(err, pgx.ErrNoRows):
		code = subject + "_not_found"
		logger.WarnContext(ctx, "no rows found",
			"err", err,
			"code", code,
		)
	case isPgErr:
		switch pgErr.SQLState() {
		case "23505":
			code = subject + "_already_exists"
			logger.WarnContext(ctx, "insert conflict",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("constraint", pgErr.ConstraintName),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "23503":
			code = subject + "_foreign_key_violation"
			logger.WarnContext(ctx, "resource relationship conflict",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("constraint", pgErr.ConstraintName),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("code", code),
				slog.Any("err", err),
			)
		// Not Null violation
		case "23502":
			code = subject + "_missing"
			logger.WarnContext(ctx, "required field missing",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("column", pgErr.ColumnName),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "22P02":
			code = subject + "_invalid"
			logger.WarnContext(ctx, "invalid value",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "40P01":
			failed = true
			code = subject + "_deadlock"
			logger.ErrorContext(ctx, "database deadlock detected",
				slog.String("sqlstate", "40P01"),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "40001":
			failed = true
			code = subject + "_serialization_failure"
			logger.ErrorContext(ctx, "database serialization failure",
				slog.String("sqlstate", "40001"),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "25P02":
			failed = true
			code = subject + "_transaction_failed"
			logger.ErrorContext(ctx, "database transaction already failed",
				slog.String("sqlstate", "25P02"),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "53300":
			failed = true
			code = "database_connections_exhausted"
			logger.ErrorContext(ctx, "database connection limit reached",
				slog.String("sqlstate", "53300"),
				slog.String("code", code),
				slog.Any("err", err),
			)
		case "57014":
			code = subject + "_query_canceled"
			logger.WarnContext(ctx, "database query canceled",
				slog.String("sqlstate", "57014"),
				slog.String("code", code),
				slog.Any("err", err),
			)
		default:
			failed = true
			code = subject + "_failed"
			logger.ErrorContext(ctx, "unexpected database error",
				slog.String("sqlstate", pgErr.SQLState()),
				slog.String("schema", pgErr.SchemaName),
				slog.String("table", pgErr.TableName),
				slog.String("routine", pgErr.Routine),
				slog.String("code", code),
				slog.Any("err", err),
			)
		}
	default:
		failed = true
		code = subject + "_failed"
		logger.ErrorContext(ctx, "unexpected database error",
			"err", err,
			"code", code,
		)
	}

	if failed {
		span.RecordError(err)
		span.SetStatus(codes.Error, code)
	}
}
