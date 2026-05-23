# Iteration cycle

End-to-end loop for shipping one scope: pick → spec → build → ship → tidy.

This doc is the single source of truth for how a scope moves through the
repo. Tooling wrappers (Claude skill, IDE snippets, etc.) defer to it.
Human contributors and any agent (Codex, Cursor, Aider, Claude) follow
the same sections.

For parallel-mode (multi-scope iteration) see [`parallel-mode.md`](parallel-mode.md).
For autonomous-mode behavioral deltas see [`autonomous.md`](autonomous.md).
Declarative reference tables (halt routing, outcome enum, cadence) live in
[`quickref.md`](quickref.md). Script interface catalog lives in [`scripts.md`](scripts.md).

> **Cross-document reference convention.** References from this doc to
> other docs use anchor names ("the pre-merge gate", "the doc-sync
> table", "the parallel-mode section") rather than section numbers.
> Internal references within the same doc may use section numbers.
> This prevents a renumbering in one doc from silently breaking links
> in another.

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
1. **Run the preflight script**:
   ```bash
   bash auto-iter/scripts/preflight.sh
   ```
   It emits a single JSON object with `clean`, `branch`, `on_main`,
   `ahead`, `behind`, `open_prs`, and `findings`. Exit code is non-zero
   when any finding is halt-class (dirty tree, off-main, behind
   origin). See [`docs/workflows/iteration-cycle/scripts.md`](scripts.md)
   for the full contract.
2. **If exit was non-zero**, surface the `findings` array as a single
   inventory halt and ask the author what to keep / stash / commit /
   close — **before** picking the scope. For parallel mode
   specifically: cross-check that no candidate scope's expected files
   overlap with files touched in any open PR (`gh pr diff <N>
   --name-only`). A surprise stale branch holding a registry file is
   far cheaper to address up front than during a rolling rebase.
3. Read `BACKLOG.md` headings so the rest of the loop has accurate
   context.
4. **Smell scan** (≤60 seconds). Surface any structural debt the recent
   ship loop has accumulated, so the next mode pick is informed rather
   than blind:

   ```bash
   # Duplicate test-helper packages — should be one shared package
   find pkg/cmd -path '*/internal/cmdtest/cmdtest.go' | wc -l

   # N-way capability switches in cmd layer (≥9 = ≥3 files of triplets)
   grep -rEo 'As[A-Z][a-zA-Z]+Client\(\)' pkg/cmd | wc -l

   # Translation-table pairs in api/backend/types.go
   grep -cE '^func (To|.*ToCLI)' api/backend/types.go

   # Unused exported interfaces (cheap heuristic — symbol declared in
   # api/backend but referenced only in its own _test.go file)
   ```

   If two or more counters jump materially since the last iteration, the
   next §1 mode pick should favour `architecture` over `iteration`.

**Exit**: clean tree on `main`, up to date with `origin/main`, workspace
inventory + smell scan acknowledged.

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
3. **Nothing emerges?** Switch to **brainstorm mode** (judgment-heavy
   model subagent). Brainstorm runs autonomously — no phone halt — and
   emits 3–4 candidate BACKLOG rows. Every emitted row must satisfy
   **all** of the following or be dropped before write:
   1. **No-overlap** — does not duplicate any ✅ row or Functionality
      Map entry. _(This is the rule that retroactively killed
      PR-TEMPLATE; without it brainstorm hallucinates plausible-but-
      redundant scopes.)_
   2. **Backend declared** — `Cloud` / `Server` / `Both`. If `Both`,
      both endpoints named.
   3. **Shape match** — declares which canonical pattern (`List*` via
      `paging.Collect`, write op with typed errors, MCP triplet). New
      shapes belong in an architecture-audit cycle, not a brainstorm.

   Emit `step2_brainstorm` to `metrics.jsonl` with `subagent_tokens`,
   `rows_added`, `rows_dropped_by_overlap`,
   `rows_dropped_by_feasibility`. Empty brainstorm runs (0 rows after
   rule application) count toward the 3-empty shutdown counter.
4. **Bundle check** — if the chosen scope's estimated diff is **<30 files**
   AND the next 🔲 row in `BACKLOG.md` is also **<30 files** AND the two
   are deeply disjoint, propose a **two-scope bundle** with a combined
   cap of **<60 files of meaningful (non-refactor) change**. Author
   confirms before proceeding.

   **Estimation cheatsheet** (use BACKLOG scope-detail rows):

   | BACKLOG element | Predicted file count |
   |---|---|
   | 1 new interface in `api/backend/client_*.go` | 1 |
   | Cloud impl + tests | 2 (skip if Server-only) |
   | Server impl + tests | 2 (skip if Cloud-only) |
   | N new commands in `pkg/cmd/<group>/` | 2N (impl + `_test.go`) |
   | MCP tools (additive in existing files) | 0 new (3 updated) |
   | Doc updates (README, BACKLOG, SKILL.md, manual-tests) | 3–4 updated |
   | New transport for new REST namespace | +2 |

   **Disjointness check (all must hold for a bundle to fire)**:
   - Different command surfaces (no shared `pkg/cmd/<group>/` tree).
   - No shared interface (neither extends the other's `client_*.go`).
   - No shared exemplar files referenced in implementation.
   - No competing manual-test scenarios.
   - Backend symmetry: both Cloud-only, both Server-only, OR both
     touch both backends.

   **Bundle anti-patterns** (BLOCKER — fall back to serial):
   - Both scopes extend the same interface.
   - Both scopes touch the same command tree at the same level.
   - One scope's design depends on the other's outcome (sequential, not
     parallel).

**Exit**: one scope (or one validated two-scope bundle) chosen, with
explicit author confirmation.

**Anti-patterns**
- Picking multiple unrelated scopes ad-hoc. Bundling is allowed **only**
  through the bundle-check rules above; everything else is one scope =
  one PR = one release.
- Inventing a scope without checking the backlog first.

## Section 2 — Write the PRD

**Goal**: a single PRD captured as a GitHub issue, scoped to the chosen
scope and only that scope.

**Actions**
1. Draft the PRD from current context — problem, in-scope, out-of-scope,
   API surface, test plan. (Automate this with your tool's PRD-drafting
   workflow if available.)
2. **Obsolescence check.** Before filing, scan `BACKLOG.md` for any 🔲
   row whose command surface would supersede or duplicate this one
   (e.g. `environment variable …` would have been superseded by a
   later `variable --scope deployment`). If found, either widen this
   scope to ship the unified form, or note in the PRD the older surface
   stays only as an alias and will be deprecated in the same PR. Two
   wired-in command trees doing the same job is a BLOCKER at §5.
3. **Expected files list.** The PRD MUST end with a `## Expected files`
   section listing every file the implementation is expected to touch
   (new + modified). This becomes the TDD subagent's contract in §3 —
   passing it to the agent up front saves the discovery round that
   ballooned cycle 9 to 47 minutes / 210k tokens. Use the §1 estimation
   cheatsheet as a starting point.

   **Files-touched budget.** If the expected list contains ≥50 entries,
   the PRD MUST include a `## Rolling refactor` section explaining
   *why* the cross-cutting blast radius is unavoidable (e.g. a new
   global flag, a transport-layer hardening that every command opts
   into). v1.35 OUT2 (91 files) is the cleanest exemplar; an accidental
   91-file PR without a stated rolling-refactor rationale is a BLOCKER
   at §5. The intent is to force the bundle-vs-split decision into the
   PRD where reviewers can see it, not bury it in the diff.
4. File it as a GitHub issue: `gh issue create --title "<scope>" --body
   "<prd>" --label prd`.
5. **Capture the issue number** — every later phase references it
   (`refs #NNN`, `Closes #NNN`).
6. Sanity-check the PRD against `BACKLOG.md` → "Architecture Contract
   (per scope)" and "Definition of Done". Those rows are non-negotiable
   and must appear in the PRD's checklist.

**Exit**: GitHub issue created, issue number captured, DoD checklist +
expected-files list embedded in the PRD.

## Section 3 — Implement (TDD)

**Goal**: green tests + green lint on a feature branch, in an isolated
worktree.

> **AUDIT_CONTINUE rule** — see [`docs/workflows/iteration-cycle/quickref.md` § AUDIT_CONTINUE](quickref.md#audit_continue--bundle-auditfix-when-the-finding-fits-in-one-pr).
> When the scope slug matches `*-AUDIT` and the audit produces exactly
> one finding that fits in ≤1 PR (≤200 LOC, ≤8 files, single
> subsystem), the TDD subagent continues into the implementation in the
> same cycle. Both BACKLOG rows flip in the same feat commit; the cycle
> emits `bundled:true` with two scope slugs in `scopes[]`. This is the
> answer to cycles 95 (UX-FLAG-AUDIT) + 96 (DEBUG-TRANSPORT-FLAG)
> shipping what should have been one cycle.

> **HARD STOP — worktree is mandatory.** Never branch off `main` inside
> the main checkout. Every iteration — single scope, bundle, parallel,
> fix, docs, chore — runs in its own worktree.
>
> ```bash
> # ❌ FORBIDDEN — pollutes the main checkout, traps in-flight work
> git checkout -b feat/<slug> origin/main
> git checkout -b fix/<slug> origin/main
>
> # ✅ REQUIRED — isolates the iteration, preserves the main checkout
> git worktree add -b feat/<slug> ../bitbottle-worktrees/<slug> origin/main
> git worktree add -b fix/<slug>  ../bitbottle-worktrees/<slug> origin/main
> ```
>
> If the orchestrator (human or agent) catches itself about to
> `git checkout -b` on `main`, that is a workflow violation — back out,
> create the worktree, and retry. PRD #372 surfaced this as the
> "Process bug" companion to four auth bugs.

**Worktree (always — even for a single scope)**
- `git worktree add -b feat/<short-slug> ../bitbottle-worktrees/<slug> origin/main`
  (or `fix/...` / `docs/...` / `chore/...` per `AGENTS.md`). All
  implementation, commits, and pushes happen **inside the worktree**.
  The main checkout stays clean — agents and humans alike never edit
  `main` directly.
- Removed in §10 (Compact). The `superpowers:using-git-worktrees` skill
  formalises the pattern; the `Task` tool also accepts
  `isolation: "worktree"` so spawned subagents get auto-cleaned worktrees
  without manual setup.
- One scope = one worktree = one branch = one PR. No drive-by refactors.
- For a bundled two-scope iteration (per §1 bundle check): one worktree,
  branch named for the combined slug — e.g. `feat/pr-reopen-and-checks`.
- For parallel mode (see [`parallel-mode.md`](parallel-mode.md)): N worktrees, one per scope,
  all created from the current `main`.

**Discipline**
- Red → green → refactor. Write a failing test first, make it pass with
  the smallest change, then refactor with the test as the safety net.
  One logical change per commit.
- Conventional Commit subjects: `feat(scope): ...`, `fix(scope): ...`,
  `docs(scope): ...`.

**Subagent brief (required when delegating TDD)**
- Pass the PRD's **Expected files** list verbatim. The subagent treats
  it as the touch boundary — additions are allowed, but if it needs to
  edit a file outside the list it must surface that as a finding before
  proceeding, not discover it silently. This caps the
  re-discovery cost that bloated cycle 9 (210k subagent tokens, 47 min)
  versus a typical 100–120k.
- Include the layer order below as a one-line ordering hint; do not
  re-list architecture. Point at `docs/agent-primer.md` instead.

**Layer order** (per `BACKLOG.md` Architecture Contract)
1. `api/backend/client.go` — new interface(s), or extend the composite
   `Client`. New types in `api/backend/types.go`.
2. `api/cloud/<domain>.go` + `_test.go`.
3. `api/server/<domain>.go` + `_test.go` — skip only if Cloud-only, and
   document why with `ErrUnsupportedOnHost`.
4. `pkg/cmd/<domain>/...` cobra commands. Every applicable command must
   support `--json` / `--jq` / `--hostname`.
5. `pkg/cmd/mcp/tools.go` registration + `handlers.go` method.

See [quickref.md](quickref.md) §Model tier per phase for dispatch guidance.

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
  Map" entries. **The flip lands in the same commit as the feat work,
  NOT in a follow-up `chore: mark X shipped in BACKLOG` PR.** Across
  cycles 77–86 in the May-17 autonomous stream, the post-merge chore-PR
  pattern produced 8 extra PRs and 4 duplicate commits on `main`
  (MCP-VALIDATION, JSON-STABILITY, ERR-EMPTY-400, and BACKEND-TYPE-STRICT
  each landed twice). One feat commit that touches both code and BACKLOG
  is the correct shape.
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

Run `docs/workflows/pre-merge-check.md` end-to-end (branch hygiene,
Conventional Commits, no build artifacts in `dist/`, `make lint` +
`go test`, the doc-sync table, the **design-judge** scan, release-please
boundaries, and the secret scan).

The **design-judge checklist** explicitly includes a dead-branch /
tautological-assignment scan — semantic no-ops like
`if x == "Y" { x = "Y" }` are not caught by `staticcheck`, `gocritic`,
or `gosimple` and have shipped in past cycles (see
`api/cloud/pr_participants.go` in PR #292 / cyc 55, fixed in this PR's
parent chore). When reviewing a diff, scan every conditional whose
then-body assigns a literal that the condition already pins.

**Order matters.** Design-judge (pre-merge-check §6) runs **locally on
the feature branch before §6 push**, not after the PR is open and CI is
green. Discovering BLOCKERs after `pr_open` forces a fix-agent round +
a second CI cycle (≈5–10 min wall + ≈80k extra subagent tokens, measured
across cycles 8 and 9). If you find yourself running design-judge in
§6, the gate ran out of order — finish it here.

Do **not** proceed past this section if any check fails. Fix the
underlying issue and re-run; never bypass.

**Exit**: pre-merge-check returns green, including design-judge BLOCKERs
resolved on the local branch.

## Section 6 — Open the PR

**Goal**: PR open against `main`, titled so release-please picks it up.

**Actions**
1. `git push -u origin <branch>`.
2. `gh pr create --base main` with:
   - **Title** — Conventional Commits prefix (`feat(scope): ...`,
     `fix(scope): ...`). Squash-merge uses the PR title as the commit
     subject; getting this wrong forces a follow-up empty commit (see
     #48 → #49 → #50).
   - **Body** — short summary and a test plan checklist mirroring the
     manual tests touched in Section 9. **The body's last line MUST be
     `Closes #<PRD-issue-number>`** so GitHub auto-closes the PRD on
     squash-merge. Verify this is present before running `gh pr create`
     (or `gh pr edit` to add post-hoc if forgotten). Section 8 verifies
     the PRD actually closed after merge — PRDs #448, #451, #454 stayed
     orphan-open in v1.90.0–v1.92.0 because their PR bodies omitted this
     keyword.
3. Wait for CI to go green. Do not request review or push more commits
   while CI is running unless something is actually broken. **Running
   design-judge here is a workflow violation** — it belongs in §5,
   before push. Catching it now means paying for a fix-agent round and
   a second CI cycle that the local gate would have prevented.

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
1. **Verify the PRD is closed.** Run `gh issue view NNN --json state | jq -r .state`.
   If `state == OPEN`, the PR body was missing the `Closes #NNN` keyword
   (see [Section 6](#section-6--open-the-pr)) — close manually with a
   comment linking the merge commit, and (autonomous mode) record
   `dispatch_violation: true` on the cycle's `step8_close_prd` metric so
   the failure mode stays visible in post-cycle analysis.
2. If a local PRD draft was written to disk during Section 2 (e.g., a
   scratch markdown file outside the repo), delete it. Do not delete
   anything tracked by git.
3. Confirm `gh issue view NNN` shows `state: CLOSED` with the close
   comment in place.

**Why this verification step exists**: PRDs #448 (REPO-EDIT), #451
(PIPELINE-STOP), #454 (SNIPPETS) all shipped in v1.90.0–v1.92.0 but stayed
open for 2+ days because their PR bodies omitted `Closes #NNN`. GitHub's
auto-close fires only when the keyword is in the PR body or squash-merge
commit body. The verification here catches the missed keyword before it
becomes an orphan-shipped PRD.

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

## Section 10 — Cleanup + compact the working session

**Goal**: remove the worktree and drop transient context now that the
iteration is durable in `main`, the release, the closed PRD, and
updated docs.

**Actions**
1. **Remove the worktree** (and any peers from a parallel batch):

   ```bash
   git worktree remove ../bitbottle-worktrees/<slug>
   ```

   The squash-merged commits are already on `main`; the local stale
   branch can be left for a periodic sweep, or deleted if you trust the
   merge (`git branch -d feat/<slug>` will refuse if the branch isn't
   reachable from `main`, which is normal for squash-merge — `-D` works
   only if Safety Net allows it).

2. **Cycle is logged automatically** — `auto-iter/scripts/log-cycle.sh`
   appends one line to `.claude/auto-iter/cycles.jsonl` at cycle end
   (per-step lines live in `metrics.jsonl`, written by `metric.sh` at
   each step). Both are gitignored runtime state and the authoritative
   record of "is the pipeline trending faster or slower." Inspect with:

   ```bash
   jq -s 'sort_by(.cycle) | .[-10:]' .claude/auto-iter/cycles.jsonl
   ```

   Earlier versions of this doc instructed manually appending to
   `auto-iter/metrics.csv`. That ledger was abandoned at cycle 14, became
   misleading, and was removed in `chore(auto-iter): loop hardening`.

3. **Compact** — if the iteration ran inside a long agent session,
   use the agent's compact / clear command so the next iteration starts
   with clean context. Everything important survives in: the merged PR,
   the release notes, `BACKLOG.md`, and `docs/manual-tests/`. Anything
   that wouldn't survive a compact either belongs in those durable places
   or didn't matter.

   For a plain shell session, just close the buffer.

**Exit**: worktree removed, session compacted; ready for the next
iteration.

## Anti-patterns (across all sections)

- **Skipping pre-merge-check** because "it's a small change". The
  squash-merge gotcha and `dist/` artifacts have hit this repo before.
- **Editing `CHANGELOG.md`** by hand. release-please owns it.
- **Pushing to `main`** directly. Always a feature branch + PR.
- **Marking a backlog row ✅** before the release PR has merged. Until
  it's published, it's not shipped.
- **Treating `--no-verify` or `git push --force` as escape hatches.**
  Diagnose the underlying failure instead.
