FILE ?= <error>
PSQL_PASSWORD ?= <error>
PSQL_DATABASE ?= <error>

.PHONY: atlas-new atlas-apply atlas-apply-dirty atlas-status test test-race

atlas-new:
	atlas migrate diff ${FILE} --to "file://internal/db/schema.sql" --dir "file://internal/db/migrations" --dev-url "docker://postgres/18/dev"

atlas-apply:
	atlas migrate apply --dir "file://internal/db/migrations" --url "postgres://postgres:${PSQL_PASSWORD}@0.0.0.0:5432/${PSQL_DATABASE}?sslmode=disable"

atlas-apply-dirty:
	atlas migrate apply --dir "file://internal/db/migrations" --url "postgres://postgres:${PSQL_PASSWORD}@0.0.0.0:5432/${PSQL_DATABASE}?sslmode=disable" --allow-dirty

atlas-status:
	atlas migrate status --dir "file://internal/db/migrations" --url "postgres://postgres:${PSQL_PASSWORD}@0.0.0.0:5432/${PSQL_DATABASE}?sslmode=disable"

include .env
export

test:
	go test -v ./internal/...

test-race:
	go test -race -v ./internal/...