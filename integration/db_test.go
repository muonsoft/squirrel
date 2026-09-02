package integration

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sq "github.com/muonsoft/squirrel"
	"github.com/stretchr/testify/require"
)

func postgresTestDSN() string {
	return os.Getenv("POSTGRES_TEST_DSN")
}

func newTestPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	dsn := postgresTestDSN()
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostgreSQL integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	return pool, ctx
}

func execSetup(t *testing.T, pool *pgxpool.Pool, ctx context.Context, setupSQL string) {
	t.Helper()

	_, err := pool.Exec(ctx, setupSQL)
	require.NoError(t, err)
}

func queryInt64s(t *testing.T, pool *pgxpool.Pool, ctx context.Context, q sq.Sqlizer) []int64 {
	t.Helper()

	sql, args, err := q.ToSql()
	require.NoError(t, err)

	rows, err := pool.Query(ctx, sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())

	return ids
}

func queryInt64StringPairs(t *testing.T, pool *pgxpool.Pool, ctx context.Context, q sq.Sqlizer) ([]int64, []string) {
	t.Helper()

	sql, args, err := q.ToSql()
	require.NoError(t, err)

	rows, err := pool.Query(ctx, sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	var ids []int64
	var names []string
	for rows.Next() {
		var id int64
		var name string
		require.NoError(t, rows.Scan(&id, &name))
		ids = append(ids, id)
		names = append(names, name)
	}
	require.NoError(t, rows.Err())

	return ids, names
}

func selectRows[T any](t *testing.T, pool *pgxpool.Pool, ctx context.Context, sql string, args ...any) []T {
	t.Helper()

	rows, err := pool.Query(ctx, sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	require.NoError(t, err)

	return results
}

func selectOneRow[T any](t *testing.T, pool *pgxpool.Pool, ctx context.Context, sql string, args ...any) T {
	t.Helper()

	rows, err := pool.Query(ctx, sql, args...)
	require.NoError(t, err)
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	require.NoError(t, err)

	return result
}

func queryScalar[T any](t *testing.T, pool *pgxpool.Pool, ctx context.Context, sql string, args ...any) T {
	t.Helper()

	var value T
	err := pool.QueryRow(ctx, sql, args...).Scan(&value)
	require.NoError(t, err)

	return value
}
