package twitter

import (
	db "kimishinu/internal/db/sqlc"
	"net/http"
)

type Twitter struct {
	ApiKey     string
	Queries    *db.Queries
	WebhookURL string
	HttpClient http.Client
}

func New(apiKey string, queries *db.Queries, webhookUrl string, httpClient http.Client) *Twitter {
	return &Twitter{
		ApiKey:     apiKey,
		Queries:    queries,
		WebhookURL: webhookUrl,
		HttpClient: httpClient,
	}
}
