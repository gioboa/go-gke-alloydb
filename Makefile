SQLC_VERSION := v1.30.0
SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
PG_DUMP ?= pg_dump
SQL_SCHEMA_FILE ?= sql/schema.sql

.PHONY: lint sqlcheck schema-dump

lint:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	go vet ./...

sqlcheck:
	$(SQLC) compile

schema-dump:
	@mkdir -p $(dir $(SQL_SCHEMA_FILE))
	@if [ -n "$$DATABASE_URL" ]; then \
		$(PG_DUMP) --schema-only --no-owner --no-privileges --file="$(SQL_SCHEMA_FILE)" "$$DATABASE_URL"; \
	else \
		: "$${DB_HOST:?set DB_HOST or DATABASE_URL}"; \
		: "$${DB_USER:?set DB_USER or DATABASE_URL}"; \
		: "$${DB_NAME:?set DB_NAME or DATABASE_URL}"; \
		PGHOST="$$DB_HOST" \
		PGPORT="$${DB_PORT:-5432}" \
		PGUSER="$$DB_USER" \
		PGPASSWORD="$$DB_PASSWORD" \
		PGDATABASE="$$DB_NAME" \
		$(PG_DUMP) --schema-only --no-owner --no-privileges --file="$(SQL_SCHEMA_FILE)"; \
	fi
