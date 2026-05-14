#!/usr/bin/env bash
# Unit test for auto-iter/scripts/lock.sh.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/lock.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== lock_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
cd "$TMPDIR"
git init -q

# Case 1: status when no lock
OUT="$(bash "$SCRIPT" status)"
if echo "$OUT" | jq -e '.held==false' >/dev/null; then
  ok "status reports held=false when no lock"
else
  fail "unexpected status output: $OUT"
fi

# Case 2: acquire fresh lock
OUT="$(bash "$SCRIPT" acquire)"
if echo "$OUT" | jq -e '.acquired==true and .age_min==0' >/dev/null; then
  ok "acquire takes fresh lock"
else
  fail "unexpected acquire output: $OUT"
fi
if [[ ! -f .claude/auto-iter/.lock ]]; then
  fail ".lock file missing after acquire"
fi

# Case 3: second acquire within window is blocked
if OUT="$(bash "$SCRIPT" acquire 2>/dev/null)"; then
  fail "expected acquire to fail when recent lock present"
else
  if echo "$OUT" | jq -e '.halt=="recent_lock"' >/dev/null; then
    ok "second acquire returns recent_lock halt"
  else
    fail "expected halt=recent_lock, got: $OUT"
  fi
fi

# Case 4: release
OUT="$(bash "$SCRIPT" release)"
if echo "$OUT" | jq -e '.released==true' >/dev/null; then
  ok "release emits released=true"
else
  fail "unexpected release output: $OUT"
fi
if [[ -f .claude/auto-iter/.lock ]]; then
  fail ".lock file should be gone after release"
fi

# Case 5: stale-lock recovery (simulate by backdating mtime)
mkdir -p .claude/auto-iter
echo "stale" > .claude/auto-iter/.lock
if [[ "$OSTYPE" == darwin* ]]; then
  touch -t 202001010000 .claude/auto-iter/.lock
else
  touch -d "2020-01-01 00:00" .claude/auto-iter/.lock
fi
OUT="$(bash "$SCRIPT" acquire)"
if echo "$OUT" | jq -e '.acquired==true and .stale_lock==true' >/dev/null; then
  ok "stale lock (>60 min) is overwritten with stale_lock=true"
else
  fail "expected stale_lock recovery, got: $OUT"
fi

# Case 6: unknown subcommand halts
if bash "$SCRIPT" frobnicate 2>/dev/null; then
  fail "expected non-zero exit on unknown subcommand"
else
  ok "unknown subcommand rejected"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: lock.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
