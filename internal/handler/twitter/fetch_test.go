package twitter

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

func (t *MyTrip) twitterReq(r *http.Request) (*http.Response, error) {
	if r.Header.Get("Authorization") != "Bearer mock_key" {
		return nil, fmt.Errorf("unexpected authorization header")
	}

	if r.URL.Host != "api.twitterapi.io" {
		return nil, fmt.Errorf("unexpected host: %s", r.URL.Host)
	}

	var body []byte
	switch r.URL.Query().Get("userName") {
	case "kimishinu":
		b, err := os.ReadFile("testdata/kimishinu.json")
		if err != nil {
			return nil, fmt.Errorf("failed to read test data: %w", err)
		}
		body = b
	case "anime_kimishinu":
		b, err := os.ReadFile("testdata/kimishinu_anime.json")
		if err != nil {
			return nil, fmt.Errorf("failed to read test data: %w", err)
		}
		body = b
	case "ROLL2__":
		b, err := os.ReadFile("testdata/roll2.json")
		if err != nil {
			return nil, fmt.Errorf("failed to read test data: %w", err)
		}
		body = b
	default:
		return nil, fmt.Errorf("unexpected username: %s", r.URL.Query().Get("userName"))
	}

	if t.Reason == "invalid_user" || t.Reason == "missing_tweets" {
		var b map[string]interface{}
		err := sonic.Unmarshal(body, &b)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal body: %v", err)
		}
		data, ok := b["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected response data shape")
		}
		data["tweets"] = []interface{}{}
		body, err = sonic.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %v", err)
		}
	}

	if t.Reason == "invalid_url" {
		r.URL = &url.URL{
			Scheme: "https",
			Host:   "api.sheena.totsuki.com",
		}
	}

	switch r.URL.Query().Get("userName") {
	case "kimishinu", "anime_kimishinu", "ROLL2__":
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	default:
		return nil, fmt.Errorf("unexpected username: %s", r.URL.Query().Get("userName"))
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
	switch r.URL.Host {
	case "api.twitterapi.io":
		return t.twitterReq(r)
	case "discord.com":
		return t.discordReq(r)
	default:
		return nil, fmt.Errorf("unexpected host: %s", r.URL.Host)
	}
}

func tk(t *testing.T, trip *MyTrip) *Server {
	t.Helper()
	pool, queries := testenv.New(t)

	httpClient := http.Client{
		Transport: trip,
	}

	h := New("mock_key", queries, "https://discord.com/api/webhooks", httpClient)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(auth.LoggerMiddleware)
		for _, user := range Users {
			r.Post("/twitter/"+user.Username, h.TWTKimishinu(user))
		}
	})

	return ServerNew(r, pool, queries)
}

func TestTWTKimishinu(t *testing.T) {
	trip := &MyTrip{}
	s := tk(t, trip)

	test := []struct {
		Name           string
		Endpoint       string
		Reason         string
		ExpectedStatus int
	}{
		{
			Name:           "kimishinu",
			Endpoint:       "/twitter/kimishinu",
			ExpectedStatus: 204,
		},
		{
			Name:           "anime_kimishinu",
			Endpoint:       "/twitter/anime_kimishinu",
			ExpectedStatus: 204,
		},
		{
			Name:           "roll2",
			Endpoint:       "/twitter/ROLL2__",
			ExpectedStatus: 204,
		},
		{
			Name:           "missing_tweets",
			Endpoint:       "/twitter/kimishinu",
			Reason:         "missing_tweets",
			ExpectedStatus: 404,
		},
		{
			Name:           "invalid_url",
			Endpoint:       "/twitter/kimishinu",
			Reason:         "invalid_url",
			ExpectedStatus: 500,
		},
		{
			Name:           "invalid_username",
			Endpoint:       "/twitter/kimishinu",
			Reason:         "invalid_user",
			ExpectedStatus: 404,
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
