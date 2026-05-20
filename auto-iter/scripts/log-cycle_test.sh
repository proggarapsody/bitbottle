#!/usr/bin/env bash
# Unit test for auto-iter/scripts/log-cycle.sh.
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

# Case 1: cycle entry happy path (scopes/prs are arrays since PRD fix)
OUT="$(bash "$SCRIPT" --cycle=55 --mode=iteration --scope=PR-PARTICIPANTS --outcome=shipped --pr=292 --release=v1.60.0)"
if echo "$OUT" | jq -e '.cycle==55 and .mode=="iteration" and .scopes==["PR-PARTICIPANTS"] and .outcome=="shipped" and .prs==[292] and .release=="v1.60.0"' >/dev/null; then
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

# Case 6: cycle entries gain metrics_steps_count derived from metrics.jsonl.
# With no metrics file: count == 0.
OUT="$(bash "$SCRIPT" --cycle=70 --mode=iteration --scope=X --outcome=shipped)"
if echo "$OUT" | jq -e '.metrics_steps_count==0' >/dev/null; then
  ok "metrics_steps_count=0 when metrics.jsonl absent"
else
  fail "expected metrics_steps_count=0 with no metrics file, got: $OUT"
fi

# Seed metrics.jsonl with 5 entries (3 with .step on cycle 71, 1 without .step,
# 1 on cycle 72) — log-cycle should count only .step entries for the queried cycle.
cat > .claude/auto-iter/metrics.jsonl <<'METRICS'
{"cycle":71,"step":"step0a_lock","ts":"2026-05-18T10:00:00Z","duration_ms":5}
{"cycle":71,"step":"step0_preflight","ts":"2026-05-18T10:00:01Z","duration_ms":200}
{"cycle":71,"step":"step1_mode_pick","ts":"2026-05-18T10:00:02Z","duration_ms":30}
{"cycle":71,"ts":"2026-05-18T10:00:03Z","note":"no step field, should not count"}
{"cycle":72,"step":"step0a_lock","ts":"2026-05-18T11:00:00Z","duration_ms":5}
METRICS
OUT="$(bash "$SCRIPT" --cycle=71 --mode=iteration --scope=Y --outcome=shipped)"
if echo "$OUT" | jq -e '.metrics_steps_count==3' >/dev/null; then
  ok "metrics_steps_count=3 (counts only .step entries for that cycle)"
else
  fail "expected metrics_steps_count=3, got: $OUT"
fi

# Stream rows do NOT carry metrics_steps_count.
OUT="$(bash "$SCRIPT" --stream=completed --max=5 --ran=5)"
if echo "$OUT" | jq -e 'has("metrics_steps_count") | not' >/dev/null; then
  ok "stream rows omit metrics_steps_count"
else
  fail "stream row should not have metrics_steps_count, got: $OUT"
fi

# Case 8: single --scope= produces 1-element array
OUT="$(bash "$SCRIPT" --cycle=80 --mode=iteration --scope=DEPLOY-KEY --outcome=shipped)"
if echo "$OUT" | jq -e '.scopes | (type=="array") and (length==1) and (.[0]=="DEPLOY-KEY")' >/dev/null; then
  ok "single --scope= produces 1-element scopes array"
else
  fail "single --scope= not an array: $OUT"
fi

# Case 9: two --scope= produce 2-element array
OUT="$(bash "$SCRIPT" --cycle=81 --mode=iteration --scope=FOO --scope=BAR --outcome=shipped)"
if echo "$OUT" | jq -e '.scopes | (type=="array") and (length==2)' >/dev/null; then
  ok "two --scope= produces 2-element scopes array"
else
  fail "two --scope= not 2-element array: $OUT"
fi

# Case 10: single --pr= produces 1-element array
OUT="$(bash "$SCRIPT" --cycle=82 --mode=iteration --scope=X --outcome=shipped --pr=123)"
if echo "$OUT" | jq -e '.prs | (type=="array") and (length==1) and (.[0]==123)' >/dev/null; then
  ok "single --pr= produces 1-element prs array"
else
  fail "single --pr= not array: $OUT"
fi

# Case 11: two --pr= produce 2-element array
OUT="$(bash "$SCRIPT" --cycle=83 --mode=iteration --scope=X --outcome=shipped --pr=111 --pr=222)"
if echo "$OUT" | jq -e '.prs | (type=="array") and (length==2)' >/dev/null; then
  ok "two --pr= produces 2-element prs array"
else
  fail "two --pr= not 2-element array: $OUT"
fi

# Case 12: cycle row includes pipeline_version, tokens, duration_active_min, duration_wall_min
# Seed a metrics.jsonl for cycle 84 with 2 steps
mkdir -p .claude/auto-iter
cat >> .claude/auto-iter/metrics.jsonl <<'METRICS2'
{"cycle":84,"step":"step0a_lock","ts":"2026-05-20T10:00:00Z","duration_ms":2000,"subagent_tokens":0}
{"cycle":84,"step":"step2_tdd","ts":"2026-05-20T10:00:02Z","duration_ms":60000,"subagent_tokens":50000}
METRICS2
# duration_wall_min requires --duration_min from caller
OUT="$(bash "$SCRIPT" --cycle=84 --mode=iteration --scope=Z --outcome=shipped --duration_min=15)"
if echo "$OUT" | jq -e '
  (.pipeline_version | type == "string") and
  (.tokens == 50000) and
  (.duration_active_min == 1.0) and
  (.duration_wall_min == 15) and
  (.duration_min == 15)
' >/dev/null; then
  ok "cycle row has pipeline_version, tokens, duration_active_min, duration_wall_min, duration_min alias"
else
  fail "cycle row missing new fields: $OUT"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: log-cycle.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
