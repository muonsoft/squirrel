#!/usr/bin/env bash
set -euo pipefail

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [MESSAGE]; with no argument, read MESSAGE from stdin" >&2
  exit 64
fi

if [[ $# -eq 1 ]]; then
  message=$1
else
  message=$(cat)
fi
if (( ${#message} > 1500 )); then
  message="${message:0:1497}..."
fi

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
mkdir -p "$repo_root/var"
timestamp=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
printf '[%s] %s\n' "$timestamp" "$message" >>"$repo_root/var/ralph-notifications.log"
printf '\n[ralph notify]\n%s\n\n' "$message" >&2

external_ok=0
external_attempted=0
if [[ -n "${RALPH_NOTIFY_CMD:-}" ]]; then
  external_attempted=1
  if [[ ! -x "$RALPH_NOTIFY_CMD" ]]; then
    echo "RALPH_NOTIFY_CMD is not executable: $RALPH_NOTIFY_CMD" >&2
    external_ok=1
  elif ! printf '%s\n' "$message" | "$RALPH_NOTIFY_CMD"; then
    echo "external notification command failed" >&2
    external_ok=1
  fi
elif command -v notify-send >/dev/null 2>&1 && [[ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ]]; then
  external_attempted=1
  if ! notify-send "muonsoft/squirrel Ralph" "$message"; then
    external_ok=1
  fi
fi

case "${RALPH_REQUIRE_EXTERNAL_NOTIFY:-0}" in
  1|true|TRUE|yes|YES)
    if (( ! external_attempted )); then
      echo "external notification is required but not configured" >&2
      exit 1
    fi
    exit "$external_ok"
    ;;
esac

# The durable local log is the default notification channel. External delivery is an
# optional additional channel unless RALPH_REQUIRE_EXTERNAL_NOTIFY is enabled.
exit 0
