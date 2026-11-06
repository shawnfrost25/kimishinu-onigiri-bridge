package twitter

import (
	"bytes"
	"io"
	db "kimishinu/internal/db/sqlc"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router  http.Handler
	pool    *pgxpool.Pool
	queries *db.Queries
}

func ServerNew(handler http.Handler, pool *pgxpool.Pool, queries *db.Queries) *Server {
	return &Server{
		router:  handler,
		pool:    pool,
		queries: queries,
	}
}

func (s *Server) do(t *testing.T, endpoint string, method string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var b io.Reader
	if body != nil {
		bodyBytes, err := sonic.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal given body")
			return nil
		}
		b = bytes.NewReader(bodyBytes)
	}
	r := httptest.NewRequest(method, endpoint, b)

	if b != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, r)
	return w
}
