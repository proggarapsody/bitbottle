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
   - BACKLOG status flip goes in the same feat commit — NOT a follow-up `chore:` PR.
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
