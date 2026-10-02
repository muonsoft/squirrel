// Command evaluation emits deterministic observations for differential comparison.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	sq "github.com/muonsoft/squirrel"
	"os"
)

type result struct {
	Name  string
	SQL   string
	Args  []any
	Error string
	Panic string
}

func observe(name string, q sq.Sqlizer) (r result) {
	r.Name = name
	defer func() {
		if p := recover(); p != nil {
			r.Panic = fmt.Sprint(p)
		}
	}()
	var err error
	r.SQL, r.Args, err = q.ToSql()
	if err != nil {
		r.Error = err.Error()
	}
	return
}

type failingSqlizer struct{}

func (failingSqlizer) ToSql() (string, []any, error) { return "", nil, errors.New("nested failure") }

func main() {
	results := []result{}
	add := func(name string, q sq.Sqlizer) { results = append(results, observe(name, q)) }
	add("zero-select", sq.SelectBuilder{})
	add("zero-insert", sq.InsertBuilder{})
	add("zero-update", sq.UpdateBuilder{})
	add("zero-delete", sq.DeleteBuilder{})
	add("zero-case", sq.CaseBuilder{})
	add("zero-cte", sq.CommonTableExpressionsBuilder{})
	add("zero-seed", sq.StatementBuilderType{}.Select("1"))
	for _, format := range []sq.PlaceholderFormat{sq.Question, sq.Dollar, sq.Colon, sq.AtP} {
		seed := sq.StatementBuilder.PlaceholderFormat(format).Where("tenant = ?", 7)
		selectBase := seed.Select("id").From("users").Where("active = ?", true)
		insertBase := sq.StatementBuilder.PlaceholderFormat(format).Insert("users").Columns("id", "name")
		updateBase := seed.Update("users").Set("active", true)
		deleteBase := seed.Delete("users")
		caseBase := sq.Case().When("active", "1")
		cteBase := sq.StatementBuilder.PlaceholderFormat(format).With("u").As(selectBase)
		for i := range 128 {
			name := fmt.Sprint(i)
			q := selectBase.Where("id > ?", i).Column(sq.Expr("?", i)).GroupBy("id").Having("count(*) > ?", i).OrderBy("id").Limit(uint64(i)).Offset(uint64(i))
			switch i % 4 {
			case 0:
				q = q.RemoveColumns().Column("id").RemoveLimit().RemoveOffset()
			case 1:
				q = q.Where(sq.Eq{"id": seed.Select("id").From("roles").Where("role = ?", i)})
			case 2:
				q = q.PrefixExpr(sq.Expr("/* ? */", i)).Suffix("/* ? */", i)
			case 3:
				q = q.FromSelect(selectBase, "u").JoinClause(sq.Expr("JOIN roles ON roles.id = ?", i))
			}
			add("select"+name, q)
			add("select-base", selectBase)
			add("insert"+name, insertBase.Values(i, "Ada").Values(i+1, "Grace").Suffix("RETURNING id"))
			add("insert-map"+name, insertBase.SetMap(map[string]any{"z": i, "a": "Ada"}))
			add("insert-base", insertBase)
			add("update"+name, updateBase.SetMap(map[string]any{"z": i, "a": "Ada"}).FromSelect(selectBase, "u").Where("id = ?", i))
			add("update-base", updateBase)
			add("delete"+name, deleteBase.Where("id = ?", i).OrderBy("id").Limit(uint64(i)).Offset(uint64(i)))
			add("delete-base", deleteBase)
			add("case"+name, caseBase.When(sq.Expr("id = ?", i), sq.Expr("?", i)).Else("0"))
			add("case-base", caseBase)
			add("cte"+name, cteBase.Cte("v").As(q).Select(selectBase))
			add("cte-update", cteBase.Update(updateBase))
			add("cte-insert", cteBase.Insert(insertBase.Values(i, "Ada")))
			add("cte-delete", cteBase.Delete(deleteBase))
		}
	}
	add("nested-error", sq.Select("id").Where(failingSqlizer{}))
	add("case-literal-error", sq.Case().When("true", 1))
	add("escaped-operator", sq.Select("id").From("docs").Where("data ?? ?", "key").PlaceholderFormat(sq.Dollar))
	add("recursive", sq.WithRecursive("u").As(sq.Select("id").From("users")).Select(sq.Select("id").From("u")).PlaceholderFormat(sq.Dollar))
	values := []any{1, "before"}
	insert := sq.Insert("users").Columns("id", "name").Values(values...)
	values[1] = "after"
	add("caller-values", insert)
	columns := []string{"id"}
	selectQ := sq.Select(columns...).From("users")
	columns[0] = "changed"
	add("caller-columns", selectQ)
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		panic(err)
	}
}
