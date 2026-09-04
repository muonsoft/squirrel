#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

started_compose=false
compose_project=squirrel-test-all
cleanup() {
  if [[ "$started_compose" == true ]]; then
    docker compose --project-name "$compose_project" -f docker-compose.test.yml down
  fi
}
trap cleanup EXIT

check_tidy() {
  local dir=$1
  (cd "$dir" && go mod tidy -diff)
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
go test -run '^$' -fuzz '^FuzzPlaceholderComposition$' -fuzztime=10s .

echo "==> root dependency policy"
bash scripts/check-root-deps.sh

echo "==> root module tidy"
check_tidy .

echo "==> root module verify"
go mod verify

echo "==> compatibility module"
(cd compatibility && go test ./...)
check_tidy compatibility

echo "==> exported API inventory"
diff -u docs/audits/baseline/muonsoft-fork-api.txt <(go run ./tools/api-surface .)

echo "==> removed API check"
if rg -n '\b(Search|Paginator|Paginate|PaginateByPage|PaginateByID|SetIDColumn|OrderByCond|OrderCond|EqNotEmpty)\b' --glob '*.go' .; then
  echo "removed API remains in Go source" >&2
  exit 1
fi

echo "==> attribution check"
rg -q 'Masterminds/squirrel' README.md UPSTREAM.md
rg -q 'n-r-w/squirrel' README.md UPSTREAM.md

echo "==> integration module tidy"
check_tidy integration

echo "==> integration tests"
if [[ -z "${POSTGRES_TEST_DSN:-}" ]]; then
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    docker compose --project-name "$compose_project" -f docker-compose.test.yml up -d --wait
    started_compose=true
    export POSTGRES_TEST_DSN='postgres://squirrel:squirrel@localhost:54329/squirrel_test?sslmode=disable'
  else
    echo "integration tests require POSTGRES_TEST_DSN or Docker Compose" >&2
    exit 1
  fi
fi
(cd integration && go test -count=1 ./...)

echo "==> all checks passed"
