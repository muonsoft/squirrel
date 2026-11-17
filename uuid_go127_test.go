//go:build go1.27

package squirrel

import (
	"database/sql/driver"
	"reflect"
	"testing"
	"uuid"
)

func TestStandardUUID(t *testing.T) {
	id := uuid.MustParse("01956596-4d67-7113-b5ff-49c868052667")
	zero := uuid.Nil()
	var nilID *uuid.UUID
	ids := []uuid.UUID{id, zero}
	tests := []struct {
		name string
		q    Sqlizer
		sql  string
		args []any
	}{
		{"eq", Eq{"id": id}, "id = ?", []any{id}},
		{"neq", NotEq{"id": id}, "id <> ?", []any{id}},
		{"pointer", Eq{"id": &id}, "id = ?", []any{id}},
		{"nil", Eq{"id": nilID}, "id IS NULL", nil},
		{"not nil", NotEq{"id": nilID}, "id IS NOT NULL", nil},
		{"zero", Eq{"id": zero}, "id = ?", []any{zero}},
		{"array of UUIDs", Eq{"id": [2]uuid.UUID{id, zero}}, "id IN (?,?)", []any{id, zero}},
		{"nil slice", Eq{"id": []uuid.UUID(nil)}, sqlFalse, []any{}},
		{"lt pointer", Lt{"id": &id}, "id < ?", []any{&id}},
		{"in pointer", In("id", &id), "id=?", []any{&id}},
		{"in nil", In("id", nilID), "id=?", []any{nilID}},
		{"not in one", NotIn("id", ids[:1]), "id<>?", []any{id}},
		{"list", Eq{"id": ids}, "id IN (?,?)", []any{id, zero}},
		{"not list", NotEq{"id": ids}, "id NOT IN (?,?)", []any{id, zero}},
		{"pointer list", Eq{"id": []*uuid.UUID{&id, nilID}}, "id IN (?,?)", []any{&id, nilID}},
		{"empty", Eq{"id": []uuid.UUID{}}, sqlFalse, []any{}},
		{"empty not", NotEq{"id": []uuid.UUID{}}, sqlTrue, []any{}},
		{"lt", Lt{"id": id}, "id < ?", []any{id}},
		{"lte", LtOrEq{"id": id}, "id <= ?", []any{id}},
		{"gt", Gt{"id": id}, "id > ?", []any{id}},
		{"gte", GtOrEq{"id": id}, "id >= ?", []any{id}},
		{"like", Like{"id": id}, "id LIKE ?", []any{id}},
		{"expr", Expr("?", id), "?", []any{id}},
		{"expr pointer", Expr("?", &id), "?", []any{&id}},
		{"expr nil", Expr("?", nilID), "?", []any{nilID}},
		{"expr list", Expr("?", ids), "?", []any{ids}},
		{"nested", Select("id").Where(Expr("? AND id = ?", Eq{"id": id}, zero)).PlaceholderFormat(Dollar), "SELECT id WHERE id = $1 AND id = $2", []any{id, zero}},
		{"in", In("id", id), "id=?", []any{id}},
		{"not in", NotIn("id", id), "id<>?", []any{id}},
		{"in one", In("id", ids[:1]), "id=?", []any{id}},
		{"in many", In("id", ids), "id=ANY(?)", []any{ids}},
		{"not in many", NotIn("id", ids), "id<>ALL(?)", []any{ids}},
		{"in empty", In("id", []uuid.UUID{}), "", nil},
		{"equal", Equal(Expr("id"), id), "(id) = ?", []any{id}},
		{"range", Range("id", id, id), "id BETWEEN ? AND ?", []any{id, id}},
		{"range one", Range("id", id, nil), "id >= ?", []any{id}},
		{"insert", Insert("t").Columns("id").Values(id), "INSERT INTO t (id) VALUES (?)", []any{id}},
		{"update", Update("t").Set("id", id), "UPDATE t SET id = ?", []any{id}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := tt.q.ToSql()
			if err != nil || sql != tt.sql || !reflect.DeepEqual(args, tt.args) {
				t.Fatalf("got %q %#v %v; want %q %#v", sql, args, err, tt.sql, tt.args)
			}
		})
	}
	for _, v := range []any{id, &id, zero, nilID} {
		got, err := driver.DefaultParameterConverter.ConvertValue(v)
		var want any = id.String()
		if v == zero {
			want = zero.String()
		}
		if v == nilID {
			want = nil
		}
		if err != nil || got != want {
			t.Fatalf("ConvertValue(%T) = %v, %v; want %v", v, got, err, want)
		}
	}
}
