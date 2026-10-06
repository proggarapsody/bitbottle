---
description: Run one autonomous iteration cycle. Picks mode (iteration | architecture-audit | brainstorm | stop) based on BACKLOG state and cycle counter, executes per docs/workflows/iteration-cycle/README.md with auto-confirmed safe halts and phone-routed risky halts via Remote Control.
allowed-tools: Bash, Read, Edit, Write, Grep, Glob, Task, TodoWrite
---

# /auto-iter — Claude Code adapter

Full iteration procedure: [`docs/workflows/iteration-cycle/README.md`](../../docs/workflows/iteration-cycle/README.md)
Autonomous-mode deltas: [`docs/workflows/iteration-cycle/autonomous.md`](../../docs/workflows/iteration-cycle/autonomous.md)
Pre-merge gate: [`docs/workflows/pre-merge-check.md`](../../docs/workflows/pre-merge-check.md)
Quickref (tables + policies): [`docs/workflows/iteration-cycle/quickref.md`](../../docs/workflows/iteration-cycle/quickref.md)
Script catalog: [`docs/workflows/iteration-cycle/scripts.md`](../../docs/workflows/iteration-cycle/scripts.md)

## Model tier mapping

| Phase tag (from quickref) | Claude model |
|---|---|
| judgment-heavy | Opus |
| code-generation | Sonnet (subagent via Task tool) |
| mechanical | Sonnet inline |

## Subagent dispatch

All heavy work dispatches via the `Task` tool:

| Phase | Dispatch | Model | Returns |
|---|---|---|---|
| **§3 TDD implementation** | `Task(isolation: "worktree")` | Sonnet subagent | commit list + `subagent_tokens` |
| **§6 design-judge** | `Task(model: "sonnet")` | Sonnet subagent | `findings_count` + `subagent_tokens` |
| **Architecture audit** | `Task(model: "opus")` | Opus subagent | finding or `audit_no_findings` |
| **Brainstorm** | `Task(model: "opus")` | Opus subagent | new BACKLOG rows |

Orchestrator-inline (NOT dispatched): §0a lock, §0 preflight, §1 mode/scope pick,
§1 bundle-check, §2 worktree creation/removal, §2 push PR + poll CI + merge,
§4 log writes, §5 release lock.

**If a step listed above as dispatched is run inline by mistake**, record `dispatch_violation: true` in the metrics line so post-cycle analysis catches it.

## In-cycle parallel block (§3.6)

After the §3.5 mechanical taste-check passes ([`autonomous.md` §3.5](../../docs/workflows/iteration-cycle/autonomous.md#35--mechanical-taste-check-post-tdd-pre-gate)), three activities run **concurrently** inside one cycle:

1. **Push branch + open PR** (orchestrator-inline) — `git push -u`, then `gh pr create` with the PR body (see § "PRD-close enforcement" below for the `Closes #<PRD>` requirement).
2. **Design-judge subagent** (`Task(model: "sonnet")`) — runs in parallel with CI. Inputs: diff URL, `docs/TASTE.md`, `docs/ARCHITECTURE.md`. Returns findings_count + subagent_tokens.
3. **Mechanical pre-merge gate** (orchestrator-inline) — §0 scope, §1 branch, §2 conventional commits, §3 no artifacts, §7 no CHANGELOG edits, §8 secrets grep. Emits one `§N name: PASS/BLOCKER` line per section.

**Convergence**: wait for (a) CI green AND (b) design-judge returned AND (c) mechanical gate clean. Then:
- If CI red → halt with check name + log URL.
- If any mechanical gate BLOCKER → halt: `❌ PR #N: <finding>`.
- If design-judge returns BLOCKER findings (layer / security) → halt.
- If design-judge returns style/informational findings → log and continue.

**Why concurrent**: CI takes ~2 min, design-judge ~2 min, mechanical gate ~5 sec. Serializing adds ~2 min per cycle. Running them in parallel keeps the cycle's wall-clock at max(CI, design-judge) ≈ 2 min.

## PRD-close enforcement

The PR body's last line MUST be `Closes #<PRD-issue-number>` so GitHub auto-closes the PRD on squash-merge. Belt-and-suspenders with the [README §6 rule](../../docs/workflows/iteration-cycle/README.md#section-6--open-the-pr): the README is the rule, the command file is the enforcement for autonomous mode.

Before `gh pr create`, programmatically append the close keyword if not already present:

```bash
PRD_NUM="$PRD_ISSUE"  # captured at §2 PRD filing
if ! echo "$PR_BODY" | grep -q "Closes #${PRD_NUM}"; then
  PR_BODY="${PR_BODY}

Closes #${PRD_NUM}"
fi
gh pr create --title "$PR_TITLE" --body "$PR_BODY"
```

**Why it matters**: PRDs #448 (REPO-EDIT), #451 (PIPELINE-STOP), #454 (SNIPPETS) all shipped in v1.90.0–v1.92.0 but stayed open for 2+ days because their PR bodies omitted the close keyword. PRD #464 was filed to fix this recurrence; this enforcement is the load-bearing half.

**Post-merge verification** (§8): `gh issue view "$PRD_NUM" --json state | jq -r .state`. If `OPEN`, close manually with a comment linking the merge commit, and record `dispatch_violation: true` on the cycle's `step8_close_prd` metric so the failure mode stays visible.

## Halt delivery

| Condition | Channel |
|---|---|
| Remote Control active | PushNotification ONLY (never also print to chat) |
| Remote Control inactive | Chat text ONLY (never attempt PushNotification) |

Never both channels. "Asked twice" is a bug.

Halt messages stay <500 chars (phone lock-screen display). Reference PR URL — never embed diff or log.

## Output compression

- Per step: one line status (e.g. `step2_tdd: 7 commits, lint+test green (Sonnet subagent, 87K tokens)`)
- Pre-merge gate: one `§N name: PASS/BLOCKER` line per section
- No narration of intermediate steps
- Invoke `caveman` skill in TDD subagent prompt for token economy (≈75% reduction in parallel mode)

## Metrics emission

Use `auto-iter/scripts/metric.sh` for ALL metric writes. Never raw `jq -nc` or `echo >>`.

```bash
T0=$(date +%s%3N)
# ... step body ...
DUR_MS=$(( $(date +%s%3N) - T0 ))
auto-iter/scripts/metric.sh --cycle="$CYCLE" --step=step2_tdd \
  --duration_ms="$DUR_MS" --subagent_tokens="${TOK:-null}"
```

## CYCLE_START cap

Capture `CYCLE_START=$(date +%s)` at §0a. Before each long-running step, check elapsed
against 7200 seconds. If exceeded → write `outcome: "halt_cycle_timeout"`, release lock,
exit cleanly. See [`autonomous.md`](../../docs/workflows/iteration-cycle/autonomous.md) for full details.

## TDD subagent prompt requirements

The TDD subagent prompt MUST include:
1. PRD issue body (with the `## Expected files` scaffold from §2).
2. Worktree path.
3. `docs/agent-primer.md` as required reading.
4. **Caveman-mode directive** for the duration of the run.
5. Anti-pattern rules (recurring BLOCKERs from past cycles):
   - Cobra `Short` ≤ 60 chars, non-empty, no trailing period.
   - Every new command adds a row to `skills/SKILL.md` AND `skills/references/<group>.md`.
   - Every new MCP tool needs `tools.go` entry + `handlers.go` method + `handlers_test.go` (triplet).
   - No raw `net/http.Client` outside `api/internal/httpx/`.
   - No `pkg/cmd/**` imports of `api/cloud` or `api/server` directly.
   - PR title MUST start with `feat(scope):` / `fix(scope):` (Conventional Commit).
   - BACKLOG→SHIPPED **move** goes in the same feat commit: cut the row + scope-detail section from `docs/backlog/BACKLOG.md`, prepend a dated entry to `docs/backlog/SHIPPED.md`. Both edits land with the code in one commit — NOT a follow-up `chore:` PR. See [`docs/backlog/SHIPPED.md`](../../docs/backlog/SHIPPED.md) "How to add an entry" for the format. `auto-iter/scripts/pre-merge-mechanical.sh` §4 blocks commits touching only BACKLOG.md or only SHIPPED.md.
6. Subagent does full red → green → refactor, runs `go test ./... -race`, commits, runs
   `bash scripts/tdd-check.sh` before returning. Does NOT push or open the PR.

## Squash-merge subject enforcement

Always pass `--subject` to override the squash-commit title:

```bash
PR_TITLE=$(gh pr view <N> --json title -q .title)
gh pr merge <N> --squash --delete-branch --subject "$PR_TITLE"
```

If the TDD subagent committed with a wrong prefix (`fix:` on a feature), surface
`tdd_commit_drift` in cycle notes rather than silently masking it.

## Pre-flight (one-time, before kicking off `/loop`)

1. Open Claude with Remote Control paired to your phone.
2. Verify `caffeinate -di` running or laptop plugged in.
3. Verify `git worktree list` clean (only main checkout).
4. Then kick off `/loop 60m /auto-iter`.
