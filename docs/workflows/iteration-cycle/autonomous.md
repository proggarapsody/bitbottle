# Autonomous-mode behavioral deltas

This file documents behavioral deltas that apply when the iteration loop
runs in **autonomous mode** (no human present each cycle). These deltas
are agent-neutral — they describe *what* the loop does differently, not
*how* a specific agent delivers it.

Full procedure: [`README.md`](README.md).
Declarative tables (halt routing, outcome enum): [`quickref.md`](quickref.md).
Agent-specific delivery details live in the agent's own command file.

---

## §0a — Re-entry lock check

**Very first check, before anything else.** Verify no other autonomous cycle
is already running.

```bash
LOCK=.claude/auto-iter/.lock
mkdir -p "$(dirname "$LOCK")"
if [ -f "$LOCK" ]; then
  LOCK_AGE_S=$(( $(date +%s) - $(stat -f %m "$LOCK" 2>/dev/null || stat -c %Y "$LOCK") ))
  if [ "$LOCK_AGE_S" -lt 3600 ]; then
    # Peer cycle still alive — log skip_in_progress and exit cleanly.
    # Do NOT increment the cycle counter.
    exit 0
  fi
  # Stale lock (>60 min) — overwrite. Record stale_lock_age_min in step0a_lock metric.
fi
echo $$ > "$LOCK"
trap 'rm -f .claude/auto-iter/.lock' EXIT
```

**Stale-lock rule**: if the lock file is older than 60 minutes, overwrite it
(assume the predecessor crashed). Record `stale_lock_age_min` in the
`step0a_lock` metrics line so post-hoc analysis can spot crash-leaks.

The lock must be **removed at every exit path** — clean completion, `halt_*`
exits, errors. The `trap` above handles this for shell-based orchestrators.
Other runtimes must implement an equivalent cleanup guarantee.

---

## §1.5 — Already-shipped check

**Before any §2 work**, verify the scope hasn't already been implemented and
just mis-labelled in BACKLOG. This check prevents re-doing shipped work.

```bash
SCOPE_TAG=$(echo "$SCOPE" | tr 'a-z' 'A-Z')
if git log --all --grep="$SCOPE_TAG" --oneline | grep -qE "^[0-9a-f]+ (feat|fix|refactor)"; then
    ALREADY_SHIPPED=1
fi
```

If matched:
1. Flip the BACKLOG row 🔲 → ✅ in a single one-line chore PR.
2. Log `outcome: "skipped_already_shipped"`.
3. Increment the cycle counter (this counts as a ran cycle).
4. Skip §2 entirely.

**Rationale**: Cycles 77 (DEPGUARD-CI) and 79 (FAKECLIENT-SAFETY) in the
May-17 stream re-did work that had shipped two days earlier (PR #338 and #334).
Without this check each took ~6 min instead of ~30 sec.

---

## §7 — Auto-merge feature PR (skip halt for non-feat commits)

In autonomous mode, the feature PR is **auto-merged without a halt** once all
gates pass:
- CI is green
- Design-judge returned no BLOCKERs
- All pre-merge gate sections (§0–§9) pass

The **only halt** in a clean feat/fix cycle is at the irreversible release-
publish step (see [quickref.md § Halt routing](quickref.md#halt-routing)).

For `refactor:` / `docs:` / `chore:` cycles there is **no halt at all** —
these don't trigger release-please, so the cycle ends after feature merge with
outcome `shipped`.

**Why no halt for feature PR merge**: the gates already vetted content. The
previous per-PR halt in earlier versions of the loop was rubber-stamp in ~99%
of cases. The halt that matters is the irreversible one (release publish),
preserved as the single mid-cycle halt.

**Release-please halt timing**: the halt fires per release-please PR, NOT per
feat cycle. When cycles run faster than release-please reacts, one
release-please PR bundles multiple `feat:` commits. See
[quickref.md § Mid-cycle halt](quickref.md#mid-cycle-halt-fires-when-release-please-opens-a-new-release-pr-not-per-feat)
for the full rationale.

---

## Auto-confirm scope (autonomous-mode §1 delta)

In autonomous mode, scope-pick is **auto-confirmed** — the loop takes the
first 🔲 BACKLOG row without halting for human confirmation. This is safe
because:
- The BACKLOG is authored and maintained by the human.
- The already-shipped check (§1.5) catches stale rows.
- The pre-merge gate catches implementation issues before merge.

Bundle-check is also auto-confirmed when all disjointness conditions from
[README.md §1](README.md#section-1--pick-the-scope) are met algorithmically.

---

## Cycle wall-clock cap

**Hard ceiling: 2 hours per cycle.** Capture `CYCLE_START` at §0a; before
each long-running step (CI wait, halt wait), check elapsed time against
7200 seconds. If exceeded:

1. Cancel any in-flight polling.
2. Write `outcome: "halt_cycle_timeout"` to `cycles.jsonl`.
3. Release the lock file.
4. Exit cleanly.

This covers the historical worst-case (~90 min) with margin. Anything past
2 hours is almost certainly stuck (hung subagent, lost CI run, missed phone
halt past 2-hour response window).

---

## Output compression

Autonomous-mode output should be terse — the orchestrator makes decisions and
routes to subagents; it does not narrate. Per-step status is one line:

```
step1: cycle 8, mode=iteration, scope=GHP-pr-reopen
step2_tdd: 7 commits, lint+test green (subagent, 87K tokens)
step2_pr_open: PR #147 opened, CI running...
```

No expanded reasoning, no narration of file lists. The metrics and cycle logs
are the durable record; verbose conversation narration is pure token cost.
