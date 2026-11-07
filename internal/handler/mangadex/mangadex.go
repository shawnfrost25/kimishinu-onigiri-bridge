package mangadex

import (
	db "kimishinu/internal/db/sqlc"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MangaDex struct {
	Pool       *pgxpool.Pool
	Queries    *db.Queries
	WebhookURL string
	HttpClient http.Client
}

func New(pool *pgxpool.Pool, queries *db.Queries, webhookURL string, httpClient http.Client) *MangaDex {
	return &MangaDex{
		Pool:       pool,
		Queries:    queries,
		WebhookURL: webhookURL,
		HttpClient: httpClient,
	}
}
