# Iteration cycle

End-to-end loop for shipping one scope: pick → spec → build → ship → tidy.

This doc is the single source of truth for how a scope moves through the
repo. Tooling wrappers (Claude skill at
`.claude/skills/iteration-cycle/`, IDE snippets, etc.) defer to it.
Human contributors and any agent (Codex, Cursor, Aider, Claude) follow
the same sections.

## Design

- **Composition over a monolith.** Each phase has one job and a clear
  exit criterion. Phases are durable pause points — safe to stop, hand
  off, or resume.
- **Halt points are explicit.** Anything with shared-state blast radius
  (PR creation, merge, release merge, closing issues) requires the
  author's go-ahead. Local edits and tests do not.
- **Failure is loud, not silent.** If an exit check fails, stop and
  report — never paper over with `--no-verify`, force pushes, or
  skipped tests. See `AGENTS.md` and `docs/workflows/pre-merge-check.md`.
- **`gh` as the design reference.** When a step's shape isn't obvious,
  look at how `gh` does the analogous thing — `reference/gh/` is checked
  in for exactly this.

## Section 0 — Preflight

**Goal**: confirm the workspace can take a new iteration.

**Actions**
1. Confirm `pwd` is the bitbottle repo root.
2. `git status` is clean (no uncommitted changes, no stale staged files).
3. `git rev-parse --abbrev-ref HEAD` is `main`.
4. `git pull --ff-only` succeeds. If not, surface the divergence and
   stop.
5. Read `BACKLOG.md` headings so the rest of the loop has accurate
   context.

**Exit**: clean tree on `main`, up to date with `origin/main`.

## Section 1 — Pick the scope

**Goal**: produce a single, named scope to ship this iteration.

**Decision flow**
1. **Open backlog?** Look at `BACKLOG.md` → `## Backlog` table. Anything
   with `Status` ≠ ✅ is open. If there are open scopes, list them with
   their one-line description and **ask the author which scope to ship**.
2. **All shipped?** Research new features. In order:
   - Re-read `BACKLOG.md` → `## Philosophy` and `## Full Functionality
     Map` for any 🚧/🆕 row added since the last consolidation.
   - Look at `reference/gh/` for nearby `gh` commands bitbottle doesn't
     have yet (parity gaps).
   - Skim recent `git log --oneline -30` for FIXMEs, deferred work, or
     "follow-up" commits.
   - Surface candidates with a one-line proposal each.
3. **Nothing emerges?** Run a stress-test interview against the rough
   ideas (the `grill-me` skill on Claude Code does this; on other
   agents, drive a similar Q&A loop) until one scope crystallises.

**Exit**: one scope chosen, with explicit author confirmation.

**Anti-patterns**
- Picking multiple scopes in one iteration. One scope = one PR = one
  release.
- Inventing a scope without checking the backlog first.

## Section 2 — Write the PRD

**Goal**: a single PRD captured as a GitHub issue, scoped to the chosen
scope and only that scope.

**Actions**
1. Draft the PRD from current context — problem, in-scope, out-of-scope,
   API surface, test plan. (The `to-prd` skill on Claude Code automates
   this; on other tooling, write the markdown by hand or via your
   equivalent.)
2. File it as a GitHub issue: `gh issue create --title "<scope>" --body
   "<prd>" --label prd`.
3. **Capture the issue number** — every later phase references it
   (`refs #NNN`, `Closes #NNN`).
4. Sanity-check the PRD against `BACKLOG.md` → "Architecture Contract
   (per scope)" and "Definition of Done". Those rows are non-negotiable
   and must appear in the PRD's checklist.

**Exit**: GitHub issue created, issue number captured, DoD checklist
embedded in the PRD.

## Section 3 — Implement (TDD)

**Goal**: green tests + green lint on a feature branch.

**Branching**
- `git checkout -b feat/<short-slug>` (or `fix/...` / `docs/...` per
  `AGENTS.md`). Never work on `main`.
- One scope = one branch = one PR. No drive-by refactors.

**Discipline**
- Red → green → refactor. Write a failing test first, make it pass with
  the smallest change, then refactor with the test as the safety net.
  One logical change per commit.
- Conventional Commit subjects: `feat(scope): ...`, `fix(scope): ...`,
  `docs(scope): ...`.

**Layer order** (per `BACKLOG.md` Architecture Contract)
1. `api/backend/client.go` — new interface(s), or extend the composite
   `Client`. New types in `api/backend/types.go`.
2. `api/cloud/<domain>.go` + `_test.go`.
3. `api/server/<domain>.go` + `_test.go` — skip only if Cloud-only, and
   document why with `ErrUnsupportedOnHost`.
4. `pkg/cmd/<domain>/...` cobra commands. Every applicable command must
   support `--json` / `--jq` / `--hostname`.
5. `pkg/cmd/mcp/tools.go` registration + `handlers.go` method.

**Verify**
- `go test ./... -race` after each green and before pushing.
- `make lint` clean before opening a PR.

**Exit**: every Definition-of-Done row for the scope is checked.

## Section 4 — Sync docs and tooling-neutral context

**Goal**: shared context reflects the new surface area before the PR
opens.

**In-repo docs** (update only what the scope actually changed)
- `README.md` — new command section, examples.
- `BACKLOG.md` — flip the row to ✅, update the "Full Functionality
  Map" entries.
- `AGENTS.md` — only if a new project-wide rule emerged (a new pattern
  for adapters, a new invariant, etc.).
- `skills/SKILL.md` — single-file Claude skill bundled with the npm
  package: add the new commands.
- `CHANGELOG.md` — **do not hand-edit**; release-please writes it from
  Conventional Commits.
- `docs/manual-tests/...` — see Section 9; only if surface area
  changed.

**Agent-local memory** (e.g., Claude's auto-memory at
`~/.claude/projects/.../memory/`, Cursor rules, etc. — outside the
repo)
- Update only invariants that have actually changed: stack, release
  pipeline, user-setup. Stale "no X in repo" hooks are worse than
  none.
- Do not add ephemeral task state — that's what plans / todos /
  issues are for.

**Exit**: `git diff` shows doc changes alongside code changes in the
same branch.

## Section 5 — Pre-merge gate

**Goal**: catch the squash-merge gotcha and the seven other classes of
issues before they touch `main`.

Run `docs/workflows/pre-merge-check.md` end-to-end. Sections 0–8 cover
branch hygiene, Conventional Commits, no build artifacts in `dist/`,
`make lint` + `go test`, doc sync, release-please boundaries, and a
secret scan.

Do **not** proceed past this section if any check fails. Fix the
underlying issue and re-run; never bypass.

**Exit**: pre-merge-check returns green.

## Section 6 — Open the PR

**Goal**: PR open against `main`, titled so release-please picks it up.

**Actions**
1. `git push -u origin <branch>`.
2. `gh pr create --base main` with:
   - **Title** — Conventional Commits prefix (`feat(scope): ...`,
     `fix(scope): ...`). Squash-merge uses the PR title as the commit
     subject; getting this wrong forces a follow-up empty commit (see
     #48 → #49 → #50).
   - **Body** — link the PRD issue (`Closes #NNN`), short summary, and a
     test plan checklist mirroring the manual tests touched in Section
     9.
3. Wait for CI to go green. Do not request review or push more commits
   while CI is running unless something is actually broken.

**Halt point**: report the PR URL to the author and wait for explicit
"merge it". The author may want to review the diff first.

## Section 7 — Merge the PR, then merge the release PR

**Goal**: change is on `main` and a release is published.

**Actions**
1. After confirmation: `gh pr merge <PR> --squash --delete-branch`.
   Confirm the squash subject still has the `feat:`/`fix:` prefix.
2. **Wait for release-please** to open (or update) a release PR
   against `main`. This is automated and usually appears within a
   minute or two. `gh pr list --label "autorelease: pending"` is a
   quick check.
3. **Halt point**: surface the release PR URL and wait for explicit
   confirmation before merging it. Releases are public and
   irreversible.
4. After confirmation: `gh pr merge <release-PR> --squash`. GoReleaser
   then builds binaries, publishes the GitHub release, pushes Docker
   images, and publishes to npm.
5. Verify: `gh release view --json tagName,publishedAt` shows the new
   tag. `npm view @proggarapsody/bitbottle version` matches.

**Exit**: new tag published, npm version bumped, both PRs closed.

## Section 8 — Close the PRD

**Goal**: the PRD issue from Section 2 is closed; the workspace has no
stale PRD artifacts.

The phrase "delete PRD" in the workflow brief means: close the GitHub
issue. GitHub issues are the durable record of the spec — they are not
deleted, they are closed with a comment linking to the released
version, which preserves history.

**Actions**
1. If the squash-merge subject didn't auto-close the issue, do it
   manually: `gh issue close NNN --comment "Shipped in v<version>."`.
2. If a local PRD draft was written to disk during Section 2 (e.g., a
   scratch markdown file outside the repo), delete it. Do not delete
   anything tracked by git.
3. Confirm `gh issue view NNN` shows `state: CLOSED` with the close
   comment in place.

**Exit**: PRD issue closed; no orphan PRD files on disk.

## Section 9 — Refresh manual tests

**Goal**: `docs/manual-tests/...` reflects the new surface area for the
next manual pass.

Manual tests are scenario-based, not per-command (see
`docs/manual-tests/README.md`). They are also documentation, so
changes must be intentional, not auto-generated.

**Decision flow**
- Did this scope add or change a user-facing flow (new command,
  changed flag, changed default, new error class)? → Update or add a
  scenario.
- Cloud-only / Server-only? → Update only the affected directory
  (`cloud/`, `server/`, or `shared/`).
- Pure refactor with no observable change? → Skip; note "no
  manual-test changes" in the PR body.

**Actions**
1. Prefer extending the closest existing scenario (e.g.,
   `tag-lifecycle.md`) over creating a new file, unless the scope is a
   genuinely new noun.
2. Keep scenarios coherent (login → action → cleanup), not exhaustive.
3. If a new file is added, link it from `docs/manual-tests/README.md`.

**Exit**: any new user-facing flow has a scenario covering it; `git
status` is clean.

## Section 10 — Compact the working session

**Goal**: drop transient context now that the iteration is durable in
`main`, the release, the closed PRD, and updated docs.

If the iteration ran inside a long agent session (Claude Code, Cursor,
etc.), use the agent's compact / clear command (`/compact` on Claude
Code) so the next iteration starts with clean context. Everything
important survives in: the merged PR, the release notes, `BACKLOG.md`,
and `docs/manual-tests/`. Anything that wouldn't survive a compact
either belongs in those durable places or didn't matter.

For a plain shell session, just close the buffer.

**Exit**: session compacted; ready for the next iteration.

## Anti-patterns (across all sections)

- **Skipping pre-merge-check** because "it's a small change". The
  squash-merge gotcha and `dist/` artifacts have hit this repo before.
- **Editing `CHANGELOG.md`** by hand. release-please owns it.
- **Pushing to `main`** directly. Always a feature branch + PR.
- **Marking a backlog row ✅** before the release PR has merged. Until
  it's published, it's not shipped.
- **Treating `--no-verify` or `git push --force` as escape hatches.**
  Diagnose the underlying failure instead.
