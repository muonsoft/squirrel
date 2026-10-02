# Typed builder implementation review

Status: implementation candidate; no merge, push, tag, or release is performed.
The baseline is the published v0.1.0 commit
`b5bb755e0097874fe2b0602b9c77b66b5412924e`, verified against origin before work.

## Design and scope

Each public builder owns a pointer to private typed state. Mutations copy the state
and append immutable typed nodes. Rendering materializes collections as ordered
slices and uses the existing SQL renderer. There is no reflection, string-keyed
property registry, global registration, or external runtime dependency.

Copy-on-write slices were prototyped first but rejected after long-chain benchmarks
showed quadratic copying. The committed collection benchmark compares this option
with persistent nodes at lengths 4, 32, and 256. At length 256, copying used about
555 KB/op versus 4 KB/op for nodes in the isolated integer-collection test. Persistent
nodes also avoid backing-array aliasing between sibling builders.

## Reproduction

```bash
# Fetch the published object if absent, without creating a local release tag:
git fetch --no-tags origin b5bb755e0097874fe2b0602b9c77b66b5412924e
bash scripts/compare-builders.sh
bash scripts/benchmark-builders.sh /tmp/squirrel-bench
go test -run '^$' -bench BenchmarkCollections -benchmem -count=5
go test -race ./...
bash scripts/test-all.sh
```

The differential harness lives in a separate `evaluation` module. The script
extracts the pinned baseline into a temporary directory and runs the same scenario
program against each implementation. It compares deterministic JSON including SQL,
argument values, error strings, and zero-value panics. It exercises all builders,
four placeholder formats, 128 branch variants per format, SetMap ordering, nested
Sqlizers, errors, caller-owned slices, and escaped operators. The full local gate
and CI matrix run this comparison. Concurrent branch/render tests cover all seven
builder types separately.

## Compatibility review

- Public functions and method signatures are unchanged; inventory changes are only
  the seven underlying builder type declarations.
- Builders remain comparable, but equality describes state identity, not SQL
  equivalence. No-op equality can differ from the old map identity.
- Conversions between concrete builders or to/from `lann/builder.Builder`, and calls
  to external `builder.Set`/`GetStruct`, are representation-dependent breaking changes.
- Zero-value behavior, final-pass placeholder numbering, and ownership of caller
  values are preserved by the exercised scenarios.
- One intentional correction is outside the equality comparison: WHERE defaults on
  statement seeds now apply only to SELECT/UPDATE/DELETE. INSERT and CTE ignore
  these defaults instead of panicking on an unknown reflected field. This has a
  dedicated regression test and migration/changelog entries.

These representation changes and the seed correction require maintainer review
before adoption. The branch is a complete candidate, not an assertion of approval.

## Benchmark evidence

Captured on Go 1.27.1, linux/amd64, AMD EPYC-Rome, with five sequential 200 ms samples
per benchmark for each implementation. The two implementations were not benchmarked
concurrently. Shared-host load still makes timing estimates noisy; these are not
production performance promises. Full raw output is in `typed-builder/`.

Median time improved in every measured scenario. Construction plus rendering uses
53–65% fewer allocations across the representative query shapes; SELECT uses 23
versus 58 allocations. Some render-only scenarios reduce allocations by less than
50%, because SQL text/argument allocation is unchanged. Long-chain construction and
branching also exceed the 50% allocation-reduction target. The benchmark cases are
representative rather than an exhaustive performance guarantee.

| Scenario | Baseline ns/op | Typed ns/op | Bytes/op baseline / typed | Allocs/op baseline / typed |
|---|---:|---:|---:|---:|
| Builders/CTE/Build | 20397 | 4300 | 2936 / 1536 | 62 / 19 |
| Builders/CTE/BuildRender | 26591 | 6124 | 5345 / 2320 | 103 / 38 |
| Builders/CTE/Render | 17277 | 4287 | 2408 / 784 | 41 / 19 |
| Builders/Case/Build | 5040 | 917.4 | 664 / 288 | 15 / 6 |
| Builders/Case/BuildRender | 5428 | 1774 | 1032 / 432 | 25 / 9 |
| Builders/Case/Render | 3258 | 903.6 | 368 / 144 | 10 / 3 |
| Builders/Delete/Build | 4938 | 557.9 | 664 / 216 | 14 / 4 |
| Builders/Delete/BuildRender | 15924 | 1629 | 1256 / 392 | 26 / 9 |
| Builders/Delete/Render | 5706 | 807.5 | 592 / 176 | 12 / 5 |
| Builders/Insert/Build | 10628 | 952 | 1672 / 320 | 35 / 7 |
| Builders/Insert/BuildRender | 19811 | 3841 | 2584 / 760 | 59 / 23 |
| Builders/Insert/Render | 15194 | 3198 | 912 / 440 | 24 / 16 |
| Builders/Nested/Build | 14418 | 4496 | 2672 / 1808 | 52 / 20 |
| Builders/Nested/BuildRender | 29358 | 11716 | 4977 / 2720 | 90 / 42 |
| Builders/Nested/Render | 16302 | 6770 | 2304 / 912 | 38 / 22 |
| Builders/Select/Build | 11688 | 3011 | 1992 / 1024 | 42 / 16 |
| Builders/Select/BuildRender | 27605 | 3455 | 2968 / 1280 | 58 / 23 |
| Builders/Select/Render | 9526 | 1740 | 976 / 256 | 16 / 7 |
| Builders/Statement/Build | 10356 | 1376 | 1248 / 640 | 27 / 9 |
| Builders/Statement/BuildRender | 22864 | 3937 | 2328 / 1024 | 45 / 19 |
| Builders/Statement/Render | 9187 | 2776 | 1080 / 384 | 18 / 10 |
| Builders/Update/Build | 9183 | 941.3 | 1136 / 296 | 23 / 5 |
| Builders/Update/BuildRender | 21299 | 3303 | 1952 / 576 | 40 / 14 |
| Builders/Update/Render | 10081 | 1885 | 816 / 280 | 17 / 9 |
| SelectBranch | 9363 | 1323 | 816 / 592 | 20 / 8 |
| SelectChain/256 | 972041 | 168959 | 101065 / 76328 | 2319 / 1030 |
| SelectChain/32 | 124014 | 21550 | 13256 / 10024 | 303 / 134 |
| SelectChain/4 | 14751 | 6393 | 2280 / 1736 | 51 / 22 |

Additional branch samples (same host/toolchain, five 200 ms runs):

| Scenario | Baseline ns/op | Typed ns/op | Bytes/op baseline / typed | Allocs/op baseline / typed |
|---|---:|---:|---:|---:|
| Builders/CTE/Branch | 8289 | 923.1 | 1312 / 432 | 31 / 7 |
| Builders/Case/Branch | 1844 | 440.9 | 360 / 192 | 10 / 4 |
| Builders/Delete/Branch | 1884 | 516.2 | 408 / 216 | 10 / 4 |
| Builders/Insert/Branch | 1811 | 409.9 | 512 / 208 | 11 / 3 |
| Builders/Nested/Branch | 1766 | 419.3 | 408 / 296 | 10 / 4 |
| Builders/Select/Branch | 1968 | 669.2 | 408 / 296 | 10 / 4 |
| Builders/Statement/Branch | 1743 | 627.5 | 408 / 296 | 10 / 4 |
| Builders/Update/Branch | 2518 | 384.7 | 488 / 208 | 10 / 2 |

## Validation

- Full `scripts/test-all.sh`: PASS on Go 1.27.1, including lint (zero issues),
  unit/race, fuzz, dependency policy, all module tidy checks, API inventory,
  differential comparison, and actual PostgreSQL execution.
- Go 1.25.14 and Go 1.26.2: unit/race, vet, differential comparison,
  compatibility fixture, focused 10-second fuzz smoke, and PostgreSQL integration:
  PASS. Fuzz runs completed 21,378 and 29,803 executions respectively.
- Root `go list -m all`: only `github.com/muonsoft/squirrel`; dependency policy now
  requires exactly that graph rather than merely rejecting known heavy dependencies.
- API inventory: only underlying representations changed; method signatures match.
- PostgreSQL used locally: 14.24, isolated in `/tmp`, with a nonempty DSN; integration
  tests were executed, not skipped. Docker pulled the configured 18.1 image but its
  runtime failed to start the container (`failed to create TTRPC connection`). Thus
  this run does not certify PostgreSQL 18.1; CI still uses 18.1 and must pass before
  adoption. No database dependencies entered the root module.
- Shell syntax and `git diff --check`: PASS.

## Adoption decision still outstanding

Review representation-dependent source incompatibilities, the seed-default
correction, raw benchmark evidence, and the CI PostgreSQL 18.1 result before
merging. Timing evidence comes from a shared host; repeat measurements on a stable
runner if a precise speedup is needed. No release version has been selected here.
