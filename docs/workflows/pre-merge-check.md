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
- If a PR exists (`gh pr view --json baseRefName,number,title,isDraft`):
  - `baseRefName` is `main`. Other base is BLOCKER.
  - `isDraft` is false (WARN if draft).

> **Note**: a PR being behind `origin/main` is **no longer a BLOCKER**.
> Branch protection has `required_status_checks.strict = false`, so
> GitHub permits merging without an up-to-date branch. Semantic-conflict
> risk is bounded for this repo (solo dev, sequential `/auto-iter`,
> squash merges); when it ever bites, main CI catches it within one
> push and a `git revert` PR is the fix. The previous "rebase or
> `gh pr update-branch <N>`" dance burned a full CI cycle per merge
> (~5.5 min) at near-zero benefit.

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
| Backlog item is now done | **Move**, don't flip: cut the row + scope-detail section from `docs/backlog/BACKLOG.md` and prepend a dated entry to `docs/backlog/SHIPPED.md`. Both edits in the same `feat:` commit as the code. The pre-merge mechanical check's §4 blocks commits touching ONLY one of those files. | BLOCKER |
| Auth, hosts.yml, or token handling | `skills/references/auth.md` | BLOCKER |
| `api/backend/{client,types,errors}.go` changed | `docs/agent-primer.md` still accurate (architecture vocabulary, invariants, exemplar-file table) | BLOCKER |
| `docs/workflows/iteration-cycle/` changed | `AGENTS.md` still references correct file paths; agent command files (e.g. `.claude/commands/auto-iter.md`) still point at correct sections | WARN |

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

### 6a. Architecture smells (BLOCKER) — recurring patterns this loop has shipped

The bullet list below is a hard checklist, added after a post-mortem of
releases v1.29.0 / v1.30.0 / v1.31.0 surfaced structural debt that the
per-PR judge had missed. Every item that fires is a BLOCKER unless the PR
description explicitly justifies the exception.

- **Repeated N-way capability switch.** If `pkg/cmd/<group>/` adds the
  same `As<X>Client → As<Y>Client → As<Z>Client` switch in three or more
  files (e.g. list/set/delete), refactor to a single
  `resolveOps(scope) <iface>` helper and call it once per command. Adding
  a fourth scope must not require editing three switches (OCP). Grep:
  `git diff origin/main...HEAD -- 'pkg/cmd/**/*.go' | grep -E 'As[A-Z][a-zA-Z]+Client\(\)' | wc -l` — ≥9 hits across ≥3 files in one PR is the trigger.
- **Per-command-tree `cmdtest` clones.** If the PR adds
  `pkg/cmd/<new>/internal/cmdtest/cmdtest.go` whose body is structurally
  identical to an existing `pkg/cmd/<other>/internal/cmdtest/cmdtest.go`,
  promote to a shared package instead of cloning.
- **Two surfaces for one job.** If the PR adds a command that overlaps
  with an open backlog item (`environment variable …` shipped before
  `variable --scope deployment`), either ship the unified form first or
  mark the older form deprecated in the same PR. Don't leave both wired
  in `pkg/cmd/root/root.go`.
- **Translation-table sprawl.** If the PR adds the third or fourth
  `ToCloud<X>` / `<X>ToCLI` pair side-by-side in `api/backend/types.go`,
  collapse them onto a typed enum with `MarshalCloud/MarshalServer/
  UnmarshalCloud/UnmarshalServer` methods.
- **Comment density above the codebase baseline.** Run
  `git diff origin/main...HEAD -- '*.go' | grep -E '^\+\s*//' | wc -l`
  against `wc -l` of added lines. If comment lines exceed ~5% of added
  Go lines and the file is not a new exported package, flag — CLAUDE.md
  prefers minimal comments and the trend across recent PRs is upward.
- **Self-referential write-op test, or unhonored backend quirk.** For a PR
  that adds or changes a write/mutation op: (a) if its test asserts only on
  stdout (e.g. `stdout 'Updated…'`) without asserting the **captured
  request body**, BLOCKER — that test cannot catch a dropped or wrong field,
  because the hand-written fake shares the code's assumptions; (b) if the op
  contradicts an applicable row in [`docs/backend-quirks.md`](../backend-quirks.md)
  (full-object Server PUT, `version` precondition, Content-Type policy),
  BLOCKER. #655 (`pr edit` wiped all reviewers; `pr request-review` 400'd)
  passed every gate precisely because the fake and the stdout-only test
  both encoded the same wrong assumption the code did.

Skip when the diff touches only docs, CI config, dependencies, or
`docs/backlog/BACKLOG.md` / `docs/backlog/SHIPPED.md`.

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
