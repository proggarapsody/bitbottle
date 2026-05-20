#!/usr/bin/env bash
# Unit test for auto-iter/scripts/cycle-summary.sh
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/cycle-summary.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== cycle-summary_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
METRICS="$TMPDIR/metrics.jsonl"

# ---- Fixture 1: full data ----
cat > "$METRICS" <<'JSONL'
{"cycle":10,"step":"step0a_lock","ts":"2026-05-20T10:00:00Z","duration_ms":1000,"subagent_tokens":0}
{"cycle":10,"step":"step2_tdd","ts":"2026-05-20T10:00:01Z","duration_ms":120000,"subagent_tokens":85000}
{"cycle":10,"step":"step2_design_judge","ts":"2026-05-20T10:02:01Z","duration_ms":60000,"subagent_tokens":40000}
{"cycle":11,"step":"step2_tdd","ts":"2026-05-20T11:00:00Z","duration_ms":90000,"subagent_tokens":50000}
JSONL

OUT="$(bash "$SCRIPT" --cycle=10 --metrics="$METRICS")"
# tokens: 0+85000+40000 = 125000
# duration_active_min: (1000+120000+60000)/60000 = 3.0 (rounded to 1 decimal)
# metrics_steps_count: 3
if echo "$OUT" | jq -e '.tokens==125000 and .duration_active_min==3.0 and .metrics_steps_count==3' >/dev/null; then
  ok "full data: tokens, duration_active_min, metrics_steps_count correct"
else
  fail "full data: unexpected JSON: $OUT"
fi

# ---- Fixture 2: partial data (missing subagent_tokens on some rows) ----
cat > "$METRICS" <<'JSONL'
{"cycle":20,"step":"step0a_lock","ts":"2026-05-20T10:00:00Z","duration_ms":2000}
{"cycle":20,"step":"step2_tdd","ts":"2026-05-20T10:00:02Z","duration_ms":90000,"subagent_tokens":70000}
{"cycle":20,"step":"step0_preflight","ts":"2026-05-20T10:00:03Z","duration_ms":5000}
JSONL

OUT="$(bash "$SCRIPT" --cycle=20 --metrics="$METRICS")"
# tokens: 0+70000+0 = 70000
# duration_active_min: (2000+90000+5000)/60000 = 1.6 (rounded to 1 decimal)
# metrics_steps_count: 3
if echo "$OUT" | jq -e '.tokens==70000 and .duration_active_min==1.6 and .metrics_steps_count==3' >/dev/null; then
  ok "partial data (missing subagent_tokens -> 0): correct"
else
  fail "partial data: unexpected JSON: $OUT"
fi

# ---- Fixture 3: no data (cycle not present in file) ----
OUT="$(bash "$SCRIPT" --cycle=99 --metrics="$METRICS")"
if echo "$OUT" | jq -e '.tokens==0 and .duration_active_min==0.0 and .metrics_steps_count==0' >/dev/null; then
  ok "no data: all zeros"
else
  fail "no data: unexpected JSON: $OUT"
fi

# ---- Fixture 4: metrics file missing ----
OUT="$(bash "$SCRIPT" --cycle=5 --metrics="$TMPDIR/nonexistent.jsonl")"
if echo "$OUT" | jq -e '.tokens==0 and .duration_active_min==0.0 and .metrics_steps_count==0' >/dev/null; then
  ok "missing file: all zeros (no crash)"
else
  fail "missing file: unexpected JSON: $OUT"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: cycle-summary.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
