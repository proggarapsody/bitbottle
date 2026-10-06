> Source: synthesized from the 30-cycle analysis (cycles 153–187) in `auto-iter/reports/analysis-2026-06-02-cycles-158-187/`. Scope confirmed with maintainer: **process-correctness only** — metric/report observability deferred to a separate PRD.

## Problem Statement

When running `/auto-iter-stream`, the orchestrator arms GitHub auto-merge as soon as the feature PR's CI turns green (~2 min). The design-judge (DJ) review runs concurrently and returns its verdict later (~3–5 min). The result: the PR squash-merges to `main` *before* DJ has finished, so any BLOCKER DJ finds arrives **after** the code is already on `main`.

Across the most recent stream (cycles 178–187), this auto-merge race produced a post-merge BLOCKER in **6 of 10 cycles** — every one a real defect (DJ false-positive rate was 0%), each requiring a follow-up fix PR and an extra release bump. The post-merge defect rate regressed 3× versus the prior stream (20% → 56%). The recurring escapes split into two classes:

1. **DJ-class defects** that DJ does catch but only after merge — a missing MCP handler test, an MCP/CLI default mismatch, a duplicated scope-switch abstraction, a custom format helper reinventing `format.ConfigFromCmd`, a zero-value `flag.Changed()` bug.
2. **Missing tier-2 integration tests** (cycles 185, 187) — new user-visible commands shipped without a `*_integration_test.go`, which `ARCHITECTURE.md` §"Test tiers" requires.

Compounding the problem: the current spec in `docs/workflows/iteration-cycle/quickref.md` **explicitly documents the race as an "Accepted tradeoff,"** and the corrective rules the maintainer has since articulated live only in gitignored Claude memory — so any agent reading the tracked spec keeps arming auto-merge on CI-green and keeps omitting integration tests. The documentation actively endorses the broken behavior.

## Solution

Make the iteration loop **gate auto-merge on the design-judge returning SHIP**, instead of racing it against CI. The orchestrator must hold the PR un-armed until DJ has produced a verdict; only a SHIP verdict permits arming `gh pr merge --auto`. A BLOCKER verdict routes to the existing fix-then-re-review path *before* any merge can happen.

This invariant is enforced by a small, testable gate rather than left to orchestrator discipline, and the tracked workflow spec is rewritten to describe the new sequence (deleting the "Accepted tradeoff" rationale that endorses the race). The two corrective rules currently stranded in gitignored memory are relocated into the tracked spec so every agent — not just one with warm memory — follows them.

Separately, the pre-merge mechanical gate gains a check that fails when a new command package lacks a tier-2 integration test, so the 185/187 omission class is caught locally before push rather than by DJ after merge.

From the maintainer's perspective: a 10-cycle stream should land with **zero** post-merge BLOCKER follow-up PRs, at the cost of ~3–5 min of DJ wall-clock added to each clean cycle (the cycles that previously "won" the race and merged early).

## User Stories

1. As the loop maintainer, I want auto-merge to be armed only after the design-judge returns SHIP, so that no DJ-detectable defect can reach `main` before review completes.
2. As the loop maintainer, I want a BLOCKER verdict to block the merge entirely, so that the fix lands in the same PR instead of a follow-up PR plus an extra release bump.
3. As the loop maintainer, I want the merge-gating invariant enforced by a script rather than by orchestrator discipline, so that the rule holds even when the orchestrator's context is degraded or summarized.
4. As the loop maintainer, I want the gate to refuse to arm when the DJ verdict artifact is missing or malformed, so that a skipped or crashed DJ step can never silently fall through to a merge.
5. As the loop maintainer, I want the tracked `quickref.md` "Accepted tradeoff" section rewritten to describe the DJ-then-arm sequence, so that an agent reading the spec implements the correct behavior.
6. As the loop maintainer, I want the merge sequence in `README.md` §3.6 and the parallel block in `.claude/commands/auto-iter.md` updated to match, so that the three spec surfaces agree.
7. As the loop maintainer, I want the auto-merge-race rule and the integration-test rule relocated from gitignored Claude memory into the tracked spec, so that the rules bind every agent and not just a session with warm memory (per `feedback_agent_rules_location.md`).
8. As the loop maintainer, I want the pre-merge mechanical gate to fail when a new command package has no `*_integration_test.go`, so that the missing-integration-test class (cycles 185, 187) is caught before push.
9. As the loop maintainer, I want the integration-test requirement stated explicitly in the TDD subagent prompt, so that the subagent writes the test in the first place rather than relying on the gate to reject it.
10. As the loop maintainer, I want the new merge gate to ship with a paired `_test.sh`, so that it matches the convention that every `auto-iter/scripts/*.sh` is independently tested.
11. As the loop maintainer running a stream, I want the per-cycle wall-clock cost of DJ-gating to be bounded and documented, so that I can decide whether the throughput tradeoff is acceptable for a given stream.
12. As an agent executing a cycle, I want a single unambiguous command that tells me whether I may arm auto-merge, so that I don't have to re-derive the ordering rule from prose each cycle.
13. As an agent executing a cycle, I want the gate's refusal to carry a machine-readable reason (blocker vs missing-verdict vs malformed), so that I route to the correct recovery path.
14. As the loop maintainer, I want a security-class BLOCKER to be impossible to merge-then-fix, so that vulnerable code never appears on `main` even transiently.
15. As the loop maintainer, I want the change to leave the existing release-publish halt untouched, so that the only behavioral change is the merge-arming order, not the release flow.
16. As the loop maintainer, I want a `refactor:`/`docs:`/`chore:` cycle (which still runs DJ but produces no release) to pass through the same gate, so that the rule has no carve-outs that could drift.
17. As the loop maintainer, I want the gate to be a no-op-cost addition for cycles that have no BLOCKER, so that the only price paid is wall-clock ordering, not extra tokens.

## Implementation Decisions

**Module 1 — `merge-gate.sh` (new deep module, `auto-iter/scripts/`).**
- Encapsulates the single ordering invariant: *auto-merge may be armed only when a SHIP verdict exists.*
- Interface: `merge-gate.sh --pr <N> --verdict <path-to-dj-verdict>`. Emits one JSON object on stdout (matching the established script convention — every script emits a single JSON object). On a SHIP verdict it performs/permits the `gh pr merge --auto --squash` arming and exits 0. On a BLOCKER, missing, or malformed verdict it exits non-zero and emits `{"armed": false, "reason": "blocker"|"missing_verdict"|"malformed_verdict", ...}` so the orchestrator routes correctly.
- The verdict artifact is the DJ subagent's structured output (it already records `findings_count` / `blocker_count`); the gate reads SHIP ⇔ `blocker_count == 0`. The DJ subagent writes this artifact to a known path in the cycle's working area.
- Rationale for a script (not pure prose): the analysis showed the prose rule was documented yet not followed because the *spec endorsed the opposite*. A gate makes the invariant load-bearing and testable, independent of orchestrator context.

**Module 2 — `pre-merge-mechanical.sh` (modify existing).**
- Add a check (a new numbered section alongside the existing §4 BACKLOG/SHIPPED check): if the diff introduces a new command package under the command tree and that package has no `*_integration_test.go`, emit a BLOCKER-class finding and fail the gate.
- Heuristic boundary: scoped to *new* user-visible command packages (the 185/187 class). Pre-existing packages without integration tests are out of scope for this check — the broader 19%-coverage backfill is a separate effort (see Out of Scope).

**Module 3 — Spec rewrite (docs, no code).**
- `docs/workflows/iteration-cycle/quickref.md`: replace the "Accepted tradeoff (auto-merge vs design-judge timing)" block (currently endorsing CI-green arming) with the DJ-then-arm sequence. Update the §"feature PR is auto-merged once CI is green AND all gates pass" line to "...once DJ returns SHIP, then CI is green."
- `docs/workflows/iteration-cycle/README.md` §3.6 / merge step: document the corrected sequence.
- `.claude/commands/auto-iter.md`: update the in-cycle parallel block to gate arming on DJ, and add the explicit integration-test requirement to the TDD subagent prompt ("if the target package already has `*_integration_test.go`, your command needs one too").

The corrected sequence (the load-bearing decision, stated precisely):

```
push feature branch
  → DJ subagent reviews → writes verdict artifact
  → merge-gate.sh --pr N --verdict <artifact>
        BLOCKER → dispatch fix subagent → re-review → re-gate   (no merge yet)
        SHIP    → gate arms `gh pr merge --auto --squash`
                  → await CI green → PR squash-merges
```

**Module 4 — Rule relocation (docs, no code).**
- Move the substance of the gitignored memory rules `feedback_auto_merge_race.md` and `feedback_integration_test_required.md` into the tracked spec (`quickref.md` anti-patterns + the relevant README/auto-iter sections). Memory may keep a one-line pointer, but the binding text lives in tracked files per `feedback_agent_rules_location.md`.

**Non-decisions / constraints carried in:**
- The orchestrator-is-shell-only rule still holds: `merge-gate.sh` and `gh` calls are shell orchestration, not code writes; all code/test authoring still dispatches to subagents.
- The release-publish halt and release-please flow are unchanged.
- `pipeline_version` should be bumped when this lands, since the step sequence changes (the analysis noted the version string has been frozen across incompatible schemas).

## Testing Decisions

A good test here asserts **external behavior** — the JSON the script emits and its exit code for a given verdict input — not the script's internal control flow. Tests feed a fixture verdict artifact and a PR number, then assert on stdout JSON + exit status. They must not call live `gh`; the arming side-effect is stubbed/guarded so tests run hermetically (the existing `await-ci_test.sh` and `pre-merge-mechanical_test.sh` establish the prior art for stubbing external calls and asserting emitted JSON).

Modules to be tested (confirmed with maintainer):

1. **`merge-gate_test.sh`** (new, paired with Module 1). Cases:
   - SHIP verdict (`blocker_count == 0`) → `{"armed": true}`, exit 0, arming command invoked exactly once.
   - BLOCKER verdict (`blocker_count > 0`) → `{"armed": false, "reason": "blocker"}`, exit non-zero, arming **not** invoked.
   - Missing verdict file → `{"armed": false, "reason": "missing_verdict"}`, exit non-zero.
   - Malformed/non-JSON verdict → `{"armed": false, "reason": "malformed_verdict"}`, exit non-zero.
   - This is the highest-value test: it pins the core invariant directly.

2. **`pre-merge-mechanical_test.sh`** (extend existing, paired with Module 2). Cases:
   - Diff adds a new command package with no `*_integration_test.go` → BLOCKER finding, gate fails.
   - Diff adds a new command package *with* `*_integration_test.go` → pass.
   - Diff touches only a pre-existing package lacking integration tests → no new finding (boundary: don't retroactively block unrelated packages).

Prior art: `auto-iter/scripts/await-ci_test.sh`, `pre-merge-mechanical_test.sh`, `metric_test.sh` — same harness, same "feed input → assert JSON + exit code" shape.

Docs modules (3 and 4) are not unit-tested; their correctness is verified by review against the corrected-sequence diagram above.

## Out of Scope

- **Metric instrumentation repair** — the epoch-as-delta corruption (cycles 157, 165), `blockers` vs `blocker_count` field drift, step-name drift, and the 168–177 zero-emission outage. Tracked separately as the observability PRD.
- **Report-validation tooling** — a script to check stream-report tables against `dataset.json` (the 178–187 report had factual errors). Separate PRD.
- **DJ fixed-context trimming** — the ~75K-token-per-cycle DJ cost reduction. Needs empirical validation that it preserves the 0% false-positive rate; separate investigation.
- **Backfilling integration tests for the existing 81% of command packages** that lack them. Module 2 only blocks *new* omissions; the backlog backfill is its own scope.
- **Recovering the lost cycles** (162, 163, 165, 166, 167) — historical data, not recoverable by this change.

## Further Notes

- This PRD reverses a decision the spec currently presents as deliberate. The original "Accepted tradeoff" reasoning was that DJ adds ~3–5 min to the 60%+ of cycles with zero BLOCKERs, so follow-up-PR-on-BLOCKER was faster for throughput. The 178–187 data refutes the premise: BLOCKERs hit 60% of cycles that stream, so the "rare BLOCKER" assumption no longer holds and the follow-up-PR tax (≈5 extra PRs, ~50 min, ~450K tokens per stream) exceeds the DJ-gating cost.
- Expected impact (from the analysis): eliminate ~4–5 follow-up fix PRs per 10-cycle stream; net wall-clock roughly neutral-to-positive once avoided fix PRs are counted.
- Triage: applied the repo's `prd` label (the project's actual PRD triage label). The skill's generic `ready-for-agent` label does not exist in this tracker.
- Full supporting analysis: `auto-iter/reports/analysis-2026-06-02-cycles-158-187/00-FINAL-REPORT.md` (Fixes 1, 2, 3).
