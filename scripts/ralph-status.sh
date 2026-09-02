#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 6 ]]; then
  echo "usage: $0 LOOP STORY ORCHESTRATION NEXT NOTES NOTIFY" >&2
  exit 64
fi

loop=$1
story=$2
orchestration=$3
next_story=$4
notes=$5
notify=$6

case "$loop" in
  RELAUNCH@next|CONTINUE|HUMAN@milestone|COMPLETE|STOPPED|STOPPED@*|BLOCKED@design) ;;
  *) echo "invalid Loop value: $loop" >&2; exit 65 ;;
esac

case "$notify" in
  sent|failed|skipped) ;;
  *) echo "invalid NOTIFY value: $notify" >&2; exit 65 ;;
esac

for value in "$loop" "$story" "$orchestration" "$next_story" "$notes" "$notify"; do
  if [[ "$value" == *$'\n'* || "$value" == *$'\r'* ]]; then
    echo "status values must be single-line" >&2
    exit 65
  fi
done

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
status_dir="$repo_root/var"
status_file="$status_dir/loop-status.txt"
mkdir -p "$status_dir"
tmp_file=$(mktemp "$status_dir/.loop-status.XXXXXX")
trap 'rm -f "$tmp_file"' EXIT

printf 'Loop: %s\nStory: %s\nOrchestration: %s\nNext: %s\nNotes: %s\nNOTIFY: %s\n' \
  "$loop" "$story" "$orchestration" "$next_story" "$notes" "$notify" >"$tmp_file"
mv "$tmp_file" "$status_file"
trap - EXIT
