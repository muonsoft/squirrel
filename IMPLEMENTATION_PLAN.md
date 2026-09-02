# v0.1.0 implementation plan

The work is split into atomic stories in `IMPLEMENTATION_TRACKER.md`. The tracker is
the execution source of truth; this document explains sequencing and review intent.

## Phase 1 — Establish evidence and identity

1. Capture reproducible baseline API, dependency, coverage, and test evidence.
2. Change project identity to `github.com/muonsoft/squirrel` and apply the supported
   Go policy.

Exit condition: provenance is reproducible and imports consistently use the new
module path.

## Phase 2 — Isolate infrastructure and remove non-core scope

3. Move PostgreSQL tests to a nested module using externally provisioned PostgreSQL.
4. Remove search and pagination policy APIs.
5. Remove conditional ordering, zero-value filtering, and stateful alias helpers.
6. Convert root tests to the standard library and minimize the root module graph.

Exit condition: the root module contains the SQL builder plus `lann/builder` only;
integration tooling cannot leak into its dependency graph.

## Phase 3 — Normalize compatibility

7. Compare exported APIs and behavior with Masterminds v1.5.4, normalize `Case`, and
   document intentional differences.

This is the first human milestone because it fixes the public API direction.

## Phase 4 — Harden composition and PostgreSQL primitives

8. Add core nested-query placeholder regressions.
9. Cover placeholder edge cases, custom Sqlizers, immutability, and errors.
10. Add focused fuzz coverage for placeholder transformation.
11. Harden retained expression helpers and PostgreSQL `IN`/`NOT IN` behavior.
12. Harden CTE, recursive/DML CTE, and `UPDATE ... FROM` composition.

Exit condition: final-pass placeholder numbering is demonstrated across every builder
and retained extension without changing the public scope.

## Phase 5 — Prove real execution and migration

13. Execute the required generated SQL scenarios against PostgreSQL through pgx.
14. Add a compile-only Masterminds migration fixture.
15. Add CI and an enforceable root dependency policy.

Exit condition: both SQL string behavior and PostgreSQL execution are continuously
verified.

## Phase 6 — Document, validate, and release

16. Replace upstream-facing documentation and produce the compatibility/audit report.
17. Run final clean-room validation and prepare the release candidate.
18. After the release gate, create the local `v0.1.0` release commit/tag.

Story 17 is the release-candidate human milestone. Story 18 is the only story allowed
to tag the first version. The loop reports `COMPLETE` after that tag is verified.

## Review model

Each story is intentionally MR-sized and has explicit paths, acceptance criteria, and
commands. Ralph executes it as one local commit so later stories can build on it. A
team using hosted review can run the same story manually on a `feature/*` branch and
open one MR without changing the story definition.
