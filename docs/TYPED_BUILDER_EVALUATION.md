# Typed builder replacement evaluation

Status: post-v0.1.0 design candidate. This document does not change the v0.1.0
decision in `docs/FOUNDATION.md`: that release keeps `github.com/lann/builder` and
`github.com/lann/ps`. Any replacement must run as a separate tracked stream after
v0.1.0 release validation and tagging.

## Summary

Replacing `lann/builder` is feasible and likely worthwhile for internal type safety,
allocation reduction, and removal of the root module's final external dependencies.
The recommended design is not a generic reimplementation of its dynamic
`string -> any` property bag. Squirrel should instead give each public builder a
typed immutable state and use generics only for small reusable copy-on-write or
persistent-collection helpers.

This is primarily an internal safety and maintainability improvement. Public methods
such as `Where(any, ...any)`, `Expr(string, ...any)`, and `Set(string, any)` remain
intentionally dynamic because they model SQL expressions and bound values.

## Current cost and failure modes

The root package currently makes 72 calls to `builder.Register`, `Set`, `Append`,
`Extend`, `Delete`, and `GetStruct` across `StatementBuilderType`, Select, Insert,
Update, Delete, Case, and CTE builders.

For every fluent mutation, `lann/builder`:

1. converts a named Squirrel builder to and from `builder.Builder` with reflection;
2. looks up a field using a string key in a heterogeneous persistent map;
3. stores appended values in a heterogeneous persistent list;
4. returns `any`, requiring a runtime type assertion at the call site.

For every `ToSql`, `GetStruct` reads the registry under a lock, allocates a new state
struct, finds its fields by name, materializes persistent lists as typed slices, and
assigns values with reflection.

Consequences include runtime-only detection of misspelled field names or mismatched
field types, registration through `init`, substantial allocation volume, and a
second representation of state solely for rendering SQL.

## Exploratory benchmark

The following directional microbenchmark was run on Go 1.26.2, linux/amd64, on an
AMD EPYC-Rome processor. Each result used five benchmark runs. The benchmark was an
uncommitted standalone harness and did not alter the repository.

Representative state construction and materialization:

| Representation | Time/op | Bytes/op | Allocs/op |
|---|---:|---:|---:|
| Current `lann/builder` map/list plus `GetStruct` | 10.1-11.8 µs | 1,808 | 41 |
| Typed state with copy-on-write slices | 0.33-0.40 µs | 80 | 3 |
| Typed state with generic persistent lists | 0.48-0.58 µs | 160 | 6 |

A second benchmark built and rendered the same ordinary SELECT shape:

| Implementation | Time/op | Bytes/op | Allocs/op |
|---|---:|---:|---:|
| Current Squirrel | 22.2-27.1 µs | 4,508 | 87 |
| Simplified typed prototype | 3.1-3.7 µs | 1,048 | 21 |

The end-to-end prototype was intentionally small and did not reproduce every nested
`Sqlizer`, CTE, error, or placeholder path. Its roughly 7x time and 4x allocation
improvement is an estimate of the opportunity, not a production performance promise.
Database latency will usually dominate these microseconds; the reduction matters most
for high-rate, bulk, or allocation-sensitive SQL generation.

## Recommended design

Use a private typed state per builder and keep every existing public method signature:

```go
type SelectBuilder struct {
    state *selectData
}

func (b SelectBuilder) Where(pred any, args ...any) SelectBuilder {
    next := cloneState(b.state)
    next.WhereParts = appendCopy(next.WhereParts, newWherePart(pred, args...))
    return SelectBuilder{state: next}
}

func appendCopy[T any](items []T, values ...T) []T {
    return append(items[:len(items):len(items)], values...)
}
```

The exact representation must be selected by benchmark:

- A pointer to immutable typed state keeps public builder values small and comparable.
- Copy-on-write slices are simple and likely best for ordinary short query chains.
- Generic persistent lists preserve O(1) branching/appends, but add nodes and require
  materialization before rendering.
- A universal `Builder[T]` with string field names is not sufficient: heterogeneous
  fields still force `any` or reflection and retain the current failure modes.

`StatementBuilderType` should become an explicit seed containing common defaults such
as placeholder format and initial WHERE parts. Its constructor methods should copy
those defaults into the typed Select/Insert/Update/Delete/CTE state rather than rely
on conversions between named types with the same underlying `builder.Builder`.

## Compatibility hazards

The public method inventory can remain unchanged, but representation-dependent Go
code may break. The implementation and migration review must cover:

- external conversions between `StatementBuilderType`, concrete Squirrel builders,
  and `lann/builder.Builder`;
- external calls to `builder.Set` or `builder.GetStruct` with Squirrel builder values;
- comparability and use of builder values as map keys;
- valid behavior of zero-value builders;
- immutable sibling reuse and concurrent rendering;
- slice backing-array aliasing after every append/extend operation;
- caller-owned values and slices retaining the same mutation semantics;
- deterministic `SetMap` ordering;
- final-pass numbering across every nested builder and custom `Sqlizer`.

The first two patterns are outside the documented Squirrel API, but they are currently
possible Go source and therefore must be called out as representational breaking
changes instead of being silently ignored.

## Separate post-v0.1.0 work stream

1. Commit reproducible benchmarks for construction, rendering, branching, and reuse
   of Select, Insert, Update, Delete, Case, CTE, and `StatementBuilderType`.
2. Add differential tests that apply identical operation sequences to the released
   v0.1.0 implementation and a typed prototype, comparing SQL, args, and errors.
3. Prototype only `SelectBuilder`, including zero values, `StatementBuilder` defaults,
   nested placeholders, and immutable sibling branches.
4. Run unit, race, fuzz, compatibility, and PostgreSQL integration suites on Go 1.25
   and the current supported Go version.
5. Review benchmark and compatibility evidence at a human decision gate.
6. If accepted, migrate the remaining builders, remove both `lann` modules, regenerate
   the API inventory, and document representation-dependent incompatibilities.

## Acceptance gate for adoption

The replacement is ready to merge only when:

- exported API inventory is unchanged unless a difference is explicitly approved;
- all differential and existing regression tests pass;
- immutable reuse and race tests pass for every builder;
- nested placeholder fuzzing and real PostgreSQL integration pass;
- representative construction/rendering benchmarks show no time regression and a
  material allocation reduction (target at least 50%);
- the root module has no runtime dependencies after `go mod tidy`;
- migration notes describe unsupported representation-dependent usage.

Until this gate is satisfied, `lann/builder` remains the compatibility reference for
immutable state behavior.
