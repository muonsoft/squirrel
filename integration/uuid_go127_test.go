//go:build go1.27

package integration

import (
	"bytes"
	"database/sql"
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib"
	sq "github.com/muonsoft/squirrel"
)

// Exercise pgx's actual codecs even when no PostgreSQL server is configured.
func TestStandardUUIDCodec(t *testing.T) {
	id := uuid.MustParse("01956596-4d67-7113-b5ff-49c868052667")
	m := pgtype.NewMap()
	for _, value := range []any{id, &id} {
		got, err := m.Encode(pgtype.UUIDOID, pgtype.BinaryFormatCode, value, nil)
		if err != nil || !bytes.Equal(got, id[:]) {
			t.Fatalf("UUID codec (%T): %x, %v", value, got, err)
		}
	}
	for _, value := range []any{[]uuid.UUID{id, uuid.Nil()}, []*uuid.UUID{&id, nil}} {
		if _, err := m.Encode(pgtype.UUIDArrayOID, pgtype.BinaryFormatCode, value, nil); err != nil {
			t.Fatalf("UUID array codec (%T): %v", value, err)
		}
	}
}

func TestStandardUUIDPostgres(t *testing.T) {
	pool, ctx := newTestPool(t)
	defer pool.Close()
	db, err := sql.Open("pgx", postgresTestDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	id := uuid.MustParse("01956596-4d67-7113-b5ff-49c868052667")
	for _, value := range []any{id, &id, uuid.Nil()} {
		q := sq.Select("id::text").FromSelect(sq.Select().Column("?::uuid AS id", value), "t").Where(sq.Eq{"id": value}).PlaceholderFormat(sq.Dollar)
		query, args, err := q.ToSql()
		if err != nil {
			t.Fatal(err)
		}
		want := id.String()
		if value == uuid.Nil() {
			want = uuid.Nil().String()
		}
		var got string
		if err := pool.QueryRow(ctx, query, args...).Scan(&got); err != nil || got != want {
			t.Fatalf("native pgx: %q %v", got, err)
		}
		if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil || got != want {
			t.Fatalf("database/sql: %q %v", got, err)
		}
	}
}
