package squirrel

import (
	"strings"
	"testing"
)

var (
	testDebugUpdateSQL    = Update("table").SetMap(Eq{"x": 1, "y": "val"})
	expectedDebugUpateSQL = "UPDATE table SET x = '1', y = 'val'"
)

func TestDebugSqlizerUpdateColon(t *testing.T) {
	t.Parallel()
	testDebugUpdateSQL.PlaceholderFormat(Colon)
	assertEqual(t, expectedDebugUpateSQL, DebugSqlizer(testDebugUpdateSQL))
}

func TestDebugSqlizerUpdateAtp(t *testing.T) {
	t.Parallel()
	testDebugUpdateSQL.PlaceholderFormat(AtP)
	assertEqual(t, expectedDebugUpateSQL, DebugSqlizer(testDebugUpdateSQL))
}

func TestDebugSqlizerUpdateDollar(t *testing.T) {
	t.Parallel()
	testDebugUpdateSQL.PlaceholderFormat(Dollar)
	assertEqual(t, expectedDebugUpateSQL, DebugSqlizer(testDebugUpdateSQL))
}

func TestDebugSqlizerUpdateQuestion(t *testing.T) {
	t.Parallel()
	testDebugUpdateSQL.PlaceholderFormat(Question)
	assertEqual(t, expectedDebugUpateSQL, DebugSqlizer(testDebugUpdateSQL))
}

var testDebugDeleteSQL = Delete("table").Where(And{
	Eq{"column": "val"},
	Eq{"other": 1},
})
var expectedDebugDeleteSQL = "DELETE FROM table WHERE (column = 'val' AND other = '1')"

func TestDebugSqlizerDeleteColon(t *testing.T) {
	t.Parallel()
	testDebugDeleteSQL.PlaceholderFormat(Colon)
	assertEqual(t, expectedDebugDeleteSQL, DebugSqlizer(testDebugDeleteSQL))
}

func TestDebugSqlizerDeleteAtp(t *testing.T) {
	t.Parallel()
	testDebugDeleteSQL.PlaceholderFormat(AtP)
	assertEqual(t, expectedDebugDeleteSQL, DebugSqlizer(testDebugDeleteSQL))
}

func TestDebugSqlizerDeleteDollar(t *testing.T) {
	t.Parallel()
	testDebugDeleteSQL.PlaceholderFormat(Dollar)
	assertEqual(t, expectedDebugDeleteSQL, DebugSqlizer(testDebugDeleteSQL))
}

func TestDebugSqlizerDeleteQuestion(t *testing.T) {
	t.Parallel()
	testDebugDeleteSQL.PlaceholderFormat(Question)
	assertEqual(t, expectedDebugDeleteSQL, DebugSqlizer(testDebugDeleteSQL))
}

var (
	testDebugInsertSQL     = Insert("table").Values(1, "test")
	expectedDebugInsertSQL = "INSERT INTO table VALUES ('1','test')"
)

func TestDebugSqlizerInsertColon(t *testing.T) {
	t.Parallel()
	testDebugInsertSQL.PlaceholderFormat(Colon)
	assertEqual(t, expectedDebugInsertSQL, DebugSqlizer(testDebugInsertSQL))
}

func TestDebugSqlizerInsertAtp(t *testing.T) {
	t.Parallel()
	testDebugInsertSQL.PlaceholderFormat(AtP)
	assertEqual(t, expectedDebugInsertSQL, DebugSqlizer(testDebugInsertSQL))
}

func TestDebugSqlizerInsertDollar(t *testing.T) {
	t.Parallel()
	testDebugInsertSQL.PlaceholderFormat(Dollar)
	assertEqual(t, expectedDebugInsertSQL, DebugSqlizer(testDebugInsertSQL))
}

func TestDebugSqlizerInsertQuestion(t *testing.T) {
	t.Parallel()
	testDebugInsertSQL.PlaceholderFormat(Question)
	assertEqual(t, expectedDebugInsertSQL, DebugSqlizer(testDebugInsertSQL))
}

var testDebugSelectSQL = Select("*").From("table").Where(And{
	Eq{"column": "val"},
	Eq{"other": 1},
})
var expectedDebugSelectSQL = "SELECT * FROM table WHERE (column = 'val' AND other = '1')"

func TestDebugSqlizerSelectColon(t *testing.T) {
	t.Parallel()
	testDebugSelectSQL.PlaceholderFormat(Colon)
	assertEqual(t, expectedDebugSelectSQL, DebugSqlizer(testDebugSelectSQL))
}

func TestDebugSqlizerSelectAtp(t *testing.T) {
	t.Parallel()
	testDebugSelectSQL.PlaceholderFormat(AtP)
	assertEqual(t, expectedDebugSelectSQL, DebugSqlizer(testDebugSelectSQL))
}

func TestDebugSqlizerSelectDollar(t *testing.T) {
	t.Parallel()
	testDebugSelectSQL.PlaceholderFormat(Dollar)
	assertEqual(t, expectedDebugSelectSQL, DebugSqlizer(testDebugSelectSQL))
}

func TestDebugSqlizerSelectQuestion(t *testing.T) {
	t.Parallel()
	testDebugSelectSQL.PlaceholderFormat(Question)
	assertEqual(t, expectedDebugSelectSQL, DebugSqlizer(testDebugSelectSQL))
}

func TestDebugSqlizer(t *testing.T) {
	t.Parallel()
	sqlizer := Expr("x = ? AND y = ? AND z = '??'", 1, "text")
	expectedDebug := "x = '1' AND y = 'text' AND z = '?'"
	assertEqual(t, expectedDebug, DebugSqlizer(sqlizer))
}

func TestDebugSqlizerErrors(t *testing.T) {
	t.Parallel()
	errorMsg := DebugSqlizer(Expr("x = ?", 1, 2)) // Not enough placeholders
	assertTrue(t, strings.HasPrefix(errorMsg, "[DebugSqlizer error: "))

	errorMsg = DebugSqlizer(Expr("x = ? AND y = ?", 1)) // Too many placeholders
	assertTrue(t, strings.HasPrefix(errorMsg, "[DebugSqlizer error: "))

	errorMsg = DebugSqlizer(Lt{"x": nil}) // Cannot use nil values with Lt
	assertTrue(t, strings.HasPrefix(errorMsg, "[ToSql error: "))
}
