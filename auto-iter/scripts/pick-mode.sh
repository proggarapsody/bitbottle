#!/usr/bin/env bash
# Pick the mode for the next /auto-iter cycle.
#
# Reads:
#   .claude/auto-iter/cycles.jsonl  (cycle counter + brainstorm history)
#   docs/backlog/BACKLOG.md          (open scope count — queue, post-reorg)
#
# Emits:
#   {"cycle":N,"mode":"iteration|architecture|brainstorm|stop",
#    "open_scopes":N,"consecutive_empty_brainstorms":N,"reason":"..."}
#
# Algorithm (canonical — see docs/workflows/iteration-cycle/quickref.md § Model tier per phase):
#   counter = max(cycle) over non-skip rows, +1
#   open    = grep 🔲 in docs/backlog/BACKLOG.md ## Backlog table
#   empties = consecutive brainstorm_added_0 rows from the tail
#
#   open==0 AND empties>=3            -> stop
#   counter>0 AND counter%5 == 0      -> architecture
#   open==0                           -> brainstorm
#   else                              -> iteration
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cd "$(repo_root)"

CYCLES_FILE=".claude/auto-iter/cycles.jsonl"
BACKLOG_FILE="docs/backlog/BACKLOG.md"

# --- counter: max cycle over rows that are NOT skip_in_progress ---
counter=1
if [[ -f "$CYCLES_FILE" ]]; then
  max=$(jq -rs '
    map(select(.cycle != null and (.outcome // "") != "skip_in_progress"))
    | (map(.cycle) | max // 0)
  ' "$CYCLES_FILE" 2>/dev/null || echo 0)
  counter=$(( max + 1 ))
fi

# --- consecutive empty brainstorms from the tail ---
empties=0
if [[ -f "$CYCLES_FILE" ]]; then
  empties=$(jq -rs '
    map(select(.mode == "brainstorm"))
    | reverse
    | [.[] | (.outcome // "")]
    | map(select(test("_0$")))
    | length
  ' "$CYCLES_FILE" 2>/dev/null || echo 0)
fi

# --- open-scope count: 🔲 lines inside the ## Backlog section ---
open_scopes=0
if [[ -f "$BACKLOG_FILE" ]]; then
  open_scopes=$(awk '
    /^## Backlog/    { inblk=1; next }
    inblk && /^## / { inblk=0 }
    inblk && /🔲/   { n++ }
    END             { print (n+0) }
  ' "$BACKLOG_FILE")
fi

# --- decision ---
mode=""; reason=""
if [[ $open_scopes -eq 0 && $empties -ge 3 ]]; then
  mode="stop"
  reason="backlog_empty_and_${empties}_consecutive_empty_brainstorms"
elif (( counter > 0 && counter % 5 == 0 )); then
  mode="architecture"
  reason="cycle_${counter}_is_multiple_of_5"
elif [[ $open_scopes -eq 0 ]]; then
  mode="brainstorm"
  reason="backlog_empty"
else
  mode="iteration"
  reason="backlog_has_${open_scopes}_open_scopes"
fi

emit_json \
  --cycle="$counter" \
  --mode="$mode" \
  --open_scopes="$open_scopes" \
  --consecutive_empty_brainstorms="$empties" \
  --reason="$reason"
