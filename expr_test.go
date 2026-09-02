package squirrel

import (
	dbsql "database/sql"
	"fmt"
	"testing"
)

func TestConcatExpr(t *testing.T) {
	t.Parallel()
	b := ConcatExpr("COALESCE(name,", Expr("CONCAT(?,' ',?)", "f", "l"), ")")
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "COALESCE(name,CONCAT(?,' ',?))"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"f", "l"}
	assertEqual(t, expectedArgs, args)
}

func TestConcatExprBadType(t *testing.T) {
	t.Parallel()
	b := ConcatExpr("prefix", 123, "suffix")
	_, _, err := b.ToSql()
	mustError(t, err)
	assertContains(t, err.Error(), "123 is not")
}

func TestEqToSql(t *testing.T) {
	t.Parallel()
	b := Eq{"id": 1}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id = ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestEqEmptyToSql(t *testing.T) {
	t.Parallel()
	sql, args, err := Eq{}.ToSql()
	mustNoError(t, err)

	expectedSql := "(1=1)"
	assertEqual(t, expectedSql, sql)
	assertEmpty(t, args)
}

func TestEqInToSql(t *testing.T) {
	t.Parallel()
	b := Eq{"id": []int{1, 2, 3}}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id IN (?,?,?)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1, 2, 3}
	assertEqual(t, expectedArgs, args)
}

func TestNotEqToSql(t *testing.T) {
	t.Parallel()
	b := NotEq{"id": 1}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id <> ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestEqNotInToSql(t *testing.T) {
	t.Parallel()
	b := NotEq{"id": []int{1, 2, 3}}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id NOT IN (?,?,?)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1, 2, 3}
	assertEqual(t, expectedArgs, args)
}

func TestEqInEmptyToSql(t *testing.T) {
	t.Parallel()
	b := Eq{"id": []int{}}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "(1=0)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{}
	assertEqual(t, expectedArgs, args)
}

func TestNotEqInEmptyToSql(t *testing.T) {
	t.Parallel()
	b := NotEq{"id": []int{}}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "(1=1)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{}
	assertEqual(t, expectedArgs, args)
}

func TestEqBytesToSql(t *testing.T) {
	t.Parallel()
	b := Eq{"id": []byte("test")}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id = ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{[]byte("test")}
	assertEqual(t, expectedArgs, args)
}

func TestLtToSql(t *testing.T) {
	t.Parallel()
	b := Lt{"id": 1}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id < ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestLtOrEqToSql(t *testing.T) {
	t.Parallel()
	b := LtOrEq{"id": 1}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id <= ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestGtToSql(t *testing.T) {
	t.Parallel()
	b := Gt{"id": 1}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id > ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestGtOrEqToSql(t *testing.T) {
	t.Parallel()
	b := GtOrEq{"id": 1}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "id >= ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestExprNilToSql(t *testing.T) {
	t.Parallel()
	var b Sqlizer
	b = NotEq{"name": nil}
	sql, args, err := b.ToSql()
	mustNoError(t, err)
	assertEmpty(t, args)

	expectedSql := "name IS NOT NULL"
	assertEqual(t, expectedSql, sql)

	b = Eq{"name": nil}
	sql, args, err = b.ToSql()
	mustNoError(t, err)
	assertEmpty(t, args)

	expectedSql = "name IS NULL"
	assertEqual(t, expectedSql, sql)
}

func TestNullTypeString(t *testing.T) {
	t.Parallel()
	var b Sqlizer
	var name dbsql.NullString

	b = Eq{"name": name}
	sql, args, err := b.ToSql()

	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "name IS NULL", sql)

	mustNoError(t, name.Scan("Name"))
	b = Eq{"name": name}
	sql, args, err = b.ToSql()

	mustNoError(t, err)
	assertEqual(t, []any{"Name"}, args)
	assertEqual(t, "name = ?", sql)
}

func TestNullTypeInt64(t *testing.T) {
	t.Parallel()
	var userID dbsql.NullInt64
	mustNoError(t, userID.Scan(nil))
	b := Eq{"user_id": userID}
	sql, args, err := b.ToSql()

	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "user_id IS NULL", sql)

	mustNoError(t, userID.Scan(int64(10)))
	b = Eq{"user_id": userID}
	sql, args, err = b.ToSql()

	mustNoError(t, err)
	assertEqual(t, []any{int64(10)}, args)
	assertEqual(t, "user_id = ?", sql)
}

func TestNilPointer(t *testing.T) {
	t.Parallel()
	var name *string = nil
	eq := Eq{"name": name}
	sql, args, err := eq.ToSql()

	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "name IS NULL", sql)

	neq := NotEq{"name": name}
	sql, args, err = neq.ToSql()

	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "name IS NOT NULL", sql)

	var ids *[]int = nil
	eq = Eq{"id": ids}
	sql, args, err = eq.ToSql()
	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "id IS NULL", sql)

	neq = NotEq{"id": ids}
	sql, args, err = neq.ToSql()
	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "id IS NOT NULL", sql)

	var ida *[3]int = nil
	eq = Eq{"id": ida}
	sql, args, err = eq.ToSql()
	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "id IS NULL", sql)

	neq = NotEq{"id": ida}
	sql, args, err = neq.ToSql()
	mustNoError(t, err)
	assertEmpty(t, args)
	assertEqual(t, "id IS NOT NULL", sql)
}

func TestNotNilPointer(t *testing.T) {
	t.Parallel()
	c := "Name"
	name := &c
	eq := Eq{"name": name}
	sql, args, err := eq.ToSql()

	mustNoError(t, err)
	assertEqual(t, []any{"Name"}, args)
	assertEqual(t, "name = ?", sql)

	neq := NotEq{"name": name}
	sql, args, err = neq.ToSql()

	mustNoError(t, err)
	assertEqual(t, []any{"Name"}, args)
	assertEqual(t, "name <> ?", sql)

	s := []int{1, 2, 3}
	ids := &s
	eq = Eq{"id": ids}
	sql, args, err = eq.ToSql()
	mustNoError(t, err)
	assertEqual(t, []any{1, 2, 3}, args)
	assertEqual(t, "id IN (?,?,?)", sql)

	neq = NotEq{"id": ids}
	sql, args, err = neq.ToSql()
	mustNoError(t, err)
	assertEqual(t, []any{1, 2, 3}, args)
	assertEqual(t, "id NOT IN (?,?,?)", sql)

	a := [3]int{1, 2, 3}
	ida := &a
	eq = Eq{"id": ida}
	sql, args, err = eq.ToSql()
	mustNoError(t, err)
	assertEqual(t, []any{1, 2, 3}, args)
	assertEqual(t, "id IN (?,?,?)", sql)

	neq = NotEq{"id": ida}
	sql, args, err = neq.ToSql()
	mustNoError(t, err)
	assertEqual(t, []any{1, 2, 3}, args)
	assertEqual(t, "id NOT IN (?,?,?)", sql)
}

func TestEmptyAndToSql(t *testing.T) {
	t.Parallel()
	sql, args, err := And{}.ToSql()
	mustNoError(t, err)

	expectedSql := "(1=1)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{}
	assertEqual(t, expectedArgs, args)
}

func TestEmptyOrToSql(t *testing.T) {
	t.Parallel()
	sql, args, err := Or{}.ToSql()
	mustNoError(t, err)

	expectedSql := "(1=0)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{}
	assertEqual(t, expectedArgs, args)
}

func TestLikeToSql(t *testing.T) {
	t.Parallel()
	b := Like{"name": "%irrel"}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "name LIKE ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"%irrel"}
	assertEqual(t, expectedArgs, args)
}

func TestNotLikeToSql(t *testing.T) {
	t.Parallel()
	b := NotLike{"name": "%irrel"}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "name NOT LIKE ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"%irrel"}
	assertEqual(t, expectedArgs, args)
}

func TestILikeToSql(t *testing.T) {
	t.Parallel()
	b := ILike{"name": "sq%"}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "name ILIKE ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"sq%"}
	assertEqual(t, expectedArgs, args)
}

func TestNotILikeToSql(t *testing.T) {
	t.Parallel()
	b := NotILike{"name": "sq%"}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "name NOT ILIKE ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"sq%"}
	assertEqual(t, expectedArgs, args)
}

func TestSqlEqOrder(t *testing.T) {
	t.Parallel()
	b := Eq{"a": 1, "b": 2, "c": 3}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "a = ? AND b = ? AND c = ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1, 2, 3}
	assertEqual(t, expectedArgs, args)
}

func TestSqlLtOrder(t *testing.T) {
	t.Parallel()
	b := Lt{"a": 1, "b": 2, "c": 3}
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "a < ? AND b < ? AND c < ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1, 2, 3}
	assertEqual(t, expectedArgs, args)
}

func TestExprEscaped(t *testing.T) {
	t.Parallel()
	b := Expr("count(??)", Expr("x"))
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "count(??)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{Expr("x")}
	assertEqual(t, expectedArgs, args)
}

func TestExprRecursion(t *testing.T) {
	t.Parallel()
	{
		b := Expr("count(?)", Expr("nullif(a,?)", "b"))
		sql, args, err := b.ToSql()
		mustNoError(t, err)

		expectedSql := "count(nullif(a,?))"
		assertEqual(t, expectedSql, sql)

		expectedArgs := []any{"b"}
		assertEqual(t, expectedArgs, args)
	}
	{
		b := Expr("extract(? from ?)", Expr("epoch"), "2001-02-03")
		sql, args, err := b.ToSql()
		mustNoError(t, err)

		expectedSql := "extract(epoch from ?)"
		assertEqual(t, expectedSql, sql)

		expectedArgs := []any{"2001-02-03"}
		assertEqual(t, expectedArgs, args)
	}
	{
		b := Expr("JOIN t1 ON ?", And{Eq{"id": 1}, Expr("NOT c1"), Expr("? @@ ?", "x", "y")})
		sql, args, err := b.ToSql()
		mustNoError(t, err)

		expectedSql := "JOIN t1 ON (id = ? AND NOT c1 AND ? @@ ?)"
		assertEqual(t, expectedSql, sql)

		expectedArgs := []any{1, "x", "y"}
		assertEqual(t, expectedArgs, args)
	}
}

func TestAggr(t *testing.T) {
	t.Parallel()
	subQuery := Select("id").From("users").Where(Eq{"company": 20})

	expectedSql := "SELECT id FROM users WHERE company = ?"
	expectedArgs := []any{20}

	// SUM
	sql, args, err := Sum(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "SUM("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)

	// AVG
	sql, args, err = Avg(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "AVG("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)

	// MAX
	sql, args, err = Max(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "MAX("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)

	// MIN
	sql, args, err = Min(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "MIN("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)

	// COUNT
	sql, args, err = Count(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "COUNT("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)

	// EXISTS
	sql, args, err = Exists(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "EXISTS ("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)

	// NOT EXISTS
	sql, args, err = NotExists(subQuery).ToSql()
	mustNoError(t, err)
	assertEqual(t, "NOT EXISTS ("+expectedSql+")", sql)
	assertEqual(t, expectedArgs, args)
}

func TestEqual(t *testing.T) {
	t.Parallel()
	q := Select("col1").
		From("table1").
		Where(
			Equal(
				Select("col2").
					From("table2"),
				2),
		)
	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedArgs := []any{2}
	assertEqual(t, expectedArgs, args)

	expectedSql := "SELECT col1 FROM table1 WHERE (SELECT col2 FROM table2) = ?"
	assertEqual(t, expectedSql, sql)
}

func TestNotEqual(t *testing.T) {
	t.Parallel()
	q := Select("col1").
		From("table1").
		Where(
			NotEqual(
				Select("col2").
					From("table2"),
				2),
		)
	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedArgs := []any{2}
	assertEqual(t, expectedArgs, args)

	expectedSql := "SELECT col1 FROM table1 WHERE (SELECT col2 FROM table2) <> ?"
	assertEqual(t, expectedSql, sql)
}

func TestGreater(t *testing.T) {
	t.Parallel()
	q := Select("col1").
		From("table1").
		Where(
			Greater(
				Select("col2").
					From("table2"),
				2),
		)
	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedArgs := []any{2}
	assertEqual(t, expectedArgs, args)

	expectedSql := "SELECT col1 FROM table1 WHERE (SELECT col2 FROM table2) > ?"
	assertEqual(t, expectedSql, sql)
}

func TestGreaterOrEqual(t *testing.T) {
	t.Parallel()
	q := Select("col1").
		From("table1").
		Where(
			GreaterOrEqual(
				Select("col2").
					From("table2"),
				2),
		)
	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedArgs := []any{2}
	assertEqual(t, expectedArgs, args)

	expectedSql := "SELECT col1 FROM table1 WHERE (SELECT col2 FROM table2) >= ?"
	assertEqual(t, expectedSql, sql)
}

func TestLess(t *testing.T) {
	t.Parallel()
	q := Select("col1").
		From("table1").
		Where(
			Less(
				Select("col2").
					From("table2"),
				2),
		)
	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedArgs := []any{2}
	assertEqual(t, expectedArgs, args)

	expectedSql := "SELECT col1 FROM table1 WHERE (SELECT col2 FROM table2) < ?"
	assertEqual(t, expectedSql, sql)
}

func TestLessOrEqual(t *testing.T) {
	t.Parallel()
	q := Select("col1").
		From("table1").
		Where(
			LessOrEqual(
				Select("col2").
					From("table2"),
				2),
		)
	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedArgs := []any{2}
	assertEqual(t, expectedArgs, args)

	expectedSql := "SELECT col1 FROM table1 WHERE (SELECT col2 FROM table2) <= ?"
	assertEqual(t, expectedSql, sql)
}

func TestIn(t *testing.T) {
	t.Parallel()
	subQuery := Select("id").From("users").Where(Eq{"company": 20})

	expectedSql := "SELECT id FROM users WHERE company = ?"

	// IN
	sql, args, err := Select("id").From("users").Where(
		And{
			In("id1", subQuery),
			In("id2", []int{1, 2, 3}),
			In("id3", []int{}),
			In("id4", []float64{1}),
			In("id5", []string{"1", "2", "3"}),
			In("id6", []bool{true, false}),
			In("id7", 1),
		}).ToSql()
	mustNoError(t, err)
	assertEqual(t, fmt.Sprintf(
		"SELECT id FROM users WHERE (id1 IN (%s) AND id2=ANY(?) AND id4=? AND id5=ANY(?) AND id6=ANY(?) AND id7=?)",
		expectedSql), sql)
	assertEqual(t, []any{
		20,
		[]int{1, 2, 3},
		float64(1),
		[]string{"1", "2", "3"},
		[]bool{true, false},
		1,
	}, args)

	// NOT IN
	sql, args, err = Select("id").From("users").Where(
		And{
			NotIn("id1", subQuery),
			NotIn("id2", []int{1, 2, 3}),
			NotIn("id3", []int{}),
			NotIn("id4", []float64{1, 2, 3}),
			NotIn("id5", []string{"1", "2", "3"}),
			NotIn("id6", []bool{true, false}),
			NotIn("id7", 1),
		}).ToSql()
	mustNoError(t, err)
	assertEqual(t, fmt.Sprintf(
		"SELECT id FROM users WHERE (id1 NOT IN (%s) AND id2<>ALL(?) AND id4<>ALL(?) AND id5<>ALL(?) AND id6<>ALL(?) AND id7<>?)",
		expectedSql), sql)
	assertEqual(t, []any{
		20,
		[]int{1, 2, 3},
		[]float64{1, 2, 3},
		[]string{"1", "2", "3"},
		[]bool{true, false},
		1,
	}, args)
}

func Test_Range(t *testing.T) {
	t.Parallel()
	sql, args, err := Range("id", 1, 10).ToSql()
	mustNoError(t, err)
	assertEqual(t, "id BETWEEN ? AND ?", sql)
	assertEqual(t, []any{1, 10}, args)

	sql, args, err = Range("id", 1, nil).ToSql()
	mustNoError(t, err)
	assertEqual(t, "id >= ?", sql)
	assertEqual(t, []any{1}, args)

	sql, args, err = Range("id", nil, 10).ToSql()
	mustNoError(t, err)
	assertEqual(t, "id <= ?", sql)
	assertEqual(t, []any{10}, args)

	sql, args, err = Range("id", nil, nil).ToSql()
	mustNoError(t, err)
	assertEmpty(t, sql)
	assertEmpty(t, args)
}

func ExampleEq() {
	sql, _, _ := Select("id", "created", "first_name").From("users").Where(Eq{
		"company": 20,
	}).ToSql()
	fmt.Println(sql)
	// Output: SELECT id, created, first_name FROM users WHERE company = ?
}

func TestNotExprToSql(t *testing.T) {
	t.Parallel()
	e := Eq{"id": 1}
	n := Not(e)
	sql, args, err := n.ToSql()
	mustNoError(t, err)

	expectedSql := "NOT (id = ?)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestNotExprNestedToSql(t *testing.T) {
	t.Parallel()
	e := Eq{"id": 1}
	n := Not(Not(e))
	sql, args, err := n.ToSql()
	mustNoError(t, err)

	expectedSql := "id = ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestCoalesceToSql(t *testing.T) {
	t.Parallel()
	b := Coalesce("value",
		Select("col1").From("table1"),
		Select("col2").From("table2"))
	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSql := "COALESCE((SELECT col1 FROM table1), (SELECT col2 FROM table2), ?)"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{"value"}
	assertEqual(t, expectedArgs, args)
}

func TestExistsAndNotExistsNestedSelect_DollarPlaceholderNumbering(t *testing.T) {
	t.Parallel()
	sb := StatementBuilder.PlaceholderFormat(Dollar)

	inner := sb.Select("1").From("s").Where("a = ?", 10)
	q1 := sb.Select("1").From("t").Where(Exists(inner)).Where("b = ?", 20)
	sql, args, err := q1.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t WHERE EXISTS (SELECT 1 FROM s WHERE a = $1) AND b = $2", sql)
	assertEqual(t, []any{10, 20}, args)

	q2 := sb.Select("1").From("t").Where(NotExists(inner)).Where("b = ?", 20)
	sql, args, err = q2.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t WHERE NOT EXISTS (SELECT 1 FROM s WHERE a = $1) AND b = $2", sql)
	assertEqual(t, []any{10, 20}, args)
}

func TestCoalesceNestedSelect_DollarPlaceholderNumbering(t *testing.T) {
	t.Parallel()
	sb := StatementBuilder.PlaceholderFormat(Dollar)

	in1 := sb.Select("x").From("a").Where("a.c = ?", 10)
	in2 := sb.Select("y").From("b").Where("b.d = ?", 20)

	co := Coalesce("fallback", in1, in2)
	q := sb.Select("id").Column(co).From("t").Where("z = ?", 30)

	sql, args, err := q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT id, COALESCE((SELECT x FROM a WHERE a.c = $1), (SELECT y FROM b WHERE b.d = $2), $3) FROM t WHERE z = $4", sql)
	assertEqual(t, []any{10, 20, "fallback", 30}, args)
}

func TestAggrNestedSelect_DollarPlaceholderNumbering(t *testing.T) {
	t.Parallel()
	sb := StatementBuilder.PlaceholderFormat(Dollar)
	inner := sb.Select("x").From("a").Where("a.c = ?", 11)

	// SUM
	q := sb.Select("id").Column(Sum(inner)).From("t").Where("b = ?", 22)
	sql, args, err := q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT id, SUM(SELECT x FROM a WHERE a.c = $1) FROM t WHERE b = $2", sql)
	assertEqual(t, []any{11, 22}, args)

	// COUNT
	q = sb.Select("id").Column(Count(inner)).From("t").Where("b = ?", 22)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT id, COUNT(SELECT x FROM a WHERE a.c = $1) FROM t WHERE b = $2", sql)
	assertEqual(t, []any{11, 22}, args)

	// MIN
	q = sb.Select("id").Column(Min(inner)).From("t").Where("b = ?", 22)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT id, MIN(SELECT x FROM a WHERE a.c = $1) FROM t WHERE b = $2", sql)
	assertEqual(t, []any{11, 22}, args)

	// MAX
	q = sb.Select("id").Column(Max(inner)).From("t").Where("b = ?", 22)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT id, MAX(SELECT x FROM a WHERE a.c = $1) FROM t WHERE b = $2", sql)
	assertEqual(t, []any{11, 22}, args)

	// AVG
	q = sb.Select("id").Column(Avg(inner)).From("t").Where("b = ?", 22)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT id, AVG(SELECT x FROM a WHERE a.c = $1) FROM t WHERE b = $2", sql)
	assertEqual(t, []any{11, 22}, args)
}

func TestComparisonsNestedSelect_DollarPlaceholderNumbering(t *testing.T) {
	t.Parallel()
	sb := StatementBuilder.PlaceholderFormat(Dollar)
	inner := sb.Select("v").From("t2").Where("w = ?", 7)

	// =
	q := sb.Select("1").From("t1").Where(Equal(inner, 5)).Where("x = ?", 9)
	sql, args, err := q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t1 WHERE (SELECT v FROM t2 WHERE w = $1) = $2 AND x = $3", sql)
	assertEqual(t, []any{7, 5, 9}, args)

	// <>
	q = sb.Select("1").From("t1").Where(NotEqual(inner, 5)).Where("x = ?", 9)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t1 WHERE (SELECT v FROM t2 WHERE w = $1) <> $2 AND x = $3", sql)
	assertEqual(t, []any{7, 5, 9}, args)

	// >
	q = sb.Select("1").From("t1").Where(Greater(inner, 5)).Where("x = ?", 9)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t1 WHERE (SELECT v FROM t2 WHERE w = $1) > $2 AND x = $3", sql)
	assertEqual(t, []any{7, 5, 9}, args)

	// >=
	q = sb.Select("1").From("t1").Where(GreaterOrEqual(inner, 5)).Where("x = ?", 9)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t1 WHERE (SELECT v FROM t2 WHERE w = $1) >= $2 AND x = $3", sql)
	assertEqual(t, []any{7, 5, 9}, args)

	// <
	q = sb.Select("1").From("t1").Where(Less(inner, 5)).Where("x = ?", 9)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t1 WHERE (SELECT v FROM t2 WHERE w = $1) < $2 AND x = $3", sql)
	assertEqual(t, []any{7, 5, 9}, args)

	// <=
	q = sb.Select("1").From("t1").Where(LessOrEqual(inner, 5)).Where("x = ?", 9)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t1 WHERE (SELECT v FROM t2 WHERE w = $1) <= $2 AND x = $3", sql)
	assertEqual(t, []any{7, 5, 9}, args)
}

func TestInNotInNestedSelect_DollarPlaceholderNumbering(t *testing.T) {
	t.Parallel()
	sb := StatementBuilder.PlaceholderFormat(Dollar)
	inner := sb.Select("id").From("ids").Where("k = ?", 10)

	q := sb.Select("1").From("t").Where(In("x", inner)).Where("y = ?", 20)
	sql, args, err := q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t WHERE x IN (SELECT id FROM ids WHERE k = $1) AND y = $2", sql)
	assertEqual(t, []any{10, 20}, args)

	q = sb.Select("1").From("t").Where(NotIn("x", inner)).Where("y = ?", 20)
	sql, args, err = q.ToSql()
	mustNoError(t, err)
	assertEqual(t, "SELECT 1 FROM t WHERE x NOT IN (SELECT id FROM ids WHERE k = $1) AND y = $2", sql)
	assertEqual(t, []any{10, 20}, args)
}
