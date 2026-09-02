# PostgreSQL integration tests

This nested Go module exercises generated SQL against a real PostgreSQL database.
It is not part of the root module dependency graph.

## Requirements

- Go toolchain matching the repository root
- PostgreSQL reachable via `POSTGRES_TEST_DSN`

## Local PostgreSQL with Docker Compose

From the repository root:

```bash
docker compose -f docker-compose.test.yml up -d --wait
export POSTGRES_TEST_DSN='postgres://squirrel:squirrel@localhost:54329/squirrel_test?sslmode=disable'
cd integration && go test ./...
docker compose -f docker-compose.test.yml down
```

The Compose file uses test-only credentials on port `54329` to avoid colliding with a
local PostgreSQL installation.

## Running tests

```bash
export POSTGRES_TEST_DSN='postgres://...'
cd integration
go test ./...
```

When `POSTGRES_TEST_DSN` is unset, tests skip with a clear message.
