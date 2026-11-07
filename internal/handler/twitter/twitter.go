package twitter

import (
	db "kimishinu/internal/db/sqlc"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Twitter struct {
	ApiKey     string
	Queries    *db.Queries
	Pool       *pgxpool.Pool
	WebhookURL string
	HttpClient http.Client
}

func New(apiKey string, queries *db.Queries, pool *pgxpool.Pool, webhookUrl string, httpClient http.Client) *Twitter {
	return &Twitter{
		ApiKey:     apiKey,
		Queries:    queries,
		Pool:       pool,
		WebhookURL: webhookUrl,
		HttpClient: httpClient,
	}
}
