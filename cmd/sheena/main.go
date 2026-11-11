package main

import (
	"context"
	"fmt"
	db "kimishinu/internal/db/sqlc"
	"kimishinu/internal/errlog"
	"kimishinu/internal/failure"
	"kimishinu/internal/job"
	"kimishinu/internal/lib"
	"kimishinu/internal/observability"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.opentelemetry.io/otel"
)

func main() {
	os.Exit(setup())
}

// REMINDER: DO UPDATE ALL THE LOGGERS SO THEY WILL TAKE CONTEXT AND USE IT (for trace_id and span_id)
// REMINDER: YOU GET RATE-LIMITED BY TWITTERAPI DUE TO THE APIKEY LEVEL

func setup() int {
	ctx := context.Background()
	l := lib.LoggerNew("kimishinu-poll")
	logger := l.With(slog.String("op", "main_setup"))
	errlog.SetLogger(ctx, logger)

	tracer := otel.Tracer("main")

	// We just created the parent span, everything starts from here
	rootCtx, parentSpan := tracer.Start(ctx, "main.start")
	defer parentSpan.End()

	otelCtx, otelSpan := tracer.Start(rootCtx, "main.otel.setup")
	defer otelSpan.End()
	otelLogCtx, otelLogSpan := tracer.Start(otelCtx, "main.log.otel.setup")
	defer otelLogSpan.End()
	loggerProv, err := observability.LoggerSetup(otelLogCtx)
	if err != nil {
		err = failure.Wrap("main.otel.log.extractor.create", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(otelLogCtx, "otel_log_extractor_create", failure.InternalGeneralError, err)
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := loggerProv.Shutdown(shutdownCtx)
		if err != nil {
			err = failure.Wrap("main.otel.log.provider.close", err)
			fmt.Fprintf(os.Stderr, "main: %v\n", err)
			errlog.Internal(otelLogCtx, "otel_log_provider_close", failure.InternalGeneralError, err)
		}
	}()

	otel.SetLoggerProvider(loggerProv)

	otelTraceCtx, otelTraceSpan := tracer.Start(otelCtx, "main.otel.trace.setup")
	defer otelTraceSpan.End()
	traceProv, err := observability.TracerSetup(otelTraceCtx)
	if err != nil {
		err = failure.Wrap("main.otel.trace.extractor.create", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(otelTraceCtx, "otel_trace_extractor_create", failure.InternalGeneralError, err)
		return 1
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := traceProv.Shutdown(shutdown)
		if err != nil {
			err = failure.Wrap("main.otel.trace.provider.close", err)
			fmt.Fprintf(os.Stderr, "main: %v\n", err)
			errlog.Internal(otelTraceCtx, "otel_trace_provider_close", failure.InternalGeneralError, err)
		}
	}()

	otel.SetTracerProvider(traceProv)

	envCtx, envSpan := tracer.Start(rootCtx, "main.environment.variable.fetch")
	defer envSpan.End()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		err := failure.EmptyValue("DATABASE_URL")
		err = failure.Wrap("main.database.url.environment.variable.fetch", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(envCtx, "DATABASE_URL", failure.InternalEmptyValueError, err)
		return 1
	}
	youtubeAPI := os.Getenv("YOUTUBE_API")
	if youtubeAPI == "" {
		err := failure.EmptyValue("YOUTUBE_API")
		err = failure.Wrap("main.youtube.api.environment.variable.fetch", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(envCtx, "YOUTUBE_API", failure.InternalEmptyValueError, err)
		return 1
	}
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		err := failure.EmptyValue("WEBHOOK_URL")
		err = failure.Wrap("main.webhook.url.environment.variable.fetch", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(envCtx, "WEBHOOK_URL", failure.InternalEmptyValueError, err)
		return 1
	}
	twitterAPI := os.Getenv("TWITTER_API")
	if twitterAPI == "" {
		err := failure.EmptyValue("TWITTER_API")
		err = failure.Wrap("main.twitter.environment.variable.fetch", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(envCtx, "TWITTER_API", failure.InternalEmptyValueError, err)
		return 1
	}

	poolCtx, poolSpan := tracer.Start(rootCtx, "main.pool.connection.create")
	defer poolSpan.End()
	conf, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		err = failure.Wrap("main.pool.connection.parse.url", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(poolCtx, "database_url_parse", failure.InternalEmptyValueError, err)
		return 1
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		err = failure.Wrap("main.pool.connection.create.from.config", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(poolCtx, "database_pool_create", failure.InternalGeneralError, err)
		return 1
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

	riverCtx, riverSpan := tracer.Start(rootCtx, "main.river.setup")
	defer riverSpan.End()
	riverSetupCtx, riverSetupSpan := tracer.Start(riverCtx, "main.river.setup.create")
	defer riverSetupSpan.End()
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		err = failure.Wrap("main.river.migrator.create", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(riverSetupCtx, "river_migrator_create", failure.InternalGeneralError, err)
		return 1
	}

	_, err = migrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil)
	if err != nil {
		err = failure.Wrap("main.river.migrator.migrate", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(riverSetupCtx, "river_migration_failed", failure.InternalGeneralError, err)
		return 1
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
		river.NewPeriodicJob(
			river.PeriodicInterval(30*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return job.MangadexPollArgs{Tier: "medium"}, nil
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
	mangadexWorker := &job.MangadexPollWorker{
		Pool:       pool,
		Queries:    queries,
		HttpClient: httpClient,
	}

	river.AddWorker(workers, mangadexWorker)
	river.AddWorker(workers, twitterWorker)
	river.AddWorker(workers, discordWorker)
	river.AddWorker(workers, youtubeWorker)

	riverRunCtx, riverRunSpan := tracer.Start(riverCtx, "main.river.run")
	defer riverRunSpan.End()
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {
				MaxWorkers: 15,
			},
		},
		PeriodicJobs: periodicJobs,
	})
	if err != nil {
		err = failure.Wrap("main.river.client.create", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(riverRunCtx, "river_client_create", failure.InternalGeneralError, err)
		return 1
	}

	// we added it a value via pointer
	youtubeWorker.RiverClient = riverClient
	twitterWorker.RiverClient = riverClient
	mangadexWorker.RiverClient = riverClient

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err = riverClient.Start(ctx)
	if err != nil {
		err = failure.Wrap("main.river.client.start", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(riverRunCtx, "river_client_start", failure.InternalGeneralError, err)
		return 1
	}
	slog.Info("River started")

	<-ctx.Done()

	slog.Info("River shutting down")
	defer func() {
		err := riverClient.Stop(ctx)
		if err != nil {
			err = failure.Wrap("main.river.client.stop", err)
			fmt.Fprintf(os.Stderr, "main: %v\n", err)
			errlog.Internal(riverRunCtx, "river_client_stop", failure.InternalGeneralError, err)
		}
	}()

	return 0
}
