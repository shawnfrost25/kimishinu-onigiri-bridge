package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"kimishinu/internal/errlog"
	"kimishinu/internal/failure"
	"log/slog"
	"net/http"
	"time"
)

type CtxKey string

const traceCtx CtxKey = "traceID"

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tb := make([]byte, 8)
		_, _ = rand.Read(tb)
		trace := hex.EncodeToString(tb)
		start := time.Now()
		log := slog.With(slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Time("started_at", start), slog.String("trace", trace))

		ctx := context.WithValue(r.Context(), traceCtx, trace)
		next.ServeHTTP(w, r.WithContext(ctx))

		log = slog.With(slog.Duration("duration", time.Since(start)))
		slog.SetDefault(log)
	})
}

func GetTrace(ctx context.Context, w http.ResponseWriter) (string, bool) {
	logger := slog.With(slog.String("op", "fetch_trace"))
	trace, ok := ctx.Value(traceCtx).(string)
	if !ok {
		err := failure.EmptyValue("trace_context")
		err = failure.Wrap("middleware.http.trace.context.fetch", err)
		errlog.Internal(*logger, w, "trace_context", failure.InternalEmptyValueError, err, "")
		return "", false
	}
	return trace, ok
}
