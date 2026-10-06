# Parallel mode (multi-scope iteration)

**Goal**: ship 2–4 disjoint scopes in one wall-clock iteration with one
release, without compromising the discipline of the serial loop.

This document describes deltas from the main procedure (sections 0–10 in
[`README.md`](README.md)). Everything not mentioned here behaves identically.

---

## When to use

- Two to four open scopes that touch **different** command surfaces and
  share **no** invariants (e.g. `CTX` + `SR` + `GHP/pr-reopen`, or
  three `RV1`–`RV6` sub-PRs).
- More than four → diminishing returns; conflict resolution starts to
  outweigh parallelism.
- One large epic that splits naturally into independent sub-PRs (e.g.
  `RV` → `RV1`–`RV6`) → use parallel mode on the sub-PRs.

## When NOT to use

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

---

## Section 1 delta — Pick *N* scopes

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
- `docs/backlog/BACKLOG.md` (row removals) + `docs/backlog/SHIPPED.md` (entries appended; see [SHIPPED.md](../../backlog/SHIPPED.md) for the format)

These conflict in every parallel run. Resolution rule for **all** of
them is **union — both sides remain.** Reviewers and merging agents
treat anything other than union as a finding.

## Section 2 delta — One PRD per scope, in parallel

Create the *N* PRDs back-to-back, each as its own GitHub issue. Capture
all *N* issue numbers up front; each subagent references its own.

## Section 3 delta — N worktrees, one per scope

§3 already requires a worktree for any scope. Parallel mode just runs N of
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

### Subagent token discipline (REQUIRED for parallel mode)

The orchestrator's prompt to each subagent **MUST** include all three:

1. **Shared primer reference.** Pass `docs/agent-primer.md` as required
   reading in lieu of re-listing the architecture; the subagent reads
   the primer once instead of re-discovering interfaces, factory shape,
   error vocabulary, and exemplar files.

2. **Model tier guidance.** Match the agent to the load:

   | Agent role | Model tier | Why |
   |---|---|---|
   | Orchestrator | judgment-heavy | Conflict resolution, batch sequencing, halt-point judgement |
   | Implementer | code-generation (judgment-heavy only if scope is genuinely complex) | Generates code; needs solid reasoning but not full judgment-heavy overhead |
   | Reviewer | mechanical | Reads a small diff against the codebase; full judgment-heavy is overkill |
   | Fix agent | code-generation | Targeted TDD changes against an existing branch |

   Halve the reviewer cost for the same outcome. If a mechanical reviewer
   surfaces a vague finding, the orchestrator can re-dispatch that
   single finding to a code-generation/judgment-heavy reviewer.

3. **Compressed-output mode (if your agent supports it).** Subagents MUST
   invoke their ultra-compressed mode (e.g. `caveman` skill for Claude) for
   the duration of the parallel run. Final reports stay correct and complete;
   conversational filler drops. Other tooling (Codex, Cursor, Aider) should
   set the equivalent "concise" or "compact" output flag. This isn't
   aesthetic — it's the single biggest token lever in parallel mode
   (≈75% reduction measured in practice).

## Section 5–6 delta — Pre-merge gate runs *N* times, halt point batches

Each PR runs its own pre-merge-check. Then surface **one** halt point
covering the whole batch:

> *"PRs ready: #X #Y #Z. CI green on all. Reviews applied on all.
> Proposed merge order: smallest blast radius first. Confirm batch
> merge."*

This replaces *N* per-PR halt points. The author still has the
opportunity to demand a per-PR review, but the default is a single
confirm.

## Section 7 delta — Sequential merge with rolling rebase

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

## Section 8 delta — Close *N* PRDs

Squash-merge auto-closes each PRD via `Closes #NNN` in its body.
Confirm all *N* are closed in one `gh issue view` round.

---

## Anti-patterns specific to parallel mode

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
