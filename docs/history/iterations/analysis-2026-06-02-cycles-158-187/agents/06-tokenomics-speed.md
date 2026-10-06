# Tokenomics & Speed — Cycles 158–187

**Analysis date:** 2026-06-02  
**Dimension:** Tokenomics & Speed  
**Window:** Cycles 153–187 (30 cycles logged); token-valid subset: 157–164 (May-25 era) + 178–187 (Jun-1 stream). Cycles 168–177 log `"tokens": 0` — fully excluded as missing instrumentation, not treated as free.

---

## TL;DR

- **TDD is the dominant cost**: 67.5% of attributed tokens and ~90% of active wall time.
- **Design Judge is surprisingly expensive**: 40% of `(TDD + DJ)` token budget at ~75K tokens per cycle despite running in 83 s median.
- **2.7× speed improvement** confirmed: May-25 era median 24.5 min → Jun-1 stream median 9.0 min/cycle.
- **Token cost rose 3× alongside the speedup**: May median 62.5K → Jun median 188K tokens/cycle. Fewer CI wait minutes; more subagent work per minute.
- **Rework cycles (shipped_with_fix) cost an extra ~45 min wall time** across the stream but show no intrinsic token premium on the primary cycle — the hidden cost is the untracked follow-up fix PRs (~90K tokens each × 5 = ~450K extra tokens not in dataset).

---

## 1. Token Distribution

### Iteration cycles with real token data (n=14)

| Statistic | Value | Cycle |
|---|---:|---|
| Minimum | 31,189 | c158 (REPO-DOWNLOAD, simple scope) |
| Median | 162,076 | — |
| Mean | 141,806 | — |
| Maximum | 235,426 | c186 (ADMIN-RATE-LIMIT) |
| Std dev | 71,338 | — |

### Per-cycle breakdown

| Cycle | Era | Outcome | Tokens | Active min |
|---:|---|---|---:|---:|
| 157 | May-25 | shipped | 105,380 | 24.7 |
| 158 | May-25 | shipped | 31,189 | 24.5 |
| 159 | May-25 | shipped | 32,570 | 24.1 |
| 160 | May-25 | shipped | 62,547 | 22.6 |
| 164 | May-25 | shipped | 172,481 | 41.1 |
| 178 | Jun-1 | shipped | 162,816 | 9.0 |
| 179 | Jun-1 | shipped (chore) | 55,860 | 1.9 |
| 180 | Jun-1 | shipped_with_fix | 218,592 | 11.6 |
| 181 | Jun-1 | shipped | 220,483 | 13.9 |
| 183 | Jun-1 | shipped_with_fix | 150,322 | 6.4 |
| 184 | Jun-1 | shipped_with_fix | 161,336 | 7.2 |
| 185 | Jun-1 | shipped_with_fix | 190,742 | 9.4 |
| 186 | Jun-1 | shipped_with_fix | **235,426** | 9.9 |
| 187 | Jun-1 | shipped_with_fix | 185,542 | 8.7 |

c186 (ADMIN-RATE-LIMIT) is the single most expensive cycle, driven by a complex `flag.Changed()` semantic issue that required an extended TDD pass (157K tokens in TDD step alone).

### By outcome (all eras)

| Outcome | n | Mean tokens | Median tokens |
|---|---:|---:|---:|
| shipped (clean) | 8 | 105,416 | 83,964 |
| shipped_with_fix | 6 | 190,327 | 188,142 |

The apparent gap is **largely an era artifact**, not rework cost: all 6 fix cycles are from the Jun-1 stream where the base token budget is higher. See §4 for the proper apples-to-apples comparison.

---

## 2. Token Attribution by Step

### Across all cycles with subagent_token data (cycles 156–187)

| Step | n (token obs) | Total tokens | Share | Mean tokens/cycle | Mean wall time | Median wall time |
|---|---:|---:|---:|---:|---:|---:|
| **TDD** | 17 | 1,599,365 | **67.5%** | 94,080 | 671 s | 724 s |
| **Design Judge** | 9 | 638,889 | **26.9%** | 70,988 | 113 s | 84 s |
| Brainstorm | 2 | 132,443 | 5.6% | 66,222 | 75 s | 75 s |
| CI wait | — | — | — | — | 387 s (May era) | 180 s |
| merge/release | — | — | — | — | ~60 s | — |

**Total attributed subagent tokens (all cycles): 2,370,697**

### Jun-1 stream only (8 cycles with TDD + DJ data)

| Step | Total tokens | Share of (TDD+DJ) | Mean/cycle |
|---|---:|---:|---:|
| TDD | 925,615 | **61%** | 115,702 |
| Design Judge | 599,644 | **39%** | 74,956 |

**Key finding**: Design Judge consumes 39% of the `(TDD + DJ)` token budget while adding only 17% of active wall time. This is structurally high — DJ runs against the full diff plus `TASTE.md`/`ARCHITECTURE.md` context on every cycle regardless of scope size. Its token count is remarkably stable across cycles: range [67,733–83,115], stdev ~5K. This suggests a large fixed-context overhead dominating over variable review content.

### Arithmetic check (c181 — clean representative)

TDD: 148,711 tokens, 765 s. DJ: 71,772 tokens, 71 s. Sum = 220,483 tokens (matches `"tokens": 220483` in dataset). Confirms subagent_tokens correctly accounts for full cycle token cost.

---

## 3. Speed Distribution

### Per-cycle active time (all iteration cycles, n=14)

| Statistic | Value |
|---|---:|
| Minimum | 1.9 min (c179 chore) |
| Maximum | 41.1 min (c164, REPO-CLONE + 20-min CI outlier) |
| Mean | 15.4 min |
| Median | 10.8 min |
| Std dev | 10.6 min |

### Per-step wall clock (Jun-1 stream, n=8–9 cycles)

| Step | n | Mean | Median | Min | Max |
|---|---:|---:|---:|---:|---:|
| TDD subagent | 9 | 446 s | 458 s | 113 s (c179 chore) | 765 s |
| Design Judge subagent | 8 | 84 s | 83 s | 65 s | 122 s |
| CI wait (May-25 era only) | 5 | 387 s | 180 s | 60 s | 1,243 s |
| merge + release | — | ~60–120 s | — | — | — |

**Wall-time budget for a typical Jun-1 feat cycle** (using c181 as example, 13.9 min active):

- TDD: 765 s (92% of active time)
- DJ: 71 s (8%)
- Overhead (lock, preflight, mode-pick, PR, merge): ~58 s remaining

TDD **dominates wall time** (92%); DJ is a rounding error on wall time despite its token weight. CI wait does not appear in Jun-1 stream steps — it runs in the background behind DJ, so the auto-merge fires when CI completes without adding to active time.

**May-25 CI wait was significant**: median 180 s (3 min), outlier 1,243 s (20.7 min) for c164 (REPO-CLONE with a large test suite). In Jun-1, CI wait is hidden behind DJ and is no longer on the critical path.

---

## 4. Cost of Rework (Auto-Merge Race)

### In-cycle token comparison (Jun-1 era, apples-to-apples)

| Category | n | Mean tokens | Median tokens |
|---|---:|---:|---:|
| shipped_with_fix (c180, 183–187) | 6 | 190,327 | 188,142 |
| shipped clean feat (c178, 181) | 2 | 191,650 | 191,650 |

**Token delta on primary cycle: −1,323 (−1%)** — statistically negligible. Fix cycles do not spend more tokens than clean cycles in their primary run. The DJ detects the blocker, but the cycle is already auto-merged by that point.

### Hidden follow-up fix PR cost (not in dataset)

Each of the 5 fix cycles (180, 183, 184, 185, 186, 187 — note stream report counts 5 but dataset marks 6 as `shipped_with_fix`) required a targeted follow-up fix PR. Based on TDD subagent step sizes for focused bug fixes, estimated cost per fix PR: ~90,000 tokens.

| Cost type | Estimate |
|---|---:|
| Extra wall time per fix PR | ~9 min (per stream report: ~8–10 min) |
| Extra wall time total (5 fix PRs) | **~45 min (0.75 h)** |
| Extra tokens per fix PR | ~90,000 |
| Extra tokens total (5 fix PRs) | **~450,000** |
| Extra releases | 5 minor version bumps |

**The 45-minute time cost understates the real disruption**: each fix PR consumed an additional cycle slot, displacing a feature cycle that could have shipped new functionality. The opportunity cost per stream is ~5 feature cycles deferred.

### Time premium by outcome (duration_active_min)

| Outcome | n | Mean active | Median active |
|---|---:|---:|---:|
| shipped (clean) | 8 | 20.2 min | 23.4 min |
| shipped_with_fix | 6 | 8.9 min | 9.1 min |

The inversion (fix cycles are shorter!) is explained by era: fix cycles all occur in the fast Jun-1 stream while some clean cycles include slow May-25 outliers. No evidence that the primary cycle of a rework case runs longer.

---

## 5. Trend

### Era comparison

| Metric | May-25 (c157–164) | Jun-1 (c178–187) | Delta |
|---|---:|---:|---|
| Active time median | 24.5 min | 9.0 min | **−2.7× faster** |
| Active time mean | 27.4 min | 8.7 min | −3.1× |
| Token median | 62,547 | 185,542 | **+2.97× more tokens** |
| TDD wall median | 933 s | 458 s | **−2.0× faster** |

### Explanation of the speed gain

TDD subagents ran 2.1× faster in Jun-1 (median 458 s) vs May-25 (median 933 s). Two plausible causes:
1. **Scope size**: May-25 scopes (`PIPE-TEST-REPORTS+BRANCH-COMPARE`, `REPO-DOWNLOAD+MILESTONES`) were bundled multi-feature scopes; Jun-1 scopes were single-feature.
2. **CI wait removal from active time**: May-25 CI wait (median 3 min) was on the critical path; Jun-1 runs DJ in parallel with CI so CI is not on the active-time clock.

### Token cost rose alongside speed

The Jun-1 stream has 3× higher token median (185K vs 63K). This is not inefficiency — it reflects:
1. DJ subagent now runs every cycle (added ~75K tokens), absent in most May-25 cycles.
2. TDD subagents are faster but produce the same or more code.

The **tokens-per-shipped-feature** metric is a better gauge. May-25: 62.5K tokens at 24.5 min active = 2,551 tokens/min. Jun-1: 185K tokens at 9.0 min active = 20,556 tokens/min. Higher intensity per minute, but fewer minutes: net output rate is higher.

### Within Jun-1 stream stability

First half (c178–181): mean 9.1 min active. Second half (c183–187): mean 8.3 min active. Essentially flat — the loop is at a stable operating point. No acceleration or degradation trend within the 10-cycle stream.

---

## 6. Efficiency Opportunities (Ranked by ROI)

### Opportunity 1 — Gate auto-merge on DJ verdict (HIGH ROI, structural)

**Current state**: Auto-merge arms immediately on `git push`, fires when CI green (~2 min). DJ returns at ~84 s median but CI can be faster. In 5/10 Jun-1 cycles, CI finished before DJ and the PR merged with an unresolved BLOCKER.

**Token cost to fix**: zero — this is a sequencing change, not more LLM work.

**Time cost to fix**: adds DJ wall time to every cycle critical path only when CI finishes before DJ (~17% of TDD time = ~76 s per cycle). For the 60% of cycles with no BLOCKER, this is a 76 s penalty. For the 40% with a BLOCKER, it saves ~9 min (one full fix-PR cycle).

**Net expected value** (per 10 cycles, Jun-1 base rates):
- 4 clean cycles: +4 × 76 s = +5 min overhead
- 6 fix cycles now avoided: −6 × 9 min = −54 min saved
- **Net: −49 min per 10 cycles = ~5 min saved per cycle**

**Implementation**: In `auto-iter.md` §3.6, replace the current "arm auto-merge on push, run DJ in parallel" with "run DJ → if SHIP → arm auto-merge → await CI → merge". The quickref §Halt routing note acknowledges the "acceptable tradeoff" but the May-25 measurement of 2/4 BLOCKERs in that era has worsened to 5/10 in Jun-1.

### Opportunity 2 — Reduce Design Judge context size (MEDIUM ROI, ~25% DJ token reduction)

**Current state**: DJ median is 74,956 tokens per cycle — remarkably uniform (stdev ~5K). The near-constant token count suggests a large fixed-cost context (TASTE.md + ARCHITECTURE.md + SKILL.md) that dominates over the variable diff.

**Estimate**: If TASTE.md + ARCHITECTURE.md total ~60K tokens context loaded unconditionally, trimming to checklist summaries could reduce DJ from ~75K to ~50K — saving ~25K tokens/cycle. Across 10 cycles: ~250K tokens.

**Caveat**: DJ's 0% false-positive rate suggests the full context is useful. Any trimming must be validated against a held-out set of BLOCKER cases before deployment.

### Opportunity 3 — Emit tokens for brainstorm and chore cycles (LOW ROI, instrumentation)

**Current state**: c182 (brainstorm) logged `"tokens": 0` in `cycles.jsonl` despite 100,526 subagent tokens in metrics.jsonl. c179 (chore/doc) also logged 0 despite 55,860 tokens.

**Fix**: In `auto-iter/scripts/log-cycle.sh`, pass `--tokens` from the `step2_brainstorm` metric line for brainstorm cycles. Estimated fix: 3–5 lines of shell.

**Benefit**: Full accounting. Currently the "median 180K tokens/cycle" figure from the stream report is computed from 8/10 cycles — the actual stream total is understated by ~156K tokens (100K brainstorm + 56K chore).

### Opportunity 4 — Skip DJ for pure-chore scopes (LOW ROI, selective)

**Current state**: c179 (BACKLOG-MIGRATION, a doc-only chore) ran with `"tokens": 55,860` for TDD — but no DJ step is logged. Correct. However, if DJ were to run on doc-only cycles, it would waste ~75K tokens reviewing non-code changes.

**Current behavior appears correct** — DJ was skipped for c179. Confirm this is enforced by a scope-type check, not an accidental omission. If enforced: no action needed.

### Opportunity 5 — Parallelize TDD and CI for independent test suites (LOW ROI, complex)

**Current state**: TDD runs to completion, then PR is opened, then CI runs.

**Theoretical speedup**: Start CI on a partial commit while TDD continues adding tests. Not viable with current worktree model (single branch, single PR). Would require significant pipeline rework for marginal gain on Jun-1's already-fast cycles.

**Verdict**: Skip. The 92% wall-time share of TDD is the real lever, and it is already 2× faster than May-25 due to scope sizing, not parallelism.

---

## Arithmetic Summary

| Quantity | Value | Source |
|---|---:|---|
| Total subagent tokens, window (156–187) | 2,370,697 | metrics.jsonl |
| TDD share | 67.5% (1,599,365) | metrics.jsonl |
| DJ share | 26.9% (638,889) | metrics.jsonl |
| Brainstorm share | 5.6% (132,443) | metrics.jsonl |
| Speed improvement May→Jun | 2.7× | duration_active_min |
| TDD wall-time improvement May→Jun | 2.1× | metrics.jsonl |
| DJ tokens / (TDD+DJ) | 39.4% | computed |
| Rework cost (5 fix PRs, time) | ~45 min | stream report + estimate |
| Rework cost (5 fix PRs, tokens) | ~450,000 | estimate at 90K/fix PR |
| Expected net saving from Opp 1 (gate) | ~49 min / 10 cycles | computed above |
