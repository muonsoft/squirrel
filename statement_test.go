package squirrel

import (
	"testing"
)

func TestStatementBuilderWhere(t *testing.T) {
	t.Parallel()
	sb := StatementBuilder.Where("x = ?", 1)

	sql, args, err := sb.Select("test").Where("y = ?", 2).ToSql()
	mustNoError(t, err)

	expectedSql := "SELECT test WHERE x = ? AND y = ?"
	assertEqual(t, expectedSql, sql)

	expectedArgs := []any{1, 2}
	assertEqual(t, expectedArgs, args)
}
