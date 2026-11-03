package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tb := make([]byte, 8)
		_, _ = rand.Read(tb)
		trace := hex.EncodeToString(tb)
		start := time.Now()
		log := slog.With(slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Time("started_at", start), slog.String("trace", trace))

		next.ServeHTTP(w, r)

		log = slog.With(slog.Duration("duration", time.Since(start)))
		slog.SetDefault(log)
	})
}
