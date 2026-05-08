# Workflow — bitbottle

The end-to-end execution rules for any change. The canonical procedure
checklists live in [`docs/workflows/`](../docs/workflows/) — this file
states the *policies* behind those procedures.

## TDD policy

**Strict by default. Red → green → refactor.**

- This applies to ~99% of work.
- Write a failing test first; make it pass with the smallest change;
  then refactor with the test as the safety net.
- One logical change per commit.

**Bypass.** Allowed only when the testing infrastructure does not yet
exist (very early prototype or spike). Bypassing must be a deliberate
decision, documented in the track's `spec.md` with the reason — never a
default. Once the testable surface lands, the bypass ends; existing code
gets retro-tests.

For Go, "test exists" means a `_test.go` file in the same package using
`factory.NewTestFactory` — never real filesystem, keyring, or network.

## Commit strategy

**Conventional Commits on every commit.** The full set:

```
feat | fix | docs | chore | refactor | test | perf | build | ci | style | revert
```

- Branch name: `feature/*` | `fix/*` | `docs/*` | `chore/*` (only these
  prefixes pass the pre-merge gate).
- Subject line: `<type>(<scope>): <subject>`. Scope is optional but
  recommended (`feat(pr): ...`).
- Bodies are encouraged for any non-trivial change — explain *why*, not
  *what* (the diff shows what).

**Squash-merge per PR.** Each PR contributes exactly one commit to
`main`. The PR title becomes that commit's subject — so the PR title
must itself be a Conventional Commit.

**Squash-merge gotcha (critical).** release-please only bumps versions
on `feat:` / `fix:` / `feat!:` subjects. If any commit on the branch is
`feat:`/`fix:` but the PR title is something else, the release won't
fire and a follow-up empty `feat:` commit is needed. History: PRs
#48 → #49 → #50.

## Code review

**Required on every PR.**

For this single-maintainer project, that means:

1. The maintainer's own self-review pass over the full diff before
   merging.
2. An automated code-reviewer agent (e.g. `code-reviewer` or
   `superpowers:requesting-code-review`) on every PR, with output
   attached to the PR or worked through in chat.

Pre-merge-check ([`docs/workflows/pre-merge-check.md`](../docs/workflows/pre-merge-check.md))
is the **BLOCKER** gate that runs alongside the human/agent review —
not a substitute for review.

## Verification checkpoints

Verification scales by frequency: cheap things every commit, more
expensive things less often.

| Checkpoint | When | What runs | Severity |
|---|---|---|---|
| Per-commit | Pre-commit hook + CI on push | `golangci-lint`, `go test ./... -race` | BLOCKER |
| Per-PR | Before merge | [`pre-merge-check`](../docs/workflows/pre-merge-check.md) sections 0–8 | BLOCKER |
| Per-track | Before track is closed | Spec checklist in `conductor/tracks/<name>/spec.md` | BLOCKER |
| Per-release | Before tagging | [`docs/manual-tests/`](../docs/manual-tests/) scenarios against real Bitbucket Cloud and Server/DC instances | BLOCKER |

There is **no** per-task verification gate — that pace is too slow for
this project. Tasks roll up to per-track verification.

## Track lifecycle

A track is the unit of work in Conductor — typically one scope from
`BACKLOG.md`, but can also be a bug or a refactor.

```
new-track  →  spec.md + plan.md committed
   │
   ├─ phase 1 (e.g., API layer) — implement under TDD
   ├─ phase 2 (e.g., commands)
   ├─ phase 3 (e.g., MCP + docs)
   │
   └─ track complete  →  pre-merge-check passes
                      →  PR opened, reviewed, squash-merged
                      →  release-please PR auto-merges
                      →  track moved to closed in tracks.md
                      →  PRD issue closed
```

For the full step-by-step (preflight, scope pick, halt points, release
merge, manual-test refresh, compact), see
[`docs/workflows/iteration-cycle.md`](../docs/workflows/iteration-cycle.md).

## Anti-patterns

- **Skipping pre-merge-check** because "it's small". The squash-merge
  gotcha and `dist/` artifacts have hit this repo before.
- **Editing `CHANGELOG.md` by hand.** release-please owns it.
- **`--no-verify` or `git push --force` as escape hatches.** Diagnose
  the underlying failure instead.
- **Marking a `BACKLOG.md` row ✅ before the release PR has merged.**
  Until it's published, it's not shipped.
- **Bypassing TDD without writing the bypass into spec.md.** Silent
  bypasses become forever-bypasses.
