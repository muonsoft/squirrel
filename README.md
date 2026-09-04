# muonsoft/squirrel

[![Go Reference](https://pkg.go.dev/badge/github.com/muonsoft/squirrel.svg)](https://pkg.go.dev/github.com/muonsoft/squirrel)
[![CI](https://github.com/muonsoft/squirrel/actions/workflows/go.yml/badge.svg)](https://github.com/muonsoft/squirrel/actions/workflows/go.yml)

`github.com/muonsoft/squirrel` is a small, maintained fork of
[Masterminds/squirrel](https://github.com/Masterminds/squirrel), based on the nested
query and PostgreSQL work in [n-r-w/squirrel](https://github.com/n-r-w/squirrel).

It is a pure SQL builder: the package turns a composable Go DSL into a SQL string and
`[]any`. It does not execute queries, scan rows, manage connections or transactions,
or define application-level search and pagination policies.

The fork exists to preserve the familiar Squirrel API while making PostgreSQL
placeholder numbering reliable across nested builders, retaining useful CTE and
`UPDATE ... FROM` support, and keeping the root dependency graph small.

## Requirements and dependencies

- Go 1.25 or newer is the v0.1.0 release target. Release validation is required on
  Go 1.25 and Go 1.26.
- The root module depends only on `github.com/lann/builder` and its small
  `github.com/lann/ps` transitive dependency.
- PostgreSQL execution dependencies live in the separate `integration` module and do
  not enter consumers' dependency graphs.

Install the builder with:

```bash
go get github.com/muonsoft/squirrel
```

## Basic usage

```go
package main

import (
	"fmt"

	sq "github.com/muonsoft/squirrel"
)

func main() {
	query := sq.Select("id", "name").
		From("users").
		Where(sq.Eq{"state": "active"}).
		OrderBy("id")

	sql, args, err := query.ToSql()
	if err != nil {
		panic(err)
	}

	fmt.Println(sql)  // SELECT id, name FROM users WHERE state = ? ORDER BY id
	fmt.Println(args) // [active]
}
```

Builders follow immutable-style composition, so a base builder can be reused to
derive independent queries.

## Placeholders and nested builders

Write logical placeholders as `?`. `PlaceholderFormat(Dollar)` performs one final
replacement pass after the complete builder tree has been composed. Nested builders
therefore share one continuous `$1 ... $N` sequence and preserve lexical argument
order.

```go
accountIDs := sq.Select("account_id").
	From("memberships").
	Where("tenant_id = ?", 42)

query := sq.Select("id", "name").
	From("users").
	Where(sq.Eq{"account_id": accountIDs}).
	Where("state = ?", "active").
	PlaceholderFormat(sq.Dollar)

sql, args, err := query.ToSql()
// sql:  SELECT id, name FROM users WHERE account_id IN
//       (SELECT account_id FROM memberships WHERE tenant_id = $1) AND state = $2
// args: []any{42, "active"}
```

The same final-pass rule applies to subqueries in `WHERE`, `Eq`, `And`/`Or`, columns,
`FROM`, joins, update values, prefixes, suffixes, and CTEs.

To write a literal PostgreSQL question-mark operator, escape the question mark by
doubling it. For example:

```go
sq.Expr("metadata ??| array[?, ?]", "priority", "owner")
```

with `Dollar` formatting becomes:

```sql
metadata ?| array[$1, $2]
```

## CTE and recursive CTE

`With` and `WithRecursive` accept any retained statement builder as a CTE body.

```go
query := sq.With("active_users").
	As(sq.Select("id").From("users").Where("state = ?", "active")).
	Select(sq.Select("id").From("active_users").Where("id > ?", 100)).
	PlaceholderFormat(sq.Dollar)

sql, args, err := query.ToSql()
// WITH active_users AS (SELECT id FROM users WHERE state = $1)
// SELECT id FROM active_users WHERE id > $2
```

Use `WithRecursive`, or call `Recursive(true)`, when the CTE body contains an anchor
and recursive term.

## DML CTE

CTE bodies and final statements may be `SELECT`, `INSERT`, `UPDATE`, or `DELETE`.
This supports PostgreSQL queue/claim patterns without falling back to a raw statement:

```go
candidate := sq.Select("id").
	From("jobs").
	Where("state = ?", "ready").
	OrderBy("id").
	Limit(1).
	Suffix("FOR UPDATE SKIP LOCKED")

updated := sq.Update("jobs AS j").
	Set("state", "claimed").
	From("candidate AS c").
	Where("j.id = c.id").
	Suffix("RETURNING j.id, j.state")

query := sq.With("candidate").As(candidate).
	Cte("updated").As(updated).
	Select(sq.Select("id", "state").From("updated")).
	PlaceholderFormat(sq.Dollar)

sql, args, err := query.ToSql()
```

## UPDATE ... FROM

`UpdateBuilder.From` and `UpdateBuilder.FromSelect` build PostgreSQL
`UPDATE ... FROM` statements and compose with nested placeholders and suffixes:

```go
query := sq.Update("accounts AS a").
	Set("state", "disabled").
	From("expired_accounts AS e").
	Where("a.id = e.id").
	Where("e.tenant_id = ?", 42).
	Suffix("RETURNING a.id").
	PlaceholderFormat(sq.Dollar)
```

## Custom Sqlizer contract

Custom SQL fragments implement:

```go
type Sqlizer interface {
	ToSql() (string, []any, error)
}
```

When a custom `Sqlizer` is nested inside another builder, it should return raw SQL
with `?` placeholders if it is expected to participate in the parent's placeholder
formatting. The parent then applies the final `Dollar`, `Colon`, or `AtP` pass.

Do not return pre-numbered `$1`, `$2`, and so on from a nested custom `Sqlizer` and
expect the parent to renumber them. Arbitrary preformatted placeholders are preserved
as literal SQL and cannot be safely reconciled with the parent's arguments.

## PostgreSQL `In` and `NotIn`

The retained `In` and `NotIn` helpers intentionally use PostgreSQL array binding for
multi-element slices:

| Input | `In` | `NotIn` |
|---|---|---|
| scalar or one item | `column = ?` | `column <> ?` |
| multiple items | `column =ANY(?)` | `column <>ALL(?)` |
| subquery | `column IN (<query>)` | `column NOT IN (<query>)` |
| empty slice | empty condition | empty condition |

The multi-element slice is passed as a single bind argument. See
[`docs/API_COMPATIBILITY.md`](docs/API_COMPATIBILITY.md) for the complete API decision
table and compatibility details.

## SQL injection boundary

Values belong in placeholders:

```go
query := sq.Select("id").From("users").Where("user_id = ?", userID)
```

Squirrel passes these values separately in `[]any`; the database driver is responsible
for binding them.

Identifiers and raw SQL are different. Inputs passed to APIs such as `From`,
`Column`, `Columns`, `OrderBy`, `GroupBy`, `Join`, `Prefix`, `Suffix`, or the SQL text
of `Expr` are inserted into the generated SQL. Never pass untrusted user input to
these positions. This package deliberately does not validate or quote arbitrary
identifiers and is not a SQL parser or sanitizer.

## Migration from Masterminds/squirrel

For typical builder-only code, change the import path:

```go
import sq "github.com/muonsoft/squirrel"
```

Core builders and expressions remain source-compatible where that does not conflict
with correctness or the pure-builder boundary. `Case` follows Masterminds v1.5.4
semantics. The following APIs are intentionally absent:

- database execution and statement-cache APIs such as `RunWith`, `Exec`, `Query`,
  `QueryRow`, `StmtCache`, and wrappers around `database/sql`;
- n-r-w application-policy helpers such as `Search`, `Paginator`, `Paginate*`,
  `SetIDColumn`, `OrderByCond`, and `EqNotEmpty`;
- stateful select table-alias helpers; use the generic `Alias(Sqlizer, string)`
  expression instead.

See [`MIGRATION.md`](MIGRATION.md) for migration examples and the complete list of
intentional incompatibilities.

## Validation and PostgreSQL tests

Root checks:

```bash
go test ./...
go test -race ./...
go vet ./...
bash scripts/check-root-deps.sh
```

PostgreSQL execution tests are in a nested module. Set `POSTGRES_TEST_DSN`, or use the
documented Compose service:

```bash
docker compose -f docker-compose.test.yml up -d --wait
POSTGRES_TEST_DSN='postgres://squirrel:squirrel@localhost:54329/squirrel_test?sslmode=disable' \
  sh -c 'cd integration && go test ./...'
docker compose -f docker-compose.test.yml down
```

See [`integration/README.md`](integration/README.md) for the exact local workflow.

## Provenance and maintenance

The implementation baseline is n-r-w/squirrel v1.6.0 at commit
`4c87dbba0f35938b7af0171bf00c4276238e7907`. Upstream changes are reviewed and ported
individually; upstream branches are not merged wholesale. See [`UPSTREAM.md`](UPSTREAM.md)
for provenance and update policy.

## License

This project is released under the [MIT License](LICENSE). The original copyright
notices and attribution are preserved.
