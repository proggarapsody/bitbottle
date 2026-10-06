# Final analysis — auto-iter cycles 153–187 (30 cycles)

**Generated:** 2026-06-02 · **Method:** 6 parallel review agents (Sonnet 4.6),
one per dimension, synthesized here. Atomic reports in [`agents/`](agents/).
Underlying data in [`dataset.json`](dataset.json) (merged from
`.claude/auto-iter/cycles.jsonl` + `metrics.jsonl`).

> **Window note:** the request was "last 30 cycles." The 30 most recent logged
> cycles span **153–187** (not 158–187 — the folder name predates the exact
> count). Cycles 162, 163, 165, 166, 167 are missing from the logs (see §3).

---

## TL;DR

- **Throughput is excellent, integrity is not.** 27 of 30 cycles shipped working
  code to `main`; the loop got **2.7× faster** (24.5 → 9.0 min/cycle) over the
  window. But only **37% of cycles have complete metrics**, 5 cycles vanished
  from the logs entirely, and the post-merge defect rate **regressed 3×**
  (20% → 56%) in the final stream.
- **One root cause dominates everything: the auto-merge race.** CI (~2 min) fires
  auto-merge before the design-judge (~3–5 min) returns its verdict. Result:
  **8 of 8 `shipped_with_fix` outcomes** in cycles 176–187, every one a *real*
  defect (DJ false-positive rate: **0%**) that escaped to `main` and needed a
  follow-up fix PR. The fix was recommended after stream 168–177, **never
  applied**, and the problem got worse.
- **The design-judge is carrying the loop.** Every escaped defect was the kind a
  careful reviewer catches in minutes (a forgotten test file, a default
  mismatch, a private helper reinventing a standard one). The architecture that
  ships is sound; *implementation discipline at push time* is degrading, and DJ
  is the only thing catching it — after merge.
- **The single highest-ROI fix is free**: gate auto-merge on DJ→SHIP instead of
  CI-green. It is a pure sequencing change (DJ already runs), would have
  prevented **all 6** recent escapes, and saves ~49 min + ~450K tokens per
  10-cycle stream.
- **The reports themselves have errors.** Three independent agents caught factual
  mistakes in the 178–187 stream report (5 vs 6 BLOCKERs, 8/10 vs 9/10 token
  emission, median 180K vs 188K) and a 13-cycle documentation gap (155–167).

---

## Scorecard

| Dimension | Grade | One-line verdict |
|---|:---:|---|
| 1. Metrics | **C** | 37% complete; 3 incompatible schemas; corrupt fields; 5 missing cycles |
| 2. Reports | **C+** | Fast & useful but factually wrong in places; no closed-loop fix tracking; 13-cycle gap |
| 3. Consistency | **C** | Tight *within* eras, chaotic *across* them; rework rate regressed 3× |
| 4. Code quality | **B** | Conventions hold; escapes are discipline lapses, not design flaws |
| 5. Guidelines | **B−** | Mostly followed; key rules live only in gitignored memory, so they don't bind |
| 6. Tokenomics & speed | **B+** | 2.7× faster, but 3× more tokens; TDD = 67% tokens / 92% wall time |

---

## The through-line: one defect, six symptoms

Every dimension's worst finding traces to the **same sequencing bug**, viewed
through a different lens:

| Lens | How the auto-merge race shows up |
|---|---|
| Consistency | 8/8 `shipped_with_fix`; rework rate 20% → 56% (3× regression) |
| Code quality | 6 real defects escaped to `main`; DJ became a post-hoc safety net |
| Guidelines | The "arm after DJ→SHIP" rule exists only in gitignored memory, so the *tracked* spec still says "arm on CI green" — agents obey the spec |
| Tokenomics | ~450K tokens + ~45 min of untracked follow-up-PR work per stream |
| Reports | Fix #1 recommended after 168–177, skipped, re-recommended after 178–187 with no "why was this not done" note |
| Metrics | All 6 blockers concentrated in 180–187; `blocker_count` proves DJ caught them — *after* the merge |

**Fixing the race collapses the worst finding in 5 of 6 dimensions at once.**

---

## Cross-validated facts (multiple agents agree)

- **DJ accuracy: 100% precision, 0% false positives** across all recorded
  blockers (metrics + code-quality + guidelines agents independently). DJ is the
  most trustworthy component in the system. The problem is *when* it runs, not
  *whether* it's right.
- **Conventions hold** (code-quality): `paging.Collect[T]`, `UseDomainErrors`
  wiring, `ContentTypeAlwaysWrite` propagation, and conventional commits are
  consistent across all new code. No architectural rot.
- **The 2.7× speedup is real but bought with tokens** (tokenomics + consistency):
  median 24.5 min/62.5K tokens (May-25 era) → 9.0 min/185.5K tokens (Jun-1).
  Faster TDD subagents + DJ added as a mandatory ~75K-token step.
- **TDD is the cost center**: 67.5% of tokens (1.6M of 2.4M), ~92% of wall time.
  Any efficiency work starts here. DJ is second at 26.9% with a near-constant
  ~75K/cycle (stdev ~5K) → large fixed context, not diff-proportional.

## Corrections to the existing 178–187 stream report

The reports agent fact-checked the prior report against `dataset.json`:

| Claim in `stream-2026-06-01-cycles-178-187.md` | Ground truth |
|---|---|
| Token emission "8/10 cycles" | **9/10** — cycle 179 logged 55,860 (non-zero) |
| Median tokens "180K" | **188.1K** (n=8 feature cycles) |
| "5 follow-up fix PRs" / "5 post-merge BLOCKERs" | **6** — cycles 180, 183, 184, 185, 186, 187. The report's own table has 6 rows but the header/footer say 5 then 6 (internal contradiction) |
| Cycle 181 "clean" | Dataset shows DJ `blocker_count=1` (fixed *pre*-merge, so not a fix-PR — but not "clean") |

These are minor individually but they show **reports are not being validated
against the logged data** before being committed.

---

## Recommended fixes (ranked by ROI)

### Fix 1 — Gate auto-merge on DJ→SHIP, not CI-green `[CRITICAL · ~20 LOC · free]`
The inverse of today's sequence. `push → await DJ verdict → if BLOCKER: fix &
re-run; else arm auto-merge → await CI → merge`. Prevents all 6 recent escapes;
saves ~49 min + ~450K tokens / 10-cycle stream; zero added LLM cost (DJ already
runs). **Must land in the *tracked* spec, not just memory.**
Files: `docs/workflows/iteration-cycle/README.md` §3.6,
`docs/workflows/iteration-cycle/quickref.md`, `.claude/commands/auto-iter.md`.

### Fix 2 — Land the memory-only rules into the tracked spec `[HIGH · docs]`
`feedback_auto_merge_race.md` and `feedback_integration_test_required.md` live
only in gitignored Claude memory — violating `feedback_agent_rules_location.md`.
Subagents follow the prompt/spec, not memory, so these rules **don't bind**.
Move them into `quickref.md` / `auto-iter.md`.

### Fix 3 — Integration-test checklist item in the TDD subagent prompt `[HIGH · 2 lines]`
Cycles 185, 187 omitted `*_integration_test.go`. Add to the TDD prompt: "If the
target package already has `*_integration_test.go`, your command needs one too —
see `pkg/cmd/repo/sync/sync_integration_test.go`." (Note: only 19% of command
packages have tier-2 tests overall — a broader gap worth a separate sweep.)

### Fix 4 — Fix metric instrumentation `[MEDIUM]`
Restore step emission (168–177 logged 0 steps; 178–187 logs only 2 of ~10).
Fix corrupt fields (cycle 157 `duration_wall_min`=29.6M; cycle 165 `step2_prd`
=1.78e12 ms — epoch-as-delta bugs). Unify field names (`blockers` vs
`blocker_count`) and step names (`step2_tdd` vs `step3_tdd`). Bump
`pipeline_version` when the schema changes — it's been a frozen string across 4
incompatible schemas.

### Fix 5 — Add a "Prior recommendations — status" section to every report `[LOW]`
Close the loop. Fix #1 was skipped silently between two streams. Each report
should state which prior fixes were applied/skipped and why.

### Fix 6 — Validate reports against `dataset.json` before commit `[LOW]`
The factual errors above are all mechanically checkable. A short script
(or a checklist step) comparing report tables to the JSON would catch them.

### Fix 7 (investigate) — Trim DJ fixed context `[LOW · needs validation]`
DJ's ~75K/cycle is dominated by fixed context (TASTE.md + ARCHITECTURE.md), not
the diff. A targeted trim could save ~250K tokens/stream — but only if the 0%
false-positive rate is preserved. Validate empirically before shipping.

---

## Per-cycle reference (153–187)

| Cycle | Scope | Outcome | Release | Tokens | Min |
|---:|---|---|---|---:|---:|
| 153 | PIPE-RERUN | shipped | — | — | 25 |
| 154 | CHERRY-PICK | shipped | — | — | 30 |
| 155 | (brainstorm +10) | brainstorm | — | — | 3 |
| 156 | — | shipped | — | — | — |
| 157 | — | shipped | — | 105.4K | 24.7 |
| 158 | — | shipped | — | 31.2K | 24.5 |
| 159 | — | shipped | — | 32.6K | 24.1 |
| 160 | — | shipped | — | 62.5K | 22.6 |
| 161 | (brainstorm +8) | brainstorm | — | 31.9K | 0.5 |
| *162–163* | — | *ran, not logged* | — | — | — |
| 164 | REPO-CLONE | shipped | — | 172.5K | 41.1 |
| *165* | — | *crashed (corrupt metric)* | — | — | — |
| *166–167* | — | *never run* | — | — | — |
| 168 | HOST-INFO | shipped | v1.126.0 | 0* | 0* |
| 169 | API-PARITY | shipped | v1.126.1 | 0* | 0* |
| 170 | OCC-AUDIT | shipped | v1.126.2 | — | — |
| 171 | CLOUD-DISCOVERY | shipped | v1.127.1 | 0* | 0* |
| 172 | SCRIPT-TRUST | shipped | — | 0* | 0* |
| 173 | FMT-CONTRACT | shipped | — | 0* | 0* |
| 174 | MCP-INPUT-VALIDATION | shipped | — | 0* | 0* |
| 175 | MCP-TAXONOMY | shipped | — | 0* | 0* |
| 176 | PR-GUARDS | shipped | — | 0* | 0* |
| 177 | CLOUD-WIRE | shipped | — | 0* | 0* |
| 178 | REF-UX | shipped | v1.128.0 | 162.8K | 9 |
| 179 | BACKLOG-MIGRATION | shipped | — | 55.9K | 1.9 |
| 180 | REPO-HOOK-SCRIPTS | **shipped_with_fix** | v1.129.0 | 218.6K | 11.6 |
| 181 | CLOUD-CODE-INSIGHTS | shipped (DJ blocker, fixed pre-merge) | v1.130.0 | 220.5K | 13.9 |
| 182 | (brainstorm +8) | brainstorm | — | — | — |
| 183 | DEPLOY-KEY-PERMISSION | **shipped_with_fix** | v1.131.0 | 150.3K | 6.4 |
| 184 | REPO-PIPELINE-VAR-VIEW | **shipped_with_fix** | v1.132.0 | 161.3K | 7.2 |
| 185 | REPO-SYNC | **shipped_with_fix** | v1.133.0 | 190.7K | 9.4 |
| 186 | ADMIN-RATE-LIMIT | **shipped_with_fix** | v1.134.0 | 235.4K | 9.9 |
| 187 | COMMIT-SEARCH | **shipped_with_fix** | v1.135.0 | 185.5K | 8.7 |

`*` = logged as zero but real work occurred (metric-emission outage 168–177).

---

## Bottom line

The loop **ships reliably and fast** — 27/30 cycles delivered working features
and the cadence more than doubled. Its weaknesses are **not in the code it
writes** (conventions hold, DJ is flawless) but in **process timing and
record-keeping**: a sequencing bug lets DJ run too late, and the
instrumentation/reporting that should catch this is itself unreliable.

**Do Fix 1 + Fix 2 before the next stream.** They are cheap, they're the same
fix viewed two ways (correct the sequence *and* put the rule where agents will
read it), and together they eliminate the defect that produced the worst finding
in five of the six dimensions reviewed.
