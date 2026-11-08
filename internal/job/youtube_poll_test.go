package job

import (
	"bytes"
	"fmt"
	"io"
	"kimishinu/internal/testenv"
	"net/http"
	"os"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type youtubeTrip struct{}

func (youtubeTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Host == "www.googleapis.com" && r.URL.Path == "/youtube/v3/playlistItems" {
		body, err := os.ReadFile("testdata/youtube/ichijinsha.json")
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body))}, nil
	}
	// If it uses http.Client and it's a Discord webhook, we mock it.
	if r.URL.Host == "discord.com" && r.URL.Path == "/api/webhooks" {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	return nil, fmt.Errorf("unexpected request: %s", r.URL)
}

func TestYoutubePollWorker(t *testing.T) {
	pool, queries := testenv.New(t)
	worker := YTNew("mock_api", pool, queries, &http.Client{Transport: youtubeTrip{}}, newRiverClient(t, pool))

	job := &river.Job[YoutubePollArgs]{
		JobRow: &rivertype.JobRow{ID: 1},
		Args:   YoutubePollArgs{Tier: "fast"},
	}
	if err := worker.Work(t.Context(), job); err != nil {
		t.Fatalf("polling YouTube: %v", err)
	}

	discordWorker := &DiscordSendWorker{
		WebhookURL: "https://discord.com/api/webhooks",
		HttpClient: http.Client{Transport: youtubeTrip{}},
	}
	discordJob := &river.Job[DiscordSendArgs]{
		JobRow: &rivertype.JobRow{ID: 2},
		Args: DiscordSendArgs{
			Source:     "youtube",
			Identifier: "video-id",
			Trace:      "test-trace",
		},
	}
	if err := discordWorker.Work(t.Context(), discordJob); err != nil {
		t.Fatalf("sending Discord job: %v", err)
	}
}
