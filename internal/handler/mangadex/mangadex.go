package mangadex

import (
	db "kimishinu/internal/db/sqlc"
	"net/http"
)

type MangaDex struct {
	Queries    *db.Queries
	WebhookURL string
	HttpClient http.Client
}

func New(queries *db.Queries, webhookURL string, httpClient http.Client) *MangaDex {
	return &MangaDex{
		Queries:    queries,
		WebhookURL: webhookURL,
		HttpClient: httpClient,
	}
}
