#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

max_iterations=${RALPH_MAX:-20}
if [[ ! "$max_iterations" =~ ^[1-9][0-9]*$ ]]; then
  echo "RALPH_MAX must be a positive integer" >&2
  exit 1
fi

if [[ -n "${AGENT_BIN:-}" ]]; then
  if ! resolved_agent=$(command -v "$AGENT_BIN"); then
    echo "AGENT_BIN is not executable or not on PATH: $AGENT_BIN" >&2
    exit 1
  fi
elif resolved_agent=$(command -v agent); then
  :
elif resolved_agent=$(command -v cursor); then
  :
else
  echo "Cursor CLI not found; set AGENT_BIN to its executable" >&2
  exit 1
fi

if ! agentmem_tools=$("$resolved_agent" mcp list-tools agentmem 2>&1); then
  echo "AgentMem MCP is unavailable; fix 'agent mcp list' before starting Ralph" >&2
  echo "$agentmem_tools" >&2
  exit 1
fi
if [[ "$agentmem_tools" != *"notify.send"* ]]; then
  echo "AgentMem MCP does not expose required tool notify.send" >&2
  exit 1
fi

if [[ -n "${AGENT_ARGS:-}" ]]; then
  # Intentional simple whitespace splitting. Use a wrapper in AGENT_BIN for arguments
  # that themselves contain spaces.
  read -r -a agent_args <<<"$AGENT_ARGS"
else
  agent_args=(
    --print
    --output-format stream-json
    --stream-partial-output
    --force
    --trust
    --approve-mcps
    --sandbox disabled
    --model composer-2.5
  )
fi

case "${RALPH_UNATTENDED:-0}" in
  1|true|TRUE|yes|YES) unattended=1 ;;
  *) unattended=0 ;;
esac
export RALPH_UNATTENDED="$unattended"

prompt=$(cat <<'PROMPT'
Use the autonomous skill at `.agents/skills/squirrel-next-task/SKILL.md` in mode ralph.
Take exactly one eligible story from `IMPLEMENTATION_TRACKER.md`, implement it, run all
of its required verification, make its one atomic local commit, notify, and always
write `var/loop-status.txt` using the skill contract before exiting. Do not push.
Send the status notification through AgentMem MCP tool `notify.send`; do not replace it
with a webhook or shell notifier.
Do not call AskQuestion, SwitchMode, enter plan-only mode, or wait for user input.
Resolve routine choices from `AGENTS.md`, `docs/FOUNDATION.md`, the implementation spec,
the tracker story, and `UPSTREAM.md`. Stop with the documented status if those sources
cannot resolve a design decision or hard blocker. In ralph mode one process means one
story; never begin the next story in this process.
PROMPT
)

run_id="$(date -u +'%Y%m%dT%H%M%SZ')-$$"
run_dir="$repo_root/var/ralph/$run_id"
status_file="$repo_root/var/loop-status.txt"
mkdir -p "$run_dir"

display_args=()
redact_next=0
for arg in "${agent_args[@]}"; do
  if (( redact_next )); then
    display_args+=("<redacted>")
    redact_next=0
  elif [[ "$arg" == "--api-key" || "$arg" == "--header" || "$arg" == "-H" ]]; then
    display_args+=("$arg")
    redact_next=1
  else
    display_args+=("$arg")
  fi
done

for ((iteration = 1; iteration <= max_iterations; iteration++)); do
  rm -f "$status_file"
  ndjson_log=$(printf '%s/iter-%03d.ndjson' "$run_dir" "$iteration")
  printf 'ralph iteration %d/%d; bin=%q; args=' "$iteration" "$max_iterations" "$resolved_agent"
  printf ' %q' "${display_args[@]}"
  printf '; unattended=%s; log=%s\n' "$unattended" "$ndjson_log"

  set +e
  "$resolved_agent" "${agent_args[@]}" "$prompt" </dev/null \
    | tee "$ndjson_log" \
    | python3 -u "$repo_root/scripts/ralph-stream.py"
  pipeline_status=("${PIPESTATUS[@]}")
  set -e

  agent_rc=${pipeline_status[0]}
  tee_rc=${pipeline_status[1]}
  decoder_rc=${pipeline_status[2]}

  if [[ ! -f "$status_file" ]]; then
    scripts/ralph-notify.sh "muonsoft/squirrel: local fallback; iteration $iteration did not create var/loop-status.txt" || true
    echo "agent did not create $status_file (agent exit $agent_rc)" >&2
    exit 3
  fi

  loop_value=$(awk -F': ' '/^Loop:/{print $2; exit}' "$status_file" | tr -d '\r')
  if [[ -z "$loop_value" ]]; then
    scripts/ralph-notify.sh "muonsoft/squirrel: local fallback; status file has no Loop field" || true
    echo "status file has no Loop value" >&2
    exit 3
  fi

  echo "status: $loop_value (file: $status_file)"

  if (( tee_rc != 0 || decoder_rc != 0 )); then
    echo "logging pipeline failed (tee=$tee_rc decoder=$decoder_rc)" >&2
    exit 3
  fi
  if (( agent_rc != 0 )); then
    case "$loop_value" in
      HUMAN@*|STOPPED|STOPPED@*|BLOCKED@design)
        echo "agent stopped with a documented gate (exit $agent_rc)" >&2
        exit 2
        ;;
      *)
        echo "agent exited $agent_rc despite non-stop status $loop_value" >&2
        exit 3
        ;;
    esac
  fi

  case "$loop_value" in
    RELAUNCH@next)
      continue
      ;;
    CONTINUE)
      echo "warning: CONTINUE is unexpected in ralph mode; relaunching" >&2
      continue
      ;;
    COMPLETE)
      echo "ralph backlog complete"
      exit 0
      ;;
    HUMAN@*|STOPPED|STOPPED@*|BLOCKED@design)
      echo "ralph paused for operator; inspect $status_file"
      exit 2
      ;;
    *)
      echo "unknown Loop value: $loop_value" >&2
      exit 3
      ;;
  esac
done

scripts/ralph-notify.sh "muonsoft/squirrel: local fallback; RALPH_MAX=$max_iterations reached without completion or gate" || true
echo "RALPH_MAX=$max_iterations reached without completion or gate" >&2
exit 4
