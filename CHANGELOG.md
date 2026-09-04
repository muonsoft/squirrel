# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — planned

First maintained release of `github.com/muonsoft/squirrel`, based on n-r-w/squirrel
v1.6.0 and preserving attribution to Masterminds/squirrel.

### Added

- Continuous final-pass placeholder numbering for nested builders.
- Regression coverage for nested SELECTs, joins, expressions, columns, update values,
  CTEs, prefixes/suffixes, escaped PostgreSQL operators, and immutable builder reuse.
- `WITH`, `WITH RECURSIVE`, DML CTE, and PostgreSQL `UPDATE ... FROM` coverage.
- PostgreSQL execution tests in an isolated nested module.
- Placeholder fuzz testing and a Masterminds migration compile fixture.
- Exported API inventory, compatibility decisions, dependency policy checks, and
  upstream provenance documentation.
- Maintainer-dispatched GitHub Release workflow with release-candidate revalidation,
  deterministic changelog finalization, and CI-owned tag creation.

### Changed

- Module path is `github.com/muonsoft/squirrel`.
- The package is explicitly limited to SQL construction; execution and scanning are
  application responsibilities.
- `Case` behavior follows Masterminds/squirrel v1.5.4 rather than n-r-w v1.6.0's
  automatic literal casting.
- Root dependencies are reduced to `github.com/lann/builder` and its `lann/ps`
  transitive dependency.
- Integration dependencies, including pgx, are isolated in `integration/go.mod`.

### Removed

- Database execution, runner, row, and statement-cache APIs inherited from
  Masterminds/squirrel.
- Search, pagination, conditional-ordering, zero-value filtering, and stateful table
  alias helpers inherited from n-r-w/squirrel.
- pgx, scany, testdock, Docker ecosystem, testify, and `golang.org/x/exp` from the
  root module graph.

### Security

- Documented that placeholders protect values only; identifiers and raw SQL remain
  trusted-input surfaces.

[Unreleased]: https://github.com/muonsoft/squirrel/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/muonsoft/squirrel/releases/tag/v0.1.0
