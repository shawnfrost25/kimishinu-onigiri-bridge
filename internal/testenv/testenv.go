package testenv

import (
	"context"
	"fmt"
	db "kimishinu/internal/db/sqlc"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	databeUrl string
	pgCont    *postgres.PostgresContainer
	queries   *db.Queries
)

func Main(m *testing.M) {
	os.Exit(start(m))
}

func start(m *testing.M) int {
	// Phase 1 (Migration file fetch)
	_, thisPath, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintf(os.Stderr, "testenv: failed to get current path\n")
		return 1
	}

	thisDirectory := filepath.Dir(thisPath)
	migrationPath := filepath.Join(thisDirectory, "..", "db", "migrations", "*.sql")

	migrationFiles, err := filepath.Glob(migrationPath)
	if err != nil || len(migrationFiles) == 0 {
		fmt.Fprintf(os.Stderr, "testenv: missing or invalid migration files at %s: %v\n", migrationPath, err)
		return 1
	}

	// Phase 2 (Create psql container)
	ctx := context.Background()
	pgc, err := postgres.Run(
		ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("kimishinu_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(migrationFiles...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to start postgres container: %v\n", err)
		return 1
	}

	pgCont = pgc
	defer func() {
		if err := pgc.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "testenv: failed to terminate postgres container: %v\n", err)
		}
	}()

	url, err := pgc.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to get database url: %v\n", err)
		return 1
	}

	databeUrl = url

	// Phase 3 (Apply river migrations too)
	conf, err := pgxpool.ParseConfig(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to parse database uri: %v\n", err)
		return 1
	}

	tempPool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to create connection pool: %v\n", err)
		return 1
	}

	migrator, err := rivermigrate.New(riverpgxv5.New(tempPool), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to create river migrator: %v", err)
		tempPool.Close()
		return 1
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to run river migration: %v", err)
		tempPool.Close()
		return 1
	}

	tempPool.Close()

	if err := pgc.Snapshot(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "testenv: failed to snapshot database: %v\n", err)
		tempPool.Close()
		return 1
	}

	return m.Run()
}

func New(t *testing.T) (*pgxpool.Pool, *db.Queries) {
	t.Helper()

	if pgCont == nil {
		t.Fatal("testenv: call testenv.Main(m) from TestMain first")
	}

	ctx := context.Background()
	if err := pgCont.Restore(ctx); err != nil {
		t.Fatalf("testenv: failed to restore database snapshot: %v", err)
	}

	pool, err := pgxpool.New(ctx, databeUrl)
	if err != nil {
		t.Fatalf("testenv: failed to create test pool: %v", err)
	}

	queries := db.New(pool)

	t.Cleanup(pool.Close)

	return pool, queries
}
