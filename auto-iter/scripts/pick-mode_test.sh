#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/pick-mode.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== pick-mode_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
cd "$TMPDIR"
git init -q -b main
mkdir -p .claude/auto-iter docs/backlog

# Helper: write a fresh fixture
write_cycles() { printf '%s\n' "$@" > .claude/auto-iter/cycles.jsonl; }
write_backlog() {
  cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Status |
|---|---|---|
MD
  for line in "$@"; do printf '%s\n' "$line" >> docs/backlog/BACKLOG.md; done
}

# Case 1: empty cycles + empty backlog -> brainstorm, counter=1
: > .claude/auto-iter/cycles.jsonl
write_backlog
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.cycle==1 and .mode=="brainstorm" and .open_scopes==0' >/dev/null; then
  ok "empty state -> brainstorm"
else
  fail "got: $OUT"
fi

# Case 2: one shipped iteration, backlog has open scopes -> iteration, counter=2
write_cycles '{"cycle":1,"mode":"iteration","outcome":"shipped"}'
write_backlog '| FOO | Foo scope | 🔲 |'
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.cycle==2 and .mode=="iteration" and .open_scopes==1' >/dev/null; then
  ok "shipped + open backlog -> iteration"
else
  fail "got: $OUT"
fi

# Case 3: cycle counter at 4 -> next is 5 -> architecture
lines=()
for c in 1 2 3 4; do lines+=("{\"cycle\":$c,\"mode\":\"iteration\",\"outcome\":\"shipped\"}"); done
write_cycles "${lines[@]}"
write_backlog '| FOO | Foo scope | 🔲 |'
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.cycle==5 and .mode=="architecture"' >/dev/null; then
  ok "cycle 5 (multiple of 5) -> architecture"
else
  fail "got: $OUT"
fi

# Case 4: skip_in_progress rows must NOT advance counter
write_cycles \
  '{"cycle":1,"mode":"iteration","outcome":"shipped"}' \
  '{"cycle":null,"outcome":"skip_in_progress"}' \
  '{"cycle":null,"outcome":"skip_in_progress"}'
write_backlog '| FOO | Foo scope | 🔲 |'
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.cycle==2' >/dev/null; then
  ok "skip_in_progress rows excluded from counter"
else
  fail "got: $OUT"
fi

# Case 5: backlog empty, three trailing empty brainstorms -> stop
write_cycles \
  '{"cycle":10,"mode":"iteration","outcome":"shipped"}' \
  '{"cycle":11,"mode":"brainstorm","outcome":"brainstorm_added_0"}' \
  '{"cycle":12,"mode":"brainstorm","outcome":"brainstorm_added_0"}' \
  '{"cycle":13,"mode":"brainstorm","outcome":"brainstorm_added_0"}'
write_backlog
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.mode=="stop" and .consecutive_empty_brainstorms==3' >/dev/null; then
  ok "empty backlog + 3 empty brainstorms -> stop"
else
  fail "got: $OUT"
fi

# Case 6: backlog empty, only 2 empty brainstorms -> brainstorm (need 3)
write_cycles \
  '{"cycle":10,"mode":"iteration","outcome":"shipped"}' \
  '{"cycle":11,"mode":"brainstorm","outcome":"brainstorm_added_3"}' \
  '{"cycle":12,"mode":"brainstorm","outcome":"brainstorm_added_0"}' \
  '{"cycle":13,"mode":"brainstorm","outcome":"brainstorm_added_0"}'
write_backlog
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.mode=="brainstorm" and .consecutive_empty_brainstorms==2' >/dev/null; then
  ok "only 2 trailing empties -> brainstorm (not stop)"
else
  fail "got: $OUT"
fi

# Case 7: backlog empty -> brainstorm even when cycle%5==0 (rare)
# Actually architecture wins per algorithm. Verify.
lines=()
for c in 1 2 3 4; do lines+=("{\"cycle\":$c,\"mode\":\"iteration\",\"outcome\":\"shipped\"}"); done
write_cycles "${lines[@]}"
write_backlog  # empty
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.cycle==5 and .mode=="architecture"' >/dev/null; then
  ok "cycle%5==0 wins over empty backlog"
else
  fail "got: $OUT"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: pick-mode.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
