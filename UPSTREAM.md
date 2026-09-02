# Upstream

This project is derived from:

- [Masterminds/squirrel](https://github.com/Masterminds/squirrel)
- [n-r-w/squirrel](https://github.com/n-r-w/squirrel)

## Baseline

- Repository: `github.com/n-r-w/squirrel`
- Release: `v1.6.0`
- Commit: `4c87dbba0f35938b7af0171bf00c4276238e7907`
- Local marker: `upstream/n-r-w-v1.6.0`
- Captured: 2026-09-02

The `main` branch was initialized from this exact commit with its full history. The
fork starts its own release sequence at `v0.1.0`; upstream release tags must not be
pushed to `origin` in bulk.

## Remotes

| Remote | URL | Purpose |
|---|---|---|
| `origin` | `https://github.com/muonsoft/squirrel.git` | maintained fork |
| `n-r-w` | `https://github.com/n-r-w/squirrel.git` | implementation baseline and review source |
| `masterminds` | `https://github.com/Masterminds/squirrel.git` | compatibility reference |

## Update policy

Upstream branches are never merged automatically. Review every change for scope,
dependency impact, compatibility, and regression coverage. Import an accepted fix
by a focused cherry-pick or manual port and preserve/add a regression test. Reference
the upstream repository and commit in the adapting commit message.

Record future reviews here with the date, reviewed commit, imported changes, and
explicitly rejected changes. Do not create a tag for every review.

## Local verification

```bash
git rev-parse upstream/n-r-w-v1.6.0^{}
git remote -v
```

The expected revision is
`4c87dbba0f35938b7af0171bf00c4276238e7907`.
