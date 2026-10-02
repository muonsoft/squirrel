package squirrel

// StatementBuilderType is the type of StatementBuilder.
type StatementBuilderType struct {
	state *statementDataState
}

type statementDataState struct {
	PlaceholderFormat PlaceholderFormat
	WhereParts        immutableList[Sqlizer]
}

func (b StatementBuilderType) clone() statementDataState {
	if b.state == nil {
		return statementDataState{}
	}
	return *b.state
}

// Select returns a SelectBuilder for this StatementBuilderType.
func (b StatementBuilderType) Select(columns ...string) SelectBuilder {
	next := selectDataState{}
	next.PlaceholderFormat = b.clone().PlaceholderFormat
	next.WhereParts = b.clone().WhereParts
	return SelectBuilder{state: &next}.Columns(columns...)
}

// Insert returns a InsertBuilder for this StatementBuilderType.
func (b StatementBuilderType) Insert(into string) InsertBuilder {
	next := insertDataState{}
	next.PlaceholderFormat = b.clone().PlaceholderFormat
	return InsertBuilder{state: &next}.Into(into)
}

// Replace returns a InsertBuilder for this StatementBuilderType with the
// statement keyword set to "REPLACE".
func (b StatementBuilderType) Replace(into string) InsertBuilder {
	next := insertDataState{}
	next.PlaceholderFormat = b.clone().PlaceholderFormat
	return InsertBuilder{state: &next}.statementKeyword("REPLACE").Into(into)
}

// Update returns a UpdateBuilder for this StatementBuilderType.
func (b StatementBuilderType) Update(table string) UpdateBuilder {
	next := updateDataState{}
	next.PlaceholderFormat = b.clone().PlaceholderFormat
	next.WhereParts = b.clone().WhereParts
	return UpdateBuilder{state: &next}.Table(table)
}

// Delete returns a DeleteBuilder for this StatementBuilderType.
func (b StatementBuilderType) Delete(from string) DeleteBuilder {
	next := deleteDataState{}
	next.PlaceholderFormat = b.clone().PlaceholderFormat
	next.WhereParts = b.clone().WhereParts
	return DeleteBuilder{state: &next}.From(from)
}

// With returns a CommonTableExpressionsBuilder for this StatementBuilderType.
func (b StatementBuilderType) With(cte string) CommonTableExpressionsBuilder {
	next := commonTableExpressionsDataState{}
	next.PlaceholderFormat = b.clone().PlaceholderFormat
	return CommonTableExpressionsBuilder{state: &next}.Cte(cte)
}

// PlaceholderFormat sets the PlaceholderFormat field for any child builders.
func (b StatementBuilderType) PlaceholderFormat(f PlaceholderFormat) StatementBuilderType {
	next := b.clone()
	next.PlaceholderFormat = f
	return StatementBuilderType{state: &next}
}

// Where adds WHERE expressions to the query.
//
// See SelectBuilder.Where for more information.
func (b StatementBuilderType) Where(pred any, args ...any) StatementBuilderType {
	next := b.clone()
	next.WhereParts = appendPersistent(next.WhereParts, newWherePart(pred, args...))
	return StatementBuilderType{state: &next}
}

// StatementBuilder is a parent builder for other builders, e.g. SelectBuilder.
//
//nolint:gochecknoglobals // common starting point for building statements
var StatementBuilder = StatementBuilderType{}.PlaceholderFormat(Question)

// Select returns a new SelectBuilder, optionally setting some result columns.
//
// See SelectBuilder.Columns.
func Select(columns ...string) SelectBuilder {
	return StatementBuilder.Select(columns...)
}

// Insert returns a new InsertBuilder with the given table name.
//
// See InsertBuilder.Into.
func Insert(into string) InsertBuilder {
	return StatementBuilder.Insert(into)
}

// Replace returns a new InsertBuilder with the statement keyword set to
// "REPLACE" and with the given table name.
//
// See InsertBuilder.Into.
func Replace(into string) InsertBuilder {
	return StatementBuilder.Replace(into)
}

// Update returns a new UpdateBuilder with the given table name.
//
// See UpdateBuilder.Table.
func Update(table string) UpdateBuilder {
	return StatementBuilder.Update(table)
}

// Delete returns a new DeleteBuilder with the given table name.
//
// See DeleteBuilder.Table.
func Delete(from string) DeleteBuilder {
	return StatementBuilder.Delete(from)
}

// With returns a new CommonTableExpressionsBuilder with the given first cte name.
//
// See CommonTableExpressionsBuilder.Cte.
func With(cte string) CommonTableExpressionsBuilder {
	return StatementBuilder.With(cte)
}

// WithRecursive returns a new CommonTableExpressionsBuilder with the RECURSIVE option and the given first cte name.
//
// See CommonTableExpressionsBuilder.Cte, CommonTableExpressionsBuilder.Recursive.
func WithRecursive(cte string) CommonTableExpressionsBuilder {
	return StatementBuilder.With(cte).Recursive(true)
}

// Case returns a new CaseBuilder.
// "what" represents case value.
func Case(what ...any) CaseBuilder {
	b := CaseBuilder{}

	switch len(what) {
	case 0:
	case 1:
		b = b.what(what[0])
	default:
		b = b.what(newPart(what[0], what[1:]...))
	}
	return b
}
