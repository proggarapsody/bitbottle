# Autonomous-loop quickref

Declarative reference for the bitbottle autonomous iteration loop. Read
this when designing or troubleshooting `/auto-iter`. Procedural details
live in agent-specific command files (e.g. `.claude/commands/auto-iter.md`);
canonical workflow lives in [`README.md`](README.md).

> **Mantra**: one `/auto-iter` invocation = one cycle = exits at end.
> `/loop` provides the cadence between cycles. Halt for anything
> irreversible or judgment-heavy; auto-confirm anything mechanical.

---

## Architecture

```
laptop (caffeinated, remote human notification channel active)
  └── /loop 30m /auto-iter
        └── /auto-iter (one cycle)
              ├── §0a re-entry lock check     → auto-iter/scripts/lock.sh
              ├── §0  preflight + workspace inventory → auto-iter/scripts/preflight.sh
              ├── §1  pick mode + scope (+ bundle-check)  ▲ scripted (planned)
              ├── §2  execute mode (iteration | architecture | brainstorm | stop)
              ├── §3  halt routing (auto-confirm safe; out-of-band notification for risky)
              ├── §4  append to cycles.jsonl + metrics.jsonl → auto-iter/scripts/{metric,log-cycle}.sh
              └── §5  release lock, exit       → auto-iter/scripts/lock.sh release
```

**Mechanical steps are scripted.** Full interface catalog in [`scripts.md`](scripts.md) — each script emits a single JSON object on stdout, exit-codes on halt-class failure, ships with a paired `_test.sh`. The orchestrator stays in the LLM only for genuine judgment phases.

Three sibling files in `.claude/auto-iter/` (all gitignored):

| File | Purpose |
|---|---|
| `cycles.jsonl` | One line per cycle, high-level outcome. Cycle counter is global, monotonic. |
| `metrics.jsonl` | One line per **step** within a cycle. Wall-clock + tokens. |
| `.lock` | Sentinel — present iff a cycle is running. Recent (<1h) lock → skip; stale (>1h) lock → assume crashed predecessor, clear. |

---

## Model tier per phase

Default to the **code-generation model**. Escalate to the **judgment-heavy model** only for genuinely judgment-heavy phases.

| Phase | Model tier | Why |
|---|---|---|
| §0a re-entry lock | mechanical | Mechanical file check |
| §0 preflight + workspace inventory | mechanical | Shell-driven, deterministic |
| §0 open-PR overlap (decision halt) | mechanical | Routes to out-of-band notification for judgment; agent itself just frames the question |
| §1 mode pick + scope pick | mechanical | Algorithm-driven; BACKLOG-driven scope-pick is deterministic |
| §1 bundle-check | mechanical | Algorithm-driven |
| §2 PRD drafting | code-generation (clone) · **judgment** (new-API/write) | Mechanical fill for pattern-clones against already-exercised endpoints. New-API or write/mutation scopes require the `## Assumptions & Evidence` gate (README §2) — a judgment phase, because an unexamined assumption here is reproduced by the fake, the test, and design-judge alike (#655). |
| §2 worktree creation | mechanical | Shell command |
| §2 TDD implementation | **code-generation** (subagent) | Code generation. Dispatch via `Task` tool with `isolation: "worktree"` — keeps orchestrator context light. Judgment-heavy model only if scope is genuinely complex (rare). |
| §2 doc sync | mechanical | Mechanical doc updates per §5 doc-sync table |
| §2 pre-merge gate | mechanical | Reads CI status, runs grep checks |
| §6 design-judge | **code-generation** (subagent) | Read-only review against `TASTE.md` + `ARCHITECTURE.md` + diff. Checklist includes a dead-branch / tautological-assignment scan (off-the-shelf linters don't catch semantic no-ops like `if x == "Y" { x = "Y" }`). Returns findings; runs in parallel with CI. |
| §2 PR open + CI wait | mechanical | Polling |
| §2 halt (ship) | mechanical | Frame the question, wait for tap |
| §2 fix-after-CI-red | **code-generation** (subagent) | Targeted code change against existing branch |
| §7 merge feature + release PRs | mechanical | `gh pr merge` mechanical |
| §7 release publish wait | mechanical | Polling `gh release view` + `npm view` |
| §8 close PRD | mechanical | `gh issue` mechanical |
| §9 manual-test refresh | mechanical | Per §9 decision flow |
| **§2 brainstorm** (when BACKLOG empty) | **judgment-heavy** | Open-ended judgment, BACKLOG-pattern matching, scope generation. Runs autonomously (no phone halt) — see brainstorm rules below. |
| **§2 architecture audit** (every 5th cycle) | **judgment-heavy** | Architectural reasoning, pattern recognition |
| §5 release lock | mechanical | File removal |

Two phases legitimately need the judgment-heavy model: brainstorm and architecture-audit. Everything else — including design-judge — is pattern-matching against a concrete checklist where the code-generation model is equally effective.

Agent-specific tool names for each tier (e.g. which Claude model maps to which tier) live in the agent's command file, not here.

---

## Halt routing

Two categories — **mid-cycle** (pause-and-continue) and **chain-breaking** (exit).

### Mid-cycle halt (fires when release-please opens a new release PR, not per feat)

| Halt | Notification format | Required reply |
|---|---|---|
| HALT — Ship | `🚀 release PR #N bundles K feat/fix → v<X.Y.Z>. ship?` | tap `ship` / `hold` |

The feature PR is **auto-merged** once CI is green AND all gates pass (no halt — gates already vetted content). The only mid-cycle halt is at the irreversible release-publish step.

**Accepted tradeoff (auto-merge vs design-judge timing).** `gh pr merge --auto`
fires on the first satisfying event, which is typically CI green (~2 min) —
faster than design-judge return (~3–5 min). When DJ finds a BLOCKER after
auto-merge fires, the BLOCKER is fixed via a follow-up PR rather than
delayed merging. Measured: 2 of 4 BLOCKER cases in the 2026-05-24 stream
(cycles 139, 143) followed this pattern. The alternative — serializing
merge after DJ — adds DJ wall-clock (~3–5 min) to every cycle, including
the 60 %+ that have zero BLOCKERs. Net: the follow-up-PR pattern is faster
for stream throughput. If a real security BLOCKER lands and the follow-up
window is unacceptable, escalate manually; do not change the default.

**The halt fires per release-please PR, NOT per feat cycle.** When cycles run back-to-back faster than release-please reacts, one release-please PR collects multiple `feat:` commits — across cycles 81–86 in the May-17 stream, six `feat:` cycles ran in 87 minutes and release-please opened a single PR (#359) that bundled all six. A halt-per-feat model would have asked the user to ship six times for one publish; instead it fired zero times because the per-cycle assumption is wrong. Match reality: halt only when release-please opens a **new** PR (different number than the previously-halted-on one).

Tap → release PR merges, GoReleaser publishes async. Tap "hold" → exit, cycle outcome `halt_release`.

For `refactor:` / `docs:` / `chore:` cycles: no halt at all — these don't trigger release-please, cycle ends after feature merge. For `feat:` / `fix:` cycles where release-please hasn't opened a new PR yet (debounce in progress): also no halt; cycle ends `shipped` with `release: null`, and the halt fires on a later cycle when the PR opens.

### Chain-breaking halts (rare — exit and wait for human)

| Halt | Notification format | Reply |
|---|---|---|
| Workspace must-resolve | `⚠️ workspace requires manual cleanup: <details>` | exit cleanly |
| Open-PR overlap | `⚠️ PR #N overlaps scope. resolve / skip / close?` | tap `resolve` / `skip` / `close` |
| Pre-merge BLOCKER | `❌ PR #N: <finding>` | tap `fix` / `override <reason>` |
| CI red | `❌ PR #N CI red: <log_url>` | tap `retry` / `fix` / `abort` |
| Merge conflict (during resolve) | `❌ PR #N merge conflict beyond union rule: <files>` | exit cleanly |
| Stop confirmation (BACKLOG empty + 3 empty brainstorms) | `🏁 confirm shutdown` | tap `confirm` / `continue` |

> **Brainstorm runs autonomously** — no halt notification. Empirically validated across cycles 34–53 (5 brainstorms, 14/16 scopes shipped, 1 false positive, 1 pending). See "Brainstorm rules" below for the constraints that keep this honest.

### Halt protocol rules

- **No diff in halt messages.** Reference PR URL, not embedded diff. Halt messages stay <500 chars so out-of-band notifications display cleanly.
- **Halt-response timeout**: 2 hours. After that, log `halt_no_response`, exit. Matches the cycle wall-clock cap (see below) so there's a single ceiling.
- **Cycle wall-clock cap**: 2 hours per cycle. Hard ceiling; orchestrator force-exits with `halt_cycle_timeout` if exceeded. Covers the historical worst-case (~90 min) with margin; anything longer is almost certainly stuck.
- **Auto-confirm**: scope-pick (BACKLOG-driven), bundle-check (algorithm-driven), workspace-clean preflight, mechanical doc-sync, secret-leak scan, build-artifacts scan, lint+test-via-CI, PRD close confirmation, manual-test refresh decision, worktree removal.

---

## Brainstorm rules

Brainstorm runs autonomously (judgment-heavy model, ~2–5 min, **target 8–10
new BACKLOG rows per run** — enough to fill one full `/auto-iter-stream 10`).
Lower targets were tried (cycles 136, 142 produced 4 each) but a 10-cycle
stream then needed 2 brainstorm cycles, ~160K extra Opus tokens. Cycle 64
empirically demonstrated 22 rows in a single 5-min brainstorm — capacity
is not the limiter, the prompt target is. To keep autonomous brainstorms
honest, every emitted row must satisfy **all** of the following or be
dropped before it lands in `docs/backlog/BACKLOG.md`:

1. **No-overlap.** Scan `docs/backlog/SHIPPED.md` + the Functionality Map in `docs/backlog/BACKLOG.md`. Reject anything redundant with already-shipped scope. _(This is the single load-bearing rule — it killed PR-TEMPLATE retroactively at cyc 39 because `repo file get` already covered it.)_
2. **Backend declared.** Each row marks `Cloud` / `Server` / `Both`. If `Both`, both endpoints must be named.
3. **Shape match.** Each row declares which canonical pattern it follows (`List*` via `paging.Collect`, write op with typed errors, MCP triplet, etc.). New shapes require a §architecture-audit cycle, not a brainstorm row.

Soft rules (judgment-heavy model uses for ordering, no auto-reject):

- Prefer scopes that exercise an under-instrumented adapter.
- Prefer scopes that mirror a recently-shipped pattern (compounds tooling familiarity).
- Avoid scopes that depend on un-released Bitbucket features.

The brainstorm emits its full output to `metrics.jsonl` as `step2_brainstorm` with `rows_added`, `rows_dropped_by_overlap`, `rows_dropped_by_feasibility`. Empty brainstorms (0 rows added after rule application) count toward the 3-empty shutdown counter.

---

## Cycle log schema (`cycles.jsonl`)

One line per cycle, append-only. **Write via [`auto-iter/scripts/log-cycle.sh`](../../../auto-iter/scripts/log-cycle.sh)**:

```bash
auto-iter/scripts/log-cycle.sh --cycle=55 --mode=iteration \
  --scope=PR-PARTICIPANTS --outcome=shipped --pr=292 --release=v1.60.0
```

Output line shape:

```jsonl
{"ts":"<ISO>","cycle":<N>,"mode":"iteration|architecture|brainstorm|stop","scopes":[<slugs>],"prs":[<numbers>],"release":"<v1.X.Y or null>","duration_min":<int>,"tokens":<int or null>,"outcome":"<see below>","bundled":<bool>,"notes":[<strings>]}
```

`cycle` is **global, monotonic** — read max from log, increment by 1.

### Outcome enum

| Value | Meaning |
|---|---|
| `shipped` | Feature PR merged, release published (or release skipped for `refactor:`/`chore:`) |
| `skipped_already_shipped` | Scope was already in main's history (§1.5 already-shipped check). Cycle wrote a 1-line BACKLOG status flip and skipped §2. Counts toward stream `ran`. |
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
| `halt_release` | User replied non-ship to the single ship halt (feature is auto-merged before this halt) |
| `halt_no_response` | Phone halt timed out (2 hr default — matches cycle cap) |
| `halt_cycle_timeout` | Cycle exceeded 2-hour wall-clock cap |
| `skip_in_progress` | Step 0a found a recent lock; another `/auto-iter` is running. **Do NOT increment cycle counter.** |

---

## Metrics log schema (`metrics.jsonl`)

One line per **step** within a cycle, append-only. **Write via [`auto-iter/scripts/metric.sh`](../../../auto-iter/scripts/metric.sh)** — it sets `ts`, validates required fields, and appends atomically:

```bash
auto-iter/scripts/metric.sh --cycle=42 --step=step2_tdd \
  --duration_ms=12345 --subagent_tokens=7000
```

Output line shape:

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
| `step2_brainstorm` | `subagent_tokens`, `rows_added`, `rows_dropped_by_overlap`, `rows_dropped_by_feasibility` |
| `step2_audit_run` | `subagent_tokens`, `findings` |
| `step2_design_judge` | `subagent_tokens`, `findings_count`, `blocker_count` |
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

### Hard sanity rules

Before emitting any metrics line, the orchestrator (or `metric.sh` itself)
must verify:

- **`duration_ms` ∈ [0, 3_600_000]**. Anything outside ≤ 1 hour is a bug —
  cycle 5 once recorded `1778303966000` (56 years) from a bad `date`
  subtraction. Typed `--argjson` catches this kind of overflow immediately;
  the bound is a defense-in-depth check.
- **`ts` is non-empty** (derived from `now | todateiso8601`; never `""`).
- **`cycle` may be `null` only for `step0a_lock` and `step0_preflight`** —
  the cycle counter isn't known until §1 mode-pick. Every other step must
  carry a real cycle number.
- **If a step ends without writing a metrics line**, the next step's
  emission must include `--previous_step_no_metric=<step_name>` so the
  omission surfaces in post-cycle analysis.

Post-cycle sanity check (run manually during break-in periods):

```bash
jq -s 'map(select(
  (.duration_ms > 3600000) or
  (.duration_ms < 0) or
  (.ts == "") or
  (.metrics_bug)
))' .claude/auto-iter/metrics.jsonl
```

Should return `[]`. Anything else is an emission bug to fix before drawing
conclusions from the metrics.

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
- ❌ Opening a separate `chore: mark X shipped in BACKLOG` PR after the feat PR merges. **The BACKLOG→SHIPPED move belongs in the feat commit itself** (see [README.md §4](README.md#section-4--sync-docs-and-tooling-neutral-context)). Across cycles 77–86 the post-merge chore-PR pattern produced 8 extra PRs and 4 duplicate commits on `main` (MCP-VALIDATION, JSON-STABILITY, ERR-EMPTY-400, BACKEND-TYPE-STRICT each landed twice).
- ❌ **Marking the row ✅ in place inside `docs/backlog/BACKLOG.md`.** Reorg (2026-05-29): a shipped scope is **moved** — cut its Up-Next table row + `### SCOPE` detail section from `docs/backlog/BACKLOG.md`, prepend a dated entry to `docs/backlog/SHIPPED.md`. Both edits land in the same `feat:` commit. The `auto-iter/scripts/pre-merge-mechanical.sh` §4 check enforces this: a commit touching ONLY `docs/backlog/BACKLOG.md` or ONLY `docs/backlog/SHIPPED.md` is the standalone-chore anti-pattern.
- ❌ Moving a row out of `docs/backlog/BACKLOG.md` before the corresponding feat PR has merged. (The move in the same commit is correct; an earlier separate PR is not.)
- ❌ Editing a `SHIPPED.md` entry after it lands. SHIPPED.md is append-only. If a shipped scope breaks, file a fresh BACKLOG entry to fix it.
- ❌ Multiple feature PRs in one cycle without §1 bundle-check or parallel-mode rules being satisfied.
- ❌ Embedding diff content in halt messages — reference PR URL only.
- ❌ Running TDD inline in the orchestrator instead of dispatching to a code-generation model `Task` subagent. Across the May-17 stream this produced `dispatch_violation:true` on cycles 77 and 79 — both were "scope already shipped historically" cases that should have been caught earlier (see §1.5 already-shipped check in [`autonomous.md`](autonomous.md)).
- ❌ Running a full §2 pipeline for a scope already present in `main`. The §1.5 already-shipped check (`git log --all --grep="$SCOPE_TAG"` matched against a `feat:`/`fix:`/`refactor:` commit) must run before TDD dispatch. Cycles 77 (DEPGUARD-CI) and 79 (FAKECLIENT-SAFETY) re-did work shipped two days earlier (PR #338 and #334).
- ❌ `git merge origin/main` at cycle boundaries on local `main`. Branch protection blocks direct pushes to `main` so the local branch is read-only at cycle boundaries — `git reset --hard origin/main` is the correct operation. The May-17 stream littered `main` with 7 `Merge remote-tracking branch` commits because the orchestrator chose merge over reset.
- ❌ `git checkout -b feat/<slug> origin/main` inside the **main checkout**. Every iteration runs in its own worktree — use `git worktree add -b feat/<slug> ../bitbottle-worktrees/<slug> origin/main`. The main checkout stays clean. See [README.md §3](README.md#section-3--implement-tdd) HARD STOP block. PRD #372 surfaced this as a process bug riding alongside four auth bugs.
- ❌ Emitting metrics via raw `echo >> metrics.jsonl` or inline `jq -nc`. **Always use [`auto-iter/scripts/metric.sh`](../../../auto-iter/scripts/metric.sh).** Prose checklists for emission discipline get skimmed past after context compaction — across cycles 81–86 in the May-17 stream, per-step metrics emission collapsed from 10 lines/cycle (cycles 77–79) to 0–1 lines/cycle. A shell invocation can't be skimmed.
- ❌ Forcing `/compact` inside `/auto-iter` — trust Claude's automatic compaction.
- ❌ Drafting a write/new-API PRD as mechanical fill, with no `## Assumptions & Evidence` section (README §2). An unverified assumption about backend behavior is reproduced identically by the hand-written fake, the test, and design-judge — all three agree, all three are wrong. This is the spec-time root of #655. New-API/write PRDs are a judgment phase; every backend-behavior claim is CITED, LEDGER (`docs/backend-quirks.md`), or a blocking `ASSUMED — UNVERIFIED` that a reality probe must settle before TDD.
- ❌ A write-op whose only test assertion is `stdout '…'`. Assert on the captured request body — the fake can't catch a dropped field it was written (by the same author, with the same assumption) to ignore. Pre-merge-check §6a blocks this.
- ❌ Splitting an `*-AUDIT` scope into "audit cycle" + "fix cycle" when the audit produced exactly one finding that fits in ≤1 PR. See § AUDIT_CONTINUE below. Cycles 95 (UX-FLAG-AUDIT) and 96 (DEBUG-TRANSPORT-FLAG) shipped what should have been one cycle, costing an extra release and ~22 min wall time.

---

## AUDIT_CONTINUE — bundle audit→fix when the finding fits in one PR

`*-AUDIT` scopes (e.g. `UX-FLAG-AUDIT`, `SCORECARD-AUDIT`) often surface findings that can be patched inside the same PR as the audit deliverable. When that's the case, the cycle must **not** end after the audit and let the loop pick the follow-up next cycle.

**Trigger**: scope slug matches `*-AUDIT` AND the audit produced **exactly one** finding AND that finding fits in ≤1 PR (heuristic: ≤200 LOC across ≤8 files, single subsystem).

**Action inside the same cycle**:
1. The audit deliverable (audit table doc, BACKLOG row for the new sub-scope) lands first as commits on the same branch.
2. The TDD subagent immediately implements the follow-up on the same branch.
3. Both BACKLOG rows (the audit and the follow-up) flip 🔲 → ✅ in the same feat commit. Conventional Commit subject: `feat(<audit-area>): audit + <one-line fix description>`.
4. Single PR, single release-please bump, single cycle entry.

**Don't trigger** when:
- Audit produced **zero** findings (cycle ends with the audit deliverable; outcome = `shipped`).
- Audit produced **two or more** findings (file each as its own BACKLOG row; current cycle ships only the audit; follow-ups are picked normally — bundling >1 fix invites scope drift).
- The follow-up fix would exceed the 1-PR heuristic (file the BACKLOG row, cycle ships the audit only).

**Cycle log shape** when AUDIT_CONTINUE fires:

```jsonl
{"cycle":N,"scopes":["UX-FLAG-AUDIT","DEBUG-TRANSPORT-FLAG"],"prs":[394],"outcome":"shipped","bundled":true,...}
```

`bundled:true` distinguishes audit-continue from a vanilla `*-AUDIT` cycle; the two scope slugs in `scopes[]` make the bundle visible to post-hoc analysis.

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
- **Auto-release without halt** — every release goes through the ship halt (the only halt left in the clean path).
- **Compaction inside the cycle** — Claude's automatic context management handles this; the cycle doesn't force `/compact`.
