---
name: squirrel-next-task
description: Execute exactly one eligible story from the muonsoft/squirrel implementation tracker, verify and commit it, then write the Ralph loop status contract. Use for autonomous backlog execution and Ralph fresh-process iterations; do not use for unrelated ad hoc changes.
---

# Execute the next Squirrel story

Operate autonomously. In Ralph mode never ask a question, switch mode, wait for input,
push Git refs, or execute more than one story.

## Load context

Read, in order:

1. `AGENTS.md`
2. `docs/FOUNDATION.md`
3. `IMPLEMENTATION_TRACKER.md` status table, common contract, and selected story
4. The specification sections referenced by that story
5. `UPSTREAM.md` when provenance or compatibility is involved

The foundation resolves routine choices. A change that contradicts it is a design
blocker, not permission to edit it.

## Select exactly one story

Use the status table as the sole mutable queue:

1. If one story is `IN_PROGRESS`, resume it. Never abandon it for a later story.
2. Otherwise select the first `TODO` story whose dependencies are all `DONE`.
3. If all stories are `DONE`, report `COMPLETE` without editing or committing.
4. Multiple `IN_PROGRESS` stories, an unmet dependency, or a dirty worktree unrelated
   to the selected story is a stop condition.

When starting a `TODO` story, change only its table status to `IN_PROGRESS` before
implementation. Preserve partial work when resuming after a crashed iteration.

## Implement and verify

- Execute only the selected story's scope and necessary proving tests/docs.
- Follow the story's acceptance criteria and required checks literally.
- Inspect existing behavior before editing and preserve unrelated user changes.
- Fix in-scope failures. Do not weaken assertions, skip required checks, or broaden API
  scope to manufacture success.
- If a required service/tool is unavailable, stop honestly. A skipped PostgreSQL suite
  never satisfies a story that requires real execution.
- For a story requiring PostgreSQL, use `POSTGRES_TEST_DSN` when set. Otherwise, if
  Docker Compose is available, provision the repository's `docker-compose.test.yml`,
  wait for readiness, export its documented test DSN, and stop only the resources this
  story started after verification. Stop at preflight only when neither path works.

On success, update the story to `DONE`, stage only related files, and make one atomic
commit with the required TASK ID. Confirm the worktree is clean. Do not push.

On failure, leave honest recoverable state (`IN_PROGRESS` for retryable work; `BLOCKED`
only for a real external/design blocker). Do not commit a knowingly failing story.

## Choose the loop result

Ralph mode is active when the invocation prompt says `mode ralph`.
`RALPH_UNATTENDED` is true only for `1`, `true`, or `yes` (case-insensitive).

- Ordinary successful story: `RELAUNCH@next`.
- Successful milestone story in attended mode: `HUMAN@milestone`.
- Successful milestone in unattended mode: `RELAUNCH@next`.
- TASK-018 successful with verified hosted release workflow and no local tag:
  `COMPLETE`; publication remains a maintainer dispatch.
- All stories already done: `COMPLETE`.
- Missing source/service/tool: `STOPPED@preflight`.
- Dirty/unexpected Git state: `STOPPED@git`.
- Required verification still failing: `STOPPED@verification`.
- Foundation/public-scope decision required: `BLOCKED@design`.

Never emit `CONTINUE` in Ralph mode.

## Notify through AgentMem and write status

Before exiting Ralph mode, always notify through the configured `agentmem` MCP server
and then atomically write `var/loop-status.txt`. Keep notes to one line and never
include secrets or SQL args.

Build a message no longer than 1500 characters:

```text
muonsoft/squirrel: <Loop>
Story: <ID> <short slug, or —>
Orchestration: <status>
Next: <next ID or —>
Notes: <one line>
```

Call AgentMem MCP tool `notify.send` exactly once with:

```json
{
  "title": "muonsoft/squirrel Ralph",
  "message": "<message above>",
  "silent": false
}
```

Do not substitute a webhook, desktop notification, or shell command for this MCP call.
Record `sent` only when the tool reports success; otherwise record `failed`. Do not
retry a failed call in the same process because delivery may have succeeded before the
error was returned. Then call:

```bash
scripts/ralph-status.sh \
  '<Loop>' '<ID and slug, or —>' '<orchestration>' \
  '<next ID or —>' '<one-line notes>' '<sent|failed>'
```

The status must reflect the actual tracker/commit state. Write it even when stopped.
If implementation tools fail, use remaining shell capability to write the status before
returning. A notification failure does not falsify a completed story, but it must be
visible as `NOTIFY: failed`. `NOTIFY: skipped` is not valid in Ralph mode.

In non-Ralph manual use, execute/commit one story and report normally; the status file
and notification contract are optional unless the caller requests them.
