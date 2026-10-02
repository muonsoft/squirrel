package squirrel

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type selectData struct {
	PlaceholderFormat PlaceholderFormat
	Prefixes          []Sqlizer
	Options           []string
	Columns           []Sqlizer
	From              Sqlizer
	Joins             []Sqlizer
	WhereParts        []Sqlizer
	GroupBys          []string
	HavingParts       []Sqlizer
	OrderByParts      []Sqlizer
	Limit             string
	Offset            string
	Suffixes          []Sqlizer
}

func (d *selectData) ToSql() (sqlStr string, args []any, err error) {
	sqlStr, args, err = d.toSqlRaw()
	if err != nil {
		return
	}

	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(sqlStr)
	return
}

func (d *selectData) writePrefixes(sql *bytes.Buffer, args []any) ([]any, error) {
	if len(d.Prefixes) == 0 {
		return args, nil
	}

	args, err := appendToSql(d.Prefixes, sql, " ", args)
	if err != nil {
		return nil, err
	}

	_, _ = sql.WriteString(" ")
	return args, nil
}

func (d *selectData) writeSelectClause(sql *bytes.Buffer, args []any) ([]any, error) {
	_, _ = sql.WriteString("SELECT ")

	if len(d.Options) > 0 {
		_, _ = sql.WriteString(strings.Join(d.Options, " "))
		_, _ = sql.WriteString(" ")
	}

	return appendToSql(d.Columns, sql, ", ", args)
}

func (d *selectData) writeFromClause(sql *bytes.Buffer, args []any) ([]any, error) {
	if d.From == nil {
		return args, nil
	}

	_, _ = sql.WriteString(" FROM ")
	return appendToSql([]Sqlizer{d.From}, sql, "", args)
}

func (d *selectData) writeJoins(sql *bytes.Buffer, args []any) ([]any, error) {
	if len(d.Joins) == 0 {
		return args, nil
	}

	_, _ = sql.WriteString(" ")
	return appendToSql(d.Joins, sql, " ", args)
}

func (d *selectData) writeWhereClause(sql *bytes.Buffer, args []any) ([]any, error) {
	if len(d.WhereParts) == 0 {
		return args, nil
	}

	_, _ = sql.WriteString(" WHERE ")
	return appendToSql(d.WhereParts, sql, " AND ", args)
}

func (d *selectData) writeGroupByClause(sql *bytes.Buffer) {
	if len(d.GroupBys) > 0 {
		_, _ = sql.WriteString(" GROUP BY ")
		_, _ = sql.WriteString(strings.Join(d.GroupBys, ", "))
	}
}

func (d *selectData) writeHavingClause(sql *bytes.Buffer, args []any) ([]any, error) {
	if len(d.HavingParts) == 0 {
		return args, nil
	}

	_, _ = sql.WriteString(" HAVING ")
	return appendToSql(d.HavingParts, sql, " AND ", args)
}

func (d *selectData) writeOrderByClause(sql *bytes.Buffer, args []any) ([]any, error) {
	if len(d.OrderByParts) == 0 {
		return args, nil
	}

	_, _ = sql.WriteString(" ORDER BY ")
	return appendToSql(d.OrderByParts, sql, ", ", args)
}

func (d *selectData) writeLimitOffset(sql *bytes.Buffer) {
	if d.Limit != "" {
		_, _ = sql.WriteString(" LIMIT ")
		_, _ = sql.WriteString(d.Limit)
	}

	if d.Offset != "" {
		_, _ = sql.WriteString(" OFFSET ")
		_, _ = sql.WriteString(d.Offset)
	}
}

func (d *selectData) writeSuffixes(sql *bytes.Buffer, args []any) ([]any, error) {
	if len(d.Suffixes) == 0 {
		return args, nil
	}

	_, _ = sql.WriteString(" ")
	return appendToSql(d.Suffixes, sql, " ", args)
}

func (d *selectData) toSqlRaw() (sqlStr string, args []any, err error) {
	if len(d.Columns) == 0 {
		return "", nil, errors.New("select statements must have at least one result column")
	}

	sql := &bytes.Buffer{}

	if args, err = d.writePrefixes(sql, args); err != nil {
		return "", nil, err
	}

	if args, err = d.writeSelectClause(sql, args); err != nil {
		return "", nil, err
	}

	if args, err = d.writeFromClause(sql, args); err != nil {
		return "", nil, err
	}

	if args, err = d.writeJoins(sql, args); err != nil {
		return "", nil, err
	}

	if args, err = d.writeWhereClause(sql, args); err != nil {
		return "", nil, err
	}

	d.writeGroupByClause(sql)

	if args, err = d.writeHavingClause(sql, args); err != nil {
		return "", nil, err
	}

	if args, err = d.writeOrderByClause(sql, args); err != nil {
		return "", nil, err
	}

	d.writeLimitOffset(sql)

	if args, err = d.writeSuffixes(sql, args); err != nil {
		return "", nil, err
	}

	return sql.String(), args, nil
}

// Builder

// SelectBuilder builds SQL SELECT statements.
type SelectBuilder struct {
	state *selectDataState
}

type selectDataState struct {
	PlaceholderFormat PlaceholderFormat
	Prefixes          immutableList[Sqlizer]
	Options           immutableList[string]
	Columns           immutableList[Sqlizer]
	From              Sqlizer
	Joins             immutableList[Sqlizer]
	WhereParts        immutableList[Sqlizer]
	GroupBys          immutableList[string]
	HavingParts       immutableList[Sqlizer]
	OrderByParts      immutableList[Sqlizer]
	Limit             string
	Offset            string
	Suffixes          immutableList[Sqlizer]
}

func (b SelectBuilder) clone() selectDataState {
	if b.state == nil {
		return selectDataState{}
	}
	return *b.state
}

func (b SelectBuilder) data() selectData {
	state := b.clone()
	return selectData{
		PlaceholderFormat: state.PlaceholderFormat,
		Prefixes:          state.Prefixes.slice(),
		Options:           state.Options.slice(),
		Columns:           state.Columns.slice(),
		From:              state.From,
		Joins:             state.Joins.slice(),
		WhereParts:        state.WhereParts.slice(),
		GroupBys:          state.GroupBys.slice(),
		HavingParts:       state.HavingParts.slice(),
		OrderByParts:      state.OrderByParts.slice(),
		Limit:             state.Limit,
		Offset:            state.Offset,
		Suffixes:          state.Suffixes.slice(),
	}
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b SelectBuilder) PlaceholderFormat(f PlaceholderFormat) SelectBuilder {
	next := b.clone()
	next.PlaceholderFormat = f
	return SelectBuilder{state: &next}
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b SelectBuilder) ToSql() (sql string, args []any, err error) {
	data := b.data()
	return data.ToSql()
}

func (b SelectBuilder) toSqlRaw() (sql string, args []any, err error) {
	data := b.data()
	return data.toSqlRaw()
}

// MustSql builds the query into a SQL string and bound args.
// It panics if there are any errors.
func (b SelectBuilder) MustSql() (sql string, args []any) {
	sql, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	return sql, args
}

// Prefix adds an expression to the beginning of the query.
func (b SelectBuilder) Prefix(sql string, args ...any) SelectBuilder {
	return b.PrefixExpr(Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query.
func (b SelectBuilder) PrefixExpr(e Sqlizer) SelectBuilder {
	next := b.clone()
	next.Prefixes = appendPersistent(next.Prefixes, e)
	return SelectBuilder{state: &next}
}

// Distinct adds a DISTINCT clause to the query.
func (b SelectBuilder) Distinct() SelectBuilder {
	return b.Options("DISTINCT")
}

// Options adds select option to the query.
func (b SelectBuilder) Options(options ...string) SelectBuilder {
	next := b.clone()
	next.Options = appendPersistent(next.Options, options...)
	return SelectBuilder{state: &next}
}

// Columns adds result columns to the query.
func (b SelectBuilder) Columns(columns ...string) SelectBuilder {
	parts := make([]Sqlizer, 0, len(columns))
	for _, str := range columns {
		parts = append(parts, newPart(str))
	}
	next := b.clone()
	next.Columns = appendPersistent(next.Columns, parts...)
	return SelectBuilder{state: &next}
}

// RemoveColumns remove all columns from query.
// Must add a new column with Column or Columns methods, otherwise
// return a error.
func (b SelectBuilder) RemoveColumns() SelectBuilder {
	next := b.clone()
	next.Columns = immutableList[Sqlizer]{}
	return SelectBuilder{state: &next}
}

// Column adds a result column to the query.
// Unlike Columns, Column accepts args which will be bound to placeholders in
// the columns string, for example:
//
//	Column("IF(col IN ("+squirrel.Placeholders(3)+"), 1, 0) as col", 1, 2, 3)
func (b SelectBuilder) Column(column any, args ...any) SelectBuilder {
	next := b.clone()
	next.Columns = appendPersistent(next.Columns, newPart(column, args...))
	return SelectBuilder{state: &next}
}

// From sets the FROM clause of the query.
func (b SelectBuilder) From(from string) SelectBuilder {
	next := b.clone()
	next.From = newPart(from)
	return SelectBuilder{state: &next}
}

// FromSelect sets a subquery into the FROM clause of the query.
func (b SelectBuilder) FromSelect(from SelectBuilder, alias string) SelectBuilder {
	next := b.clone()
	next.From = Alias(from, alias)
	return SelectBuilder{state: &next}
}

// JoinClause adds a join clause to the query.
func (b SelectBuilder) JoinClause(pred any, args ...any) SelectBuilder {
	next := b.clone()
	next.Joins = appendPersistent(next.Joins, newPart(pred, args...))
	return SelectBuilder{state: &next}
}

// Join adds a JOIN clause to the query.
func (b SelectBuilder) Join(join string, rest ...any) SelectBuilder {
	return b.JoinClause("JOIN "+join, rest...)
}

// LeftJoin adds a LEFT JOIN clause to the query.
func (b SelectBuilder) LeftJoin(join string, rest ...any) SelectBuilder {
	return b.JoinClause("LEFT JOIN "+join, rest...)
}

// RightJoin adds a RIGHT JOIN clause to the query.
func (b SelectBuilder) RightJoin(join string, rest ...any) SelectBuilder {
	return b.JoinClause("RIGHT JOIN "+join, rest...)
}

// InnerJoin adds a INNER JOIN clause to the query.
func (b SelectBuilder) InnerJoin(join string, rest ...any) SelectBuilder {
	return b.JoinClause("INNER JOIN "+join, rest...)
}

// CrossJoin adds a CROSS JOIN clause to the query.
func (b SelectBuilder) CrossJoin(join string, rest ...any) SelectBuilder {
	return b.JoinClause("CROSS JOIN "+join, rest...)
}

// Where adds an expression to the WHERE clause of the query.
//
// Expressions are ANDed together in the generated SQL.
//
// Where accepts several types for its pred argument:
//
// nil OR "" - ignored.
//
// string - SQL expression.
// If the expression has SQL placeholders then a set of arguments must be passed
// as well, one for each placeholder.
//
// map[string]any OR Eq - map of SQL expressions to values. Each key is
// transformed into an expression like "<key> = ?", with the corresponding value
// bound to the placeholder. If the value is nil, the expression will be "<key>
// IS NULL". If the value is an array or slice, the expression will be "<key> IN
// (?,?,...)", with one placeholder for each item in the value. These expressions
// are ANDed together.
//
// Where will panic if pred isn't any of the above types.
func (b SelectBuilder) Where(pred any, args ...any) SelectBuilder {
	if pred == nil || pred == "" {
		return b
	}
	next := b.clone()
	next.WhereParts = appendPersistent(next.WhereParts, newWherePart(pred, args...))
	return SelectBuilder{state: &next}
}

// GroupBy adds GROUP BY expressions to the query.
func (b SelectBuilder) GroupBy(groupBys ...string) SelectBuilder {
	next := b.clone()
	next.GroupBys = appendPersistent(next.GroupBys, groupBys...)
	return SelectBuilder{state: &next}
}

// Having adds an expression to the HAVING clause of the query.
//
// See Where.
func (b SelectBuilder) Having(pred any, rest ...any) SelectBuilder {
	next := b.clone()
	next.HavingParts = appendPersistent(next.HavingParts, newWherePart(pred, rest...))
	return SelectBuilder{state: &next}
}

// OrderByClause adds ORDER BY clause to the query.
func (b SelectBuilder) OrderByClause(pred any, args ...any) SelectBuilder {
	next := b.clone()
	next.OrderByParts = appendPersistent(next.OrderByParts, newPart(pred, args...))
	return SelectBuilder{state: &next}
}

// OrderBy adds ORDER BY expressions to the query.
func (b SelectBuilder) OrderBy(orderBys ...string) SelectBuilder {
	for _, orderBy := range orderBys {
		b = b.OrderByClause(orderBy)
	}

	return b
}

// Limit sets a LIMIT clause on the query.
func (b SelectBuilder) Limit(limit uint64) SelectBuilder {
	next := b.clone()
	next.Limit = strconv.FormatUint(limit, 10)
	return SelectBuilder{state: &next}
}

// RemoveLimit removes LIMIT clause allowing access to all records.
func (b SelectBuilder) RemoveLimit() SelectBuilder {
	next := b.clone()
	next.Limit = ""
	return SelectBuilder{state: &next}
}

// Offset sets a OFFSET clause on the query.
func (b SelectBuilder) Offset(offset uint64) SelectBuilder {
	next := b.clone()
	next.Offset = strconv.FormatUint(offset, 10)
	return SelectBuilder{state: &next}
}

// RemoveOffset removes OFFSET clause.
func (b SelectBuilder) RemoveOffset() SelectBuilder {
	next := b.clone()
	next.Offset = ""
	return SelectBuilder{state: &next}
}

// Suffix adds an expression to the end of the query.
func (b SelectBuilder) Suffix(sql string, args ...any) SelectBuilder {
	return b.SuffixExpr(Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query.
func (b SelectBuilder) SuffixExpr(e Sqlizer) SelectBuilder {
	next := b.clone()
	next.Suffixes = appendPersistent(next.Suffixes, e)
	return SelectBuilder{state: &next}
}

// With adds a CTE (Common Table Expression) to the query.
func (b SelectBuilder) With(cteName string, cte SelectBuilder) SelectBuilder {
	return b.PrefixExpr(cte.Prefix(fmt.Sprintf("WITH %s AS (", cteName)).Suffix(")"))
}
