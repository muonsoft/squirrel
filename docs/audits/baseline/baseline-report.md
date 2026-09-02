# n-r-w v1.6.0 baseline report

Captured 2026-09-02 before fork implementation changes.

## Baseline identity

| Field | Value |
|---|---|
| Repository | `github.com/n-r-w/squirrel` |
| Release | `v1.6.0` |
| Commit | `4c87dbba0f35938b7af0171bf00c4276238e7907` |
| Local marker | `upstream/n-r-w-v1.6.0` |

Verified with:

```bash
git rev-parse upstream/n-r-w-v1.6.0^{}
# 4c87dbba0f35938b7af0171bf00c4276238e7907
```

## Module graph

Root `go.mod` and the full module list are saved as:

- `n-r-w-v1.6.0-go.mod`
- `n-r-w-v1.6.0-modules.txt`

Direct root requirements at baseline: `scany`, `pgx`, `lann/builder`, `testdock`,
`testify`, and `golang.org/x/exp`.

## Exported API inventories

| Reference | Inventory file | Symbol count |
|---|---|---:|
| Masterminds `v1.5.4` (`d8eb51b`) | `masterminds-v1.5.4-api.txt` | 273 lines |
| n-r-w `v1.6.0` (`4c87dbb`) | `n-r-w-v1.6.0-api.txt` | 211 lines |

Generation procedure is documented in `README.md`.

## Tests

Commands run at baseline (library source and `go.mod` unchanged):

```bash
go test ./...
go test -race ./...
go test -coverprofile=/tmp/cover.out ./...
```

### Unit tests (`go test ./...`)

| Result | Details |
|---|---|
| **PASS** | `ok github.com/n-r-w/squirrel` |

Default `go test ./...` does not build `itests/` (build tag `itest`). That package
requires Docker-backed PostgreSQL via `testdock` and is excluded from this baseline
unit run by design, not by failure.

### Race detector (`go test -race ./...`)

| Result | Details |
|---|---|
| **PASS** | `ok github.com/n-r-w/squirrel` |

### Coverage (`go test -coverprofile=... ./...`)

| Package | Coverage |
|---|---:|
| `github.com/n-r-w/squirrel` | 91.6% of statements |
| `github.com/n-r-w/squirrel/tools/api-surface` | 0.0% (audit tool, no tests) |
| **Total** | **84.3%** |

### Integration tests (`itests/`, build tag `itest`)

Not executed in this baseline capture. They depend on a live PostgreSQL instance
provisioned through `testdock`/Docker and are orthogonal to root unit-test health.
A skipped or unavailable integration environment is **not** counted as a unit-test
failure.

## Source integrity check

Library Go sources and module files were not modified by TASK-001:

```bash
git diff upstream/n-r-w-v1.6.0 -- ':(top,glob)*.go' go.mod go.sum
# (empty)
```

Audit tooling under `tools/api-surface/`, `scripts/generate-api-inventory.sh`, and
this directory are the only additions.

## Root dependency minimization (TASK-006)

Captured 2026-09-02 after converting root tests from testify to standard-library
helpers and running `go mod tidy`.

| State | Direct requires | Full module list |
|---|---|---|
| Before TASK-006 | `lann/builder`, `testify` | `task-006-modules-before.txt` (17 modules) |
| After TASK-006 | `lann/builder` only | `task-006-modules-after.txt` (3 modules) |

Removed from the root graph: `testify` and its transitive test-only dependencies
(`go-spew`, `difflib`, `yaml.v3`, `check.v1`, and related tooling modules).

Verification after TASK-006:

```bash
go mod tidy
go mod verify
go test ./...
go test -race ./...
go list -m all
```
