#!/usr/bin/env bash
# Unit test for scripts/auto-iter/log-cycle.sh.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/log-cycle.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== log-cycle_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
cd "$TMPDIR"
git init -q
mkdir -p .claude/auto-iter

# Case 1: cycle entry happy path
OUT="$(bash "$SCRIPT" --cycle=55 --mode=iteration --scope=PR-PARTICIPANTS --outcome=shipped --pr=292 --release=v1.60.0)"
if echo "$OUT" | jq -e '.cycle==55 and .mode=="iteration" and .scope=="PR-PARTICIPANTS" and .outcome=="shipped" and .pr==292 and .release=="v1.60.0"' >/dev/null; then
  ok "cycle entry emits typed JSON"
else
  fail "unexpected JSON: $OUT"
fi

# Case 2: stream entry happy path
OUT="$(bash "$SCRIPT" --stream=started --max=5 --ran=0)"
if echo "$OUT" | jq -e '.stream=="started" and .max==5 and .ran==0' >/dev/null; then
  ok "stream entry emits typed JSON"
else
  fail "unexpected stream JSON: $OUT"
fi

# Case 3: cycle without outcome is rejected
if bash "$SCRIPT" --cycle=99 --mode=brainstorm 2>/dev/null; then
  fail "expected non-zero exit when --outcome missing on cycle"
else
  ok "rejects cycle row missing --outcome"
fi

# Case 4: neither --cycle nor --stream rejected
if bash "$SCRIPT" --mode=brainstorm 2>/dev/null; then
  fail "expected non-zero exit when both --cycle and --stream absent"
else
  ok "rejects row with neither --cycle nor --stream"
fi

# Case 5: two appended lines total
if [[ "$(wc -l < .claude/auto-iter/cycles.jsonl | tr -d ' ')" == "2" ]]; then
  ok "appended both rows (cycle + stream)"
else
  fail "expected 2 lines in cycles.jsonl"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: log-cycle.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
