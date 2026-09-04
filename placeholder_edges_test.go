package squirrel

import (
	"errors"
	"strings"
	"testing"
)

func TestPrefixBodySuffixDollarPlaceholders(t *testing.T) {
	t.Parallel()

	q := Select("event_id").
		Prefix("WITH scope AS (SELECT ? AS tenant_key)", "pfx-9001").
		From("events").
		Where("tenant_id = ?", "body-9002").
		Suffix("FOR UPDATE SKIP LOCKED").
		PlaceholderFormat(Dollar)

	assertNestedPlaceholderSQL(
		t,
		q,
		"WITH scope AS (SELECT $1 AS tenant_key) SELECT event_id FROM events WHERE tenant_id = $2 FOR UPDATE SKIP LOCKED",
		[]any{"pfx-9001", "body-9002"},
	)
}

func TestPostgreSQLJSONOperatorEscapes_Dollar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		where    string
		args     []any
		wantSQL  string
		wantArgs []any
	}{
		{
			name:     "escaped question mark operator",
			where:    "meta ?? 'format'",
			args:     nil,
			wantSQL:  "SELECT id FROM nodes WHERE meta ? 'format'",
			wantArgs: nil,
		},
		{
			name:     "escaped json exists any",
			where:    "meta->'tags' ??| array[?, ?]",
			args:     []any{"json-any-9101", "json-any-9102"},
			wantSQL:  "SELECT id FROM nodes WHERE meta->'tags' ?| array[$1, $2]",
			wantArgs: []any{"json-any-9101", "json-any-9102"},
		},
		{
			name:     "escaped json exists all",
			where:    "meta->'keys' ??& array[?, ?]",
			args:     []any{"json-all-9201", "json-all-9202"},
			wantSQL:  "SELECT id FROM nodes WHERE meta->'keys' ?& array[$1, $2]",
			wantArgs: []any{"json-all-9201", "json-all-9202"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := Select("id").From("nodes").PlaceholderFormat(Dollar)
			if len(tc.args) > 0 {
				q = q.Where(tc.where, tc.args...)
			} else {
				q = q.Where(tc.where)
			}

			assertNestedPlaceholderSQL(t, q, tc.wantSQL, tc.wantArgs)
		})
	}
}

func TestBuilderReuseDoesNotMutateSiblings(t *testing.T) {
	t.Parallel()

	base := Select("id").
		From("users").
		Where("active = ?", true)

	alice := base.Where("name = ?", "alice-9301")
	bob := base.Where("name = ?", "bob-9302")

	assertNestedPlaceholderSQL(
		t,
		base.PlaceholderFormat(Dollar),
		"SELECT id FROM users WHERE active = $1",
		[]any{true},
	)
	assertNestedPlaceholderSQL(
		t,
		alice.PlaceholderFormat(Dollar),
		"SELECT id FROM users WHERE active = $1 AND name = $2",
		[]any{true, "alice-9301"},
	)
	assertNestedPlaceholderSQL(
		t,
		bob.PlaceholderFormat(Dollar),
		"SELECT id FROM users WHERE active = $1 AND name = $2",
		[]any{true, "bob-9302"},
	)
}

type questionMarkSqlizer struct {
	sql  string
	args []any
}

func (q questionMarkSqlizer) ToSql() (string, []any, error) {
	return q.sql, q.args, nil
}

type preformattedDollarSqlizer struct {
	sql  string
	args []any
}

func (p preformattedDollarSqlizer) ToSql() (string, []any, error) {
	return p.sql, p.args, nil
}

type failingSqlizer struct{}

func (failingSqlizer) ToSql() (string, []any, error) {
	return "", nil, errors.New("custom sqlizer failed")
}

func TestCustomSqlizerQuestionMarkPlaceholders(t *testing.T) {
	t.Parallel()

	custom := questionMarkSqlizer{
		sql:  "payload->'kind' ??| array[?, ?]",
		args: []any{"kind-a-9401", "kind-b-9402"},
	}

	q := Select("id").
		From("events").
		Where(custom).
		Where("state = ?", "parent-9403").
		PlaceholderFormat(Dollar)

	assertNestedPlaceholderSQL(
		t,
		q,
		"SELECT id FROM events WHERE payload->'kind' ?| array[$1, $2] AND state = $3",
		[]any{"kind-a-9401", "kind-b-9402", "parent-9403"},
	)
}

func TestCustomSqlizerPreformattedDollarNotRenumbered(t *testing.T) {
	t.Parallel()

	custom := preformattedDollarSqlizer{
		sql:  "tenant_id = $1",
		args: []any{"nested-9501"},
	}

	q := Select("id").
		From("events").
		Where(custom).
		Where("state = ?", "parent-9502").
		PlaceholderFormat(Dollar)

	sql, args, err := q.ToSql()
	mustNoError(t, err)

	// Preformatted "$1" from the nested Sqlizer is preserved; the parent only
	// renumbers its own question-mark placeholders.
	assertEqual(t, "SELECT id FROM events WHERE tenant_id = $1 AND state = $1", sql)
	assertEqual(t, []any{"nested-9501", "parent-9502"}, args)
}

func TestBuilderValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("select without columns", func(t *testing.T) {
		t.Parallel()

		_, _, err := Select().From("users").ToSql()
		assertValidationError(t, err, "select statements must have at least one result column")
	})

	t.Run("insert without table", func(t *testing.T) {
		t.Parallel()

		_, _, err := Insert("").Values(1).ToSql()
		assertValidationError(t, err, "insert statements must specify a table")
	})

	t.Run("insert without values or select", func(t *testing.T) {
		t.Parallel()

		_, _, err := Insert("users").ToSql()
		assertValidationError(t, err, "insert statements must have at least one set of values or select clause")
	})

	t.Run("insert nested select without columns", func(t *testing.T) {
		t.Parallel()

		_, _, err := Insert("users").Select(Select().From("accounts")).ToSql()
		assertValidationError(t, err, "select statements must have at least one result column")
	})

	t.Run("cte without body", func(t *testing.T) {
		t.Parallel()

		_, _, err := With("scope").Select(Select("id").From("users")).ToSql()
		assertValidationError(t, err, "common table expressions statements must have at least one label and subquery")
	})

	t.Run("cte without final statement", func(t *testing.T) {
		t.Parallel()

		_, _, err := With("scope").As(Select("id").From("users")).ToSql()
		assertValidationError(t, err, "common table expressions must one of the following final statement")
	})

	t.Run("unsupported concat expression type", func(t *testing.T) {
		t.Parallel()

		_, _, err := ConcatExpr("prefix", 123, "suffix").ToSql()
		assertValidationError(t, err, "is not a string or Sqlizer")
	})

	t.Run("invalid where part type", func(t *testing.T) {
		t.Parallel()

		_, _, err := newWherePart(123).ToSql()
		assertValidationError(t, err, "expected string-keyed map or string")
	})

	t.Run("invalid part type", func(t *testing.T) {
		t.Parallel()

		_, _, err := newPart(123).ToSql()
		assertValidationError(t, err, "expected string or Sqlizer")
	})

	t.Run("nested failing sqlizer in where", func(t *testing.T) {
		t.Parallel()

		_, _, err := Select("id").
			From("users").
			Where(failingSqlizer{}).
			ToSql()
		assertValidationError(t, err, "custom sqlizer failed")
	})

	t.Run("nested failing sqlizer in prefix", func(t *testing.T) {
		t.Parallel()

		_, _, err := Select("id").
			PrefixExpr(failingSqlizer{}).
			From("users").
			ToSql()
		assertValidationError(t, err, "custom sqlizer failed")
	})
}

func assertValidationError(t *testing.T, err error, wantFragment string) {
	t.Helper()
	mustError(t, err)
	assertContains(t, err.Error(), wantFragment)
	assertErrorOmitsSensitiveArgs(t, err)
}

func assertErrorOmitsSensitiveArgs(t *testing.T, err error) {
	t.Helper()
	msg := err.Error()
	for _, sensitive := range []string{"secret-", "password", "token-"} {
		if strings.Contains(msg, sensitive) {
			t.Fatalf("error message must not include sensitive argument values: %q", msg)
		}
	}
}
