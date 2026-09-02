package squirrel

import (
	"testing"
)

func TestCaseWithVal(t *testing.T) {
	t.Parallel()
	caseStmt := Case("number").
		When("1", "one").
		When("2", "two").
		Else(Expr("?", "big number"))

	qb := Select().
		Column(caseStmt).
		From("table")
	sql, args, err := qb.ToSql()

	mustNoError(t, err)

	expectedSql := "SELECT CASE number " +
		"WHEN 1 THEN one " +
		"WHEN 2 THEN two " +
		"ELSE ? " +
		"END " +
		"FROM table"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"big number"}
	assertEqual(t, expectedArgs, args)
}

func TestCaseWithComplexVal(t *testing.T) {
	t.Parallel()
	caseStmt := Case("? > ?", 10, 5).
		When("true", "'T'")

	qb := Select().
		Column(Alias(caseStmt, "complexCase")).
		From("table")
	sql, args, err := qb.ToSql()

	mustNoError(t, err)

	expectedSql := "SELECT (CASE ? > ? " +
		"WHEN true THEN 'T' " +
		"END) AS complexCase " +
		"FROM table"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{10, 5}
	assertEqual(t, expectedArgs, args)
}

func TestCaseWithNoVal(t *testing.T) {
	t.Parallel()
	caseStmt := Case().
		When(Eq{"x": 0}, "x is zero").
		When(Expr("x > ?", 1), Expr("CONCAT('x is greater than ', ?)", 2))

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	mustNoError(t, err)

	expectedSql := "SELECT CASE " +
		"WHEN x = ? THEN x is zero " +
		"WHEN x > ? THEN CONCAT('x is greater than ', ?) " +
		"END " +
		"FROM table"

	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{0, 1, 2}
	assertEqual(t, expectedArgs, args)
}

func TestCaseWithExpr(t *testing.T) {
	t.Parallel()
	caseStmt := Case(Expr("x = ?", true)).
		When("true", Expr("?", "it's true!")).
		Else("42")

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	mustNoError(t, err)

	expectedSql := "SELECT CASE x = ? " +
		"WHEN true THEN ? " +
		"ELSE 42 " +
		"END " +
		"FROM table"

	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{true, "it's true!"}
	assertEqual(t, expectedArgs, args)
}

func TestMultipleCase(t *testing.T) {
	t.Parallel()
	caseStmtNoval := Case(Expr("x = ?", true)).
		When("true", Expr("?", "it's true!")).
		Else("42")
	caseStmtExpr := Case().
		When(Eq{"x": 0}, "'x is zero'").
		When(Expr("x > ?", 1), Expr("CONCAT('x is greater than ', ?)", 2))

	qb := Select().
		Column(Alias(caseStmtNoval, "case_noval")).
		Column(Alias(caseStmtExpr, "case_expr")).
		From("table")

	sql, args, err := qb.ToSql()

	mustNoError(t, err)

	expectedSql := "SELECT " +
		"(CASE x = ? WHEN true THEN ? ELSE 42 END) AS case_noval, " +
		"(CASE WHEN x = ? THEN 'x is zero' WHEN x > ? THEN CONCAT('x is greater than ', ?) END) AS case_expr " +
		"FROM table"

	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{
		true, "it's true!",
		0, 1, 2,
	}
	assertEqual(t, expectedArgs, args)
}

func TestCaseWithNoWhenClause(t *testing.T) {
	t.Parallel()
	caseStmt := Case("something").
		Else("42")

	qb := Select().Column(caseStmt).From("table")

	_, _, err := qb.ToSql()

	mustError(t, err)

	assertEqual(t, "case expression must contain at lease one WHEN clause", err.Error())
}

func TestCaseBuilderMustSql(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestCaseBuilderMustSql should have panicked!")
		}
	}()
	Case("").MustSql()
}

// TestCaseMastermindsSearchedCase documents Masterminds v1.5.4 semantics for simple
// searched CASE expressions: string WHEN/THEN fragments are embedded as SQL text.
func TestCaseMastermindsSearchedCase(t *testing.T) {
	t.Parallel()
	caseStmt := Case().
		When("status = 'active'", "'active'").
		Else("'inactive'")

	sql, args, err := Select().Column(caseStmt).From("users").ToSql()
	mustNoError(t, err)

	expectedSQL := "SELECT CASE " +
		"WHEN status = 'active' THEN 'active' " +
		"ELSE 'inactive' " +
		"END " +
		"FROM users"
	assertEqual(t, expectedSQL, sql)
	assertEqual(t, []any(nil), args)
}

// TestCaseMastermindsSimpleCase documents Masterminds v1.5.4 simple CASE: the
// compared value and literal branches are embedded as SQL text.
func TestCaseMastermindsSimpleCase(t *testing.T) {
	t.Parallel()
	caseStmt := Case("id").
		When("1", "2").
		When("2", "'text'").
		Else("4")

	sql, args, err := Select("id", "name").
		From("users").
		Where(caseStmt).
		ToSql()
	mustNoError(t, err)

	expectedSQL := "SELECT id, name FROM users WHERE CASE id " +
		"WHEN 1 THEN 2 " +
		"WHEN 2 THEN 'text' " +
		"ELSE 4 " +
		"END"
	assertEqual(t, expectedSQL, sql)
	assertEqual(t, []any(nil), args)
}
