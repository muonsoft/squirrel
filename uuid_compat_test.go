package squirrel

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

type uuidShapedArray [16]byte

func (uuidShapedArray) String() string               { return "not a UUID" }
func (uuidShapedArray) MarshalText() ([]byte, error) { return []byte("not a UUID"), nil }

type uuidValuer [16]byte

func (uuidValuer) Value() (driver.Value, error) { return "valuer UUID", nil }

func TestUUIDClassificationCompatibility(t *testing.T) {
	for _, value := range []any{[16]byte{}, uuidShapedArray{}, []int{1, 2}, [2]int{1, 2}} {
		sql, args, err := (Eq{"id": value}).ToSql()
		rv := reflect.ValueOf(value)
		want := make([]any, rv.Len())
		for i := range rv.Len() {
			want[i] = rv.Index(i).Interface()
		}
		if err != nil || sql != "id IN ("+Placeholders(rv.Len())+")" || !reflect.DeepEqual(args, want) {
			t.Fatalf("%T: %s %#v %v", value, sql, args, err)
		}
	}
	for _, value := range []any{[]byte{1, 2}, uuidValuer{}} {
		sql, args, err := (Eq{"id": value}).ToSql()
		want := value
		if _, ok := value.(driver.Valuer); ok {
			want = "valuer UUID"
		}
		if err != nil || sql != "id = ?" || !reflect.DeepEqual(args, []any{want}) {
			t.Fatalf("%T: %s %#v %v", value, sql, args, err)
		}
	}
}
