package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sq "github.com/muonsoft/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpec15PostgreSQLExecution proves every integration scenario from spec §15
// against a real PostgreSQL database through pgx bind arguments.
func TestSpec15PostgreSQLExecution(t *testing.T) {
	pool, ctx := newTestPool(t)

	t.Run("simple select", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_simple (id serial PRIMARY KEY, name text NOT NULL);
INSERT INTO spec15_simple (name) VALUES ('alpha'), ('beta');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_simple") })

		q := sq.Select("name").
			From("spec15_simple").
			Where(sq.Eq{"id": 1}).
			PlaceholderFormat(sq.Dollar)

		sql, args := mustSQL(t, q)
		name := queryScalar[string](t, pool, ctx, sql, args...)
		assert.Equal(t, "alpha", name)
	})

	t.Run("nested select", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_nested (id serial PRIMARY KEY, name text NOT NULL, active boolean NOT NULL);
INSERT INTO spec15_nested (name, active) VALUES ('keep', true), ('drop', false);
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_nested") })

		inner := sq.Select("id", "name").
			From("spec15_nested").
			Where(sq.Eq{"active": true})

		q := sq.Select("name").
			FromSelect(inner, "sub").
			Where(sq.Gt{"sub.id": 0}).
			OrderBy("sub.id").
			PlaceholderFormat(sq.Dollar)

		sql, args, err := q.ToSql()
		require.NoError(t, err)

		names := queryStrings(t, pool, ctx, sql, args...)
		assert.Equal(t, []string{"keep"}, names)
	})

	t.Run("correlated exists", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_exists_parent (id serial PRIMARY KEY, name text NOT NULL);
CREATE TABLE spec15_exists_child (id serial PRIMARY KEY, parent_id bigint NOT NULL, tag text NOT NULL);
INSERT INTO spec15_exists_parent (name) VALUES ('p1'), ('p2');
INSERT INTO spec15_exists_child (parent_id, tag) VALUES (1, 'hit'), (2, 'miss');
`)
		t.Cleanup(func() {
			execSetup(t, pool, ctx, `
DROP TABLE IF EXISTS spec15_exists_child;
DROP TABLE IF EXISTS spec15_exists_parent;
`)
		})

		exists := sq.Exists(
			sq.Select("1").
				From("spec15_exists_child c").
				Where(sq.Expr("c.parent_id = p.id")).
				Where(sq.Eq{"c.tag": "hit"}),
		)

		q := sq.Select("p.name").
			From("spec15_exists_parent p").
			Where(exists).
			PlaceholderFormat(sq.Dollar)

		sql, args := mustSQL(t, q)
		names := queryStrings(t, pool, ctx, sql, args...)
		assert.Equal(t, []string{"p1"}, names)
	})

	t.Run("cte", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_cte_src (id serial PRIMARY KEY, state text NOT NULL);
INSERT INTO spec15_cte_src (state) VALUES ('open'), ('closed'), ('open');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_cte_src") })

		openRows := sq.Select("id").
			From("spec15_cte_src").
			Where(sq.Eq{"state": "open"})

		q := sq.With("open_rows").As(openRows).
			Select(sq.Select("id").From("open_rows").OrderBy("id")).
			PlaceholderFormat(sq.Dollar)

		ids := queryInt64s(t, pool, ctx, q)
		assert.Equal(t, []int64{1, 3}, ids)
	})

	t.Run("recursive cte", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_tree (id int PRIMARY KEY, parent_id int, name text NOT NULL);
INSERT INTO spec15_tree (id, parent_id, name) VALUES
	(1, NULL, 'root'),
	(2, 1, 'child-a'),
	(3, 1, 'child-b'),
	(4, 2, 'grandchild');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_tree") })

		anchor := sq.Select("id", "parent_id", "name").
			From("spec15_tree").
			Where("parent_id IS NULL")
		recursive := sq.Select("n.id", "n.parent_id", "n.name").
			From("spec15_tree n").
			Join("tree t ON n.parent_id = t.id")
		body := sq.ConcatExpr(anchor, " UNION ALL ", recursive)

		q := sq.With("tree").
			Recursive(true).
			As(body).
			Select(sq.Select("name").From("tree").Where("parent_id IS NOT NULL").OrderBy("id")).
			PlaceholderFormat(sq.Dollar)

		sql, args := mustSQL(t, q)
		names := queryStrings(t, pool, ctx, sql, args...)
		assert.Equal(t, []string{"child-a", "child-b", "grandchild"}, names)
	})

	t.Run("update from", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_accounts (id serial PRIMARY KEY, balance numeric NOT NULL);
CREATE TABLE spec15_ledger (account_id bigint NOT NULL, delta numeric NOT NULL);
INSERT INTO spec15_accounts (id, balance) VALUES (1, 100), (2, 50);
INSERT INTO spec15_ledger (account_id, delta) VALUES (1, 25), (2, 10);
`)
		t.Cleanup(func() {
			execSetup(t, pool, ctx, `
DROP TABLE IF EXISTS spec15_ledger;
DROP TABLE IF EXISTS spec15_accounts;
`)
		})

		totals := sq.Select("account_id").
			Column(sq.Alias(sq.Sum(sq.Expr("delta")), "total")).
			From("spec15_ledger").
			GroupBy("account_id")

		q := sq.Update("spec15_accounts a").
			Set("balance", sq.Expr("a.balance + sub.total")).
			FromSelect(totals, "sub").
			Where("a.id = sub.account_id").
			Suffix("RETURNING a.id, a.balance").
			PlaceholderFormat(sq.Dollar)

		type row struct {
			ID      int64   `db:"id"`
			Balance float64 `db:"balance"`
		}
		sql, args := mustSQL(t, q)
		rows := selectRows[row](t, pool, ctx, sql, args...)

		balances := map[int64]float64{}
		for _, r := range rows {
			balances[r.ID] = r.Balance
		}
		assert.InEpsilon(t, 125.0, balances[1], 0.001)
		assert.InEpsilon(t, 60.0, balances[2], 0.001)
	})

	t.Run("dml cte", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_jobs (id serial PRIMARY KEY, status text NOT NULL);
INSERT INTO spec15_jobs (status) VALUES ('pending'), ('pending'), ('done');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_jobs") })

		candidate := sq.Select("id").
			From("spec15_jobs").
			Where(sq.Eq{"status": "pending"}).
			OrderBy("id").
			Limit(1).
			Suffix("FOR UPDATE SKIP LOCKED")
		picked := sq.Select("id").From("candidate")
		updated := sq.Update("spec15_jobs j").
			Set("status", "running").
			From("picked p").
			Where("j.id = p.id").
			Suffix("RETURNING j.id, j.status")

		q := sq.With("candidate").As(candidate).
			Cte("picked").As(picked).
			Cte("updated").As(updated).
			Select(sq.Select("id", "status").From("updated").Where(sq.Eq{"status": "running"})).
			PlaceholderFormat(sq.Dollar)

		type jobRow struct {
			ID     int64  `db:"id"`
			Status string `db:"status"`
		}
		sql, args := mustSQL(t, q)
		rows := selectRows[jobRow](t, pool, ctx, sql, args...)
		require.Len(t, rows, 1)
		assert.Equal(t, int64(1), rows[0].ID)
		assert.Equal(t, "running", rows[0].Status)

		remaining := queryScalar[int](t, pool, ctx, "SELECT COUNT(*) FROM spec15_jobs WHERE status = 'pending'")
		assert.Equal(t, 1, remaining)
	})

	t.Run("for update skip locked", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_lock_rows (id serial PRIMARY KEY, token text NOT NULL);
INSERT INTO spec15_lock_rows (token) VALUES ('a'), ('b');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_lock_rows") })

		q := sq.Select("id", "token").
			From("spec15_lock_rows").
			OrderBy("id").
			Limit(1).
			Suffix("FOR UPDATE SKIP LOCKED").
			PlaceholderFormat(sq.Dollar)

		type row struct {
			ID    int64  `db:"id"`
			Token string `db:"token"`
		}
		sql, args := mustSQL(t, q)
		got := selectOneRow[row](t, pool, ctx, sql, args...)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "a", got.Token)
	})

	t.Run("insert returning", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_insert (id serial PRIMARY KEY, label text NOT NULL);
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_insert") })

		q := sq.Insert("spec15_insert").
			Columns("label").
			Values("new-row").
			Suffix("RETURNING id, label").
			PlaceholderFormat(sq.Dollar)

		type row struct {
			ID    int64  `db:"id"`
			Label string `db:"label"`
		}
		sql, args := mustSQL(t, q)
		got := selectOneRow[row](t, pool, ctx, sql, args...)
		assert.NotZero(t, got.ID)
		assert.Equal(t, "new-row", got.Label)
	})

	t.Run("update returning", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_update (id serial PRIMARY KEY, score int NOT NULL);
INSERT INTO spec15_update (score) VALUES (1);
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_update") })

		q := sq.Update("spec15_update").
			Set("score", 99).
			Where(sq.Eq{"id": 1}).
			Suffix("RETURNING id, score").
			PlaceholderFormat(sq.Dollar)

		type row struct {
			ID    int64 `db:"id"`
			Score int   `db:"score"`
		}
		sql, args := mustSQL(t, q)
		got := selectOneRow[row](t, pool, ctx, sql, args...)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, 99, got.Score)
	})

	t.Run("delete returning", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_delete (id serial PRIMARY KEY, label text NOT NULL);
INSERT INTO spec15_delete (label) VALUES ('gone');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_delete") })

		q := sq.Delete("spec15_delete").
			Where(sq.Eq{"label": "gone"}).
			Suffix("RETURNING id").
			PlaceholderFormat(sq.Dollar)

		sql, args := mustSQL(t, q)
		id := queryScalar[int64](t, pool, ctx, sql, args...)
		assert.Equal(t, int64(1), id)

		count := queryScalar[int](t, pool, ctx, "SELECT COUNT(*) FROM spec15_delete")
		assert.Equal(t, 0, count)
	})

	t.Run("in and not in", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_in (id serial PRIMARY KEY, bucket text NOT NULL);
INSERT INTO spec15_in (bucket) VALUES ('a'), ('b'), ('c');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_in") })

		sub := sq.Select("id").From("spec15_in").Where(sq.Eq{"bucket": "c"})

		inQ := sq.Select("bucket").
			From("spec15_in").
			Where(sq.In("id", []int64{1, 2})).
			OrderBy("bucket").
			PlaceholderFormat(sq.Dollar)
		inSQL, inArgs := mustSQL(t, inQ)
		inNames := queryStrings(t, pool, ctx, inSQL, inArgs...)
		assert.Equal(t, []string{"a", "b"}, inNames)

		notInQ := sq.Select("bucket").
			From("spec15_in").
			Where(sq.NotIn("id", sub)).
			OrderBy("bucket").
			PlaceholderFormat(sq.Dollar)
		notInSQL, notInArgs := mustSQL(t, notInQ)
		notInNames := queryStrings(t, pool, ctx, notInSQL, notInArgs...)
		assert.Equal(t, []string{"a", "b"}, notInNames)
	})

	t.Run("any and all typed slices", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_uuid_rows (id uuid PRIMARY KEY, label text NOT NULL);
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_uuid_rows") })

		idA := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		idB := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		idC := uuid.MustParse("33333333-3333-3333-3333-333333333333")

		insert := sq.Insert("spec15_uuid_rows").
			Columns("id", "label").
			Values(idA, "a").
			Values(idB, "b").
			Values(idC, "c").
			PlaceholderFormat(sq.Dollar)
		insertSQL, insertArgs := mustSQL(t, insert)
		_, err := pool.Exec(ctx, insertSQL, insertArgs...)
		require.NoError(t, err)

		anyQ := sq.Select("label").
			From("spec15_uuid_rows").
			Where(sq.In("id", []uuid.UUID{idA, idB})).
			OrderBy("label").
			PlaceholderFormat(sq.Dollar)
		anySQL, anyArgs := mustSQL(t, anyQ)
		anyLabels := queryStrings(t, pool, ctx, anySQL, anyArgs...)
		assert.Equal(t, []string{"a", "b"}, anyLabels)

		allQ := sq.Select("label").
			From("spec15_uuid_rows").
			Where(sq.NotIn("id", []uuid.UUID{idC})).
			OrderBy("label").
			PlaceholderFormat(sq.Dollar)
		allSQL, allArgs := mustSQL(t, allQ)
		allLabels := queryStrings(t, pool, ctx, allSQL, allArgs...)
		assert.Equal(t, []string{"a", "b"}, allLabels)
	})

	t.Run("json operators", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_json (id serial PRIMARY KEY, meta jsonb NOT NULL);
INSERT INTO spec15_json (meta) VALUES
	('{"format":"json","tags":["alpha","beta"],"keys":["k1","k2"]}'),
	('{"format":"xml","tags":["gamma"],"keys":["k1"]}');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_json") })

		hasFormat := sq.Select("id").
			From("spec15_json").
			Where("meta ?? 'format'").
			PlaceholderFormat(sq.Dollar)
		formatIDs := queryInt64s(t, pool, ctx, hasFormat)
		assert.Equal(t, []int64{1, 2}, formatIDs)

		anyTags := sq.Select("id").
			From("spec15_json").
			Where("meta->'tags' ??| array[?, ?]", "alpha", "gamma").
			OrderBy("id").
			PlaceholderFormat(sq.Dollar)
		anyIDs := queryInt64s(t, pool, ctx, anyTags)
		assert.Equal(t, []int64{1, 2}, anyIDs)

		allKeys := sq.Select("id").
			From("spec15_json").
			Where("meta->'keys' ??& array[?, ?]", "k1", "k2").
			PlaceholderFormat(sq.Dollar)
		allIDs := queryInt64s(t, pool, ctx, allKeys)
		assert.Equal(t, []int64{1}, allIDs)
	})

	t.Run("case", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_case (id serial PRIMARY KEY, status text NOT NULL);
INSERT INTO spec15_case (status) VALUES ('active'), ('inactive'), ('pending');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_case") })

		label := sq.Case("status").
			When(sq.Expr("?", "active"), sq.Expr("?", "enabled")).
			When(sq.Expr("?", "pending"), sq.Expr("?", "waiting")).
			Else(sq.Expr("?", "disabled"))

		q := sq.Select().
			Column(sq.Alias(label, "label")).
			From("spec15_case").
			OrderBy("id").
			PlaceholderFormat(sq.Dollar)

		sql, args := mustSQL(t, q)
		labels := queryStrings(t, pool, ctx, sql, args...)
		assert.Equal(t, []string{"enabled", "disabled", "waiting"}, labels)
	})

	t.Run("coalesce", func(t *testing.T) {
		execSetup(t, pool, ctx, `
CREATE TABLE spec15_coalesce (id serial PRIMARY KEY, nickname text, display text);
INSERT INTO spec15_coalesce (nickname, display) VALUES ('nick', NULL), (NULL, 'name');
`)
		t.Cleanup(func() { execSetup(t, pool, ctx, "DROP TABLE IF EXISTS spec15_coalesce") })

		q := sq.Select().
			Column(sq.Alias(sq.Coalesce("unknown", sq.Expr("nickname"), sq.Expr("display")), "label")).
			From("spec15_coalesce").
			OrderBy("id").
			PlaceholderFormat(sq.Dollar)

		sql, args := mustSQL(t, q)
		labels := queryStrings(t, pool, ctx, sql, args...)
		assert.Equal(t, []string{"nick", "name"}, labels)
	})
}

func mustSQL(t *testing.T, q sq.Sqlizer) (string, []any) {
	t.Helper()
	sql, args, err := q.ToSql()
	require.NoError(t, err)
	return sql, args
}

func queryStrings(t *testing.T, pool *pgxpool.Pool, ctx context.Context, sql string, args ...any) []string {
	t.Helper()

	rows, err := pool.Query(ctx, sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		require.NoError(t, rows.Scan(&value))
		values = append(values, value)
	}
	require.NoError(t, rows.Err())

	return values
}
