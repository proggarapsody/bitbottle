# Autonomous-loop quickref

Declarative reference for the bitbottle autonomous iteration loop. Read
this when designing or troubleshooting `/auto-iter`. Procedural details
live in the gitignored `.claude/commands/auto-iter.md`; canonical
workflow lives in [`docs/workflows/iteration-cycle.md`](../workflows/iteration-cycle.md).

> **Mantra**: one `/auto-iter` invocation = one cycle = exits at end.
> `/loop` provides the cadence between cycles. Halt for anything
> irreversible or judgment-heavy; auto-confirm anything mechanical.

---

## Architecture

```
laptop (caffeinated, Claude desktop with Remote Control paired)
  └── /loop 30m /auto-iter
        └── /auto-iter (one cycle)
              ├── §0a re-entry lock check     (.claude/auto-iter/.lock)
              ├── §0  preflight + workspace inventory
              ├── §1  pick mode + scope (+ bundle-check)
              ├── §2  execute mode (iteration | architecture | brainstorm | stop)
              ├── §3  halt routing (auto-confirm safe; phone for risky)
              ├── §4  append to cycles.jsonl + metrics.jsonl
              └── §5  release lock, exit
```

Three sibling files in `.claude/auto-iter/` (all gitignored):

| File | Purpose |
|---|---|
| `cycles.jsonl` | One line per cycle, high-level outcome. Cycle counter is global, monotonic. |
| `metrics.jsonl` | One line per **step** within a cycle. Wall-clock + tokens. |
| `.lock` | Sentinel — present iff a cycle is running. Recent (<1h) lock → skip; stale (>1h) lock → assume crashed predecessor, clear. |

---

## Model tier per phase

Default to **Sonnet**. Escalate to **Opus** only for genuinely judgment-heavy phases.

| Phase | Model | Why |
|---|---|---|
| §0a re-entry lock | Sonnet | Mechanical file check |
| §0 preflight + workspace inventory | Sonnet | Shell-driven, deterministic |
| §0 open-PR overlap (decision halt) | Sonnet | Routes to phone for judgment; agent itself just frames the question |
| §1 mode pick + scope pick | Sonnet | Algorithm-driven; BACKLOG-driven scope-pick is deterministic |
| §1 bundle-check | Sonnet | Algorithm-driven |
| §2 PRD drafting | Sonnet | Mechanical fill from BACKLOG scope-detail |
| §2 worktree creation | Sonnet | Shell command |
| §2 TDD implementation | **Sonnet** (subagent) | Code generation. Dispatch via `Task` tool with `isolation: "worktree"` — keeps orchestrator context light. Opus only if scope is genuinely complex (rare). |
| §2 doc sync | Sonnet | Mechanical doc updates per §5 doc-sync table |
| §2 pre-merge gate | Sonnet | Reads CI status, runs grep checks |
| §6 design-judge | **Opus** (subagent) | Read-only review against `TASTE.md` + `ARCHITECTURE.md` + diff. Returns findings; runs in parallel with CI. |
| §2 PR open + CI wait | Sonnet | Polling |
| §2 halts (HALT 1, HALT 2) | Sonnet | Frame the question, wait for tap |
| §2 fix-after-CI-red | **Sonnet** (subagent) | Targeted code change against existing branch |
| §7 merge feature + release PRs | Sonnet | `gh pr merge` mechanical |
| §7 release publish wait | Sonnet | Polling `gh release view` + `npm view` |
| §8 close PRD | Sonnet | `gh issue` mechanical |
| §9 manual-test refresh | Sonnet | Per §9 decision flow |
| **§2 brainstorm Q&A** (when BACKLOG empty) | **Opus** | Open-ended judgment, multi-turn dialogue |
| **§2 architecture audit** (every 5th cycle) | **Opus** | Architectural reasoning, pattern recognition |
| §5 release lock | Sonnet | File removal |

Two phases legitimately need Opus: brainstorm and architecture-audit. Everything else is mechanical or a routed halt where Sonnet is fine.

---

## Halt routing

Two categories — **mid-cycle** (pause-and-continue) and **chain-breaking** (exit).

### Mid-cycle halts (every successful `feat:` cycle has 2 of these)

| Halt | Phone format | Required reply |
|---|---|---|
| HALT 1 — Feature merge | `✅ PR #N <scope> ready. ship?` | tap `ship` / `hold` |
| HALT 2 — Release merge | `🚀 v<X.Y.Z> ready. ship?` | tap `ship` / `hold` |

Tap → cycle resumes. Tap "hold" → exit, cycle outcome `halt_feature_merge` / `halt_release`.

### Chain-breaking halts (rare — exit and wait for human)

| Halt | Phone format | Reply |
|---|---|---|
| Workspace must-resolve | `⚠️ workspace requires manual cleanup: <details>` | exit cleanly |
| Open-PR overlap | `⚠️ PR #N overlaps scope. resolve / skip / close?` | tap `resolve` / `skip` / `close` |
| Pre-merge BLOCKER | `❌ PR #N: <finding>` | tap `fix` / `override <reason>` |
| CI red | `❌ PR #N CI red: <log_url>` | tap `retry` / `fix` / `abort` |
| Merge conflict (during resolve) | `❌ PR #N merge conflict beyond union rule: <files>` | exit cleanly |
| Brainstorm question (when BACKLOG empty) | `💭 <question>` | free-form reply |
| Stop confirmation (BACKLOG empty + 3 empty brainstorms) | `🏁 confirm shutdown` | tap `confirm` / `continue` |

### Halt protocol rules

- **No diff in halt messages.** Reference PR URL, not embedded diff. Halt messages stay <500 chars so phone push notifications display cleanly.
- **Halt-response timeout**: 4 hours. After that, log `halt_no_response`, exit. Don't poll forever.
- **Auto-confirm**: scope-pick (BACKLOG-driven), bundle-check (algorithm-driven), workspace-clean preflight, mechanical doc-sync, secret-leak scan, build-artifacts scan, lint+test-via-CI, PRD close confirmation, manual-test refresh decision, worktree removal.

---

## Cycle log schema (`cycles.jsonl`)

One line per cycle, append-only:

```jsonl
{"ts":"<ISO>","cycle":<N>,"mode":"iteration|architecture|brainstorm|stop","scopes":[<slugs>],"prs":[<numbers>],"release":"<v1.X.Y or null>","duration_min":<int>,"tokens":<int or null>,"outcome":"<see below>","bundled":<bool>,"notes":[<strings>]}
```

`cycle` is **global, monotonic** — read max from log, increment by 1.

### Outcome enum

| Value | Meaning |
|---|---|
| `shipped` | Feature PR merged, release published (or release skipped for `refactor:`/`chore:`) |
| `audit_no_findings` | Architecture mode, no findings → no PR opened |
| `brainstorm_added_<N>` | Brainstorm yielded N new BACKLOG rows |
| `shutdown_confirmed` | Stop mode confirmed |
| `skipped_overlapping_pr` | Step 0B "skip" reply |
| `closed_overlapping_pr` | Step 0B "close" reply |
| `resolved_overlapping_pr` | Step 0B "resolve" reply led to a successful ship |
| `halt_workspace_must_resolve` | Dirty tree / not-on-main / divergence |
| `halt_merge_conflict` | Step 0B "resolve" hit a non-union conflict |
| `halt_pre_merge_blocker` | Pre-merge gate found a BLOCKER finding |
| `halt_ci_red` | CI failed |
| `halt_feature_merge` | User replied non-ship to feature halt |
| `halt_release` | User replied non-ship to release halt |
| `halt_no_response` | Phone halt timed out (4 hr default) |
| `halt_cycle_timeout` | Cycle exceeded wall-clock cap |
| `skip_in_progress` | Step 0a found a recent lock; another `/auto-iter` is running. **Do NOT increment cycle counter.** |

---

## Metrics log schema (`metrics.jsonl`)

One line per **step** within a cycle, append-only:

```jsonl
{"cycle":<N>,"step":"<name>","ts":"<ISO start>","duration_ms":<int>,"...optional...":"..."}
```

**Mandatory**: `cycle`, `step`, `ts`, `duration_ms`. Optional fields per step:

| Step name | Optional fields |
|---|---|
| `step0a_lock` | `lock_age_min` |
| `step0_preflight` | `inventory_findings_count` |
| `step0_open_pr_overlap` | `pr`, `decision` |
| `step1_mode_pick` | `mode`, `scope` (or `scopes` if bundled) |
| `step2_prd` | `prd_issue` |
| `step2_worktree` | `worktree_path`, `branch` |
| `step2_tdd` | `subagent_tokens`, `commits_count` |
| `step2_pre_merge_gate` | `findings_count` |
| `step2_pr_open` | `pr` |
| `step2_ci_wait` | `ci_minutes` |
| `step2_halt1` | `halt_response` |
| `step2_release_pr_wait` | `release_pr` |
| `step2_halt2` | `halt_response` |
| `step2_release_publish` | `version`, `npm_published` |
| `step5_cleanup` | (none) |

### Token-accounting honesty

- `subagent_tokens`: ✅ reliable (Task tool returns it).
- `orchestrator_tokens` (at cycle level in `cycles.jsonl`): ⚠️ best-effort, often `null`.
- Wall-clock `duration_ms`: ✅ always reliable. The primary performance signal.

### Sample `jq` queries

```bash
# Total wall-clock for cycle 47 (seconds)
jq -s 'map(select(.cycle==47)) | (map(.duration_ms) | add) / 1000' \
  .claude/auto-iter/metrics.jsonl

# Average TDD duration across all cycles
jq -s 'map(select(.step=="step2_tdd")) | (map(.duration_ms) | add / length) / 1000' \
  .claude/auto-iter/metrics.jsonl

# Subagent token cost per cycle
jq -s 'group_by(.cycle) | map({cycle: .[0].cycle, tokens: (map(.subagent_tokens // 0) | add)})' \
  .claude/auto-iter/metrics.jsonl

# Top 10 slowest steps
jq -s 'sort_by(-.duration_ms) | .[:10] | map({cycle, step, duration_s: (.duration_ms/1000)})' \
  .claude/auto-iter/metrics.jsonl

# Halt-response time distribution
jq -s 'map(select(.step | startswith("step2_halt")) | .duration_ms / 1000)' \
  .claude/auto-iter/metrics.jsonl
```

---

## Anti-patterns

- ❌ Auto-merging release PR without an explicit phone-tap confirmation.
- ❌ Skipping pre-merge-check or design-judge "because the change is small".
- ❌ Editing `CHANGELOG.md`, `.release-please-manifest.json`, or `<!-- x-release-please-version -->` lines on a non-release branch.
- ❌ Force-pushing or `--no-verify`.
- ❌ Marking a `BACKLOG.md` row ✅ before the release PR has merged.
- ❌ Multiple feature PRs in one cycle without §1 bundle-check or §11 parallel-mode rules being satisfied.
- ❌ Embedding diff content in halt messages — reference PR URL only.
- ❌ Running TDD inline in the orchestrator instead of dispatching to a Sonnet `Task` subagent.
- ❌ Forcing `/compact` inside `/auto-iter` — trust Claude's automatic compaction.

---

## Cadence

| Cadence | Use when |
|---|---|
| `/loop 20m /auto-iter` | NOT recommended — many cycles overflow, halts cluster, collision risk. |
| `/loop 30m /auto-iter` | At-desk attentive; aggressive but viable. |
| **`/loop 60m /auto-iter`** | **Recommended for unattended runs.** |
| `/loop 2-4h /auto-iter` | Conservative; long uninterrupted blocks. |

`/auto-iter` is one cycle per invocation; `/loop` provides the chain. If you see lots of `skip_in_progress` entries in `cycles.jsonl`, the cadence is too tight — bump it.

---

## What's NOT in this loop

- **Manual review of every diff** — design-judge + pre-merge gate are the firewall. If those gates pass, the loop trusts and ships. (Tighten the gates if drift appears.)
- **Auto-resolution of merge conflicts beyond the union rule** — if a real conflict surfaces (during `resolve` of an overlapping PR), the cycle halts.
- **Brainstorming on autopilot** — brainstorm mode requires phone interaction.
- **Auto-release without halt** — every release goes through HALT 2.
- **Compaction inside the cycle** — Claude's automatic context management handles this; the cycle doesn't force `/compact`.
