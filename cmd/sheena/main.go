package main

import (
	"context"
	db "kimishinu/internal/db/sqlc"
	"kimishinu/internal/job"
	"kimishinu/internal/lib"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
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

	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(10*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return job.YoutubePollArgs{Tier: "fast"}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(1*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return job.YoutubePollArgs{Tier: "slow"}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(10*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return job.TwitterPollArgs{Tier: "fast"}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(30*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return job.TwitterPollArgs{Tier: "medium"}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(1*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return job.TwitterPollArgs{Tier: "slow"}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}
	workers := river.NewWorkers()
	// Think of this process as actually giving the values to "Worker" - so it can run normally
	discordWorker := &job.DiscordSendWorker{
		WebhookURL: webhookURL,
		HttpClient: httpClient,
	}
	youtubeWorker := &job.YoutubePollWorker{
		ApiKey:     youtubeAPI,
		Pool:       pool,
		Queries:    queries,
		HttpClient: httpClient,
		// To avoid the chicken-and-egg problem, we simply let it "nil" for now, and we will add it later
	}
	twitterWorker := &job.TwitterPollWorker{
		ApiKey:     twitterAPI,
		Pool:       pool,
		Queries:    queries,
		HttpClient: httpClient,
	}

	river.AddWorker(workers, twitterWorker)
	river.AddWorker(workers, discordWorker)
	river.AddWorker(workers, youtubeWorker)

	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers:      workers,
		PeriodicJobs: periodicJobs,
	})
	if err != nil {
		panic(err)
	}

	// we added it a value via pointer
	youtubeWorker.RiverClient = riverClient
	twitterWorker.RiverClient = riverClient

	err = riverClient.Start(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() {
		err := riverClient.Stop(context.Background())
		if err != nil {
			panic(err)
		}
	}()
}
