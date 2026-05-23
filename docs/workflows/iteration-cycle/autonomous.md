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

## §3.5 — Mechanical taste-check (post-TDD, pre-gate)

After the TDD subagent returns and before opening the PR, run a mechanical
shell sweep inside the worktree to catch recurring BLOCKERs that the inline
TDD checklist has historically failed to prevent. This is the agent-neutral
mechanical encoding of the rules in [`docs/TASTE.md`](../../TASTE.md) that can
be expressed as `grep` patterns.

**Why mechanical, not checklist**: across cycles 23–24, the inline checklist
was treated as scenery; recurring BLOCKERs landed in PRs anyway. A script
can't be skimmed — it exits non-zero or it doesn't.

```bash
# Run inside the TDD worktree:
cd "$WORKTREE"
violations=""

# 1. Cobra Short ≤ 60 chars + non-empty + no trailing period
#    (skip MCP tool-registration files; they have no cobra Short field)
while IFS= read -r f; do
    [[ "$f" =~ /tools_.*\.go$ || "$f" =~ /handlers.*\.go$ ]] && continue
    short=$(grep -oE 'Short:\s*"[^"]*"' "$f" | sed 's/Short:\s*"//;s/"$//')
    [ -z "$short" ] && violations+="$f: empty Short\n" && continue
    [ "${#short}" -gt 60 ] && violations+="$f: Short=${#short} chars (>60)\n"
    echo "$short" | grep -qE '\.$' && violations+="$f: Short has trailing period\n"
done < <(git diff --name-only origin/main...HEAD -- 'pkg/cmd/**/*.go' ':!*_test.go')

# 2. New command files have SKILL.md + skills/references/<group>.md entries
while IFS= read -r f; do
    verb=$(basename "$f" .go)
    group=$(echo "$f" | awk -F/ '{print $3}')
    grep -q "$group $verb" skills/SKILL.md \
        || violations+="$f: skills/SKILL.md missing '$group $verb' entry\n"
    [ -f "skills/references/${group}.md" ] && \
        ! grep -q "$verb" "skills/references/${group}.md" && \
        violations+="$f: skills/references/${group}.md missing '$verb' entry\n"
done < <(git diff --name-only --diff-filter=A origin/main...HEAD -- 'pkg/cmd/**/*.go' ':!*_test.go' ':!**/*_fields.go')

# 3. Forbidden imports
git diff origin/main...HEAD -- 'pkg/cmd/**/*.go' | \
    grep -E '^\+.*"github.com/[^"]*/(api/cloud|api/server)' \
    && violations+="forbidden: pkg/cmd/** importing api/cloud or api/server directly\n"

# 4. No raw net/http outside api/internal/httpx.
#    Use xargs grep on filtered files — do NOT use `git diff | grep '"net/http"'`,
#    which can't tell which diff hunk a matching line belongs to (false positives).
#    Also exclude test files (legitimate to import net/http for httptest).
git diff origin/main...HEAD --name-only -- '*.go' | grep -v '_test\.go' | \
    xargs grep -l '"net/http"' 2>/dev/null | grep -v 'api/internal/httpx' | \
    grep -q . && violations+="forbidden: net/http used outside api/internal/httpx\n"

if [ -n "$violations" ]; then
    echo "TASTE-CHECK BLOCKERS:"
    echo -e "$violations"
    # → dispatch a fix-agent (code-generation tier) with this specific violation
    #   list (NOT the full PRD; targeted fixes only). Re-run this script after
    #   the fix-agent returns. If still violations after 1 fix attempt, halt
    #   for human triage.
fi
```

**False-positive prevention rules** (each is load-bearing — earlier versions
of the sweep produced spurious failures without them):
- Skip MCP tool-registration files (`tools_*.go`, `handlers*.go`) — they have
  no cobra `Short` field and are not cobra commands.
- The `net/http` check must exclude `*_test.go` (legitimate to import for
  `httptest`).
- For `net/http`, use `xargs grep -l` on filtered files, never raw
  `git diff | grep '"net/http"'` — the latter can't tell which file a
  matching line belongs to.

**Metric**: emit `step2_taste_check` with `violations_count` (and
`fixagent_dispatched: true|false`).

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

## §7 — Release publish (trust-don't-wait override)

In autonomous mode, the release PR merge command returning is treated as
**cycle end**. The cycle does NOT wait for GoReleaser to publish binaries or
for `npm publish` to complete.

Canonical [README.md §7](README.md#section-7--merge-the-pr-then-merge-the-release-pr)
step 5 says: "Verify: `gh release view --json tagName,publishedAt` shows the
new tag." That verification is the **serial** procedure. Autonomous mode
overrides it:

```bash
gh pr merge "$RELEASE_PR" --squash
# Log "release": "v1.X.Y" in cycles.jsonl. Do NOT poll await-publish.sh.
# Cycle ends; outcome = "shipped".
```

**Why no wait**: GoReleaser publishes reliably ~3–5 min after the release-PR
merge. Blocking the cycle on verification is ~3–5 min of pure idle per
release-firing cycle (~9–15 min per stream of 3). If GoReleaser fails (rare),
the GitHub Actions tab surfaces it via the normal channels — the cycle is
"merged, publish happens async", not "failed".

**The human verifies manually whenever they want**: `npm view @proggarapsody/bitbottle version`
or check the Actions tab. The cycle log records the merge, not the publish.

**When this override does NOT apply**: serial (`/iteration-cycle`) mode, where
a human is present and the verification adds confidence at near-zero
opportunity cost.

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
