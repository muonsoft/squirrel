package compatibility_test

import (
	"testing"

	sq "github.com/muonsoft/squirrel"
)

// Typical Masterminds migration patterns compile and exercise ToSql without a
// database connection.
func TestSelectPatterns(t *testing.T) {
	t.Parallel()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sub := sq.Select("account_id").
		From("memberships").
		Where("tenant_key = ?", "sub-tenant")

	q := psql.Select("u.id", "u.name").
		From("users u").
		Join("orders o ON o.user_id = u.id").
		LeftJoin("profiles p ON p.user_id = u.id").
		Where(sq.Eq{"u.active": true}).
		Where(sq.And{
			sq.GtOrEq{"u.score": 10},
			sq.Lt{"u.score": 100},
		}).
		Where(sq.Or{
			sq.Eq{"u.role": "admin"},
			sq.Expr("u.role = ?", "owner"),
		}).
		Where(sq.Expr("u.account_id IN (?)", sub)).
		FromSelect(
			sq.Select("id", "name").From("archived_users").Where(sq.Eq{"deleted": false}),
			"au",
		).
		Prefix("/* select-prefix */").
		Suffix("LIMIT 50")

	sql, args, err := q.ToSql()
	if err != nil {
		t.Fatalf("Select ToSql: %v", err)
	}
	if sql == "" || len(args) == 0 {
		t.Fatalf("Select ToSql returned empty output: sql=%q args=%v", sql, args)
	}
}

func TestInsertPatterns(t *testing.T) {
	t.Parallel()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	q := psql.Insert("users").
		Columns("name", "email", "score").
		Values("alice", "alice@example.com", 42).
		Values("bob", "bob@example.com", 17).
		Prefix("/* insert-prefix */").
		Suffix("ON CONFLICT (email) DO NOTHING")

	sql, args, err := q.ToSql()
	if err != nil {
		t.Fatalf("Insert ToSql: %v", err)
	}
	if sql == "" || len(args) != 6 {
		t.Fatalf("Insert ToSql: sql=%q args=%v", sql, args)
	}
}

func TestUpdatePatterns(t *testing.T) {
	t.Parallel()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	q := psql.Update("users u").
		Set("name", "updated").
		SetMap(sq.Eq{"active": true, "score": 99}).
		From("departments d").
		Where(sq.And{
			sq.Eq{"u.department_id": sq.Expr("d.id")},
			sq.Expr("d.region = ?", "west"),
		}).
		Prefix("/* update-prefix */").
		Suffix("RETURNING u.id, u.name")

	sql, args, err := q.ToSql()
	if err != nil {
		t.Fatalf("Update ToSql: %v", err)
	}
	if sql == "" || len(args) == 0 {
		t.Fatalf("Update ToSql returned empty output: sql=%q args=%v", sql, args)
	}
}

func TestDeletePatterns(t *testing.T) {
	t.Parallel()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	q := psql.Delete("sessions").
		Where(sq.Or{
			sq.Lt{"expires_at": sq.Expr("NOW()")},
			sq.Eq{"revoked": true},
		}).
		Prefix("/* delete-prefix */").
		Suffix("RETURNING id")

	sql, args, err := q.ToSql()
	if err != nil {
		t.Fatalf("Delete ToSql: %v", err)
	}
	if sql == "" {
		t.Fatalf("Delete ToSql returned empty sql")
	}
	_ = args
}

func TestCaseAndExprPatterns(t *testing.T) {
	t.Parallel()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	caseExpr := sq.Case("status").
		When("1", "'active'").
		When("0", "'inactive'").
		Else(sq.Expr("?", "unknown"))

	q := psql.Select().
		Column(sq.Alias(caseExpr, "status_label")).
		Column(sq.Expr("COALESCE(nickname, ?)", "anonymous")).
		From("users").
		Where(sq.Eq{"id": 7})

	sql, args, err := q.ToSql()
	if err != nil {
		t.Fatalf("Case/Expr ToSql: %v", err)
	}
	if sql == "" || len(args) == 0 {
		t.Fatalf("Case/Expr ToSql returned empty output: sql=%q args=%v", sql, args)
	}
}

func TestPackageLevelBuilders(t *testing.T) {
	t.Parallel()

	// Direct package-level constructors used by many Masterminds consumers.
	selectSQL, selectArgs, err := sq.Select("id").From("items").Where(sq.Eq{"id": 1}).ToSql()
	if err != nil {
		t.Fatalf("sq.Select ToSql: %v", err)
	}
	if selectSQL == "" || len(selectArgs) != 1 {
		t.Fatalf("sq.Select: sql=%q args=%v", selectSQL, selectArgs)
	}

	insertSQL, insertArgs, err := sq.Insert("items").Columns("name").Values("widget").ToSql()
	if err != nil {
		t.Fatalf("sq.Insert ToSql: %v", err)
	}
	if insertSQL == "" || len(insertArgs) != 1 {
		t.Fatalf("sq.Insert: sql=%q args=%v", insertSQL, insertArgs)
	}

	updateSQL, updateArgs, err := sq.Update("items").Set("name", "gadget").Where(sq.Eq{"id": 1}).ToSql()
	if err != nil {
		t.Fatalf("sq.Update ToSql: %v", err)
	}
	if updateSQL == "" || len(updateArgs) != 2 {
		t.Fatalf("sq.Update: sql=%q args=%v", updateSQL, updateArgs)
	}

	deleteSQL, deleteArgs, err := sq.Delete("items").Where(sq.Eq{"id": 1}).ToSql()
	if err != nil {
		t.Fatalf("sq.Delete ToSql: %v", err)
	}
	if deleteSQL == "" || len(deleteArgs) != 1 {
		t.Fatalf("sq.Delete: sql=%q args=%v", deleteSQL, deleteArgs)
	}
}
