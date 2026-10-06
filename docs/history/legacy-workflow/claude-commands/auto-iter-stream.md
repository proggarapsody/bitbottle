---
description: Run a bounded chain of /auto-iter cycles back-to-back. Each cycle completes (including release), then immediately starts the next, up to max. Stream stops on max reached OR on any halt_* outcome.
allowed-tools: Bash, Read, Edit, Write, Grep, Glob, Task, TodoWrite
---

# /auto-iter-stream — Claude Code adapter

Chain-loop mechanic: run N `/auto-iter` cycles back-to-back with no inter-cycle wait.
Lock/halt/cadence/metrics: see `.claude/commands/auto-iter.md`.
Full procedure: `docs/workflows/iteration-cycle/README.md`.

## Argument parsing

`max` = `$1` (first arg). Validation:
- Missing or empty → error, exit immediately. **Do not default.**
- Not a positive integer → error, exit.
- `> 10` → error, exit. (Hard ceiling; raise manually if genuinely needed.)

Announce: `stream started — max=<N>, ran=0`.

## Pre-flight

```bash
LOCK=.claude/auto-iter/.lock
if [ -f "$LOCK" ]; then
    LOCK_AGE_MIN=$(( ($(date +%s) - $(stat -f %m "$LOCK" 2>/dev/null || stat -c %Y "$LOCK")) / 60 ))
    [ "$LOCK_AGE_MIN" -lt 60 ] && exit  # abort — another cycle is running
fi
```

Log stream start: `{"ts":"<ISO>","stream":"started","max":<N>,"ran":0}`

## Loop logic

```
ran = 0
while ran < max:
    execute one full /auto-iter cycle (per .claude/commands/auto-iter.md)
    cycle_outcome = <last line in cycles.jsonl written by this cycle>

    if cycle_outcome starts with "halt_" OR cycle_outcome == "skip_in_progress":
        final_outcome = "stream_halted"; break

    if cycle_outcome == "shutdown_confirmed":
        final_outcome = "stream_shutdown"; break

    ran += 1

if ran == max: final_outcome = "stream_max_reached"

append: {"ts":"<ISO>","stream":"completed","stream_started":"<ISO>","max":<N>,"ran":<ran>,"final_outcome":"<final_outcome>"}
```

## `stream_*` outcomes

| Value | Meaning |
|---|---|
| `stream_max_reached` | ran = max, all cycles shipped |
| `stream_halted` | a `halt_*` stopped the chain |
| `stream_shutdown` | BACKLOG empty + 3 empty brainstorms |
| `stream_aborted_lock_held` | pre-flight found a fresh lock |

## Halt routing

Mid-cycle halts (`halt_release`, CI red, pre-merge BLOCKER) are handled inside each
`/auto-iter` cycle. `ship` reply → cycle ships, **stream continues**. `hold` reply →
`halt_release` outcome → **stream stops**.

Workspace overlap (`resolve`/`skip`/`close`) → stream continues regardless of reply.

## Anti-patterns

- ❌ `/auto-iter-stream 10+` unattended — chain risk compounds.
- ❌ Running `/auto-iter-stream` AND `/loop /auto-iter` simultaneously.
- ❌ `max > 5` on Claude Max plan — fair-use throttling kicks in.
- ❌ Treating mid-cycle halts as stream-breakers — use `hold` to stop.
