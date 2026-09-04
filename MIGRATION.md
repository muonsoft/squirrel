# Migration guide

This guide covers migration from `github.com/Masterminds/squirrel` v1.5.4 and from
the n-r-w/squirrel v1.6.0 baseline to `github.com/muonsoft/squirrel` v0.1.0.

## Change the module import

Replace the old import path:

```go
import sq "github.com/Masterminds/squirrel"
```

with:

```go
import sq "github.com/muonsoft/squirrel"
```

Typical pure-builder use of `Select`, `Insert`, `Update`, `Delete`, `Expr`, `Eq`,
`And`, `Or`, `Case`, `StatementBuilder`, joins, `FromSelect`, `Prefix`, and `Suffix`
continues to compile after this import-only change. The nested `compatibility` module
contains a compile fixture for these patterns.

## Database execution is intentionally absent

This fork only builds SQL and arguments. It does not provide the Masterminds execution
and statement-cache surface, including:

- `RunWith` and runner types;
- `Exec`, `ExecContext`, and `ExecWith` variants;
- `Query`, `QueryContext`, `QueryRow`, and their `*With` variants;
- `StmtCache`, `NewStmtCache`, and related wrappers;
- `WrapStdSql` and `WrapStdSqlCtx`.

Replace builder-owned execution:

```go
// Not available in this fork:
// rows, err := sq.Select("id").From("users").RunWith(db).Query()
```

with explicit construction and execution in the application layer:

```go
query, args, err := sq.Select("id").
	From("users").
	Where("state = ?", "active").
	PlaceholderFormat(sq.Dollar).
	ToSql()
if err != nil {
	return err
}

rows, err := pool.Query(ctx, query, args...)
```

Use `database/sql`, pgx, or another driver directly. Connection management,
transactions, scanning, and retries remain application concerns.

## Placeholder composition

Nested builders should be passed as `Sqlizer` values rather than converted to SQL
early. Apply `PlaceholderFormat` to the outermost statement:

```go
subquery := sq.Select("account_id").
	From("memberships").
	Where("tenant_id = ?", tenantID)

query, args, err := sq.Select("id").
	From("users").
	Where(sq.Eq{"account_id": subquery}).
	Where("state = ?", state).
	PlaceholderFormat(sq.Dollar).
	ToSql()
```

The result uses a single continuous sequence across the whole query. Do not call
`ToSql` on the child and interpolate its already-numbered `$N` SQL into the parent.

Custom nested `Sqlizer` implementations should return raw `?` placeholders. The
parent cannot renumber arbitrary preformatted `$N` placeholders.

## `Case` compatibility

The fork restores Masterminds v1.5.4 semantics for `Case` parts. Plain strings are SQL
fragments rather than bind values. Use SQL text for literals or `Expr` for explicit
binding:

```go
expression := sq.Case("state").
	When("'ready'", "1").
	Else("0")

boundExpression := sq.Case().
	When(sq.Expr("score > ?", 10), sq.Expr("?", "high")).
	Else(sq.Expr("?", "normal"))
```

The n-r-w v1.6.0 behavior that automatically cast plain numeric and string Go values
in `THEN` and `ELSE` is not retained.

## APIs removed from n-r-w/squirrel

Application policy does not belong in the core builder. Replace these helpers in the
calling application:

| Removed API | Migration |
|---|---|
| `Search` | Compose explicit `Where`, `Like`, or `ILike` expressions. |
| `Paginator`, `Paginate*`, `SetIDColumn` | Compose `Where`, `OrderBy`, `Limit`, and `Offset`. |
| `OrderByCond` and related types | Map accepted sort keys to trusted SQL identifiers in application code, then call `OrderBy`. |
| `EqNotEmpty` | Decide which zero values should be omitted before constructing `Eq`. |
| stateful `SelectBuilder.Alias` helpers | Use `Alias(Sqlizer, string)` for expressions and write trusted table aliases explicitly. |

These removals prevent business rules and identifier mappings from becoming hidden
builder behavior.

## PostgreSQL extensions retained from n-r-w

The fork retains:

- final-pass nested placeholder handling;
- `WITH` and `WITH RECURSIVE`;
- CTE bodies and final statements for SELECT/INSERT/UPDATE/DELETE;
- `UPDATE ... FROM` and `FromSelect`;
- `Exists`, `NotExists`, `Not`, scalar comparison helpers, `Coalesce`, and aggregate
  expressions;
- PostgreSQL-oriented `In` and `NotIn` helpers.

For a multi-element slice, `In("id", ids)` emits `id =ANY(?)` and passes the slice as
one argument; `NotIn` emits `<>ALL(?)`. This differs from builders that expand a slice
into `IN (?, ?, ?)`. Empty slices produce an empty condition, and subqueries use
ordinary `IN (<subquery>)`/`NOT IN (<subquery>)` forms.

## Security review during migration

Continue placing values in placeholders. Do not pass untrusted strings as SQL text or
identifiers to `From`, `Column`, `Columns`, `OrderBy`, `GroupBy`, `Join`, `Prefix`,
`Suffix`, or `Expr`. The builder does not sanitize or quote arbitrary identifiers.

## Compatibility reference

The complete exported-symbol comparison and rationale for each difference are in
[`docs/API_COMPATIBILITY.md`](docs/API_COMPATIBILITY.md). Provenance and the upstream
update policy are in [`UPSTREAM.md`](UPSTREAM.md).
