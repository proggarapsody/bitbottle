# Iteration cycle

End-to-end loop for shipping one scope: pick → spec → build → ship → tidy.

This doc is the single source of truth for how a scope moves through the
repo. Tooling wrappers (Claude skill at
`.claude/skills/iteration-cycle/`, IDE snippets, etc.) defer to it.
Human contributors and any agent (Codex, Cursor, Aider, Claude) follow
the same sections.

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
1. Confirm `pwd` is the bitbottle repo root.
2. `git status` is clean (no uncommitted changes, no stale staged files).
3. `git rev-parse --abbrev-ref HEAD` is `main`.
4. `git pull --ff-only` succeeds. If not, surface the divergence and
   stop.
5. **Inventory the workspace** — surface everything that could bite
   later in one halt, not trickle-discovered mid-iteration:

   ```bash
   # Open PRs the author owns — any blocker for the planned scopes?
   gh pr list --author @me --state open \
     --json number,title,headRefName,isDraft,updatedAt

   # Local branches whose remote is gone or that have drifted from main
   git for-each-ref --format='%(refname:short) %(upstream:track)' refs/heads/ \
     | grep -E '\[gone\]|\[behind' || true

   # Untracked paths the author may not have noticed
   git status --porcelain | awk '/^\?\?/{print $2}'
   ```

   If any of these are non-empty, surface them as a single inventory
   and ask the author what to keep / stash / commit / close —
   **before** picking the scope. For parallel mode specifically: cross-
   check that no candidate scope's expected files overlap with files
   touched in any open PR (`gh pr diff <N> --name-only`). A surprise
   PR-#133 or stale branch holding a registry file is far cheaper to
   address up front than during a rolling rebase.
6. Read `BACKLOG.md` headings so the rest of the loop has accurate
   context.
7. **Smell scan** (≤60 seconds). Surface any structural debt the recent
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
3. **Nothing emerges?** Switch to **brainstorm mode** (Opus subagent).
   Brainstorm runs autonomously — no phone halt — and emits 3–4
   candidate BACKLOG rows. Every emitted row must satisfy **all** of
   the following or be dropped before write:
   1. **No-overlap** — does not duplicate any ✅ row or Functionality
      Map entry. _(This is the rule that retroactively killed
      PR-TEMPLATE; without it brainstorm hallucinates plausible-but-
      redundant scopes.)_
   2. **Backend declared** — `Cloud` / `Server` / `Both`. If `Both`,
      both endpoints named.
   3. **Shape match** — declares which canonical pattern (`List*` via
      `paging.Collect`, write op with typed errors, MCP triplet). New
      shapes belong in an architecture-audit cycle, not a brainstorm.
   4. **Pointer estimate** — 1 / 2 / 3 / 5. Anything >3 must be
      decomposed.

   Emit `step1_brainstorm` to `metrics.jsonl` with `subagent_tokens`,
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
   API surface, test plan. (The `to-prd` skill on Claude Code automates
   this; on other tooling, write the markdown by hand or via your
   equivalent.)
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

**Worktree (always — even for a single scope)**
- `git worktree add -b feat/<short-slug> ../bitbottle-worktrees/<slug> main`
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
- For parallel mode (§11): N worktrees, one per scope, all created from
  the current `main`.

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
   - **Body** — link the PRD issue (`Closes #NNN`), short summary, and a
     test plan checklist mirroring the manual tests touched in Section
     9.
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

2. **Append to `docs/auto-iter/metrics.csv`** — one row per cycle:

   ```
   cycle,version,scope,pr,merged_at_utc,wall_minutes,ci_seconds,loc_added,loc_deleted,files_touched,subagent_tokens_k
   ```

   Pull from `gh release view`, `gh pr view <N> --json
   mergedAt,additions,deletions,changedFiles`, and `gh run view <run-id>
   --json createdAt,updatedAt`. `subagent_tokens_k` is optional (leave
   empty if you weren't tracking it). Commit the row on the chore branch
   that ships the next iteration's backlog updates — never on the
   feature branch. The series is the source of truth for "is the
   pipeline trending faster or slower"; we stopped relying on memory
   after the v1.29–v1.31 post-mortem.

3. **Compact** — if the iteration ran inside a long agent session
   (Claude Code, Cursor, etc.), use the agent's compact / clear command
   (`/compact` on Claude Code) so the next iteration starts with clean
   context. Everything important survives in: the merged PR, the
   release notes, `BACKLOG.md`, and `docs/manual-tests/`. Anything that
   wouldn't survive a compact either belongs in those durable places or
   didn't matter.

   For a plain shell session, just close the buffer.

**Exit**: worktree removed, session compacted; ready for the next
iteration.

## Section 11 — Parallel mode (multi-scope iteration)

**Goal**: ship 2–4 disjoint scopes in one wall-clock iteration with one
release, without compromising the discipline of the serial loop.

This section describes deltas from sections 0–10. Everything not
mentioned here behaves identically.

### When to use

- Two to four open scopes that touch **different** command surfaces and
  share **no** invariants (e.g. `CTX` + `SR` + `GHP/pr-reopen`, or
  three `RV1`–`RV6` sub-PRs).
- More than four → diminishing returns; conflict resolution starts to
  outweigh parallelism.
- One large epic that splits naturally into independent sub-PRs (e.g.
  `RV` → `RV1`–`RV6`) → use parallel mode on the sub-PRs.

### When NOT to use

- Scopes that share an invariant or a refactor boundary (e.g. two
  scopes both rewriting the factory).
- Scopes where one's design depends on another's outcome.
- Cross-cutting work (`EX`, `T`) — those touch every command and must
  serialize.
- **Ambiguous phrasing.** If the author says "all these scopes in one
  iteration", "ship these *N* scopes together", or similar wording that
  doesn't clearly demand parallelism, default to **serial** and ask for
  explicit confirmation before launching parallel agents. Parallel mode
  burns ~5–10× the tokens of a serial cycle; misfires are expensive.

### Section 1 delta — Pick *N* scopes

After the author confirms the scope list, list the conflict zones the
author should expect. The repo's "registry files" (one-liner additions
that every scope makes) are:

- `pkg/cmd/root/root.go` (subcommand registration)
- `pkg/cmd/mcp/tools.go` (MCP tool registration)
- `pkg/cmd/mcp/handlers.go` + `handlers_test.go` (MCP handlers + tests)
- `api/backend/client.go` (interface declarations + composite `Client`)
- `api/backend/types.go` (new domain types)
- `api/backend/host_unsupported_test.go` (capability-gating coverage)
- `README.md` (new sections, additive)
- `BACKLOG.md` (status flips)

These conflict in every parallel run. Resolution rule for **all** of
them is **union — both sides remain.** Reviewers and merging agents
treat anything other than union as a finding.

### Section 2 delta — One PRD per scope, in parallel

Create the *N* PRDs back-to-back, each as its own GitHub issue. Capture
all *N* issue numbers up front; each subagent references its own.

### Section 3 delta — N worktrees, one per scope

§3 already requires a worktree for any scope (since this doc was
revised — see §3 "Worktree (always)"). Parallel mode just runs N of
them simultaneously:

```bash
git worktree add -b feat/<slug-1> ../bitbottle-worktrees/<slug-1> main
git worktree add -b feat/<slug-2> ../bitbottle-worktrees/<slug-2> main
git worktree add -b feat/<slug-3> ../bitbottle-worktrees/<slug-3> main
```

Each implementer subagent works only inside its assigned worktree.
File ownership is enforced structurally: a subagent that edits another
worktree's tree is misbehaving. The `Task` tool's
`isolation: "worktree"` parameter automates this for orchestrator-driven
dispatch; Conductor's "track" model applies it natively per track.

#### Subagent token discipline (REQUIRED for parallel mode)

The orchestrator's prompt to each subagent **MUST** include all three:

1. **Shared primer reference.** Pass `docs/agent-primer.md` as required
   reading in lieu of re-listing the architecture; the subagent reads
   the primer once instead of re-discovering interfaces, factory shape,
   error vocabulary, and exemplar files.

2. **Model tier guidance.** Match the agent to the load:

   | Agent role | Model tier | Why |
   |---|---|---|
   | Orchestrator (you) | Opus | Conflict resolution, batch sequencing, halt-point judgement |
   | Implementer | Sonnet (Opus only if scope is genuinely complex) | Generates code; needs solid reasoning but not full Opus |
   | Reviewer | Haiku | Reads a small diff against the codebase; full Opus is overkill |
   | Fix agent | Sonnet | Targeted TDD changes against an existing branch |

   Halve the reviewer cost for the same outcome. If a Haiku reviewer
   surfaces a vague finding, the orchestrator can re-dispatch that
   single finding to a Sonnet/Opus reviewer.

3. **Compressed-output mode (if your agent supports it).** Claude
   subagents MUST invoke their `caveman` skill (or equivalent
   ultra-compressed mode) for the duration of the parallel run. Final
   reports stay correct and complete; conversational filler drops.
   Other tooling (Codex, Cursor, Aider) should set the equivalent
   "concise" or "compact" output flag. This isn't aesthetic — it's the
   single biggest token lever in parallel mode (≈75% reduction
   measured in practice).

### Section 5–6 delta — Pre-merge gate runs *N* times, halt point batches

Each PR runs its own pre-merge-check. Then surface **one** halt point
covering the whole batch:

> *"PRs ready: #X #Y #Z. CI green on all. Reviews applied on all.
> Proposed merge order: smallest blast radius first. Confirm batch
> merge."*

This replaces *N* per-PR halt points. The author still has the
opportunity to demand a per-PR review, but the default is a single
confirm.

### Section 7 delta — Sequential merge with rolling rebase

PRs cannot squash-merge in parallel — `main` advances after each merge
and the next PR's working tree must include the merged change.

1. **Merge order**: smallest blast radius first (fewest registry-file
   touches). In practice that means Cloud-only-optional-interface
   scopes before both-backend scopes before composite-`Client`-extending
   scopes.

2. **Rolling rebase**: after each merge, `cd` into each remaining
   worktree and `git merge origin/main --no-ff`. Resolve any registry-
   file conflicts via the union rule above. Push. CI re-runs (~75s).
   Merge when green.

3. The release-please PR halt point (irreversible) stays per-iteration
   — one release for the whole batch, with all *N* `feat:` commits
   bundled in one `CHANGELOG.md` block.

### Section 8 delta — Close *N* PRDs

Squash-merge auto-closes each PRD via `Closes #NNN` in its body.
Confirm all *N* are closed in one `gh issue view` round.

### Anti-patterns specific to parallel mode

- **Forgetting the rolling rebase.** PR-2 will hit conflicts on merge
  if its branch is still based on the pre-PR-1 main.
- **Halting per PR instead of batching.** Wastes round-trips and
  context; the user already authorised the batch.
- **Treating registry-file conflicts as bugs.** They are structural.
  Resolve by union and move on.
- **Over-paralleling.** Five+ scopes is usually slower end-to-end than
  serial because rebase loops + CI cycles dominate.
- **Skipping the primer / model-tier / caveman guidance.** All three
  exist because they were measured savings, not aesthetics.

**Exit**: *N* PRs merged, one release published, *N* PRDs closed, one
batch halt-point used.

## Anti-patterns (across all sections)

- **Skipping pre-merge-check** because "it's a small change". The
  squash-merge gotcha and `dist/` artifacts have hit this repo before.
- **Editing `CHANGELOG.md`** by hand. release-please owns it.
- **Pushing to `main`** directly. Always a feature branch + PR.
- **Marking a backlog row ✅** before the release PR has merged. Until
  it's published, it's not shipped.
- **Treating `--no-verify` or `git push --force` as escape hatches.**
  Diagnose the underlying failure instead.
