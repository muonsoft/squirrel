# Project foundation

This file records decisions that implementation agents may rely on without asking a
human. Changes to these decisions are design changes and require a human gate.

## Identity and release policy

- Repository and Go module: `github.com/muonsoft/squirrel`.
- Package name: `squirrel`.
- Baseline: `n-r-w/squirrel v1.6.0`, commit
  `4c87dbba0f35938b7af0171bf00c4276238e7907`.
- First maintained release: `v0.1.0`; do not reuse upstream version numbering.
- Development model: releasable `main`, atomic feature/fix changes, SemVer releases.

## Product boundary

The root module converts a Go SQL-building DSL into SQL plus `[]any`. It does not
execute SQL or own application policies. Preserve familiar Masterminds core APIs,
n-r-w nested placeholder composition, CTE/DML CTE, `UPDATE ... FROM`, and small
dependency-free SQL expressions. Remove search, pagination, conditional ordering,
zero-value filtering, and stateful table-alias conveniences.

Do not add an ORM, mapper, executor, `database/sql`/pgx wrapper, transaction layer,
migration engine, identifier sanitizer, repository abstraction, or replacement for
`lann/builder` in v0.1.0.

## Compatibility policy

Priority order:

1. Correct SQL and argument ordering.
2. Typical Masterminds/squirrel v1.5.4 source compatibility.
3. Minimal root dependency graph.
4. PostgreSQL extensions.
5. Convenience helpers.

Masterminds database execution APIs remain intentionally absent. Do not preserve an
n-r-w-only helper when it violates the product boundary. For altered legacy behavior,
especially `Case`, prefer Masterminds compatibility unless tests demonstrate a
correctness problem; document every intentional difference.

## Dependencies and Go

- Keep `github.com/lann/builder` and its `github.com/lann/ps` transitive dependency.
- Root `go.mod` must contain no pgx, scany, testdock, Docker ecosystem, testify, or
  `golang.org/x/exp` dependency at v0.1.0.
- Put pgx and PostgreSQL execution tests in a nested `integration` module.
- Provision PostgreSQL outside Go tests; use `POSTGRES_TEST_DSN` and Docker Compose/CI
  service containers.
- Target minimum Go 1.25 and test both Go 1.25 and Go 1.26. A task may raise this only
  with documented code or infrastructure evidence and a human gate.

## Git and automation

- Ralph stories commit to the current local branch; they never push.
- One story produces one atomic, MR-sized commit.
- Milestones are after the compatibility decision and release-candidate validation.
- In attended mode milestones pause the loop. `RALPH_UNATTENDED=1` bypasses milestone
  pauses but never bypasses a hard blocker.
- Local agents and scripts never create or push release tags. A maintainer dispatches
  the GitHub Actions Release workflow from current `main`; after revalidation it may
  push one changelog-only release commit and publish the source-only GitHub Release,
  which creates `v0.1.0` at the verified commit.
