SQLC_VERSION := v1.30.0
SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
PG_DUMP ?= pg_dump
SQL_SCHEMA_FILE ?= sql/schema.sql

.PHONY: lint sqlcheck schema-dump env-up setup setup-mac setup-windows setup-wsl setup-unsupported app-up app-down db-up db-down db-logs db-shell app-logs app-run app-stop

UNAME_S := $(shell uname -s 2>/dev/null || echo "")
UNAME_R := $(shell uname -r 2>/dev/null || echo "")

ifeq ($(OS),Windows_NT)
ifeq ($(findstring microsoft,$(shell echo $(UNAME_R) | tr A-Z a-z),microsoft)
SETUP_TARGET := setup-wsl
else
SETUP_TARGET := setup-windows
endif
else ifeq ($(UNAME_S),Darwin)
SETUP_TARGET := setup-mac
else
SETUP_TARGET := setup-unsupported
endif

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

db-up:
	docker compose up --build -d

db-down:
	docker compose down

db-shell:
	docker compose exec postgres psql -U "$${DB_USER:-postgres}" -d "$${DB_NAME:-app}"

db-logs:
	docker compose logs -f

env-up:
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo ".env created from env.example"; \
	else \
		echo ".env already exists"; \
	fi

app-up:
	docker compose up -d app

app-down:
	docker compose stop app

app-stop:
	docker compose stop app

app-logs:
	docker compose logs -f app

app-run:
	@if [ ! -f .env ]; then \
		cp env.example .env; \
	fi
	set -a; \
	. ./.env; \
	set +a; \
	go run .

setup:
	@echo "Running $(SETUP_TARGET) setup"
	@$(MAKE) $(SETUP_TARGET)

setup-unsupported:
	@echo "Unsupported host for automatic setup."
	@echo "Detected OS=$(OS) UNAME_S=$(UNAME_S) UNAME_R=$(UNAME_R)"
	@echo "Please install Go + Docker manually, create .env from env.example, then run make db-up."

setup-mac:
	@echo "Setup (macOS):"
	@command -v go >/dev/null 2>&1 && echo "go: $(shell command -v go)" || echo "Missing: brew install go"
	@command -v docker >/dev/null 2>&1 && echo "docker: $(shell command -v docker)" || echo "Missing: brew install --cask docker"
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo ".env created from env.example"; \
	else \
		echo ".env already exists"; \
	fi
	@echo "Run: make db-up"

setup-wsl:
	@echo "Setup (Windows via WSL):"
	@command -v wsl.exe >/dev/null 2>&1 && echo "wsl: available" || echo "Enable WSL from Windows: wsl --install"
	@command -v go >/dev/null 2>&1 && echo "go: $(shell command -v go)" || echo "Install Go in WSL (eg. sudo apt-get install -y golang-go)"
	@command -v docker >/dev/null 2>&1 && echo "docker: $(shell command -v docker)" || echo "Install Docker in WSL or use Docker Desktop with WSL integration"
	@command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && echo "docker compose: installed" || echo "Enable Docker Compose (plugin)"
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo ".env created from env.example"; \
	else \
		echo ".env already exists"; \
	fi
	@echo "Tip: enable Docker Desktop WSL integration for this distro, then run: make db-up"

setup-windows:
	@echo "Setup (Windows):"
	@echo "Run this in PowerShell as Administrator."
	@if command -v go >/dev/null 2>&1; then \
		echo "go: installed"; \
	else \
		echo "Install Go: winget install -e --id GoLang.Go"; \
	fi
	@if command -v docker >/dev/null 2>&1; then \
		echo "docker: installed"; \
	else \
		echo "Install Docker Desktop: winget install -e --id Docker.DockerDesktop"; \
	fi
	@if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		echo "docker compose: installed"; \
	else \
		echo "Enable Docker Compose (bundled with recent Docker Desktop)"; \
	fi
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo ".env created from env.example"; \
	else \
		echo ".env already exists"; \
	fi
	@echo "Then run: make db-up"
