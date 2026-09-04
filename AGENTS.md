# Project agent rules

This repository is the maintained `github.com/muonsoft/squirrel` fork.

## Sources of truth

Read these before changing the library:

1. `docs/FOUNDATION.md` — fixed scope and architectural decisions.
2. `docs/API_COMPATIBILITY.md` — public compatibility contract.
3. `UPSTREAM.md` — provenance and upstream policy.
4. `docs/release-checklist.md` — release procedure and verification.

The post-v0.1.0 builder replacement is a separate design stream described in
`docs/TYPED_BUILDER_EVALUATION.md`.

## Non-negotiable boundaries

- Keep the package name `squirrel` and target module path `github.com/muonsoft/squirrel`.
- The root module is a pure SQL builder. Do not add database execution, scanning,
  connection, transaction, pagination-policy, or search-policy APIs.
- Preserve final-pass placeholder numbering for nested builders.
- Keep `github.com/lann/builder` through v0.1.0. Any replacement requires its own
  reviewed and evidence-gated change stream.
- PostgreSQL integration dependencies belong only in the nested `integration` module.
- Do not merge upstream wholesale, push branches/tags, publish GitHub Releases, or
  rewrite published history. The maintainer-dispatched Release workflow is the only
  path authorized to push its changelog-only commit and create a release tag.
- Preserve MIT licensing and upstream attribution.

## Work discipline

- Preserve unrelated user changes.
- Keep changes focused and commits atomic.
- Update public documentation and `CHANGELOG.md` when behavior changes.
- Run checks proportional to the change and the full release gate before publication.
- Never weaken tests or dependency policy merely to make a check pass.
- Never create `v0.1.0` locally; GitHub creates it when the maintainer-dispatched
  Release workflow publishes the verified release.
