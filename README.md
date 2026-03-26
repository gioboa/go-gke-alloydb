# Go and Local PostgreSQL

Simple Go API backed by PostgreSQL.

## Requirements

- Go 1.26+
- Docker + Docker Compose

## Quick setup

1. Clone and enter repo.
2. Run setup helper:

```bash
make setup
```

- macOS: checks Homebrew dependencies and creates `.env` from `env.example`
- Windows (native): checks Go + Docker Desktop via `winget`
- Windows (WSL): checks WSL + Go/Docker in WSL and creates `.env`
- Other OS: prints manual setup message

## Start with Docker (recommended)

Start everything:

```bash
make db-up
```

Check app logs:

```bash
make app-logs
```

Stop all services:

```bash
make db-down
```

Stop only app:

```bash
make app-stop
```

## Start app locally (without Docker)

This uses local `.env` values.

```bash
make app-run
```

## Test

When app is running on port `8080`:

```bash
curl http://localhost:8080/healthz/live
curl http://localhost:8080/healthz/ready
curl http://localhost:8080/ping
curl http://localhost:8080/db/ping
```

Regions API examples:

```bash
curl "http://localhost:8080/regions?limit=10"
curl -X POST http://localhost:8080/regions \
  -H "Content-Type: application/json" \
  -d '{"region_id":1,"region_name":"N. America"}'
curl -X GET http://localhost:8080/regions/1
```

## DB helper

- Open DB shell:

```bash
make db-shell
```

- Export schema manually:

```bash
make schema-dump
```

## Notes

- App uses `.env` values in local/dev workflows (`DB_*` or `DATABASE_URL`).
- `docker-compose.yml` uses:
  - PostgreSQL: `postgres:17`, database `app`, user/password `postgres`
  - App container gets `DB_HOST=postgres`

## Troubleshooting

### DB not reachable / `/db/ping` fails

1. Check Postgres is healthy:

```bash
make db-logs
```

2. Verify DB envs in `.env` are correct (for local shell run, `DB_HOST=localhost`).

3. Restart compose:

```bash
make db-down
make db-up
```

### App starts but `ready` is 503

```bash
curl http://localhost:8080/healthz/ready
```

- `200` means DB is reachable and queries pass.
- If not, check `make db-logs` and `make db-shell`.

### Port already in use

- 8080 already used:

```bash
lsof -i :8080
```

Change host port in [docker-compose.yml](docker-compose.yml) and rerun `make db-up`.

- 5432 already used:

```bash
lsof -i :5432
```

Use an external Postgres or update `postgres.ports` in compose.
