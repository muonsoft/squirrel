# Masterminds migration compile fixture

This nested Go module exercises typical Masterminds/squirrel consumer patterns after
changing the import path to `github.com/muonsoft/squirrel`. It compiles and runs
lightweight `ToSql()` smoke tests; it does not execute SQL against a database.

## Running tests

```bash
cd compatibility
go test ./...
```

## Intentional incompatibilities

The following Masterminds or n-r-w APIs are intentionally absent or changed in the
fork. See [`docs/API_COMPATIBILITY.md`](../docs/API_COMPATIBILITY.md) for the full
symbol table and rationale.

| Area | Status |
|---|---|
| `RunWith`, `Exec`, `Query`, `StmtCache`, and related execution helpers | removed |
| `Search`, `Paginator`, `Paginate*`, `SetIDColumn` | removed |
| `OrderByCond`, `EqNotEmpty`, stateful table-alias helpers | removed |
| `Case` THEN/ELSE typed literal auto-cast (n-r-w behavior) | not preserved; Masterminds semantics restored |

Do not add intentionally failing source files for removed APIs; document differences
here and in `docs/API_COMPATIBILITY.md` instead.
