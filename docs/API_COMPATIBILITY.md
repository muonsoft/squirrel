# API compatibility

Exported API inventory and compatibility decisions for the muonsoft/squirrel v0.1.0 fork.

Inventories:

| Source | File | Git ref |
|---|---|---|
| Masterminds/squirrel | `docs/audits/baseline/masterminds-v1.5.4-api.txt` | `v1.5.4` |
| n-r-w/squirrel | `docs/audits/baseline/n-r-w-v1.6.0-api.txt` | `upstream/n-r-w-v1.6.0` |
| muonsoft/squirrel | `docs/audits/baseline/muonsoft-fork-api.txt` | current `main` |

Regenerate the fork inventory:

```bash
go run ./tools/api-surface . > docs/audits/baseline/muonsoft-fork-api.txt
```

## Compatibility policy

Priority order (from `docs/FOUNDATION.md`):

1. Correct SQL and argument ordering.
2. Typical Masterminds/squirrel v1.5.4 source compatibility.
3. Minimal root dependency graph.
4. PostgreSQL extensions from n-r-w.
5. Convenience helpers.

Database execution APIs (`RunWith`, `Exec`, `Query`, `StmtCache`, and related types)
remain intentionally absent. Application helpers (`Search`, `Paginator`, conditional
ordering, `EqNotEmpty`, stateful table aliases) are removed.

## `Case` behavior

The fork restores Masterminds v1.5.4 `Case` semantics:

- `When` and `Else` arguments that are plain strings or `Sqlizer` values are embedded
  through `newPart` — string fragments become SQL text, not bind parameters.
- Typed literals (`int`, `float`, `bool`) are not auto-cast; pass them as SQL text
  strings (for example `"2"`) or use `Expr` for bind parameters.
- n-r-w v1.6.0 auto-`CAST(? AS …)` for numeric and string THEN/ELSE values is
  intentionally not preserved.

See `case_test.go` (`TestCaseMastermindsSearchedCase`, `TestCaseMastermindsSimpleCase`)
for regression coverage.

## Symbol table

| Symbol | Masterminds | n-r-w | Fork | Decision | Rationale |
|---|---:|---:|---:|---|---|
| `const Asc` | no | yes | no | removed | Application helper removed per product boundary |
| `const Desc` | no | yes | no | removed | Application helper removed per product boundary |
| `const OrderNullsFirst` | no | yes | no | removed | Application helper removed per product boundary |
| `const OrderNullsLast` | no | yes | no | removed | Application helper removed per product boundary |
| `const OrderNullsUndefined` | no | yes | no | removed | Application helper removed per product boundary |
| `const PaginatorTypeByID` | no | yes | no | removed | Application helper removed per product boundary |
| `const PaginatorTypeByPage` | no | yes | no | removed | Application helper removed per product boundary |
| `const PaginatorTypeUndefined` | no | yes | no | removed | Application helper removed per product boundary |
| `func Alias` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Avg` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Case` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Coalesce` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func ConcatExpr` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Count` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Cte` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func DebugSqlizer` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Delete` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Equal` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func ExecContextWith` | yes | no | no | removed | Database execution API intentionally absent |
| `func ExecWith` | yes | no | no | removed | Database execution API intentionally absent |
| `func Exists` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Expr` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Greater` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func GreaterOrEqual` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func In` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Insert` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Less` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func LessOrEqual` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Max` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Min` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func NewStmtCache` | yes | no | no | removed | Database execution API intentionally absent |
| `func NewStmtCacheProxy` | yes | no | no | removed | Database execution API intentionally absent |
| `func NewStmtCacher` | yes | no | no | removed | Database execution API intentionally absent |
| `func Not` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func NotEqual` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func NotExists` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func NotIn` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func PaginatorByID` | no | yes | no | removed | Application helper removed per product boundary |
| `func PaginatorByPage` | no | yes | no | removed | Application helper removed per product boundary |
| `func Placeholders` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func QueryContextWith` | yes | no | no | removed | Database execution API intentionally absent |
| `func QueryRowContextWith` | yes | no | no | removed | Database execution API intentionally absent |
| `func QueryRowWith` | yes | no | no | removed | Database execution API intentionally absent |
| `func QueryWith` | yes | no | no | removed | Database execution API intentionally absent |
| `func Range` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Replace` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Select` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func Sum` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func Update` | yes | yes | yes | keep | Masterminds-compatible core API |
| `func With` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func WithRecursive` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `func WrapStdSql` | yes | no | no | removed | Database execution API intentionally absent |
| `func WrapStdSqlCtx` | yes | no | no | removed | Database execution API intentionally absent |
| `method (*Row)` | yes | no | no | removed | Database execution API intentionally absent |
| `method (*StmtCache)` | yes | no | no | removed | Database execution API intentionally absent |
| `method (And)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (CaseBuilder)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (CommonTableExpressionsBuilder)` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `method (DeleteBuilder)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (Direction)` | no | yes | no | removed | Application helper removed per product boundary |
| `method (Eq)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (EqNotEmpty)` | no | yes | no | removed | Application helper removed per product boundary |
| `method (Gt)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (GtOrEq)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (ILike)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (InsertBuilder)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (Like)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (Lt)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (LtOrEq)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (NotEq)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (NotILike)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (NotLike)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (Or)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (OrderNullsType)` | no | yes | no | removed | Application helper removed per product boundary |
| `method (Paginator)` | no | yes | no | removed | Application helper removed per product boundary |
| `method (SelectBuilder)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (StatementBuilderType)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `method (UpdateBuilder)` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type And` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type BaseRunner` | yes | no | no | removed | Database execution API intentionally absent |
| `type CaseBuilder` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type CommonTableExpressionsBuilder` | no | yes | yes | keep | PostgreSQL extension retained from n-r-w |
| `type DBProxy` | yes | no | no | removed | Database execution API intentionally absent |
| `type DBProxyBeginner` | yes | no | no | removed | Database execution API intentionally absent |
| `type DBProxyContext` | yes | no | no | removed | Database execution API intentionally absent |
| `type DeleteBuilder` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type Direction` | no | yes | no | removed | Application helper removed per product boundary |
| `type Eq` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type EqNotEmpty` | no | yes | no | removed | Application helper removed per product boundary |
| `type Execer` | yes | no | no | removed | Database execution API intentionally absent |
| `type ExecerContext` | yes | no | no | removed | Database execution API intentionally absent |
| `type Gt` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type GtOrEq` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type ILike` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type InsertBuilder` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type Like` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type Lt` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type LtOrEq` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type NotEq` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type NotILike` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type NotLike` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type Or` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type OrderByCondOption` | no | yes | no | removed | Application helper removed per product boundary |
| `type OrderCond` | no | yes | no | removed | Application helper removed per product boundary |
| `type OrderNullsType` | no | yes | no | removed | Application helper removed per product boundary |
| `type Paginator` | no | yes | no | removed | Application helper removed per product boundary |
| `type PaginatorType` | no | yes | no | removed | Application helper removed per product boundary |
| `type PlaceholderFormat` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type Preparer` | yes | no | no | removed | Database execution API intentionally absent |
| `type PreparerContext` | yes | no | no | removed | Database execution API intentionally absent |
| `type QueryRower` | yes | no | no | removed | Database execution API intentionally absent |
| `type QueryRowerContext` | yes | no | no | removed | Database execution API intentionally absent |
| `type Queryer` | yes | no | no | removed | Database execution API intentionally absent |
| `type QueryerContext` | yes | no | no | removed | Database execution API intentionally absent |
| `type Row` | yes | no | no | removed | Database execution API intentionally absent |
| `type RowScanner` | yes | no | no | removed | Database execution API intentionally absent |
| `type Runner` | yes | no | no | removed | Database execution API intentionally absent |
| `type RunnerContext` | yes | no | no | removed | Database execution API intentionally absent |
| `type SelectBuilder` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type Sqlizer` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type StatementBuilderType` | yes | yes | yes | keep | Masterminds-compatible core API |
| `type StdSql` | yes | no | no | removed | Database execution API intentionally absent |
| `type StdSqlCtx` | yes | no | no | removed | Database execution API intentionally absent |
| `type StmtCache` | yes | no | no | removed | Database execution API intentionally absent |
| `type UpdateBuilder` | yes | yes | yes | keep | Masterminds-compatible core API |
| `var AtP` | yes | yes | yes | keep | Masterminds-compatible core API |
| `var Colon` | yes | yes | yes | keep | Masterminds-compatible core API |
| `var Dollar` | yes | yes | yes | keep | Masterminds-compatible core API |
| `var NoContextSupport` | yes | no | no | removed | Database execution API intentionally absent |
| `var Question` | yes | yes | yes | keep | Masterminds-compatible core API |
| `var RunnerNotQueryRunner` | yes | no | no | removed | Database execution API intentionally absent |
| `var RunnerNotSet` | yes | no | no | removed | Database execution API intentionally absent |
| `var StatementBuilder` | yes | yes | yes | keep | Masterminds-compatible core API |

## Intentional incompatibilities vs Masterminds v1.5.4

| Area | Masterminds | Fork | Rationale |
|---|---|---|---|
| Database execution | `RunWith`, `Exec`, `Query`, `StmtCache`, runner types | absent | Pure SQL builder boundary |
| CTE / DML CTE | not present | `With`, `WithRecursive`, `CommonTableExpressionsBuilder` | PostgreSQL extension from n-r-w |
| Expression helpers | limited set | `Equal`, `Exists`, `In`, aggregates, etc. | PostgreSQL extension from n-r-w |
| `UPDATE ... FROM` | not present | `UpdateBuilder.From` | PostgreSQL extension from n-r-w |
| Application helpers | some in n-r-w only | removed (`Search`, `Paginator`, …) | Product boundary |

## Removed n-r-w-only symbols

The following n-r-w v1.6.0 exports are intentionally absent:

- `Search`, `Paginator`, `Paginate*`, `SetIDColumn`, stateful `SelectBuilder.Alias`
- `OrderCond`, `OrderByCond`, `EqNotEmpty`
- All database execution and `StmtCache` APIs
