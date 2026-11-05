package main

import (
	"context"
	"kimishinu/internal/auth"
	db "kimishinu/internal/db/sqlc"
	"kimishinu/internal/handler/youtube"
	"kimishinu/internal/lib"
	"log/slog"
	"net/http"
	"os"

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
	httpClient := http.Client{}

	y := youtube.New(youtubeAPI, webhookURL, queries, httpClient)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(auth.LoggerMiddleware)

	r.Group(func(r chi.Router) {
		for _, channel := range youtube.Channels {
			r.Post("/youtube/"+channel.HandlerName, y.FetchKimishinu(channel))
		}
	})

	const addr string = "0.0.0.0:8080"
	print("Listening on port 8080")
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
