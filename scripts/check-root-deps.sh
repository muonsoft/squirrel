#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

# The post-v0.1.0 typed builder has no external runtime dependencies.
# Database and test infrastructure belong in nested modules.
modules=$(go list -m all)
if [[ "$modules" != "github.com/muonsoft/squirrel" ]]; then
  echo "root module must have no external dependencies:" >&2
  printf '%s\n' "$modules" >&2
  exit 1
fi

echo "root dependency policy: ok"
