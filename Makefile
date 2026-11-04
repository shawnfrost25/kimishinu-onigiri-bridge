Update ?= <error>
PSQL_Password ?= <error>
PSQL_Database ?= <error>

.PHONY: atlas-new

atlas-new:
	atlas migrate diff ${Update} --to "file://internal/db/schema.sql" --dir "file://internal/db/migrations" --dev-url "docker://postgres/18/dev"

atlas-apply:
	atlas migrate apply --dir "file://migrations" --url "postgres://postgres:${PSQL_Password}@0.0.0.0:5432/${PSQL_Database}?sslmode=disable"