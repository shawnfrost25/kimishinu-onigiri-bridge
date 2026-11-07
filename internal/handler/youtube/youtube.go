package youtube

import (
	db "kimishinu/internal/db/sqlc"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Youtube struct {
	ApiKey     string
	WebhookURL string
	Pool       *pgxpool.Pool
	Queries    *db.Queries
	HttpClient http.Client
}

func New(apiKey string, webhookURL string, pool *pgxpool.Pool, queries *db.Queries, httpClient http.Client) *Youtube {
	return &Youtube{
		ApiKey:     apiKey,
		WebhookURL: webhookURL,
		Pool:       pool,
		Queries:    queries,
		HttpClient: httpClient,
	}
}
