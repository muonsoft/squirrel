package squirrel

import (
	"reflect"
	"sync"
	"testing"
)

func TestTypedBuilderConcurrentBranches(t *testing.T) {
	type branchCase struct {
		name   string
		base   Sqlizer
		branch func(int) Sqlizer
	}
	selectBase := Select("id").From("users").Where("tenant = ?", 1)
	insertBase := Insert("users").Columns("id").Values(1)
	updateBase := Update("users").Set("active", true).Where("tenant = ?", 1)
	deleteBase := Delete("users").Where("tenant = ?", 1)
	caseBase := Case().When("active", "1").Else("0")
	cteBase := With("u").As(selectBase).Select(selectBase)
	seed := StatementBuilder.PlaceholderFormat(Dollar).Where("tenant = ?", 1)
	cases := []branchCase{
		{"select", selectBase, func(i int) Sqlizer { return selectBase.Where("id = ?", i).Column("name").Suffix("FOR UPDATE") }},
		{"insert", insertBase, func(i int) Sqlizer { return insertBase.Values(i).Suffix("RETURNING id") }},
		{"update", updateBase, func(i int) Sqlizer { return updateBase.Set("id", i).Where("id <> ?", i) }},
		{"delete", deleteBase, func(i int) Sqlizer { return deleteBase.Where("id = ?", i).OrderBy("id") }},
		{"case", caseBase, func(i int) Sqlizer { return caseBase.When(Expr("id = ?", i), "2") }},
		{"cte", cteBase, func(i int) Sqlizer { return cteBase.Cte("v").As(Select("?").Column(Expr("?", i))).Select(selectBase) }},
		{"seed", seed.Select("id"), func(i int) Sqlizer { return seed.Where("id = ?", i).Select("id") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wantSQL, wantArgs, err := tc.base.ToSql()
			if err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for i := range 32 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					q := tc.branch(i)
					first, args, err := q.ToSql()
					if err != nil {
						t.Error(err)
						return
					}
					for range 8 {
						got, gotArgs, err := q.ToSql()
						if err != nil || got != first || !reflect.DeepEqual(args, gotArgs) {
							t.Errorf("branch changed: %q %v %v", got, gotArgs, err)
						}
						got, gotArgs, err = tc.base.ToSql()
						if err != nil || got != wantSQL || !reflect.DeepEqual(wantArgs, gotArgs) {
							t.Errorf("base changed: %q %v %v", got, gotArgs, err)
						}
					}
				}()
			}
			wg.Wait()
		})
	}
}

func TestStatementWhereDefaultsOnlyApplyToWhereStatements(t *testing.T) {
	seed := StatementBuilder.PlaceholderFormat(Dollar).Where("tenant = ?", 7)
	insert := seed.Insert("users").Columns("id").Values(1)
	got, args, err := insert.ToSql()
	if err != nil || got != "INSERT INTO users (id) VALUES ($1)" || !reflect.DeepEqual(args, []any{1}) {
		t.Fatalf("insert: %q %v %v", got, args, err)
	}
	got, args, err = seed.With("u").As(Select("id").From("users")).Select(Select("id").From("u")).ToSql()
	if err != nil || got != "WITH u AS (SELECT id FROM users) SELECT id FROM u" || len(args) != 0 {
		t.Fatalf("cte: %q %v %v", got, args, err)
	}
}
