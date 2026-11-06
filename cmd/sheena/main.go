package main

import (
	"context"
	"kimishinu/internal/auth"
	db "kimishinu/internal/db/sqlc"
	"kimishinu/internal/handler/mangadex"
	"kimishinu/internal/handler/twitter"
	"kimishinu/internal/handler/youtube"
	"kimishinu/internal/lib"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	lib.LoggerNew(slog.LevelDebug)
	_ = godotenv.Load("../../.env")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("Missing 'DATABASE_URL' inside .env")
	}
	youtubeAPI := os.Getenv("YOUTUBE_API")
	if youtubeAPI == "" {
		panic("Missing 'YOUTUBE_API' inside .env")
	}
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		panic("Missing 'WEBHOOK_URL'")
	}
	twitterAPI := os.Getenv("TWITTER_API")
	if twitterAPI == "" {
		panic("Missing 'TWITTER_API' inside .env")
	}

	conf, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		panic("Failed to parse database url")
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		panic("Failed to create connection pool")
	}
	defer pool.Close()

	queries := db.New(pool)
	httpClient := http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   20,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	y := youtube.New(youtubeAPI, webhookURL, queries, httpClient)
	t := twitter.New(twitterAPI, queries, webhookURL, httpClient)
	m := mangadex.New(queries, webhookURL, httpClient)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(auth.LoggerMiddleware)

	r.Group(func(r chi.Router) {
		for _, channel := range youtube.Channels {
			r.Post("/youtube/"+channel.HandlerName, y.YTKimishinu(channel))
		}
		for _, user := range twitter.Users {
			r.Post("/twitter/"+user.Handlername, t.TWTKimishinu(user))
		}
		for _, manga := range mangadex.Mangas {
			r.Post("/mangadex/"+manga.HandlerName, m.MDKimishinu(manga))
		}
	})

	const addr string = "0.0.0.0:8080"
	print("Listening on port 8080")
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
