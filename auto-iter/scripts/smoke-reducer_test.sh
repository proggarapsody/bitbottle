#!/usr/bin/env bash
# Unit test for auto-iter/scripts/smoke-reducer.sh.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/smoke-reducer.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== smoke-reducer_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
cd "$TMPDIR"
git init -q
mkdir -p .claude/auto-iter

# ---------------------------------------------------------------------------
# Seed a cycles.jsonl with several rows spanning two releases.
# ---------------------------------------------------------------------------
seed_cycles() {
  cat > .claude/auto-iter/cycles.jsonl <<'JSONL'
{"ts":"2026-06-01T10:00:00Z","cycle":1,"mode":"iteration","scopes":["FOO"],"prs":[10],"release":"v1.130.0","outcome":"shipped","smoke":"pending"}
{"ts":"2026-06-02T10:00:00Z","cycle":2,"mode":"iteration","scopes":["BAR"],"prs":[11],"release":"v1.131.0","outcome":"shipped","smoke":"pending"}
{"ts":"2026-06-03T10:00:00Z","cycle":3,"mode":"iteration","scopes":["BAZ"],"prs":[12],"release":"v1.130.0","outcome":"shipped","smoke":"pending"}
{"ts":"2026-06-04T10:00:00Z","cycle":4,"mode":"brainstorm","scopes":[],"prs":[],"release":null,"outcome":"brainstorm_added_9"}
JSONL
}

# Case 1: stamp smoke=passed on v1.130.0 -> 2 rows updated, v1.131.0 row unchanged
seed_cycles
OUT="$(bash "$SCRIPT" --version=v1.130.0 --result=passed)"
if echo "$OUT" | jq -e '.updated==2 and .version=="v1.130.0" and .result=="passed"' >/dev/null; then
  ok "stamps 2 rows for v1.130.0 -> result=passed"
else
  fail "unexpected output: $OUT"
fi

# Verify cycles.jsonl was rewritten correctly:
# rows 1 and 3 (v1.130.0) should have smoke=passed
# row 2 (v1.131.0) should still have smoke=pending
# row 4 (release=null) unchanged (no smoke field originally, still absent)
ROW1="$(jq -r 'select(.cycle==1) | .smoke' .claude/auto-iter/cycles.jsonl)"
ROW2="$(jq -r 'select(.cycle==2) | .smoke' .claude/auto-iter/cycles.jsonl)"
ROW3="$(jq -r 'select(.cycle==3) | .smoke' .claude/auto-iter/cycles.jsonl)"

if [[ "$ROW1" == "passed" ]]; then
  ok "cycle 1 (v1.130.0) stamped smoke=passed"
else
  fail "cycle 1 smoke=$ROW1, expected passed"
fi
if [[ "$ROW2" == "pending" ]]; then
  ok "cycle 2 (v1.131.0) smoke unchanged (pending)"
else
  fail "cycle 2 smoke=$ROW2, expected pending (unchanged)"
fi
if [[ "$ROW3" == "passed" ]]; then
  ok "cycle 3 (v1.130.0) stamped smoke=passed"
else
  fail "cycle 3 smoke=$ROW3, expected passed"
fi

# Case 2: stamp smoke=failed on v1.131.0 -> 1 row updated
seed_cycles
OUT="$(bash "$SCRIPT" --version=v1.131.0 --result=failed)"
if echo "$OUT" | jq -e '.updated==1 and .result=="failed"' >/dev/null; then
  ok "stamps 1 row for v1.131.0 -> result=failed"
else
  fail "unexpected output: $OUT"
fi
ROW2="$(jq -r 'select(.cycle==2) | .smoke' .claude/auto-iter/cycles.jsonl)"
if [[ "$ROW2" == "failed" ]]; then
  ok "cycle 2 stamped smoke=failed"
else
  fail "cycle 2 smoke=$ROW2, expected failed"
fi

# Case 3: version not in file -> 0 rows updated, file unchanged
seed_cycles
OUT="$(bash "$SCRIPT" --version=v9.99.0 --result=passed)"
if echo "$OUT" | jq -e '.updated==0' >/dev/null; then
  ok "unknown version -> 0 rows updated"
else
  fail "unexpected output: $OUT"
fi

# Case 4: missing --version -> non-zero exit + halt envelope
if bash "$SCRIPT" --result=passed 2>/dev/null; then
  fail "expected non-zero exit when --version missing"
else
  OUT="$(bash "$SCRIPT" --result=passed 2>/dev/null || true)"
  if echo "$OUT" | jq -e '.halt=="missing_required_field"' >/dev/null; then
    ok "missing --version -> halt missing_required_field"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 5: missing --result -> non-zero exit + halt envelope
if bash "$SCRIPT" --version=v1.130.0 2>/dev/null; then
  fail "expected non-zero exit when --result missing"
else
  OUT="$(bash "$SCRIPT" --version=v1.130.0 2>/dev/null || true)"
  if echo "$OUT" | jq -e '.halt=="missing_required_field"' >/dev/null; then
    ok "missing --result -> halt missing_required_field"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 6: invalid --result value -> non-zero exit + halt envelope
if bash "$SCRIPT" --version=v1.130.0 --result=maybe 2>/dev/null; then
  fail "expected non-zero exit for invalid --result"
else
  OUT="$(bash "$SCRIPT" --version=v1.130.0 --result=maybe 2>/dev/null || true)"
  if echo "$OUT" | jq -e '.halt=="invalid_result"' >/dev/null; then
    ok "invalid --result -> halt invalid_result"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 7: no cycles.jsonl -> 0 rows updated, no error
rm -f .claude/auto-iter/cycles.jsonl
OUT="$(bash "$SCRIPT" --version=v1.130.0 --result=passed)"
if echo "$OUT" | jq -e '.updated==0' >/dev/null; then
  ok "absent cycles.jsonl -> 0 updated, no error"
else
  fail "unexpected output: $OUT"
fi

# Case 8: atomic write — file should not be left as a tmp file on success
seed_cycles
bash "$SCRIPT" --version=v1.130.0 --result=passed >/dev/null
tmp_count="$(find .claude/auto-iter -name '*.tmp.*' 2>/dev/null | wc -l | tr -d ' ')"
if [[ "$tmp_count" -eq 0 ]]; then
  ok "no tmp files left after successful run (atomic swap)"
else
  fail "tmp files left behind: $tmp_count"
fi

# Case 9: file line count preserved after rewrite
seed_cycles
line_count_before="$(wc -l < .claude/auto-iter/cycles.jsonl | tr -d ' ')"
bash "$SCRIPT" --version=v1.130.0 --result=passed >/dev/null
line_count_after="$(wc -l < .claude/auto-iter/cycles.jsonl | tr -d ' ')"
if [[ "$line_count_before" == "$line_count_after" ]]; then
  ok "line count preserved after rewrite ($line_count_before lines)"
else
  fail "line count changed: before=$line_count_before after=$line_count_after"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: smoke-reducer.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
