package youtube

import (
	"bytes"
	"fmt"
	"io"
	"kimishinu/internal/auth"
	"kimishinu/internal/testenv"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/go-chi/chi/v5"
)

type MyTrip struct {
	Reason string
}

func (t *MyTrip) youtubeReq(r *http.Request) (*http.Response, error) {
	if r.Header.Get("Accept") != "application/json" {
		return nil, fmt.Errorf("Provided type is not json: \nProvided status: %s", r.Header.Get("Content-Type"))
	}

	if r.URL.Host != "www.googleapis.com" || r.URL.Path != "/youtube/v3/playlistItems" {
		return nil, fmt.Errorf("Provided url doesn't match the expected one: \nProvided url: %s", r.URL.String())
	}

	if r.Header.Get("X-Goog-Api-Key") != "mock_api" {
		return nil, fmt.Errorf("Provided api doesn't match the expected one")
	}

	body, err := os.ReadFile("testdata/ichijinsha.json")
	if err != nil {
		return nil, fmt.Errorf("Failed to read path: %q", "testdata/ichijinsha.json")
	}

	if t.Reason == "missing_items" {
		var b map[string]interface{}
		err = sonic.Unmarshal(body, &b)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal body: %v", err)
		}
		b["items"] = []interface{}{}
		body, err = sonic.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal body: %v", err)
		}
	}

	if t.Reason == "url" {
		r.URL = &url.URL{
			Path: "www.kimishinu.news",
			Host: "/invalid/path",
		}
	}

	switch r.URL.Query().Get("playlistId") {
	case "UU2QDgMJWRQHEJYyVX3Rt08w", "UUyUhtF50BuUjr2jOhxF3IjQ", "UUevPX0DXUiKmgH8o19NrE5w":
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	default:
		return nil, fmt.Errorf("Forbidden playlist ID: \nProvided id: %s", r.URL.Query().Get("playlistId"))
	}
}

func (t *MyTrip) discordReq(r *http.Request) (*http.Response, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Provided type is not json: \nProvided status: %s", r.Header.Get("Content-Type"))
	}

	if r.URL.String() != "https://discord.com/api/webhooks" {
		return nil, fmt.Errorf("Provided webhook url doesn't match the expected one")
	}

	return &http.Response{
		Status:     "204 No Content",
		StatusCode: 204,
		Body:       io.NopCloser(bytes.NewReader([]byte("No Content"))),
	}, nil
}

func (t *MyTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	switch {
	case r.URL.Host == "www.googleapis.com" && r.URL.Path == "/youtube/v3/playlistItems":
		return t.youtubeReq(r)
	case r.URL.Host == "discord.com" && r.URL.Path == "/api/webhooks":
		return t.discordReq(r)
	default:
		return nil, fmt.Errorf("Unknown request: \nProvided url: %s", r.URL.String())
	}
}

func y(t *testing.T, trip *MyTrip) *Server {
	pool, queries := testenv.New(t)

	mockClient := http.Client{
		Transport: trip,
	}

	h := New("mock_api", "https://discord.com/api/webhooks", queries, mockClient)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(auth.LoggerMiddleware)
		for _, ch := range Channels {
			r.Post("/youtube/"+ch.HandlerName, h.YTKimishinu(ch))
		}
	})

	return ServerNew(r, pool, queries)
}

func TestYTKimishinu(t *testing.T) {
	trip := &MyTrip{}
	s := y(t, trip)

	test := []struct {
		Name           string
		Endpoint       string
		Reason         string
		ExpectedStatus int
	}{
		{
			Name:           "Successfully fetching Ichijinsha",
			Endpoint:       "/youtube/ichijinsha",
			ExpectedStatus: http.StatusNoContent,
		},
		{
			Name:           "Failing due to invalid path",
			Endpoint:       "/youtube/ichijinsha",
			Reason:         "url",
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			Name:           "Failing due to missing items",
			Endpoint:       "/youtube/ichijinsha",
			Reason:         "missing_items",
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:           "Successfully fetching ReoNa",
			Endpoint:       "/youtube/reona",
			ExpectedStatus: http.StatusNoContent,
		},
		{
			Name:           "Successfully fetching Infinite Channel",
			Endpoint:       "/youtube/infinite",
			ExpectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range test {
		t.Run(tt.Name, func(t *testing.T) {
			trip.Reason = tt.Reason
			w := s.do(t, tt.Endpoint, http.MethodPost, nil)

			if w.Code != tt.ExpectedStatus {
				t.Fatalf("Failed: \nProvided status: %d \nExpected status: %d \nResponse: %s", w.Code, tt.ExpectedStatus, w.Body.String())
			}
		})
	}
}
