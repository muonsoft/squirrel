# Builder differential harness

Run `bash scripts/compare-builders.sh` from the repository root. The script runs
this program against the pinned published v0.1.0 commit and the working tree, using
temporary module replacements. It compares SQL, JSON-compatible arguments, errors,
and recovered zero-value panics. It fails on any mismatch. Network access may be
needed to populate Go's checksum/dependency cache on the baseline run.

The harness intentionally exercises documented builder methods, not conversions to
`lann/builder` or undocumented cross-builder casts. WHERE defaults on INSERT/CTE
seeds are covered by a separate regression test because their old reflection panic
is intentionally removed. See `docs/audits/typed-builder-report.md` for limitations.
