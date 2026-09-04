#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

# Forbidden module path fragments in the root module graph (spec §11.2 and §22).
forbidden_patterns=(
  'github.com/jackc/pgx'
  'github.com/georgysavva/scany'
  'github.com/n-r-w/testdock'
  'github.com/ory/dockertest'
  'github.com/docker/'
  'github.com/golang-migrate/'
  'github.com/pressly/goose'
  'github.com/stretchr/testify'
  'golang.org/x/exp'
  'github.com/go-sql-driver/mysql'
  'go.mongodb.org/mongo-driver'
)

violations=()
while IFS= read -r mod; do
  for pattern in "${forbidden_patterns[@]}"; do
    if [[ "$mod" == *"$pattern"* ]]; then
      violations+=("$mod (matched $pattern)")
      break
    fi
  done
done < <(go list -m all)

if [[ ${#violations[@]} -gt 0 ]]; then
  echo "root module graph contains forbidden dependencies:" >&2
  printf '  %s\n' "${violations[@]}" >&2
  exit 1
fi

echo "root dependency policy: ok"
