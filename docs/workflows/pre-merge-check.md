# Pre-merge check

Run before merging any branch into `main`. Stop at the first **BLOCKER**.
For **WARN**, list it but continue. End with a punch list and a verdict.

Don't fix issues automatically — surface them; the author decides.

This doc is the single source of truth for pre-merge gates. Tooling
wrappers (Claude skill at `.claude/skills/pre-merge-check/`, GitHub
checks, etc.) defer to it.

> **Cross-document reference convention.** References from this doc to
> other docs use anchor names ("the pre-merge gate", "the doc-sync
> table", "the parallel-mode section") rather than section numbers.
> Internal references within the same doc may use section numbers.
> This prevents a renumbering in one doc from silently breaking links
> in another.

---

## 0. Scope first

```bash
git rev-parse --abbrev-ref HEAD          # current branch
git fetch origin main --quiet
git diff --name-only origin/main...HEAD  # files changed
git log --format='%s' origin/main..HEAD  # commit subjects
```

Cache the diff file list — later checks key off it. Don't re-run.

## 1. Branch & tree hygiene — BLOCKER

- Branch name matches `^(feature|fix|docs|chore)/`. `main` itself is a
  BLOCKER (never merge from main into main).
- `git status --porcelain` is empty.
- Branch is not behind `origin/main`:
  `git rev-list --count HEAD..origin/main` is `0`. If non-zero, rebase or
  `gh pr update-branch <N>`.
- If a PR exists (`gh pr view --json baseRefName,number,title,isDraft`):
  - `baseRefName` is `main`. Other base is BLOCKER.
  - `isDraft` is false (WARN if draft).

## 2. Conventional Commits + PR title — BLOCKER

Every commit subject must match
`^(feat|fix|docs|chore|refactor|test|perf|build|ci|style|revert)(\(.+\))?!?: `.

**Squash-merge gotcha (critical).** GitHub's squash uses the **PR title**
as the commit subject. Release-please only bumps on `feat:` / `fix:` /
`feat!:`. Therefore:

- If any commit on the branch is `feat:` or `fix:`, the PR title must
  also start with `feat:` / `fix:` / `feat!:`. Otherwise the release
  won't fire and a follow-up empty `feat:` commit is needed (history:
  PRs #48 → #49 → #50). BLOCKER.
- If only `docs:` / `chore:` / `refactor:` / `test:` commits, a PR title
  with the same prefix is fine — no release expected.
- `feat!:` or `BREAKING CHANGE` footer → confirm with the author this
  is intentional (major bump).

## 3. Build artifacts & repo cleanliness — BLOCKER

- No tracked files under `dist/`: `git ls-files dist/` empty.
- No tracked files > 1 MB:
  `git ls-files | xargs -I{} du -k "{}" 2>/dev/null | awk '$1>1024'` empty.
- The compiled `bitbottle` binary at repo root is gitignored but easy to
  accidentally `git add`. Check `git ls-files | grep -E '^bitbottle$'`
  is empty.
- No `.DS_Store`, `*.log`, `coverage.out` tracked in this branch's diff.

## 4. Lint & tests — BLOCKER

If a PR exists and has been pushed, **trust the CI run** — don't re-run
locally:

```bash
gh pr view --json statusCheckRollup \
  -q '[.statusCheckRollup[]|{name,conclusion}]'
```

Every check must be `SUCCESS`. If any is `FAILURE` / `IN_PROGRESS`,
surface the failing check name and log URL; don't try to fix locally
without first reading the CI failure.

Re-running `make lint && make test -race` against a SHA whose CI is
already green is the third run for the same code (TDD → CI → here) and
adds ~5 minutes for no signal.

**Local fallback** — only when there is no PR yet (e.g. running the gate
before `git push`):

```bash
make lint    # golangci-lint
make test    # go test ./... -race
```

Both must exit 0. Surface offending file:line for lint, exact test name
for failures — don't paraphrase. `go vet ./...` is part of `make test`;
no separate run needed.

## 5. Doc sync (conditional) — BLOCKER on required docs

Match against the changed-file list from §0. Multiple rules can fire.

| If diff touches… | Verify also updated | Severity |
|---|---|---|
| `pkg/cmd/<group>/<new-file>.go` (new command) | `pkg/cmd/<group>/<group>.go` registers it (or self-registers via `pkg/cmdregistry`); **both** `api/cloud/` and `api/server/` have the operation; `pkg/cmd/mcp/` has a matching tool if it's a Bitbucket op | BLOCKER |
| New command in `pkg/cmd/<group>/` | `skills/SKILL.md` extended with the new command | BLOCKER |
| New / changed flag in `pkg/cmd/**` | `skills/references/{auth,pr,repos,api}.md` reflects it (the curated overview — exhaustive flag detail stays in `bitbottle <cmd> --help`) | BLOCKER |
| User-visible flow change not covered by an existing manual smoke | `docs/manual-tests/` — extend the relevant smoke (`cloud/pr-happy-path.md`, `server/pr-happy-path.md`, `shared/multi-host.md`) or add a fresh scenario file | WARN |
| User-visible UX change (output format, new subcommand) | `README.md` | BLOCKER |
| Branch strategy / commit / release / setup change | `CONTRIBUTING.md`, `AGENTS.md` | BLOCKER |
| New invariant or pattern (transport policy, paging helper, etc.) | `AGENTS.md` "Key rules for AI agents" | WARN |
| Backlog item is now done | `BACKLOG.md` (mark complete / remove) | BLOCKER |
| Auth, hosts.yml, or token handling | `skills/references/auth.md` | BLOCKER |
| `api/backend/{client,types,errors}.go` changed | `docs/agent-primer.md` still accurate (architecture vocabulary, invariants, exemplar-file table) | BLOCKER |
| `docs/workflows/iteration-cycle.md` changed | `AGENTS.md` and `.claude/skills/iteration-cycle/SKILL.md` (per-user, gitignored — flag as WARN; contributor's responsibility) still reference correct section numbers | WARN |

If `skills/references/*.md` was edited, also check `skills/SKILL.md`'s
router table still points at the right file.

## 6. Design-judge — BLOCKER on principle violations

For PRs that introduce new commands, interfaces, packages, transports, MCP
tools, or error sites, sanity-check against the two principle docs:

- [`docs/TASTE.md`](../TASTE.md) — UX (gh philosophy, standard flags,
  TTY-aware output, error format), agentic skill experience, MCP tool shape.
- [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) — SOLID, layered structure,
  composite + optional interfaces, deep modules, design decisions, gh
  references.

For each new surface, either cite an exemplar file from the codebase
demonstrating the principle is followed, or cite a violation with file:line
and the principle violated. No vague "feels off" — every finding must point
at code.

Skip when the diff touches only docs, CI config, dependencies, or
`BACKLOG.md`.

## 7. Release-please boundaries — BLOCKER

Release-please owns these. **Never hand-edit on a feature branch:**

- `CHANGELOG.md`
- `.release-please-manifest.json`
- Any line tagged `<!-- x-release-please-version -->` (currently in
  `skills/SKILL.md` lines 8 and 56).

If the diff touches any of these, BLOCKER unless the branch *is* a
release-please branch (named `release-please--*`).

## 8. Secret leak scan — BLOCKER on any hit

```bash
git diff origin/main...HEAD -- . ':(exclude)reference/' \
  | grep -nEi '(BBPAT|BBToken|app[_-]?password|x-token-auth|ATATT|BB_TOKEN=|hosts\.yml)' \
  || echo "clean"
```

Reject any literal corporate hostnames, internal logins, or personal
emails that belong only in local config — they must never land in
commits. (For this repo: the maintainer's Bitbucket Server host /
username should not appear in any tracked file.)

## 9. Final report

```
## Pre-merge check: <branch> → main

BLOCKERS (N):
  - <one-line, file:line where applicable>

WARNINGS (M):
  - <one-line>

CHANGED: <count> files, <±lines>
COMMITS: <count>; PR title: "<title>" — release impact: <none|patch|minor|major>

Verdict: <READY TO MERGE | NOT READY — fix BLOCKERS>
```

Don't summarize what the branch *does* — that's the PR description's
job. This report is purely about merge readiness.
