# Report dimension audit — cycles 158–187 (2026-06-02)

Dimension: **REPORTS** — quality, accuracy, and usefulness of post-stream analysis reports.
Audited reports: `stream-2026-05-24-cycles-135-144.md`, `stream-2026-05-25-cycles-145-154.md`,
`stream-2026-05-29-cycles-168-177.md`, `stream-2026-06-01-cycles-178-187.md`.
Ground truth: `analysis-2026-06-02-cycles-158-187/dataset.json` and `.claude/auto-iter/cycles.jsonl`.

---

## TL;DR

Four stream reports cover cycles 135–187, with a hard 13-cycle documentation gap at 155–167
that has no report and 5 cycles (162, 163, 165, 166, 167) missing from the underlying data
entirely. Naming convention compliance is 100%. Section compliance improved sharply across
the four reports (0/6 required sections in the earliest two vs 6/6 in 168–177), but the most
recent report (178–187) regresses on two sections. The 178–187 report contains four factual
errors traceable to ground-truth data: the follow-up-fix count is 5 but should be 6; the
token-emission rate is 8/10 but should be 9/10 (or 10/10 counting metric-step tokens); the
median token figure is 180K but the correct value is 188.1K; cycle 181 is described as
"clean" but had a DJ BLOCKER caught pre-merge. Recommendations across all reports are
well-ranked and actionable but lack a closed-loop section: no report explicitly confirms
which prior fixes were applied before the stream it describes.

---

## Coverage map

| Cycle range | Report file | Status |
|---|---|---|
| 135–144 | `stream-2026-05-24-cycles-135-144.md` | Documented |
| 145–154 | `stream-2026-05-25-cycles-145-154.md` | Documented (cycles 153–154 are the tail) |
| **155–167** | *(none)* | **UNDOCUMENTED GAP — 13 cycles** |
| 168–177 | `stream-2026-05-29-cycles-168-177.md` | Documented |
| 178–187 | `stream-2026-06-01-cycles-178-187.md` | Documented |

### Gap detail: cycles 155–167

Thirteen consecutive cycles (2026-05-25 → 2026-05-29) have no stream report. Of these:

| Cycle | Scope (from dataset.json / cycles.jsonl) | Data present? |
|---|---|---|
| 155 | brainstorm (+10 scopes) | Yes |
| 156 | PIPE-CONFIG+SSH-KEY-SERVER | Yes |
| 157 | PIPE-TEST-REPORTS+BRANCH-COMPARE | Yes |
| 158 | REPO-DOWNLOAD+MILESTONES | Yes |
| 159 | ISSUE-VERSIONS+WORKSPACE-PROJECT-CRUD | Yes |
| 160 | REPO-MIRROR+WORKSPACE-PERMS | Yes |
| 161 | brainstorm (+8 scopes) | Yes |
| 162 | unknown | **No — missing from dataset.json and cycles.jsonl** |
| 163 | unknown | **No — missing from dataset.json and cycles.jsonl** |
| 164 | REPO-CLONE | Yes |
| 165 | unknown | **No — missing from dataset.json and cycles.jsonl** |
| 166 | unknown | **No — missing from dataset.json and cycles.jsonl** |
| 167 | unknown | **No — missing from dataset.json and cycles.jsonl** |

Five cycles (162, 163, 165, 166, 167) are unrecoverable from available data: no record exists
in either `dataset.json` or `cycles.jsonl`. Eight cycles have structured data and could
support a belated report, but the five dark cycles mean any such report would need to note
an unrecoverable sub-gap.

The gap matters for trend analysis: the 168–177 report's historical comparison table skips
directly from 145–154 data to 168–177, leaving the 155–167 window as an unknown era when
interpreting whether dispatch discipline was consistent before cycles 168–169.

---

## Naming convention compliance

All four report files conform exactly to the `<kind>-<YYYY-MM-DD>-<descriptor>.md` convention
specified in `auto-iter/reports/README.md`. The date in each filename matches the UTC start of
the analyzed cycle window (not the report generation date), as required.

| File | Kind | Date | Descriptor | Compliant |
|---|---|---|---|---|
| `stream-2026-05-24-cycles-135-144.md` | `stream` | 2026-05-24 | `cycles-135-144` | Yes |
| `stream-2026-05-25-cycles-145-154.md` | `stream` | 2026-05-25 | `cycles-145-154` | Yes |
| `stream-2026-05-29-cycles-168-177.md` | `stream` | 2026-05-29 | `cycles-168-177` | Yes |
| `stream-2026-06-01-cycles-178-187.md` | `stream` | 2026-06-01 | `cycles-178-187` | Yes |

---

## Section compliance

Required sections per README: TL;DR, Throughput, Code quality, Predictability & consistency,
Recommended fixes, Bottom line.

| Report | TL;DR | Throughput | Code quality | Predictability & consistency | Recommended fixes | Bottom line | Score |
|---|---|---|---|---|---|---|---|
| 135–144 | No (opens with "Data quality preface") | No (labeled "Per-cycle facts") | No (labeled "Confirmed structural findings") | No (embedded in findings) | Yes | Yes | 2/6 |
| 145–154 | No (opens with "Data quality preface") | No (labeled "Per-cycle facts") | No (labeled "Design-judge outcomes") | No (not labeled) | Yes (labeled "Recommendations") | Yes | 2/6 |
| 168–177 | Yes | Yes | Yes | Yes (labeled "Predictability and algorithm consistency") | Yes | Yes | **6/6** |
| 178–187 | Yes | Yes | No (replaced by "Design-judge accuracy") | No (replaced by "Recurring patterns") | Yes | Yes | 4/6 |

### Notes

- The 135–144 and 145–154 reports predate the current README section spec. The README's
  "minimum sections" list was codified after those reports were written, so the non-compliance
  is retrospective rather than a regression.
- The 178–187 report notably drops two sections that 168–177 had, replacing them with
  purpose-specific sections ("Design-judge accuracy" and "Recurring patterns"). Those
  replacements are substantively richer for their respective topics but break the predictable
  structure a reader expects when scanning multiple reports.
- The 168–177 report is the only one that achieves full section compliance.

---

## Accuracy audit

### Report: `stream-2026-06-01-cycles-178-187.md`

Ground truth: `dataset.json` cycle records with `_metric_steps` and top-level `tokens` fields.

| Claim | Report says | Ground truth | Verdict |
|---|---|---|---|
| Token emission rate | "8/10 cycles" | 9/10 cycles have non-zero top-level `tokens` field (only cycle 182 is 0); cycle 182 additionally has 100.5K subagent tokens in `step2_brainstorm` | **WRONG** — should be 9/10 (or 10/10 counting step-level) |
| Median tokens/cycle | "180K (n=8)" | n=8 feature cycles: 150.3K, 161.3K, 162.8K, 185.5K, 190.7K, 218.6K, 220.5K, 235.4K → median = (185.5+190.7)/2 = **188.1K** | **WRONG** — off by 8.1K (~4.5%) |
| Follow-up fix PRs | "5 (cycles 180, 183–187)" | 180, 183, 184, 185, 186, 187 = **6 cycles** with `outcome=shipped_with_fix` | **WRONG** — the list (180, 183–187) contains 6 cycles, not 5 |
| Post-merge BLOCKER rate | "5/10 (50%)" | 6/10 = 60% | **WRONG** — cascades from fix-count error |
| Regression vs prior era | "+150% regression" | 6/10 vs 2/10 = +200% regression | **WRONG** — cascades from fix-count error |
| Auto-merge race pattern | "5 of 10 cycles" | 6 of 10 cycles | **WRONG** — cascades |
| Cycle 178 tokens | 162.8K | 162,816 | Correct (within rounding) |
| Cycle 180 tokens | 218.6K | 218,592 | Correct |
| Cycle 181 tokens | 220.5K | 220,483 | Correct |
| Cycle 182 tokens | 100.5K | top-level=0; step2_brainstorm subagent_tokens=100,526 | Correct if sourced from step-level; inconsistent with top-level field |
| Cycle 183–187 tokens | 150.3K, 161.3K, 190.7K, 235.4K, 185.5K | match dataset within rounding | Correct |
| Cycle 181 description | "clean" | DJ `step2_design_judge`: findings_count=3, blocker_count=1; outcome=`shipped` | **MISLEADING** — DJ found 1 BLOCKER, fixed pre-merge; not "clean" |
| Release count | "8 releases (v1.128.0 → v1.135.0)" | 8 distinct release versions (178, 180, 181, 183, 184, 185, 186, 187) | Correct |
| DJ section header | "All 5 post-merge BLOCKERs were real issues" | 6 cycles with shipped_with_fix | **WRONG** — should be 6 |
| DJ section footer | "All 6 individual BLOCKERs were confirmed real issues" | Cycle 185 has blocker_count=2 in dataset, all others=1 → total from dataset=7 | Partial discrepancy — 6 is what the BLOCKER table lists; dataset suggests 7 |

#### Internal inconsistency in the DJ section

The report uses three different numbers for the same count in the same section:
- Header: "All **5** post-merge BLOCKERs were real issues"
- BLOCKER table: **6 rows** (cycles 180, 183, 184, 185, 186, 187)
- Footer: "All **6** individual BLOCKERs were confirmed"

The root cause: the author intended to say "5 cycles with follow-up fix PRs, containing 6
individual BLOCKERs" but the cycle count itself is wrong (6 cycles, not 5), and cycle 185
had 2 individual BLOCKERs (blocker_count=2 in dataset) while the table shows only 1 row
for cycle 185.

### Report: `stream-2026-05-29-cycles-168-177.md`

Data for cycles 168–177 is sparse in `cycles.jsonl` (all show `tokens=0`,
`duration_active_min=0`, minimal metric steps). The report acknowledges this explicitly
("Token data was not emitted to cycles.jsonl for any of cycles 168–177"). Claims in this
report that are drawn from the user-supplied factual record (subagent tokens) rather than
automated logging cannot be independently verified from `dataset.json`. No factual
contradictions were found between the report and available logged data.

One caveat: the report's release count of "10 releases (v1.126.0 → v1.127.5)" cannot be
confirmed from `dataset.json` or `cycles.jsonl` because cycles 172–177 have no release
version logged. This is a data-quality gap in the underlying log, not a report fabrication.

### Reports: `stream-2026-05-24-cycles-135-144.md` and `stream-2026-05-25-cycles-145-154.md`

Both reports include explicit "Data quality preface" sections that warn where claims are
statistically weak (n=3, n=4). The 135–144 report even enumerates claims that were removed
from draft because the evidence didn't support them. This is exemplary epistemic hygiene.
No fact-checkable claims in these two reports were contradicted by available data.

---

## Usefulness assessment

### Actionability of recommendations

| Report | Fix count | Ranked? | Files specified? | Effort estimated? | ROI label? |
|---|---|---|---|---|---|
| 135–144 | 5 | Yes (evidence-weighted) | Partially | No | No |
| 145–154 | 5 | Yes (evidence + fix size) | Partially | No | No |
| 168–177 | 4 | Yes (ROI) | Yes (filename + section) | Yes (LOC) | Yes |
| 178–187 | 3 | Yes (ROI) | Yes (filename + section) | Yes (LOC) | Yes |

The 168–177 and 178–187 reports are substantially more actionable than their predecessors:
each fix names the exact files and sections that need changing, gives a line-of-code estimate,
and assigns a ROI label (HIGH/MEDIUM/LOW). The 135–144 and 145–154 recommendations are
still useful but require the reader to go find the files themselves.

### Closed-loop tracking (were prior fixes acknowledged?)

No report contains an explicit "Previous recommendations — status" section. Closed-loop
coverage is implicit and incomplete:

| Fix | Source report | Status in next report |
|---|---|---|
| Taste-check FP skip | 135–144 (Fix 1) | Never mentioned in any subsequent report |
| Auto-merge ordering | 135–144 (Fix 2) | 145–154 implicitly confirms it worked (0 race events); 168–177 notes regression |
| Token emission | 135–144 (Fix 3) | 145–154 and 168–177 both note still absent; 178–187 shows improvement but doesn't say when/how it was fixed |
| DJ-then-wait gate | 168–177 (Fix 1) | 178–187 shows race worsened (6/10 vs 2/10); the fix was not applied; 178–187 repeats the same recommendation nearly verbatim |
| Dispatch enforcement | 168–177 (Fix 2) | 178–187 notes 0 dispatch violations but doesn't credit this fix |
| Socket-drop recovery | 168–177 (Fix 4) | Not mentioned in 178–187 |

The DJ-then-wait gate fix is the starkest gap: it was the #1 priority in 168–177, was not
applied before the next stream, the problem worsened substantially (from 2 to 6 post-merge
BLOCKERs), and 178–187 re-recommends an almost identical fix without noting it was
previously recommended and skipped. A reader comparing the two reports in sequence would
see this, but someone reading only the latest report would not know it had already been
flagged once.

### Trend visibility

The historical comparison tables added in 168–177 and 178–187 are a significant usefulness
improvement over the earlier reports. However, the comparison tables implicitly skip the
155–167 era (the undocumented gap), which means the reader cannot tell whether the dispatch
violations in cycles 168–169 were a new regression or continuation of a multi-stream pattern.

---

## Recommendations

1. **Write a belated report for cycles 155–167** (or explicitly acknowledge the gap in the
   analysis-2026-06-02 analysis). Eight of thirteen cycles have structured data; the five dark
   cycles (162, 163, 165, 166, 167) should be noted as unrecoverable. Without this, the
   168–177 historical comparison table has an invisible 13-cycle blind spot.

2. **Fix the four factual errors in `stream-2026-06-01-cycles-178-187.md`**:
   - Token emission: 8/10 → 9/10 (cycle 179 top-level `tokens=55860` is non-zero)
   - Median tokens: 180K → 188.1K (recalculate from 8 feature cycles)
   - Follow-up fix count: 5 → 6 (cycles 180, 183, 184, 185, 186, 187)
   - Cycle 181 description: "clean" → "DJ BLOCKER caught pre-merge; 3 findings incl. 1 BLOCKER; fixed before merge"

3. **Add a "Prior recommendations — status" section** to every stream report. Even two lines
   suffices: `Fix 1 from 168–177 (DJ gate): NOT applied — race worsened to 6/10 cycles`.
   Without this, recurring recommendations create no accountability signal.

4. **Standardize sections in future reports** to match the 168–177 template (6/6 sections).
   The 178–187 report dropped "Code quality" and "Predictability & consistency" in favor of
   domain-specific sections. The information is present but the section names no longer match
   the README spec, which creates friction for automated or structured reading.

5. **Require the token-emission source to be noted** when tokens are sourced from
   step-level data vs top-level fields. Cycle 182 in the 178–187 report shows 100.5K in the
   table but `tokens=0` in `cycles.jsonl`; the report's Fix 3 recommends populating the
   top-level field, but the body of the report treats step-level tokens as equivalent
   without flagging the distinction.
