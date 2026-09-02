# Ralph loop operator guide

The external loop starts a fresh Cursor Agent process for each tracker story. It does
not resume chats and it never performs more than one story in a process.

## Prerequisites

- Bash, Python 3, `tee`, `awk`, Git, Go, and Cursor Agent CLI (`agent` preferred).
- Cursor authentication (`agent status`) and access to model `composer-2.5`.
- Connected Cursor MCP server `agentmem` exposing `notify.send`.
- Docker Compose or `POSTGRES_TEST_DSN` before PostgreSQL-required stories.
- A clean local branch based on the prepared repository. The loop commits but never
  pushes.

## Start

```bash
bash scripts/ralph-loop.sh
RALPH_UNATTENDED=1 bash scripts/ralph-loop.sh
RALPH_MAX=50 bash scripts/ralph-loop.sh
```

Defaults use the exact non-Fast model ID `composer-2.5` and arguments:

```text
--print --output-format stream-json --stream-partial-output
--force --trust --approve-mcps --sandbox disabled --model composer-2.5
```

`--print` and closed stdin are essential; otherwise the CLI can remain in an
interactive chat and the next iteration will never start.

## Environment

| Variable | Default | Meaning |
|---|---|---|
| `RALPH_MAX` | `20` | Maximum fresh processes in one loop run |
| `RALPH_UNATTENDED` | `0` | `1/true/yes` bypasses milestone pauses, never hard blockers |
| `AGENT_BIN` | first `agent`, then `cursor` | Cursor Agent executable |
| `AGENT_ARGS` | arguments above | Full whitespace-split override; use a wrapper for embedded spaces |

Each Cursor process sends its story status through AgentMem MCP tool `notify.send` with
title `muonsoft/squirrel Ralph` and `silent: false`. Verify availability with:

```bash
agent mcp list
agent mcp list-tools agentmem
```

`var/ralph-notifications.log` and the terminal are local fallbacks only for failures of
the outer Bash loop where no healthy Cursor process remains to invoke AgentMem.

## Exit codes

| Code | Meaning | Operator action |
|---:|---|---|
| 0 | Backlog completed and TASK-018 verified `v0.1.0` | Review local tag and push explicitly if desired |
| 1 | Invalid setup, missing CLI, or unavailable AgentMem `notify.send` | Fix environment and rerun |
| 2 | Human milestone or documented blocker | Read `var/loop-status.txt`, fix/approve, rerun |
| 3 | Missing/invalid status or agent/logging protocol failure | Inspect status and raw NDJSON |
| 4 | `RALPH_MAX` reached without completion/gate | Inspect tracker, increase limit if healthy |

After exit 2, rerunning starts a fresh process. A completed milestone remains `DONE`, so
the next eligible story is selected; an interrupted `IN_PROGRESS` story is resumed.

## Artifacts

- Status contract: `var/loop-status.txt` (deleted before every iteration).
- Raw Cursor NDJSON: `var/ralph/<UTC-run-id>/iter-NNN.ndjson`.
- Outer-loop fallback alerts: `var/ralph-notifications.log`.
- Console: decoded live assistant/tool progress.

All `var/` artifacts are ignored by Git. Do not treat a missing status file as success;
the loop exits 3 specifically to prevent stale or fabricated progression.

## Safety notes

The requested defaults deliberately grant Cursor unsandboxed, force-approved tool
access. Run only in this repository with appropriate host credentials. The skill
forbids push, upstream merges, destructive history edits, and multi-story sessions,
but the CLI flags themselves provide broad local authority.
