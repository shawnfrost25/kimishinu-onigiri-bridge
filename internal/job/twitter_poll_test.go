package job

import (
	"bytes"
	"context"
	"io"
	"kimishinu/internal/testenv"
	"net/http"
	"os"
	"testing"
)

type twitterTrip struct{}

func (twitterTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	body, err := os.ReadFile("testdata/twitter/" + r.URL.Query().Get("userName") + ".json")
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body))}, nil
}

func TestTwitterPollWorker(t *testing.T) {
	pool, queries := testenv.New(t)
	worker := &TwitterPollWorker{
		Pool:        pool,
		Queries:     queries,
		RiverClient: newRiverClient(t, pool),
		HttpClient:  http.Client{Transport: twitterTrip{}},
		ApiKey:      "mock_key",
	}

	if err := worker.TWTKimishinu(context.Background(), Users[1]); err != nil {
		t.Fatalf("polling Twitter: %v", err)
	}
}
