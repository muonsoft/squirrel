package squirrel

import (
	"strconv"
	"testing"
)

var benchmarkBuilder Sqlizer
var benchmarkSQL string
var benchmarkArgs []any

func BenchmarkBuilders(b *testing.B) {
	factories := map[string]func() Sqlizer{
		"Select": func() Sqlizer {
			return Select("id", "name").From("users").Where("active = ?", true).OrderBy("id").Limit(20)
		},
		"Insert": func() Sqlizer { return Insert("users").Columns("id", "name").Values(1, "Ada").Values(2, "Grace") },
		"Update": func() Sqlizer { return Update("users").Set("name", "Ada").Where("id = ?", 1) },
		"Delete": func() Sqlizer { return Delete("users").Where("id = ?", 1) },
		"Case":   func() Sqlizer { return Case().When("active", "1").Else("0") },
		"CTE": func() Sqlizer {
			return With("u").As(Select("id").From("users").Where("active = ?", true)).Select(Select("id").From("u"))
		},
		"Statement": func() Sqlizer {
			return StatementBuilder.PlaceholderFormat(Dollar).Where("tenant = ?", 1).Select("id").From("users")
		},
		"Nested": func() Sqlizer {
			return Select("id").From("users").Where(Eq{"id": Select("user_id").From("roles").Where("role = ?", "admin")}).PlaceholderFormat(Dollar)
		},
	}
	for name, factory := range factories {
		b.Run(name, func(b *testing.B) {
			b.Run("Build", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					benchmarkBuilder = factory()
				}
			})
			query := factory()
			b.Run("Branch", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					switch base := query.(type) {
					case SelectBuilder:
						benchmarkBuilder = base.Where("id = ?", 1)
					case InsertBuilder:
						benchmarkBuilder = base.Values(3, "Lin")
					case UpdateBuilder:
						benchmarkBuilder = base.Set("id", 1)
					case DeleteBuilder:
						benchmarkBuilder = base.Where("id = ?", 1)
					case CaseBuilder:
						benchmarkBuilder = base.When("false", "2")
					case CommonTableExpressionsBuilder:
						benchmarkBuilder = base.Cte("v").As(Select("1"))
					}
				}
			})
			b.Run("Render", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var err error
					benchmarkSQL, benchmarkArgs, err = query.ToSql()
					if err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("BuildRender", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					var err error
					benchmarkSQL, benchmarkArgs, err = factory().ToSql()
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkSelectChain(b *testing.B) {
	for _, length := range []int{4, 32, 256} {
		b.Run(strconv.Itoa(length), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				q := Select("id").From("users")
				for i := range length {
					q = q.Where("id <> ?", i)
				}
				benchmarkBuilder = q
			}
		})
	}
}

func BenchmarkSelectBranch(b *testing.B) {
	base := Select("id").From("users").Where("tenant = ?", 1)
	b.ReportAllocs()
	for b.Loop() {
		benchmarkBuilder = base.Where("id = ?", 2)
		benchmarkBuilder = base.Where("id = ?", 3)
	}
}
