#!/usr/bin/env bash
# Unit test for auto-iter/scripts/metric.sh.
# Verifies: emits valid JSON, appends to metrics.jsonl, rejects missing required fields.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/metric.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== metric_test ==="

# Run inside a temporary git repo so .claude/auto-iter is sandboxed.
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
cd "$TMPDIR"
git init -q
mkdir -p .claude/auto-iter

# Case 1: happy path
OUT="$(bash "$SCRIPT" --cycle=42 --step=step2_tdd --duration_ms=12345 --subagent_tokens=7000)"
if echo "$OUT" | jq -e '.cycle==42 and .step=="step2_tdd" and .duration_ms==12345 and .subagent_tokens==7000 and (.ts | type=="string")' >/dev/null; then
  ok "emits JSON with typed integers and string ts"
else
  fail "unexpected JSON: $OUT"
fi

# Case 2: file appended
if [[ "$(wc -l < .claude/auto-iter/metrics.jsonl | tr -d ' ')" == "1" ]]; then
  ok "appended single line to metrics.jsonl"
else
  fail "expected 1 line in metrics.jsonl"
fi

# Case 3: missing --cycle
if bash "$SCRIPT" --step=foo 2>/dev/null; then
  fail "expected non-zero exit on missing --cycle"
else
  ok "rejects missing --cycle"
fi

# Case 4: missing --step
if bash "$SCRIPT" --cycle=5 2>/dev/null; then
  fail "expected non-zero exit on missing --step"
else
  ok "rejects missing --step"
fi

# Case 5: string fields stay strings
OUT="$(bash "$SCRIPT" --cycle=43 --step=step2_audit_run --findings='ServerCapabilities dead interface')"
if echo "$OUT" | jq -e '.findings=="ServerCapabilities dead interface"' >/dev/null; then
  ok "string fields stay quoted"
else
  fail "string field mis-typed: $OUT"
fi

# Case 6: 2 lines total now
if [[ "$(wc -l < .claude/auto-iter/metrics.jsonl | tr -d ' ')" == "2" ]]; then
  ok "subsequent runs append, do not truncate"
else
  fail "expected 2 lines in metrics.jsonl"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: metric.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
