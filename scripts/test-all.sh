#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

started_compose=false
cleanup() {
  if [[ "$started_compose" == true ]]; then
    docker compose -f docker-compose.test.yml down
  fi
}
trap cleanup EXIT

check_tidy() {
  local dir=$1
  (cd "$dir" && go mod tidy)
  if ! git diff --exit-code "$dir/go.mod" "$dir/go.sum"; then
    echo "$dir: go.mod/go.sum not tidy" >&2
    exit 1
  fi
}

echo "==> format check"
unformatted=$(gofmt -l .)
if [[ -n "$unformatted" ]]; then
  echo "gofmt needed for:" >&2
  echo "$unformatted" >&2
  exit 1
fi

echo "==> vet"
go vet ./...

if command -v golangci-lint >/dev/null 2>&1; then
  echo "==> lint"
  golangci-lint run
fi

echo "==> unit tests"
go test ./...

echo "==> race tests"
go test -race ./...

echo "==> fuzz smoke"
go test -run '^$' -fuzz '^FuzzPlaceholderComposition$' -fuzztime=10s ./...

echo "==> root dependency policy"
bash scripts/check-root-deps.sh

echo "==> root module tidy"
check_tidy .

echo "==> root module verify"
go mod verify

echo "==> compatibility module"
(cd compatibility && go test ./...)
check_tidy compatibility

echo "==> integration module tidy"
check_tidy integration

echo "==> integration tests"
if [[ -z "${POSTGRES_TEST_DSN:-}" ]]; then
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    docker compose -f docker-compose.test.yml up -d --wait
    started_compose=true
    export POSTGRES_TEST_DSN='postgres://squirrel:squirrel@localhost:54329/squirrel_test?sslmode=disable'
  fi
fi
(cd integration && go test ./...)

echo "==> all checks passed"
