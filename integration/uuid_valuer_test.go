package integration

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid/v5"
	sq "github.com/muonsoft/squirrel"
)

func TestGofrsUUIDValuer(t *testing.T) {
	id := uuid.Must(uuid.FromString("01956596-4d67-7113-b5ff-49c868052667"))
	for _, value := range []any{id, &id} {
		query, args, err := (sq.Eq{"id": value}).ToSql()
		if err != nil || query != "id = ?" || !reflect.DeepEqual(args, []any{id.String()}) {
			t.Fatalf("%s %#v %v", query, args, err)
		}
	}
	query, args, err := (sq.Eq{"id": []uuid.UUID{id}}).ToSql()
	if err != nil || query != "id IN (?)" || !reflect.DeepEqual(args, []any{id}) {
		t.Fatalf("%s %#v %v", query, args, err)
	}
}
