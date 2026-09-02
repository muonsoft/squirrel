#!/usr/bin/env bash
set -euo pipefail

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [MESSAGE]; local fallback only, read MESSAGE from stdin when omitted" >&2
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

# Story notifications are delivered by the Cursor process through AgentMem MCP. This
# script intentionally provides only a durable local fallback for outer-loop failures
# where no healthy agent remains to invoke an MCP tool.
exit 0
