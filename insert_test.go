package squirrel

import (
	"testing"
)

func TestInsertBuilderToSql(t *testing.T) {
	t.Parallel()
	b := Insert("").
		Prefix("WITH prefix AS ?", 0).
		Into("a").
		Options("DELAYED", "IGNORE").
		Columns("b", "c").
		Values(1, 2).
		Values(3, Expr("? + 1", 4)).
		Suffix("RETURNING ?", 5)

	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSQL := "WITH prefix AS ? " +
		"INSERT DELAYED IGNORE INTO a (b,c) VALUES (?,?),(?,? + 1) " +
		"RETURNING ?"
	assertEqual(t, expectedSQL, sql)

	expectedArgs := []any{0, 1, 2, 3, 4, 5}
	assertEqual(t, expectedArgs, args)
}

func TestInsertBuilderToSqlErr(t *testing.T) {
	t.Parallel()
	_, _, err := Insert("").Values(1).ToSql()
	mustError(t, err)

	_, _, err = Insert("x").ToSql()
	mustError(t, err)
}

func TestInsertBuilderMustSql(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestInsertBuilderMustSql should have panicked!")
		}
	}()
	Insert("").MustSql()
}

func TestInsertBuilderPlaceholders(t *testing.T) {
	t.Parallel()
	b := Insert("test").Values(1, 2)

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assertEqual(t, "INSERT INTO test VALUES (?,?)", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assertEqual(t, "INSERT INTO test VALUES ($1,$2)", sql)
}

func TestInsertBuilderSetMap(t *testing.T) {
	t.Parallel()
	b := Insert("table").SetMap(Eq{"field1": 1, "field2": 2, "field3": 3})

	sql, args, err := b.ToSql()
	mustNoError(t, err)

	expectedSQL := "INSERT INTO table (field1,field2,field3) VALUES (?,?,?)"
	assertEqual(t, expectedSQL, sql)

	expectedArgs := []any{1, 2, 3}
	assertEqual(t, expectedArgs, args)
}

func TestInsertBuilderSelect(t *testing.T) {
	t.Parallel()
	sb := Select("field1").From("table1").Where(Eq{"field1": 1})
	ib := Insert("table2").Columns("field1").Select(sb)

	sql, args, err := ib.ToSql()
	mustNoError(t, err)

	expectedSQL := "INSERT INTO table2 (field1) SELECT field1 FROM table1 WHERE field1 = ?"
	assertEqual(t, expectedSQL, sql)

	expectedArgs := []any{1}
	assertEqual(t, expectedArgs, args)
}

func TestInsertBuilderReplace(t *testing.T) {
	t.Parallel()
	b := Replace("table").Values(1)

	expectedSQL := "REPLACE INTO table VALUES (?)"

	sql, _, err := b.ToSql()
	mustNoError(t, err)

	assertEqual(t, expectedSQL, sql)
}

func TestInsertSelect_DollarPlaceholderNumberingConflict(t *testing.T) {
	t.Parallel()
	b := StatementBuilder.PlaceholderFormat(Dollar)

	sub := b.Select("a").From("src").Where("x = ?", 1)

	q := b.Insert("dst").
		Columns("a").
		Select(sub).
		Suffix("RETURNING id = ?", 2)

	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedSQL := "INSERT INTO dst (a) SELECT a FROM src WHERE x = $1 RETURNING id = $2"
	assertEqual(t, expectedSQL, sql)
	assertEqual(t, []any{1, 2}, args)
}

func TestInsertValuesNestedSelect_DollarPlaceholderNumberingConflict(t *testing.T) {
	t.Parallel()
	b := StatementBuilder.PlaceholderFormat(Dollar)

	inner := b.Select("y").From("t2").Where("y = ?", 7)

	q := b.Insert("t1").
		Columns("x").
		Values(inner).
		Suffix("RETURNING z = ?", 8)

	sql, args, err := q.ToSql()
	mustNoError(t, err)

	expectedSQL := "INSERT INTO t1 (x) VALUES (SELECT y FROM t2 WHERE y = $1) RETURNING z = $2"
	assertEqual(t, expectedSQL, sql)
	assertEqual(t, []any{7, 8}, args)
}
