# Guidelines Audit — Cycles 153–187
**Dimension**: Adherence to documented workflow rules and explicit user feedback  
**Generated**: 2026-06-02  
**Analyst**: review-agent (Claude Sonnet 4.6)

---

## TL;DR

Across 30 cycles (153–187), six distinct compliance dimensions were audited. **Three rules showed sustained violations**: metric emission collapsed to 0–2 steps/cycle for cycles 168–187 (vs. 10 expected), cycles 168–169 had confirmed dispatch violations, and the integration-test requirement (ARCHITECTURE.md tier-2) was missing from TDD subagent prompts, causing DJ BLOCKERs in cycles 185 and 187. **The auto-merge "race" (5/10 cycles 178–187) is not a rule violation**: the old documented rule in `quickref.md` explicitly accepted the "arm on CI green, fix via follow-up PR" tradeoff; the new "arm after DJ" rule was written reactively to memory but never landed in tracked files, creating a documentation-vs-behavior gap. Cycle 179 (BACKLOG-MIGRATION standalone chore) is a genuine but mitigated violation: the §4 gate in `pre-merge-mechanical.sh` only catches single-file commits and silently passed the two-file doc-only PR.

---

## Compliance Matrix

| Rule | Source | Cycles Checked | Verdict | Violations |
|---|---|---|---|---|
| **1. Mandatory step compliance** (lock → preflight → PRD → TDD → DJ → pre-merge → merge → release → cleanup) | `quickref.md` §Metrics schema; `README.md` §0–§10 | 153–187 | **FAIL** | Metric emission collapsed: 0 steps logged for cycles 169–177; 2 steps for 178–187 (vs. ~10 expected). Cycles 162, 163, 165 ran TDD work but were never appended to `cycles.jsonl`. |
| **2. BACKLOG→SHIPPED discipline** (move in same feat commit, not standalone chore) | `README.md` §4; `quickref.md` anti-patterns; `pre-merge-mechanical.sh` §4 | 153–187 | **CONDITIONAL FAIL** | Cycle 179 (BACKLOG-MIGRATION, PR #625): standalone `chore(backlog)` PR touching only `docs/backlog/BACKLOG.md` and `docs/backlog/SHIPPED.md`, with no feat code. Gate §4 did **not** catch it (§4 only blocks `file_count == 1`; cycle 179 had `file_count == 2`). Earlier: cycle 147 had confirmed standalone chore PR #507 (standalone `BACKLOG` flip noted in stream 145–154 report). |
| **3. Auto-merge gating** ("arm after DJ returns SHIP, not on CI green") | `feedback_auto_merge_race.md` (memory, 2026-06-01) | 176–187 | **NOT APPLICABLE at time of cycles** | Old `quickref.md` rule (last updated 2026-05-29) explicitly endorsed arm-on-CI-green with follow-up PR. Cycles 176–177 (2 follow-up PRs) and 178–187 (5 follow-up PRs) followed the **documented** behavior. The new rule exists only in memory and has not been landed in `quickref.md`, `README.md`, or `auto-iter.md`. This is a documentation gap, not a behavioral failure. |
| **4. Integration-test requirement** (`*_integration_test.go` required per new command) | `feedback_integration_test_required.md` (memory, 2026-06-01); `docs/ARCHITECTURE.md` §Test tiers | 185, 187 | **FAIL** | Cycle 185 (REPO-SYNC, PR #643): missing integration test — DJ BLOCKER, fix PR required. Cycle 187 (COMMIT-SEARCH, PR #651): missing `search_integration_test.go` — DJ BLOCKER, fix PR #7bfe6b6 required. The requirement exists in `ARCHITECTURE.md` but is **absent from the TDD subagent prompt** in `.claude/commands/auto-iter.md`. |
| **5. Orchestrator-dispatch rule** (all code writes go to subagents, orchestrator is shell-only) | `feedback_orchestrator_inline_work.md`; `quickref.md` anti-patterns | 168–187 | **PASS with 2 violations corrected** | Cycle 168 (HOST-INFO): orchestrator applied DJ dead-branch fix inline. Cycle 169 (API-PARITY): full TDD cycle done inline, no DJ run. Both logged as dispatch violations in stream 168–177 report. Corrected by cycle 172; cycles 172–187 show no dispatch violations. Cycles 178–187: dispatch violations = 0. |
| **6. Release-halt-in-stream rule** (skip `ship?` halt in stream mode) | `feedback_release_halt_stream.md` | 168–187 | **PASS** | Both streams (168–177 and 178–187): zero release halts recorded. Stream 168–177: user pre-authorized no halts; stream 178–187: same. No `step2_halt` or `step2_halt2` metric lines appear in cycles.jsonl for either stream. Correct. |
| **7. No internal hostnames in public artifacts** | `feedback_no_internal_hostnames.md` (2026-05-19) | 153–187 | **PRE-EXISTING DEBT** | `pkg/cmd/mcp/handlers_test.go` contains `git.moscow.alfaintra.net` at lines 166, 168, 190. However, this was introduced on 2026-04-26 (before the rule was established 2026-05-19) in `feat: add pipeline and branch commands`. No cycle in 153–187 added new hostname instances; subsequent modifications (cycles 174–175) inherited the existing lines. Not a new violation from the audited window — existing technical debt. |
| **8. Agent-rules-location rule** (new algorithm rules in tracked `auto-iter/` or `docs/workflows/`, not memory) | `feedback_agent_rules_location.md` | Post-178–187 | **FAIL** | Two new rules generated as feedback exist only in gitignored memory: (a) `feedback_auto_merge_race.md` — the "arm after DJ" gate rule; (b) `feedback_integration_test_required.md` — the explicit integration test requirement. Neither has been landed in `quickref.md`, `README.md`, or `auto-iter.md`. Per the rule: "any rule that should govern multiple agent harnesses must live under tracked `auto-iter/` or `docs/workflows/`." |

---

## Mandatory Step Compliance — Detail

### Steps expected per iteration cycle (from `quickref.md` §Metrics schema)

The schema lists these required/expected steps for a full iteration cycle:
`step0a_lock`, `step0_preflight`, `step1_mode_pick`, `step2_prd`, `step2_worktree`, `step2_tdd`, `step2_pre_merge_gate`, `step2_design_judge`, `step2_pr_open`, `step2_ci_wait`, `step2_release_pr_wait`, `step2_release_publish`, `step5_cleanup`.

### Observed emission per era

| Era | Typical step count | Missing steps |
|---|---|---|
| Cycles 155–160 | 8–10 | None significant |
| Cycles 161–165 | 1–4 | Partial (lock/preflight often missing) |
| Cycle 168 | 1 (`step_ship` only) | All preflight, TDD, DJ, CI, release steps |
| Cycles 169–177 | 0 | Complete collapse — all steps missing |
| Cycles 178–187 | 1–2 (`step2_tdd` + `step2_design_judge`) | All §0 (lock/preflight), §1 (mode/scope), §2 (PRD), PR open, CI wait, release, cleanup |

### Missing from `cycles.jsonl`

Cycles 162, 163, and 165 appear in `metrics.jsonl` with partial step data but have **no entry in `cycles.jsonl`**. The `log-cycle.sh` call at cycle end was not executed. This means those cycles have no durable outcome record.

---

## BACKLOG→SHIPPED Discipline — Detail

### What the rule says

`README.md` §4: "Both edits land in the same commit as the feat work, NOT in a follow-up `chore:` PR."  
`quickref.md` anti-patterns: "Opening a separate `chore: mark X shipped in BACKLOG` PR after the feat PR merges."  
`pre-merge-mechanical.sh` §4: blocks commits touching ONLY one of BACKLOG.md or SHIPPED.md.

### Cycle 179 (BACKLOG-MIGRATION, PR #625)

- Commit: `chore(backlog): BACKLOG-MIGRATION — sweep shipped scope details to SHIPPED.md`  
- Files changed: `docs/backlog/BACKLOG.md` (629 deletions) + `docs/backlog/SHIPPED.md` (637 insertions)  
- No Go code, no test files, no release bump.

**Strict reading**: this is a standalone chore PR touching only BACKLOG + SHIPPED without feat code — the pattern the rule prohibits.  

**Mitigating context**: BACKLOG-MIGRATION was itself a queued BACKLOG scope (added by the `docs(backlog): split into queue` commit as "deferred so each move can be its own bisectable chore commit"). It represents a **bulk migration** of 50+ already-shipped scopes that had accumulated before the SHIPPED.md file existed. There is no "corresponding feat commit" to piggyback on.

**Gate gap confirmed**: `pre-merge-mechanical.sh` §4 checks `file_count == 1`. Cycle 179 had `file_count == 2`, so the gate passed silently. The anti-pattern applies to two-file doc-only commits as much as single-file ones, but the gate does not enforce it.

**Verdict**: Technical violation of the letter of the rule; mitigated by intentional design of the migration scope. The gate gap (file_count threshold) is the more actionable finding.

---

## Auto-Merge Gating — Reactive Rule vs. Documented Behavior

### What the quickref said at cycle time (last updated 2026-05-29)

`quickref.md` §Halt routing "Accepted tradeoff":
> `gh pr merge --auto` fires on the first satisfying event, which is typically CI green (~2 min) — faster than design-judge return (~3–5 min). When DJ finds a BLOCKER after auto-merge fires, the BLOCKER is fixed via a follow-up PR rather than delayed merging. [...] Net: the follow-up-PR pattern is faster for stream throughput.

### What the new feedback says (written 2026-06-01, post-stream)

`feedback_auto_merge_race.md`: arm auto-merge only AFTER DJ returns SHIP. The old behavior is now explicitly wrong.

### Assessment

The loop behavior in cycles 176–177 and 178–187 was **consistent with the documented spec**. The 5/10 follow-up PRs in stream 178–187 are a **consequence of the documented tradeoff**, not a rule violation at cycle time. The new feedback effectively invalidates the "Accepted tradeoff" section of `quickref.md` — but that invalidation has not been written into the tracked spec files.

**Documentation-vs-behavior gap**: the new rule exists in `feedback_auto_merge_race.md` (gitignored memory) but the contradicting old rule remains in `quickref.md` (tracked). Any agent reading `quickref.md` as the spec would continue arming auto-merge on CI green — which is correct per the spec, incorrect per the new feedback.

---

## Integration-Test Rule — Confirmed Violations

### Cycles 185 and 187

Both cited in `feedback_integration_test_required.md`:
- **Cycle 185 (REPO-SYNC, PR #643)**: DJ returned 2 BLOCKERs including missing integration test. Fix PR landed post-merge. `ARCHITECTURE.md` §"Test tiers" tier-2 is clear: integration tests required for every new command.
- **Cycle 187 (COMMIT-SEARCH, PR #651)**: DJ returned BLOCKER for missing `search_integration_test.go`. Fix PR `fix(commit): add search_integration_test.go` (`7bfe6b6`) landed post-merge.

### Root cause

`ARCHITECTURE.md` §Test tiers states the requirement. The TDD subagent prompt (`.claude/commands/auto-iter.md` §TDD subagent prompt requirements) does not explicitly mention `*_integration_test.go`. Subagents read the PRD and primer but do not independently re-read the full ARCHITECTURE.md for each task. The omission of an explicit directive in the prompt is the direct cause.

The fix is 2 lines in `.claude/commands/auto-iter.md` and equivalent wording in the tracked `docs/workflows/iteration-cycle/README.md` §3 subagent brief — neither has been updated as of 2026-06-02.

---

## Orchestrator-Dispatch Violations — Detail

### Confirmed violations (cycles 168–169, stream 168–177)

**Cycle 168 (HOST-INFO)**: DJ returned a dead-branch finding (`isCloud` conditional always-false). Orchestrator edited the file inline instead of dispatching a fix-agent. Logged in stream report as dispatch violation.

**Cycle 169 (API-PARITY)**: Orchestrator performed the full TDD implementation inline (txtar + Long description edits). No DJ subagent dispatched. Logged in stream report.

**Stream 178–187**: 0 dispatch violations. The rule was re-internalized after the cycle 168–169 correction and held for 20 subsequent cycles.

### Pattern note

Both violations in cycles 168–169 occurred at stream open when the orchestrator had high context (user had specified which issues to open). The stream 168–177 report notes: "High-context opening may lower dispatch discipline." No structural guard prevents this — the rule is currently policy only, not mechanically enforced.

---

## Release Halt in Stream — Compliant

Both streams correctly skipped the `ship?` halt:
- Stream 168–177: user explicitly pre-authorized ("for this 10 cycles don't ask me").
- Stream 178–187: user pre-authorized at stream start.

No `step2_halt` or `step2_halt2` metric lines appear for cycles 168–187. Release PRs auto-merged without halts. This is the correct behavior per `feedback_release_halt_stream.md`.

---

## Documentation-vs-Behavior Gaps

| Gap | Doc state | Actual behavior | Risk |
|---|---|---|---|
| **Auto-merge gate sequence** | `quickref.md` says: arm on CI green, accept follow-up. `feedback_auto_merge_race.md` says: arm after DJ. | Loop follows quickref (arm on CI green). | Any agent reading the spec will continue the old behavior. 50% follow-up PR rate in next stream if unchanged. |
| **Integration-test explicit requirement** | `ARCHITECTURE.md` implies it; TDD subagent prompt in `auto-iter.md` does not mention it. | TDD subagents skip integration tests when not prompted. | Continued DJ BLOCKERs post-merge each time a new command lands without the directive. |
| **§4 gate — two-file doc-only commits** | `pre-merge-mechanical.sh` §4 only blocks `file_count == 1`. Quickref says move must land with feat code. | Two-file doc-only chore commits (like cycle 179) pass §4 silently. | Bulk doc migrations or split BACKLOG+SHIPPED moves can bypass the gate. |
| **Metric emission breadth** | `quickref.md` requires ~10 step rows per cycle via `metric.sh`. | Streams 168–187 emit 0–2 steps/cycle. | Cross-cycle performance analysis is impossible; orchestrator tokens always `null`. |
| **Cycle log completeness** | `autonomous.md` §10: `log-cycle.sh` appends at cycle end. | Cycles 162, 163, 165 have metric steps but no `cycles.jsonl` entry. | Dataset gaps; cycle counter reliability uncertain for that era. |

---

## Repeat Offenders

| Failure mode | First seen | Cycles | Status |
|---|---|---|---|
| **Metric emission collapse** | Cycles 77–86 (May-17 stream) | 168–177 (complete), 178–187 (partial) | Unresolved — partial improvement but still far below spec |
| **Auto-merge race → follow-up PR** | Cycles 135–144 | 176–177, 178–187 (5/10) | Rule change written to memory; not yet in tracked files |
| **Missing integration test** | Cycle 185 | 185, 187 | Rule written to memory; not yet in TDD subagent prompt |
| **Standalone BACKLOG chore PR** | Cycle 77–86 era; cycle 147 | Cycle 179 | §4 gate gap (2-file threshold); technical debt |
| **Dispatch violation at stream open** | Cycles 77–86 (inline TDD) | 168–169 | Corrected mid-stream; held for 18 subsequent cycles |

---

## Recommendations

Ranked by ROI (highest first):

### R1 — Land the new auto-merge gate rule in tracked files (CRITICAL, ~20 LOC)

The `feedback_auto_merge_race.md` rule must be written into `quickref.md` (replace/remove the "Accepted tradeoff" section), `README.md` §3.6, and `.claude/commands/auto-iter.md` §In-cycle parallel block. Until this lands, any future stream will continue the old behavior and produce follow-up PRs at the 50% rate observed in 178–187.

**Files**: `docs/workflows/iteration-cycle/quickref.md` (replace "Accepted tradeoff" block), `docs/workflows/iteration-cycle/README.md` §3.5–3.6, `.claude/commands/auto-iter.md` §In-cycle parallel block.

### R2 — Add integration-test directive to TDD subagent prompt (HIGH, 2 lines)

Add to `.claude/commands/auto-iter.md` §TDD subagent prompt requirements:
```
7. Create `<command>_integration_test.go` using httptest — canonical pattern:
   `pkg/cmd/repo/sync/sync_integration_test.go`. At least 2 tests (success + error).
   This is tier-2 per ARCHITECTURE.md §Test tiers — not optional.
```
Also add equivalent wording to `docs/workflows/iteration-cycle/README.md` §3 subagent brief per the agent-rules-location rule.

**Files**: `.claude/commands/auto-iter.md`, `docs/workflows/iteration-cycle/README.md`.

### R3 — Fix §4 gate to catch two-file doc-only commits (MEDIUM, ~5 LOC)

`pre-merge-mechanical.sh` §4 should block commits where ALL changed files are in `docs/backlog/` (BACKLOG.md and/or SHIPPED.md) and there is no Go or non-doc file in the same commit, regardless of whether `file_count` is 1 or 2. This closes the cycle 179 gate gap.

**File**: `auto-iter/scripts/pre-merge-mechanical.sh` §4 block (~5 lines).

### R4 — Restore full metric emission in stream mode (MEDIUM, mechanical)

The collapse from 10 steps/cycle (era 153–165) to 2 steps/cycle (era 178–187) is unexplained. The stream is faster (~12 min/cycle) — it's plausible that compressed context caused emission skips. The spec requires `metric.sh` calls at each step boundary; the orchestrator must not skip them under time pressure. Add a post-cycle sanity check (already in `quickref.md` §Hard sanity rules) to the stream wrapper.

### R5 — Replace internal hostname in `handlers_test.go` with a generic placeholder (LOW, ~3 lines)

`pkg/cmd/mcp/handlers_test.go` lines 166, 168, 190 reference `git.moscow.alfaintra.net`. This predates the rule (added 2026-04-26), but the rule says existing test fixtures should not carry it. Replace with `bitbucket.example.com` or `example.bitbucket.server` so the secret scan in `pre-merge-check.md` §8 doesn't need a permanent carve-out.

**File**: `pkg/cmd/mcp/handlers_test.go` (3 lines).

---

## Bottom Line

The documented workflow was largely followed where it was unambiguous. The two critical gaps are both of the form "new feedback rule exists in memory but not in tracked spec files" — auto-merge gating (R1) and integration tests (R2). Neither was a rule violation at cycle time because the spec files said otherwise. This is the most important finding: **feedback rules written only to memory are invisible to any agent reading the tracked spec**. Until R1 and R2 land in `quickref.md` and `auto-iter.md`, the behavior they prohibit will recur in the next stream.
