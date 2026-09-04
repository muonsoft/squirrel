# v0.1.0 implementation tracker

This file is the single source of truth for autonomous implementation. Stories are
ordered and dependency-linked. The status table is mutable; story definitions are
not changed merely to make implementation easier.

## Status model

Allowed values: `TODO`, `IN_PROGRESS`, `DONE`, `BLOCKED`.

- Select the first `IN_PROGRESS` story, otherwise the first `TODO` story whose
  dependencies are all `DONE`.
- A fresh agent executes exactly one story.
- Set `IN_PROGRESS` before implementation and `DONE` only after every acceptance check
  and required command succeeds.
- A process crash leaves `IN_PROGRESS`; the next process resumes that same story.
- Use `BLOCKED` only with a corresponding loop status and concrete notes. Never skip a
  blocked dependency.

## Backlog

| ID | Status | Depends on | Milestone after DONE | Story |
|---|---|---|---:|---|
| TASK-001 | DONE | — | no | Capture reproducible baseline evidence |
| TASK-002 | DONE | TASK-001 | no | Adopt muonsoft module identity and Go policy |
| TASK-003 | DONE | TASK-002 | no | Isolate PostgreSQL integration module |
| TASK-004 | DONE | TASK-003 | no | Remove search and pagination policy APIs |
| TASK-005 | DONE | TASK-004 | no | Remove remaining application helpers |
| TASK-006 | DONE | TASK-005 | no | Minimize root tests and dependencies |
| TASK-007 | DONE | TASK-006 | yes | Normalize Masterminds API compatibility |
| TASK-008 | DONE | TASK-007 | no | Add core nested-placeholder regressions |
| TASK-009 | DONE | TASK-008 | no | Harden placeholder edges and errors |
| TASK-010 | DONE | TASK-009 | no | Add placeholder fuzz tests |
| TASK-011 | DONE | TASK-010 | no | Harden retained expression helpers |
| TASK-012 | DONE | TASK-011 | no | Harden CTE and UPDATE FROM composition |
| TASK-013 | DONE | TASK-012 | no | Prove SQL execution in PostgreSQL |
| TASK-014 | DONE | TASK-013 | no | Add Masterminds migration compile fixture |
| TASK-015 | DONE | TASK-014 | no | Enforce CI and dependency policy |
| TASK-016 | DONE | TASK-015 | no | Complete public documentation and audit report |
| TASK-017 | DONE | TASK-016 | yes | Validate the v0.1.0 release candidate |
| TASK-018 | TODO | TASK-017 | final | Create local v0.1.0 release |

## Common story contract

Every story must:

1. Read `AGENTS.md`, `docs/FOUNDATION.md`, the full story below, and referenced spec
   sections before editing.
2. Confirm dependencies are `DONE` and the worktree contains no unrelated changes.
3. Change only the selected story's scope, including tests/docs necessary to prove it.
4. Preserve licensing, attribution, pure-builder scope, and unrelated user work.
5. Run the story's required commands. Fix failures within scope; do not weaken checks.
6. Update this table to `DONE` and create one atomic commit using subject
   `<type>: <story outcome> (<TASK-ID>)`. Do not push.
7. Leave the worktree clean except ignored Ralph artifacts.

If a required external service is unavailable, retain honest `IN_PROGRESS` state and
stop with a preflight status. If a solution requires changing `docs/FOUNDATION.md` or
the public scope, do not improvise: stop with `BLOCKED@design`.

---

## TASK-001 — Capture reproducible baseline evidence

Objective: preserve evidence about both upstream APIs and the exact n-r-w baseline
before implementation changes make comparison harder.

Scope:

- Add a small standard-library Go AST command or shell wrapper under `tools/api-surface`
  or `scripts/` that emits a deterministic exported-symbol inventory.
- Save inventories for Masterminds `v1.5.4` and n-r-w `v1.6.0` under
  `docs/audits/baseline/`, clearly recording how each was generated.
- Save the n-r-w root `go.mod`, `go list -m all`, baseline commit, unit/race result,
  and coverage summary in a concise baseline report. Large raw logs do not belong in
  Git.
- Do not alter library Go source, root dependencies, or upstream tests in this story.

Acceptance:

- Inventories include exported types, functions, variables/constants, and exported
  methods with signatures, in stable order.
- The report distinguishes a test failure from a missing external integration service.
- `git diff upstream/n-r-w-v1.6.0 -- ':(top,glob)*.go' go.mod go.sum` remains
  empty; audit tooling outside the root package is allowed.

Required checks:

```bash
go test ./...
go test -race ./...
git rev-parse upstream/n-r-w-v1.6.0^{}
```

## TASK-002 — Adopt muonsoft module identity and Go policy

Objective: make the baseline consistently identify as the maintained fork before code
or nested modules depend on it.

Scope:

- Change the root module path to `github.com/muonsoft/squirrel` and update repository
  self-imports.
- Set the root Go directive to minimum Go 1.25 as fixed in `docs/FOUNDATION.md`; do not
  add a `toolchain` directive without an evidenced need.
- Remove or adapt upstream configuration that falsely claims Go 1.26 is the minimum.
- Keep functional source behavior and dependency versions unchanged in this story.

Acceptance:

- No tracked source/config import references `github.com/n-r-w/squirrel` except
  provenance/audit documents.
- Package name remains `squirrel`.
- Both Go 1.25-compatible compilation and the current Go toolchain are represented in
  the report; if Go 1.25 is unavailable locally, record that for CI rather than
  changing the decision.

Required checks:

```bash
go mod edit -json
go test ./...
go vet ./...
```

## TASK-003 — Isolate PostgreSQL integration module

Objective: make PostgreSQL execution tests independent of the root dependency graph.

Scope:

- Replace root `itests/` with nested module `integration/` using its own `go.mod` and
  `go.sum`, `github.com/jackc/pgx/v5`, and a local `replace` to the root module.
- Remove scany, testdock/dockertest, and in-Go Docker provisioning. Use ordinary pgx
  `Scan` and `POSTGRES_TEST_DSN`; tests must skip with a clear reason when it is absent.
- Add `docker-compose.test.yml` and `integration/README.md` for deterministic local
  PostgreSQL provisioning. Do not embed credentials intended for non-test systems.
- Existing integration intent may be ported, but application helpers scheduled for
  removal must not gain new coverage.

Acceptance:

- `go test ./...` at root never enters the integration module.
- `cd integration && go test ./...` compiles and either passes against PostgreSQL or
  skips cleanly without `POSTGRES_TEST_DSN`.
- Root module graph no longer contains pgx, scany, testdock, or Docker dependencies
  after `go mod tidy`; testify/x-exp may remain only until their scheduled stories.

Required checks:

```bash
go mod tidy
go test ./...
(cd integration && go mod tidy && go test ./...)
go list -m all
```

## TASK-004 — Remove search and pagination policy APIs

Objective: remove application-level search and pagination semantics from the builder.

Scope:

- Remove `Search`, `Paginator`, `PaginatorType`, constructors/accessors,
  `Paginate`, `PaginateByPage`, `PaginateByID`, `SetIDColumn`, related constants/data
  fields, and tests/usages.
- Preserve generic `Where`, `OrderBy`, `Limit`, and `Offset` behavior.
- Do not redesign SelectBuilder or touch unrelated expression helpers.

Acceptance:

- Removed symbols and their hidden state are absent from non-audit Go source.
- Select behavior without pagination is unchanged and unit tests cover ordinary
  Limit/Offset composition.
- Root and integration modules compile after intentional call sites are removed.

Required checks:

```bash
go test ./...
(cd integration && go test ./...)
rg 'Search|Paginator|Paginate|SetIDColumn' --glob '*.go'
```

The final `rg` must find no API implementation or active usage.

## TASK-005 — Remove remaining application helpers

Objective: finish the scope cleanup and eliminate the code-level need for x/exp.

Scope:

- Remove `OrderCond`, `OrderByCond`, `OrderByCondOption`, related null-order mapping,
  duplicate filtering, and tests.
- Remove `EqNotEmpty` and tests that encode zero-value omission policy.
- Remove stateful Select table-alias helpers that rewrite columns/group/order, while
  preserving the generic expression `Alias(Sqlizer, string)`.
- Remove `golang.org/x/exp` imports and dependency.

Acceptance:

- No removed symbol remains in active Go source.
- Generic alias expressions and normal OrderBy/Eq behavior remain tested.
- Root module graph contains no `golang.org/x/exp`.

Required checks:

```bash
gofmt -w *.go
go mod tidy
go test ./...
(cd integration && go test ./...)
go list -m all
```

## TASK-006 — Minimize root tests and dependencies

Objective: make the root module dependency graph match the production SQL-builder
boundary even after tests are tidied.

Scope:

- Convert root tests from testify to standard `testing` helpers without reducing
  assertions or diagnostic quality.
- Run root tidy and remove every obsolete direct/indirect dependency.
- Add a concise dependency snapshot to the baseline audit showing before/after.
- Do not remove `github.com/lann/builder` or rewrite immutable builder internals.

Acceptance:

- Root `go.mod` directly requires only `github.com/lann/builder`; its only expected
  transitive module is `github.com/lann/ps` (plus the Go module itself).
- Root imports and module graph contain no forbidden dependency listed in spec §11.2.
- Unit and race tests preserve coverage of retained behavior.

Required checks:

```bash
go mod tidy
go mod verify
go test ./...
go test -race ./...
go list -m all
```

## TASK-007 — Normalize Masterminds API compatibility

Objective: make an explicit, regression-tested decision for every exported API and the
known `Case` behavior difference.

Scope:

- Regenerate the fork exported inventory using TASK-001 tooling and create
  `docs/API_COMPATIBILITY.md` with a complete Masterminds/n-r-w/fork symbol table and
  decision for each difference.
- Compare Masterminds v1.5.4 and current behavior, especially string/numeric values in
  searched/simple `Case`; restore Masterminds-compatible behavior by default and add
  regressions.
- Preserve intentional absence of RunWith/Exec/Query/StmtCache and all removed
  application helpers. Do not reintroduce them to improve a numeric compatibility score.

Acceptance:

- Every exported symbol in either upstream inventory has a documented keep/remove/add
  decision.
- Typical Masterminds core builders and Case patterns compile and behave as documented.
- All intentional incompatibilities have rationale tied to foundation/spec scope.

Required checks:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Milestone: after successful completion use `HUMAN@milestone` in attended mode and
`RELAUNCH@next` in unattended mode.

## TASK-008 — Add core nested-placeholder regressions

Objective: prove single final-pass Dollar numbering through the main builder graph.

Scope:

- Add a clearly named regression suite covering spec §§14.1–14.8: basic Dollar,
  nested SELECT in Where and Eq, JoinClause/join subquery, FromSelect, Column subquery,
  Update Set subquery, and at least three levels of And/Or nesting.
- In every case assert exact SQL and exact argument order with values unique enough to
  diagnose transposition.
- Fix defects only through the existing `rawSqlizer`/`nestedToSql` architecture; do not
  independently renumber nested `$N` strings.

Acceptance:

- Tests demonstrate continuous `$1..$N` numbering across parent/child boundaries.
- All builders that support raw composition use `toSqlRaw`/`nestedToSql` consistently.
- No public API expansion is introduced solely for tests.

Required checks:

```bash
go test ./...
go test -race ./...
```

## TASK-009 — Harden placeholder edges and errors

Objective: cover composition edges that commonly corrupt placeholders or conceal
invalid builders.

Scope:

- Cover spec §§14.12–14.15: Prefix/body/Suffix ordering, PostgreSQL escaped JSON
  operators (`??`, `??|`, `??&` as supported), repeated immutable builder reuse, and a
  custom Sqlizer that returns raw `?`.
- Document/test that arbitrary preformatted `$N` from a custom Sqlizer is not magically
  renumbered.
- Test and clarify errors from spec §17: missing SELECT columns, invalid Insert shape,
  invalid/empty CTE, unsupported expressions/nested Sqlizers, and placeholder errors.
- Errors must add useful context without formatting sensitive arguments.

Acceptance:

- Escaped question marks consume no arguments and Dollar sequences have no gaps.
- Reusing a base builder cannot mutate sibling queries.
- Error improvements do not introduce a custom hierarchy or expose argument values.

Required checks:

```bash
go test ./...
go test -race ./...
```

## TASK-010 — Add placeholder fuzz tests

Objective: continuously explore placeholder transformation/composition without trying
to implement a SQL parser.

Scope:

- Add one focused Go fuzz target, `FuzzPlaceholderComposition`, for replacement and
  representative nested composition.
- Seed escaped `??`, ordinary placeholders, nested builders, prefixes/suffixes, and
  zero/multiple argument cases.
- Assert no panic, no gaps/duplicates in valid Dollar sequences, correct placeholder
  count for generated valid cases, and that escaped operators consume no argument.
- Keep default fuzz runs deterministic and fast; CI smoke duration is wired later.

Acceptance:

- The fuzz target passes its seed corpus as an ordinary unit test.
- A short local fuzz run completes without failure or unbounded corpus artifacts in Git.

Required checks:

```bash
go test ./...
go test -run '^$' -fuzz '^FuzzPlaceholderComposition$' -fuzztime=10s ./...
```

## TASK-011 — Harden retained expression helpers

Objective: validate retained generic/PostgreSQL expressions against the same nested
placeholder invariant.

Scope:

- Cover Exists/NotExists, Equal comparison family, Not including double-NOT, Coalesce,
  and retained aggregate helpers with nested builders and exact args.
- Specify and test `In`/`NotIn` for empty, scalar, one-item slice, multi-item typed
  slice, subquery, and PostgreSQL ANY/ALL behavior retained from n-r-w.
- Keep uuid/pgx-specific values in integration tests; root tests may use local named
  slice types and standard-library values only.
- Record the non-obvious PostgreSQL optimization in API docs as an architectural
  concern; do not redesign the API in this story.

Acceptance:

- Helper output and edge semantics are explicit and regression-tested.
- Root module remains free of PostgreSQL driver dependencies.
- All nested expressions participate in final parent placeholder formatting.

Required checks:

```bash
go test ./...
go test -race ./...
go list -m all
```

## TASK-012 — Harden CTE and UPDATE FROM composition

Objective: prove PostgreSQL structural extensions compose across all DML forms.

Scope:

- Cover multiple CTEs, recursive anchor/recursive/final args, and CTE bodies containing
  Select, Insert, Update, and Delete.
- Add the production DML-CTE pattern: locked candidate Select, Update From candidate
  with Returning, and final Select from updated rows, with args at every level.
- Cover standalone `UPDATE ... FROM` with subqueries, CTEs, and suffix Returning.
- Validate errors for CTE without body/final statement without broad API redesign.

Acceptance:

- Exact SQL has one continuous Dollar sequence across all CTE and final statements.
- Exact args preserve lexical SQL order.
- Existing CTE and Update APIs remain source-compatible unless TASK-007 documented an
  intentional difference.

Required checks:

```bash
go test ./...
go test -race ./...
```

## TASK-013 — Prove SQL execution in PostgreSQL

Objective: prove generated SQL and bind arguments work in real PostgreSQL through pgx.

Scope:

- Build deterministic fixtures and execution tests for every scenario in spec §15:
  simple/nested Select, correlated Exists, CTE, recursive CTE, Update From, DML CTE,
  `FOR UPDATE SKIP LOCKED`, Insert/Update/Delete Returning, In/NotIn and ANY/ALL, JSON
  operators, Case, and Coalesce.
- Add typed slice coverage including `[]uuid.UUID` using a dependency located only in
  the integration module if needed.
- Tests must isolate data, clean up safely, and use `POSTGRES_TEST_DSN` only.
- When the environment does not provide the DSN, start `docker-compose.test.yml`, wait
  for PostgreSQL readiness, use its documented DSN, and stop only that Compose project
  after the checks.

Acceptance:

- All scenarios execute rather than only comparing SQL strings.
- Bind args are exercised through pgx and returned rows/effects are asserted.
- Root `go.mod` and graph remain unchanged.

Required checks:

```bash
(cd integration && go test ./...)
git diff --exit-code HEAD^ -- go.mod go.sum
```

If neither a configured PostgreSQL DSN nor working Docker Compose is available, this
story is a hard preflight stop and must not be marked `DONE` based on skipped tests.

## TASK-014 — Add Masterminds migration compile fixture

Objective: prove typical consumers compile after changing only their import path.

Scope:

- Add a small nested compile-test module under `compatibility/` with a local replace to
  the root and core examples from spec §24.
- Exercise Select, Insert, Update, Delete, Eq/And/Or/Expr, StatementBuilder Dollar,
  Case, joins, FromSelect, and common Prefix/Suffix usage.
- Include negative incompatibilities only as documentation, not intentionally failing
  source files.

Acceptance:

- `go test ./...` inside the fixture compiles typical patterns.
- Fixture dependencies do not enter root `go.mod`.
- Intentional execution API and removed-helper differences link to
  `docs/API_COMPATIBILITY.md`.

Required checks:

```bash
(cd compatibility && go mod tidy && go test ./...)
go test ./...
```

## TASK-015 — Enforce CI and dependency policy

Objective: turn all release-critical checks into reproducible local and GitHub CI
commands without polluting module dependencies.

Scope:

- Add `scripts/check-root-deps.sh` that fails for exact forbidden dependency categories
  from spec §22 while allowing pgx only in `integration`.
- Add a local aggregate validation script for format check, vet, unit, race, fuzz smoke,
  nested-module compile/tests, tidy-cleanliness, and module verification.
- Replace/adapt upstream GitHub Actions for Go 1.25 and 1.26, with PostgreSQL service
  integration execution and separately installed lint only if existing config remains
  useful. Pin actions to stable major versions.
- CI tidy checks must fail on generated diffs rather than silently commit them.

Acceptance:

- Root dependency policy detects representative forbidden module names and passes the
  actual graph.
- CI runs unit/vet on supported Go, race, fuzz smoke, compatibility fixture, and real
  PostgreSQL integration.
- CI tools do not enter root or integration `go.mod` solely for tooling.

Required checks:

```bash
bash scripts/check-root-deps.sh
bash -n scripts/*.sh
go test ./...
(cd compatibility && go test ./...)
```

## TASK-016 — Complete public documentation and audit report

Objective: make the maintained fork usable and its compatibility/safety boundaries
unambiguous before release validation.

Scope:

- Rewrite `README.md` for the muonsoft fork with all content required by spec §19:
  rationale, pure-builder scope, Go/dependency policy, nested placeholders, CTE/DML CTE,
  Update From, custom Sqlizer, migration summary, removed API, and SQL injection boundary.
- Add `MIGRATION.md`, `CHANGELOG.md` with an Unreleased v0.1.0 section, and complete
  `docs/API_COMPATIBILITY.md`/audit outputs with post-cleanup inventory and dependency
  diff.
- Preserve `LICENSE` and all upstream attribution. Update `UPSTREAM.md` only with facts
  discovered during implementation.
- Add the final report headings required by spec §29 in
  `docs/audits/v0.1.0-report.md`, leaving test results explicitly pending TASK-017.

Acceptance:

- Examples compile or are covered by tests; no removed API is advertised.
- Value placeholders versus unsafe untrusted identifiers/raw SQL are clearly separated.
- Documentation names both upstream projects and all intentional incompatibilities.

Required checks:

```bash
go test ./...
(cd compatibility && go test ./...)
rg 'github.com/n-r-w/squirrel' README.md MIGRATION.md
```

Any remaining n-r-w path in user-facing docs must be provenance, not an import example.

## TASK-017 — Validate the v0.1.0 release candidate

Objective: produce auditable evidence that the repository is releasable from a clean
checkout before any release tag exists.

Scope:

- Run the aggregate validation from TASK-015, including real PostgreSQL integration;
  use Docker Compose if no external `POSTGRES_TEST_DSN` is configured.
- Verify root and nested module tidy state, format, vet, unit, race, fuzz smoke,
  dependency policy, compatibility fixture, integration suite, attribution, API
  inventory, and absence of removed APIs.
- Fill exact commands/results and remaining risks in `docs/audits/v0.1.0-report.md`.
- Do not change public behavior merely to pass validation; fixes outside validation
  bookkeeping must keep this story `IN_PROGRESS` until revalidated.

Acceptance:

- All Definition of Done items in spec §27 are checked with evidence.
- `git status --short` is clean after the story commit.
- Local tag `v0.1.0` does not yet exist.

Required checks:

```bash
bash scripts/test-all.sh
git status --short
git tag --list v0.1.0
```

Milestone: after successful completion use `HUMAN@milestone` in attended mode and
`RELAUNCH@next` in unattended mode. Missing Docker/PostgreSQL is a hard preflight stop.

## TASK-018 — Create local v0.1.0 release

Objective: finalize the first maintained version after release-candidate approval.

Scope:

- Confirm TASK-017 is `DONE`, its commit is reachable from HEAD, validation evidence is
  complete, and the worktree is clean.
- Promote the CHANGELOG Unreleased entry to `v0.1.0` dated with the current UTC date.
- Mark this story `DONE`, create one release commit, then create annotated local tag
  `v0.1.0` pointing at that commit.
- Do not push the branch or any tag. Do not push upstream tags to origin.

Acceptance:

- `git describe --exact-match --tags HEAD` returns `v0.1.0`.
- `git show v0.1.0` shows the release commit and annotation.
- Worktree is clean and every tracker story is `DONE`.

Required checks:

```bash
git status --short
git describe --exact-match --tags HEAD
git tag --list 'v0.1.0' --format='%(objecttype) %(refname:short)'
```

On success write `Loop: COMPLETE`. If the tag already exists at another commit, stop
instead of moving or deleting it.
