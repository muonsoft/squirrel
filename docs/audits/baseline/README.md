# Baseline audit artifacts

Captured before fork implementation changes (TASK-001). These files preserve the
exported API surface and dependency graph of the n-r-w baseline and the Masterminds
compatibility reference.

## API inventories

| File | Source | Git ref | Commit |
|---|---|---|---|
| `masterminds-v1.5.4-api.txt` | Masterminds/squirrel | `v1.5.4` | `d8eb51bf129800f02602eaef1a44c220a69ccc36` |
| `n-r-w-v1.6.0-api.txt` | n-r-w/squirrel | `upstream/n-r-w-v1.6.0` | `4c87dbba0f35938b7af0171bf00c4276238e7907` |
| `muonsoft-fork-api.txt` | muonsoft/squirrel | post-cleanup fork | regenerated after TASK-007 |

Each line has the form `<kind> <name> <signature>` where `kind` is one of `const`,
`func`, `method`, `type`, or `var`. Entries are sorted by kind, then name, then
signature. Multiline signatures (interfaces and structs) span multiple lines until the
next entry.

## Regeneration

From the repository root:

```bash
bash scripts/generate-api-inventory.sh v1.5.4 docs/audits/baseline/masterminds-v1.5.4-api.txt
bash scripts/generate-api-inventory.sh upstream/n-r-w-v1.6.0 docs/audits/baseline/n-r-w-v1.6.0-api.txt
go run ./tools/api-surface . > docs/audits/baseline/muonsoft-fork-api.txt
```

The script uses `git archive` to extract non-test `*.go` sources at the requested ref,
then runs the standard-library AST tool at `tools/api-surface/main.go`.

## Module snapshots

- `n-r-w-v1.6.0-go.mod` — root `go.mod` at the baseline commit.
- `n-r-w-v1.6.0-modules.txt` — `go list -m all` from a clean checkout at baseline.
- `task-006-modules-before.txt` — fork root graph before test dependency cleanup.
- `task-006-modules-after.txt` — post-cleanup graph, containing only the fork,
  `github.com/lann/builder`, and `github.com/lann/ps`.

See `baseline-report.md` for test and coverage results captured at the same point.
