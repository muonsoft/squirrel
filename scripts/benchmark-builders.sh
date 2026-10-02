#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
output=${1:?usage: benchmark-builders.sh OUTPUT_DIRECTORY}
mkdir -p "$output"
output=$(cd "$output" && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
git -C "$repo_root" archive b5bb755e0097874fe2b0602b9c77b66b5412924e | tar -x -C "$scratch"
cp "$repo_root/builder_benchmark_test.go" "$scratch/"
# Run sequentially to avoid competition between the two implementations.
(cd "$scratch" && go test -run '^$' -bench 'Benchmark(Builders|Select)' -benchmem -benchtime=200ms -count=5) > "$output/baseline.txt"
(cd "$repo_root" && go test -run '^$' -bench 'Benchmark(Builders|Select)' -benchmem -benchtime=200ms -count=5) > "$output/typed.txt"
