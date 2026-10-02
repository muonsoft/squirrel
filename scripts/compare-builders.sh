#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# Published v0.1.0; fetch this object with --no-tags if it is absent locally.
baseline=b5bb755e0097874fe2b0602b9c77b66b5412924e
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
mkdir "$scratch/baseline" "$scratch/evaluation"
git -C "$repo_root" archive "$baseline" | tar -x -C "$scratch/baseline"
cp "$repo_root/evaluation/"*.go "$repo_root/evaluation/go.mod" "$scratch/evaluation/"
cd "$scratch/evaluation"
go mod edit -replace="github.com/muonsoft/squirrel=$scratch/baseline"
go mod tidy
go run . > "$scratch/baseline.json"
go mod edit -replace="github.com/muonsoft/squirrel=$repo_root"
go mod tidy
go run . > "$scratch/typed.json"
cmp "$scratch/baseline.json" "$scratch/typed.json"
echo 'builder differential comparison: identical SQL, arguments, errors, and panics'
