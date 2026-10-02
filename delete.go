package squirrel

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

type deleteData struct {
	PlaceholderFormat PlaceholderFormat
	Prefixes          []Sqlizer
	From              string
	WhereParts        []Sqlizer
	OrderBys          []string
	Limit             string
	Offset            string
	Suffixes          []Sqlizer
}

func (d *deleteData) toSqlRaw() (sqlStr string, args []any, err error) {
	if d.From == "" {
		err = errors.New("delete statements must specify a From table")
		return "", nil, err
	}

	sql := &bytes.Buffer{}

	if len(d.Prefixes) > 0 {
		args, err = appendToSql(d.Prefixes, sql, " ", args)
		if err != nil {
			return "", nil, err
		}

		_, _ = sql.WriteString(" ")
	}

	_, _ = sql.WriteString("DELETE FROM ")
	_, _ = sql.WriteString(d.From)

	if len(d.WhereParts) > 0 {
		_, _ = sql.WriteString(" WHERE ")
		args, err = appendToSql(d.WhereParts, sql, " AND ", args)
		if err != nil {
			return "", nil, err
		}
	}

	if len(d.OrderBys) > 0 {
		_, _ = sql.WriteString(" ORDER BY ")
		_, _ = sql.WriteString(strings.Join(d.OrderBys, ", "))
	}

	if d.Limit != "" {
		_, _ = sql.WriteString(" LIMIT ")
		_, _ = sql.WriteString(d.Limit)
	}

	if d.Offset != "" {
		_, _ = sql.WriteString(" OFFSET ")
		_, _ = sql.WriteString(d.Offset)
	}

	if len(d.Suffixes) > 0 {
		_, _ = sql.WriteString(" ")
		args, err = appendToSql(d.Suffixes, sql, " ", args)
		if err != nil {
			return "", nil, err
		}
	}

	return sql.String(), args, nil
}

func (d *deleteData) ToSql() (sqlStr string, args []any, err error) {
	s, a, e := d.toSqlRaw()
	if e != nil {
		return "", nil, e
	}
	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(s)
	return sqlStr, a, err
}

// Builder

// DeleteBuilder builds SQL DELETE statements.
type DeleteBuilder struct {
	state *deleteDataState
}

type deleteDataState struct {
	PlaceholderFormat PlaceholderFormat
	Prefixes          immutableList[Sqlizer]
	From              string
	WhereParts        immutableList[Sqlizer]
	OrderBys          immutableList[string]
	Limit             string
	Offset            string
	Suffixes          immutableList[Sqlizer]
}

func (b DeleteBuilder) clone() deleteDataState {
	if b.state == nil {
		return deleteDataState{}
	}
	return *b.state
}

func (b DeleteBuilder) data() deleteData {
	state := b.clone()
	return deleteData{
		PlaceholderFormat: state.PlaceholderFormat,
		Prefixes:          state.Prefixes.slice(),
		From:              state.From,
		WhereParts:        state.WhereParts.slice(),
		OrderBys:          state.OrderBys.slice(),
		Limit:             state.Limit,
		Offset:            state.Offset,
		Suffixes:          state.Suffixes.slice(),
	}
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b DeleteBuilder) PlaceholderFormat(f PlaceholderFormat) DeleteBuilder {
	next := b.clone()
	next.PlaceholderFormat = f
	return DeleteBuilder{state: &next}
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b DeleteBuilder) ToSql() (sql string, args []any, err error) {
	data := b.data()
	return data.ToSql()
}

// MustSql builds the query into a SQL string and bound args.
// It panics if there are any errors.
func (b DeleteBuilder) MustSql() (sql string, args []any) {
	sql, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	return sql, args
}

// Prefix adds an expression to the beginning of the query.
func (b DeleteBuilder) Prefix(sql string, args ...any) DeleteBuilder {
	return b.PrefixExpr(Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query.
func (b DeleteBuilder) PrefixExpr(e Sqlizer) DeleteBuilder {
	next := b.clone()
	next.Prefixes = appendPersistent(next.Prefixes, e)
	return DeleteBuilder{state: &next}
}

// From sets the table to be deleted from.
func (b DeleteBuilder) From(from string) DeleteBuilder {
	next := b.clone()
	next.From = from
	return DeleteBuilder{state: &next}
}

// Where adds WHERE expressions to the query.
//
// See SelectBuilder.Where for more information.
func (b DeleteBuilder) Where(pred any, args ...any) DeleteBuilder {
	next := b.clone()
	next.WhereParts = appendPersistent(next.WhereParts, newWherePart(pred, args...))
	return DeleteBuilder{state: &next}
}

// OrderBy adds ORDER BY expressions to the query.
func (b DeleteBuilder) OrderBy(orderBys ...string) DeleteBuilder {
	next := b.clone()
	next.OrderBys = appendPersistent(next.OrderBys, orderBys...)
	return DeleteBuilder{state: &next}
}

// Limit sets a LIMIT clause on the query.
func (b DeleteBuilder) Limit(limit uint64) DeleteBuilder {
	next := b.clone()
	next.Limit = strconv.FormatUint(limit, 10)
	return DeleteBuilder{state: &next}
}

// Offset sets a OFFSET clause on the query.
func (b DeleteBuilder) Offset(offset uint64) DeleteBuilder {
	next := b.clone()
	next.Offset = strconv.FormatUint(offset, 10)
	return DeleteBuilder{state: &next}
}

// toSqlRaw builds SQL with raw placeholders ("?") without applying PlaceholderFormat.
func (b DeleteBuilder) toSqlRaw() (sql string, args []any, err error) {
	data := b.data()
	return data.toSqlRaw()
}

// Suffix adds an expression to the end of the query.
func (b DeleteBuilder) Suffix(sql string, args ...any) DeleteBuilder {
	return b.SuffixExpr(Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query.
func (b DeleteBuilder) SuffixExpr(e Sqlizer) DeleteBuilder {
	next := b.clone()
	next.Suffixes = appendPersistent(next.Suffixes, e)
	return DeleteBuilder{state: &next}
}
