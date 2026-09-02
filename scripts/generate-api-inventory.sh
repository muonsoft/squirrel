#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 <git-ref> <output-file>" >&2
  exit 64
}

[[ $# -eq 2 ]] || usage

ref=$1
out=$2

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

git -C "$repo_root" archive "$ref" -- '*.go' | tar -x -C "$tmp_dir"

go run "$repo_root/tools/api-surface/main.go" "$tmp_dir" >"$out"
