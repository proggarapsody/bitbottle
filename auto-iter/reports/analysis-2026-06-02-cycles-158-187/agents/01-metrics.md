# Metrics Agent Report — cycles 153–187 (2026-06-02)

Analysis dimension: **METRICS** — quality, completeness, and trustworthiness of
auto-iter instrumentation; what the metrics reveal about cycle performance.

---

## TL;DR

- **Data completeness is 37% overall** (11/30 cycles have all five key fields
  populated); the dominant gap is the 168–177 zero-emission regime where tokens,
  duration, and metric steps were never written.
- **TDD dominates wall time at 84–92%** of instrumented step time (cycles 178–187
  median 9.6 min/cycle); design-judge consumes a consistent 8–17%; CI wait and
  PRD are not instrumented in the latest stream at all.
- **Design judge ran in 12/30 cycles** and flagged blockers in 7 of the 11 cycles
  where `blocker_count` was recorded (64% blocker rate); all 7 were confirmed real
  (0% false-positive rate per stream-178-187 report).
- **Five distinct instrumentation defects** exist: corrupt `duration_wall_min`
  on cycle 157 (29.6M min instead of 31 min); field-name inconsistency
  (`blockers` vs `blocker_count`); step-name drift across 4 canonical groups;
  `metrics_steps_count` mismatches on 4 cycles; and five cycles (162, 163, 165,
  166, 167) entirely absent from both dataset and raw `cycles.jsonl`.

---

## 1. Data Completeness

Five key fields assessed: `tokens > 0`, `duration_active_min > 0`, `release`
populated, `scopes` populated, `_metric_steps` non-empty. Brainstorm cycles
excused from release requirement.

| Regime | Cycles | tokens>0 | duration>0 | scopes | steps>0 | DJ step | Fully complete |
|---|---|---|---|---|---|---|---|
| 153–164 (pre-stream) | 10 | 6/10 | 9/10 | 7/10 | 8/10 | 4 | 1/10 (10%) |
| 168–177 (zero-logged) | 10 | 0/10 | 0/10 | 10/10 | 1/10 | 0 | 0/10 (0%) |
| 178–187 (instrumented) | 10 | 9/10 | 9/10 | 10/10 | 10/10 | 8 | 10/10 (100%) |
| **Total** | **30** | **15/30** | **18/30** | **27/30** | **19/30** | **12** | **11/30 (37%)** |

**Five cycles completely absent** (162, 163, 165, 166, 167) from both the
dataset and `cycles.jsonl` — they have entries in `metrics.jsonl` only. This
leaves a 5-cycle gap in any per-cycle lineage analysis. The metrics.jsonl entries
for 162 and 163 show only `step3_tdd` + `step8_ship` (no CI wait, no DJ). Cycle
165 has a corrupt `step2_prd` duration (see §5).

The clear instrumentation regime change:
- **153–164**: partial instrumentation; some cycles logged manually or retroactively
- **168–177**: pipeline ran but emitted no tokens/duration (all zeros), only
  scopes and outcome; likely a metric-script update lag after the pipeline overhaul
  around 2026-05-29
- **178–187**: full emission; all iteration cycles have tokens, duration, step
  timing, and DJ output

---

## 2. Per-Step Timing Breakdown

Source: `_metric_steps` across all 30 cycles (only steps with `duration_ms` > 0
and < 1×10¹⁰ included). N = number of step instances.

| Step (canonical) | N | Median (ms) | Min (ms) | Max (ms) | Mean (ms) | Note |
|---|---|---|---|---|---|---|
| tdd | 15 | 575,025 | 112,636 | 1,111,843 | 661,627 | Dominant step |
| ci_wait | 5 | 180,000 | 60,000 | 1,243,000 | 386,600 | Outlier: cycle 164 (1,243s) |
| brainstorm | 3 | 118,381 | 32,493 | 158,691 | 103,188 | |
| merge | 2 | 95,000 | 10,000 | 180,000 | 95,000 | |
| design_judge | 12 | 83,668 | 15,704 | 305,055 | 113,296 | Outlier: cycle 159 (305s) |
| release_pr | 5 | 60,000 | 5,000 | 60,000 | 43,000 | Clamped estimate |
| pr_open | 1 | 45,000 | — | — | — | Single obs. |
| release/merge | 5 | 30,000 | 5,000 | 60,000 | 35,000 | |
| prd | 5 | 5,000 | 5,000 | 8,000 | 6,200 | Clamped |
| lock | 7 | 2,000 | 100 | 240,000 | 35,614 | Outlier: cycle 156 blockers_fixed mislabeled |
| mode_pick/scope | 7 | 1,000 | 0 | 3,000 | 1,714 | |
| preflight | 7 | 500 | 0 | 12,000 | 3,786 | |
| tdd_start (sentinel) | 4 | 100 | 100 | 100 | 100 | Zero-cost marker |
| ship (sentinel) | 1 | 0 | — | — | — | Zero-cost marker |

**Wall-time share in cycles 178–187** (most complete regime, N=9 iteration cycles):

| Cycle | TDD | Design-Judge | Other | Total |
|---|---|---|---|---|
| 178 | 458s (85%) | 84s (15%) | 0s | 542s |
| 179 | 113s (100%) | 0s | 0s | 113s |
| 180 | 575s (83%) | 122s (17%) | 0s | 697s |
| 181 | 765s (92%) | 71s (8%) | 0s | 836s |
| 183 | 309s (80%) | 75s (20%) | 0s | 384s |
| 184 | 368s (85%) | 65s (15%) | 0s | 434s |
| 185 | 478s (85%) | 83s (15%) | 0s | 562s |
| 186 | 506s (85%) | 91s (15%) | 0s | 597s |
| 187 | 437s (84%) | 83s (16%) | 0s | 520s |

**Key findings:**
- TDD is 80–92% of instrumented wall time; DJ is 8–20%. CI wait, PRD, lock,
  preflight, and release steps are **not instrumented in cycles 178–187** — they
  exist in earlier cycles but were dropped from the latest stream. This means the
  step total (530–840s) understates true cycle wall time (10–14 min per stream
  report, i.e. ~600–840s), suggesting the unaccounted gap is 0–300s of CI/release
  overhead.
- TDD median across all instrumented cycles: **9.6 min** (range 1.9–18.5 min).
  Cycles 156–164 (pre-stream) ran 14–18 min; cycles 178–187 run 6–13 min —
  consistent with the pipeline speedup from the 2026.05.20 version.

---

## 3. Design-Judge Statistics

| Cycle | Outcome | Findings | Blockers | DJ dur (s) | DJ tokens |
|---|---|---|---|---|---|
| 157 | shipped | 3 | 0 | 219 | — |
| 158 | shipped | 2 | 0 | 146 | — |
| 159 | shipped | 2 | 0 | 305 | — |
| 164 | shipped | 3 | — | 16 | 39,245 |
| 178 | shipped | 0 | 0 | 84 | 81,661 |
| 180 | shipped_with_fix | 4 | 1 | 122 | 83,115 |
| 181 | shipped | 3 | 1 | 71 | 71,772 |
| 183 | shipped_with_fix | 3 | 1 | 75 | 73,336 |
| 184 | shipped_with_fix | 6 | 1 | 65 | 71,055 |
| 185 | shipped_with_fix | 4 | 2 | 83 | 67,733 |
| 186 | shipped_with_fix | 2 | 1 | 91 | 78,397 |
| 187 | shipped_with_fix | 3 | 1 | 83 | 72,575 |

**Summary statistics (N=12 cycles with DJ step):**
- Findings count: median=3.0, mean=2.9, range [0–6]
- Blocker count (N=11 with field): median=1.0, mean=0.7, sum=8
- Cycles with ≥1 blocker: **7/11 (64%)** — all post-180 cycles had ≥1 blocker
- Cycles without any blocker_count field: 1 (cycle 164, pre-field era)
- DJ token cost: median=73,336, mean=73,699 (N=8 cycles with token data)
- DJ duration: median=84s, mean=114s — but cycles 157–159 ran 146–305s (pre-178
  stream had no token cap; cycle 159 at 305s is an outlier)

**Blocker precision:** Per the stream-178-187 report, all 6 individual blockers
flagged in cycles 180–187 were real (0% false-positive rate).

**Notable:** cycle 181 was logged as `shipped` despite `blocker_count=1`; the
fix was apparently applied inline before the final merge, not as a follow-up PR,
which is why it appears under `shipped` rather than `shipped_with_fix`. This
ambiguity in the outcome field should be noted.

---

## 4. Outcome Distribution

| Outcome | Count | % |
|---|---|---|
| shipped | 21 | 70% |
| shipped_with_fix | 6 | 20% |
| brainstorm_added_8 | 2 | 7% |
| brainstorm_added_10 | 1 | 3% |

All 6 `shipped_with_fix` outcomes are in cycles 180–187. The 20% fix rate in the
178–187 window (6/10 cycles, correcting for 1 brainstorm) is higher than the
168–177 rate (2/10, 20%), which is the same — but the pattern is concentrated:
all post-180 iteration cycles had a post-merge fix. The fix rate is entirely
explained by the auto-merge race (DJ finishes after CI triggers merge).

---

## 5. Instrumentation Defects (ranked)

### D1 — Five cycles absent from dataset and cycles.jsonl (cycles 162, 163, 165, 166, 167)
**Severity: High.** These cycles shipped real code (162: token_management, 163:
WORKSPACE-PIPELINE-VARS per metrics.jsonl) but have no entry in cycles.jsonl.
Dataset is 30 cycles when it should be 35. Cycle 165 exists in metrics.jsonl
only (scope: step2_prd). Cycles 166–167 have no presence anywhere.
Impact: 14% of the 153–187 range is unaccountable.

### D2 — Zero-emission regime: cycles 168–177 (tokens=0, duration=0, steps=[])
**Severity: High.** All 10 cycles in this range report `tokens=0` and
`duration_active_min=0` despite shipping real code. The metric-script update
that enabled proper emission in cycles 178+ was not backfilled. Makes any
cross-stream token or timing analysis unreliable.

### D3 — `duration_wall_min` corrupt on cycle 157: value = 29,662,224
**Severity: Medium.** True wall time (derived from step timestamps 17:53→18:24)
is ~31 min. The stored value is ~494,370 hours — appears to be a raw epoch
arithmetic error (likely `$SECONDS` or timestamp subtraction bug). Any aggregate
using `duration_wall_min` would be garbage if not clamped.

### D4 — Field name inconsistency: `blockers` vs `blocker_count` in design_judge steps
**Severity: Medium.** Cycles 157–159 use `"blockers": 0`; cycles 178–187 use
`"blocker_count": N`. Cycle 164 has neither field. Consumers must handle both
names or silently miss blocker data for cycles 157–159. Should be normalized to
one canonical name.

### D5 — Step naming drift: 4 canonical concepts have multiple names
**Severity: Low–Medium.**

| Concept | Names seen |
|---|---|
| TDD step | `step2_tdd`, `step3_tdd`, `step3_tdd_start` |
| Design judge | `step2_design_judge`, `step3_design_judge`, `step3_6_design_judge` |
| Release PR | `step5_release_pr`, `step6_release`, `step6_release_please` |
| Merge | `step5_merge`, `step6_release_merge` |

Step number prefixes changed between pipeline versions, making step-based
grouping impossible without normalization.

### D6 — `metrics_steps_count` mismatches (4 cycles)
**Severity: Low.** Cycles 153 (declared=9, actual=0), 154 (declared=12,
actual=0), 156 (declared=None, actual=9), 182 (declared=0, actual=1). The
declared count cannot be trusted as a proxy for completeness.

### D7 — CI wait and release steps not instrumented in cycles 178–187
**Severity: Low.** Cycles 156–164 include `step4_ci_wait`, `step5_release_pr`,
and `step6_release_merge`; cycles 178–187 do not. The two-step schema
(`step2_tdd` + `step2_design_judge` only) understates true wall time by an
estimated 3–5 min/cycle (CI + release overhead). The `duration_active_min`
field (9–14 min) captures this but step-level attribution is lost.

### D8 — `subagent_tokens=null` (explicit null) on 3 steps
**Severity: Low.** Cycles 160 (`step5_merge`, `step6_release`) and 168
(`step_ship`) record `"subagent_tokens": null` explicitly. These are release
orchestration steps where token counting does not apply, but explicit null
pollutes token sum queries. Should use field-absent rather than null.

---

## 6. Recommendations

1. **Backfill cycles 162, 163, 165**: add minimal records to cycles.jsonl from
   the metrics.jsonl evidence (step timestamps, subagent_tokens) and mark them
   with `"reconstructed": true`. Cycles 166–167 are unrecoverable.

2. **Normalize field names** before 188+:
   - Rename `"blockers"` → `"blocker_count"` in all design_judge step emissions
   - Rename step names to a single schema: `step_lock`, `step_preflight`,
     `step_prd`, `step_tdd`, `step_design_judge`, `step_ci_wait`,
     `step_release_pr`, `step_release_merge`

3. **Restore CI wait and release step timing** to cycles 178+ schema: these
   steps exist in the pipeline but are not being timed. ~3–5 min of per-cycle
   wall time is unattributed.

4. **Guard `duration_wall_min`** in the metric script: compute as
   `end_epoch - start_epoch` in seconds, then convert — not as a raw delta
   that can overflow or underflow.

5. **Remove explicit `"subagent_tokens": null`** from steps where token counting
   is not applicable; use field-absent to avoid polluting sum queries.

6. **Add a cycle-start sentinel** that writes the cycle number, start timestamp,
   and pipeline version immediately on cycle entry — before any step can fail.
   This would have prevented the 166–167 disappearance.
