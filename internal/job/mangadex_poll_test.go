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

type mangadexTrip struct{}

func (mangadexTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	body, err := os.ReadFile("testdata/mangadex/kimishinu_manga.json")
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body))}, nil
}

func TestMangadexPollWorker(t *testing.T) {
	pool, queries := testenv.New(t)
	worker := &MangadexPollWorker{
		Pool:        pool,
		Queries:     queries,
		RiverClient: newRiverClient(t, pool),
		HttpClient:  http.Client{Transport: mangadexTrip{}},
	}

	if err := worker.MDKimishinu(context.Background(), Mangas[0]); err != nil {
		t.Fatalf("polling MangaDex: %v", err)
	}
}
