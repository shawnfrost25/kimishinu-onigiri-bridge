package mangadex

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

func (t *MyTrip) mangaReq(r *http.Request) (*http.Response, error) {
	body, err := os.ReadFile("testdata/kimishinu_manga.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read test data: %w", err)
	}

	switch t.Reason {
	case "no_data":
		var response map[string]any
		if err := sonic.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal body: %w", err)
		}
		response["data"] = []any{}
		body, err = sonic.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
	case "result_not_ok":
		var response map[string]any
		if err := sonic.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal body: %w", err)
		}
		response["result"] = "error"
		body, err = sonic.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
	case "invalid_response":
		body = []byte("{")
	}

	if t.Reason == "invalid_url" {
		r.URL = &url.URL{Scheme: "https", Host: "api.sheena.totsuki.com"}
	}
	if r.URL.Host != "api.mangadex.org" {
		return nil, fmt.Errorf("unexpected host: %s", r.URL.Host)
	}

	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (t *MyTrip) discordReq(r *http.Request) (*http.Response, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("provided type is not json: %s", r.Header.Get("Content-Type"))
	}

	if r.URL.String() != "https://discord.com/api/webhooks" {
		return nil, fmt.Errorf("provided webhook url doesn't match the expected one")
	}

	return &http.Response{
		Status:     "204 No Content",
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}, nil
}

func (t *MyTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	switch r.URL.Host {
	case "api.mangadex.org":
		return t.mangaReq(r)
	case "discord.com":
		return t.discordReq(r)
	default:
		return nil, fmt.Errorf("unexpected host: %s", r.URL.Host)
	}
}

func mk(t *testing.T, trip *MyTrip) *Server {
	t.Helper()
	pool, queries := testenv.New(t)

	httpClient := http.Client{Transport: trip}
	h := New(queries, "https://discord.com/api/webhooks", httpClient)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(auth.LoggerMiddleware)
		for _, manga := range Mangas {
			r.Post("/mangadex/"+manga.HandlerName, h.MDKimishinu(manga))
		}
	})

	return ServerNew(r, pool, queries)
}

func TestMDKimishinu(t *testing.T) {
	trip := &MyTrip{}
	s := mk(t, trip)

	tests := []struct {
		Name           string
		Reason         string
		ExpectedStatus int
	}{
		{Name: "success", ExpectedStatus: http.StatusNoContent},
		{Name: "no_data", Reason: "no_data", ExpectedStatus: http.StatusNotFound},
		{Name: "result_not_ok", Reason: "result_not_ok", ExpectedStatus: http.StatusInternalServerError},
		{Name: "invalid_response", Reason: "invalid_response", ExpectedStatus: http.StatusInternalServerError},
		{Name: "invalid_url", Reason: "invalid_url", ExpectedStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			trip.Reason = tt.Reason
			w := s.do(t, "/mangadex/kimishinu_main", http.MethodPost, nil)

			if w.Code != tt.ExpectedStatus {
				t.Fatalf("Failed: \nProvided status: %d \nExpected status: %d \nResponse: %s", w.Code, tt.ExpectedStatus, w.Body.String())
			}
		})
	}
}
