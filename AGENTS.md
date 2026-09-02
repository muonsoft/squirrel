# Project agent rules

This repository is the maintained `github.com/muonsoft/squirrel` fork described in
[`squirrel_minimal_fork_implementation_spec.md`](squirrel_minimal_fork_implementation_spec.md).

## Sources of truth

Read these before changing the library:

1. `docs/FOUNDATION.md` — fixed scope and decisions.
2. `squirrel_minimal_fork_implementation_spec.md` — full acceptance criteria.
3. `IMPLEMENTATION_TRACKER.md` — ordered, atomic implementation stories.
4. `UPSTREAM.md` — provenance and upstream policy.

When invoked by `scripts/ralph-loop.sh`, follow
`.agents/skills/squirrel-next-task/SKILL.md` exactly and execute one tracker story only.

## Non-negotiable boundaries

- Keep the package name `squirrel` and target module path `github.com/muonsoft/squirrel`.
- The root module is a pure SQL builder. Do not add database execution, scanning,
  connection, transaction, pagination-policy, or search-policy APIs.
- Preserve final-pass placeholder numbering for nested builders.
- Keep `github.com/lann/builder` for v0.1.0.
- PostgreSQL integration dependencies belong only in the nested `integration` module.
- Do not merge upstream wholesale, push branches/tags, or rewrite published history.
- Preserve MIT licensing and upstream attribution.

## Work discipline

- Preserve unrelated user changes.
- Keep each tracker story atomic and commit it only after its acceptance checks pass.
- Do not mark a story `DONE` if required verification did not run successfully.
- Never weaken tests or dependency policy merely to make a check pass.
- Do not create the `v0.1.0` tag before the release story authorizes it.
