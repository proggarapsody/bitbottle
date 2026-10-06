# Agent 03 — Consistency & Predictability
## Cycles 153–187 · Analysis date: 2026-06-02

---

## TL;DR

The loop is **not predictably consistent** across cycles 153–187. Three distinct sub-eras exist with incompatible step schemas, missing logs, and sharply different rework rates. The single recurring root cause is the **auto-merge race** (CI ~2 min completes before DJ ~3–5 min; auto-merge fires; BLOCKER lands post-merge). This pattern appeared in 0 of 8 iteration cycles in era 153–161, 2 of 10 in era 168–177, and **5–6 of 9 in era 178–187 (56–67%)**. The trend is sharply worsening. Duration variance is high cross-era (CV 64%) but tight within each era. Pipeline version is superficially frozen at `2026.05.20` for all 35 cycles while the underlying step schema changed three times.

---

## 1. Outcome Stability

### Sub-era breakdown (iteration cycles only, brainstorm excluded)

| Sub-era | Cycles (iteration) | Shipped clean | Shipped-with-fix | Rework rate | Trend |
|---|---|---:|---:|---:|---|
| 153–161 (May 25 stream) | 7 | 7 | 0 | **0%** | — baseline |
| 164 (isolated) | 1 | 1 | 0 | **0%** | — |
| 168–177 (May 29–Jun 1 stream) | 10 | 8 | 2 | **20%** | ↑ regression |
| 178–187 (Jun 1 stream) | 9 | 3–4 | 5–6 | **56–67%** | ↑↑ sharply worse |

**Overall 153–187 rework rate (iteration cycles)**: 7–8 of 27 = **26–30%**.

### Cycle 181 ambiguity

Cycle 181 (CLOUD-CODE-INSIGHTS) is logged as `outcome=shipped` in cycles.jsonl but the DJ metrics record `blocker_count=1`. The stream report classifies it as "shipped | clean." This implies the BLOCKER was caught and fixed pre-merge within the same cycle — the only cycle in 178–187 where the DJ-then-arm gate succeeded by timing luck. If counted as "shipped-with-fix" the era 178–187 rework rate is 6/9 = **67%**; if excluded it is 5/9 = **56%**.

### Era 168–177 detail

Cycles 168–169 shipped but with **dispatch violations** (orchestrator worked inline rather than dispatching subagents). The violations do not show as `shipped_with_fix` in the log but represent a process failure. Post-correction (cycles 172–177), 6 of 6 were dispatch-clean. The 2 `shipped_with_fix` entries (176, 177) are both auto-merge races.

### Auto-merge race: is it worsening?

| Era | Post-merge BLOCKER cycles | Total iteration cycles | Rate |
|---|---:|---:|---:|
| 153–161 | 0 | 7 | 0% |
| 168–177 | 2 | 10 | 20% |
| 178–187 | 5 | 9 | 56% |

The rate tripled from 168–177 to 178–187 despite no workflow spec change. Contributing factor: the 178–187 stream ran faster (~10 min/cycle vs ~20 min/cycle), which may have left less margin for DJ to complete before CI; the scope complexity (MCP handlers, integration tests, permission contracts) also raised the probability that a BLOCKER existed in a given cycle.

---

## 2. Duration Variance

Only eras 153–161 and 178–187 have reliable `duration_active_min` data. Era 168–177 logs all durations as `0`.

### Era 153–161 (iteration cycles: 153, 154, 157, 158, 159, 160)

| Stat | Value |
|---|---|
| Mean | 25.1 min |
| Std dev | 2.5 min |
| CV | **10%** (very tight) |
| Min | 22.6 min (cycle 160) |
| Max | 30.0 min (cycle 154) |

### Era 178–187 (iteration cycles: 178, 179, 180, 181, 183, 184, 185, 186, 187)

| Stat | Value |
|---|---|
| Mean | 8.7 min |
| Std dev | 3.4 min |
| CV | **39%** (moderate spread) |
| Min | 1.9 min (cycle 179, chore scope) |
| Max | 13.9 min (cycle 181) |

### Outlier: cycle 164

Cycle 164 (REPO-CLONE) ran 41.1 min active — **6.3 standard deviations above the era 153–161 mean**. Metrics show the TDD subagent ran 1,093 s (~18 min), followed by an unusually large `step4_ci_wait` of 1,243 s (~21 min). The CI wait accounts for most of the excess. This is a legitimate outlier driven by scope complexity and CI queue time, not a pipeline fault.

### Cross-era comparison

| Cross-era stat | Value |
|---|---|
| Combined mean (all eras with data, n=16) | 16.9 min |
| Combined std dev | 10.7 min |
| Combined CV | **64%** |

The high combined CV is almost entirely explained by the era shift: era 153–161 cycles average ~25 min while era 178–187 cycles average ~9 min. Within each era separately, the pipeline is much more predictable (CV 10% and 39% respectively).

---

## 3. Pipeline Version and Step-Sequence Consistency

### Pipeline version

`pipeline_version` is **frozen at `2026.05.20`** for all 30 dataset entries and all 35 cycles in the jsonl files. Cycle 170 lacks the field entirely (only `bundled:false` present). This field does not discriminate between meaningfully different pipeline behaviours — three distinct step schemas coexisted under the same version string.

### Step schema evolution (4 generations in 35 cycles)

| Generation | Cycles | Step count | Key steps |
|---|---|---|---|
| A (full) | 153–161 | 10 | `lock` → `preflight` → `mode_pick` → `prd_file` → `tdd_start` → `tdd` → `design_judge` → `ci_wait` → `release_pr` → `release_merge` |
| B (legacy) | 162–165 | 4–5 | `preflight` → `scope` → `prd` → `tdd` → `step8_ship` |
| C (collapsed) | 168–177 | 0–1 | `step_ship` only (or nothing); logging disabled |
| D (simplified) | 178–187 | 2 | `step2_tdd` → `step2_design_judge` |

Notably:
- Generation B appears to be an **older schema** that survived in parallel during the May 25 session (cycles 162–165 overlap temporally with 156–161 and use different step names).
- Generation C represents a **complete logging regression** — 9 of 10 cycles in era 168–177 have zero metric steps recorded; token data is absent for all 10.
- Generation D restores minimal logging (2 steps) and re-enables token emission (8/10 cycles emit real tokens).

**DJ presence by schema**:

| Generation | Cycles with DJ step | Total iteration cycles |
|---|---|---|
| A | 3/6 (cycles 157–159 yes; 153, 154, 160 no) | 6 |
| B | 1/3 (cycle 164 yes) | 3 |
| C | 0/10 | 10 |
| D | 8/9 (all except chore cycle 179) | 9 |

DJ was absent from cycles 153, 154 (era A) and entirely absent from era C. This means that 10 of 28 iteration cycles (36%) ran without any DJ gate at all.

---

## 4. Cycle Numbering Integrity

### Observed gaps vs expected range 153–187

The 35-cycle range 153–187 contains **5 absent cycle numbers**: 162, 163, 165, 166, 167.

### Classification

| Cycle(s) | In cycles.jsonl | In metrics.jsonl | Classification |
|---|---|---|---|
| 162 | No | Yes (tdd + step8_ship=shipped) | **Log emit failure** — cycle ran and shipped; `log-cycle.sh` did not write; metrics only |
| 163 | No | Yes (preflight + scope + tdd + step8_ship=shipped) | Same as 162 |
| 165 | No | Yes (scope + prd [corrupted] + step8_ship=shipped) | **Crash + corrupt** — `step2_prd` duration logged as 1,779,745,231,000 ms (~20,600 days, obviously bogus); shipped outcome unreliable; `log-cycle.sh` never ran |
| 166 | No | No | **Never ran** or both files failed — no data anywhere |
| 167 | No | No | Same as 166 |

Cycle 168 appears **twice** in cycles.jsonl: once as a 2026-05-26 manual/legacy entry (PR #568, v1.120.0) and once as the 2026-05-29 stream entry (HOST-INFO, v1.126.0). These are different cycles given the same number — a **cycle counter collision** during the transition between May 25 session and May 29 stream.

### What the gaps imply

1. Cycles 162–163 shipped real work that is tracked in GitHub PRs but not in the cycle audit trail — the SHIPPED.md coverage is incomplete for this period.
2. Cycle 165 represents a crash that may have emitted partial or incorrect commits to main; the reliability of whatever shipped is uncertain.
3. Cycles 166–167 represent two abandoned sessions — the pipeline crashed after 165 and was restarted only on 2026-05-29 at cycle 168. A ~8-hour outage window (2026-05-25 22:07 to 2026-05-26 05:30) is implied.
4. The duplicate cycle-168 entry is a counter integrity violation: the same number labels two independent cycles with different scope, release, and PRs.

**Reliability implication**: 4–5 of the 35 cycles in range (11–14%) have no audit-quality log entry.

---

## 5. Recurring Failure Clusters

### Cluster A: Auto-merge race (n=8, worsening)

Root mechanism: CI (~2 min) completes and fires auto-merge before DJ (~3–5 min) returns BLOCKER verdict.

Sub-clusters by defect type:

| Sub-cluster | Cycles | Finding |
|---|---|---|
| Missing integration test | 185, 187 | `*_integration_test.go` absent; required by ARCHITECTURE.md §Test tiers |
| Missing handler/test coverage | 180 | MCP handler test triplet missing |
| MCP API contract mismatch | 183 | `add_deploy_key` permission default `""` vs CLI default `"read"` |
| Implementation logic bug | 184, 186 | Handler scope duplication; zero-value flag.Changed() bug |
| URL/error mapping | 176, 177 | Missing `url.PathEscape`; wrong error code mapping |

All 8 BLOCKERs were confirmed real (0% false-positive rate on BLOCKER severity). The race is purely a sequencing problem — DJ always found the right answer; it arrived too late.

The pattern has appeared in every stream since 168, was present but rare in 168–177 (20%), and is dominant in 178–187 (56–67%). **No mitigation has been applied between streams** — the recommended "arm auto-merge only after DJ VERDICT" fix from the 168–177 report was not implemented before 178–187 ran.

### Cluster B: Dispatch violations (n=2, corrected)

Cycles 168–169: orchestrator performed inline code edits instead of dispatching subagents. Corrected by cycle 172; did not recur in 172–187. Self-limiting pattern.

### Cluster C: Log emit failures / crashes (n=3–4, historical)

Cycles 162–163: `log-cycle.sh` did not write cycle records after shipped cycles. Cycle 165: corrupted metric duration field; crash before log emit. Cycles 166–167: session abandoned. This cluster is concentrated in the May 25 transition period and has not recurred in eras 168–187.

---

## 6. Recommendations (ranked by impact)

### R1: Gate auto-merge on DJ VERDICT (CRITICAL — would have prevented 7–8 of 8 failures)

Change the sequence from:
```
push → arm auto-merge → await CI → merge  (DJ running in parallel, may arrive post-merge)
```
to:
```
push → await DJ VERDICT → if BLOCKER: push fix → re-run
                         ↓ SHIP
              arm auto-merge → await CI → merge
```

Files: `.claude/commands/auto-iter.md` (~5 lines), `docs/workflows/iteration-cycle/README.md` §3.5–3.6, `docs/workflows/iteration-cycle/quickref.md` anti-patterns. Effort: ~20 lines. Estimated elimination: 5–6 follow-up PRs per 10-cycle stream.

This fix was recommended in the 168–177 stream report. Its absence during 178–187 accounts for the regression from 20% to 56% post-merge BLOCKER rate.

### R2: Require `*_integration_test.go` in TDD subagent prompt (HIGH — 2 of 8 race failures)

Add to mandatory TDD subagent checklist: create `<command>_integration_test.go` using httptest; cite `sync_integration_test.go` as canonical pattern. File: `.claude/commands/auto-iter.md`. Effort: ~2 lines.

### R3: Bump pipeline_version on schema changes (MEDIUM — observability)

`pipeline_version: 2026.05.20` is meaningless for discriminating generations A/B/C/D. Either increment the version string on each schema change or replace it with an explicit schema field (`step_schema: A | B | C | D`). File: `auto-iter/scripts/log-cycle.sh`. Effort: ~1 line + schema definition.

### R4: Validate cycle counter before emit (MEDIUM — prevents collision)

`log-cycle.sh` should verify that the cycle number being emitted does not already exist in cycles.jsonl (or that the new entry is from the same session). The cycle-168 collision was harmless but could obscure real analysis bugs in future. Effort: ~5 lines in `auto-iter/scripts/log-cycle.sh`.

### R5: Crash recovery protocol for metric corruption (LOW — historical)

The cycle-165 corrupted duration (1.78T ms) and the missing 162–163 cycle entries indicate that `metrics.sh` and `log-cycle.sh` have no error handling for partial writes. Add a validation step that rejects obviously bogus durations (> 24h = 86,400,000 ms) and retries the log write on failure. Effort: ~10 lines.

---

## Appendix: Data completeness summary

| Metric | Cycles 153–161 | Cycles 168–177 | Cycles 178–187 |
|---|---|---|---|
| Cycle entries in jsonl | 9/9 | 10/10 | 10/10 |
| Token data emitted | 3/7 iter (partial) | 0/10 | 8/10 |
| Metric steps recorded | 6/7 iter | 1/10 | 8/9 iter |
| DJ step logged | 3/7 iter | estimated 8/10 (not in metrics.jsonl) | 8/9 iter |
| `duration_active_min` present | 7/9 | 0/10 | 9/10 |
| `pipeline_version` present | 8/9 | 9/10 (est.) | 10/10 |
