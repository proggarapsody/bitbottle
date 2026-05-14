#!/usr/bin/env bash
# Test for await-ci.sh — uses a fake `gh` on PATH to script the polling sequence.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/await-ci.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== await-ci_test ==="

TMPDIR=$(mktemp -d)
STUBDIR=$(mktemp -d)
trap "rm -rf '$TMPDIR' '$STUBDIR'" EXIT
cd "$TMPDIR"
git init -q -b main

# Build a programmable fake gh. Each call reads the next response from
# $STUBDIR/responses (one JSON blob per line, popped front).
cat > "$STUBDIR/gh" <<'GH'
#!/usr/bin/env bash
# Only handle the one form we use.
RESP_FILE="$STUB_RESPONSES"
if [[ ! -f "$RESP_FILE" ]]; then echo '[]'; exit 0; fi
head -1 "$RESP_FILE"
tail -n +2 "$RESP_FILE" > "$RESP_FILE.tmp" && mv "$RESP_FILE.tmp" "$RESP_FILE"
GH
chmod +x "$STUBDIR/gh"
export PATH="$STUBDIR:$PATH"
export STUB_RESPONSES="$STUBDIR/responses"

# Case 1: one poll, all pass
echo '[{"name":"Build","state":"SUCCESS","bucket":"pass"},{"name":"Test","state":"SUCCESS","bucket":"pass"}]' > "$STUB_RESPONSES"
OUT="$(bash "$SCRIPT" 123 --interval-sec=1 --timeout-min=1)"
if echo "$OUT" | jq -e '.pr==123 and .all_passed==true and (.failed | length)==0' >/dev/null; then
  ok "all pass on first poll"
else
  fail "got: $OUT"
fi

# Case 2: one pending → second pass returns done with one fail
cat > "$STUB_RESPONSES" <<JSON
[{"name":"Build","state":"PENDING","bucket":"pending"},{"name":"Test","state":"SUCCESS","bucket":"pass"}]
[{"name":"Build","state":"FAILURE","bucket":"fail"},{"name":"Test","state":"SUCCESS","bucket":"pass"}]
JSON
OUT="$(bash "$SCRIPT" 456 --interval-sec=1 --timeout-min=1)"
if echo "$OUT" | jq -e '.pr==456 and .all_passed==false and (.failed | index("Build") != null)' >/dev/null; then
  ok "pending → fail surfaces failed check"
else
  fail "got: $OUT"
fi

# Case 3: SKIPPED counted as passing the gate (auto-merge / release skip)
echo '[{"name":"Build","state":"SUCCESS","bucket":"pass"},{"name":"auto-merge","state":"SKIPPED","bucket":"skipping"}]' > "$STUB_RESPONSES"
OUT="$(bash "$SCRIPT" 789 --interval-sec=1 --timeout-min=1)"
if echo "$OUT" | jq -e '.all_passed==true and (.skipped | index("auto-merge") != null)' >/dev/null; then
  ok "skipped checks accepted, surfaced separately"
else
  fail "got: $OUT"
fi

# Case 4: timeout — always pending, very short timeout
cat > "$STUB_RESPONSES" <<'JSON'
[{"name":"Build","state":"PENDING","bucket":"pending"}]
[{"name":"Build","state":"PENDING","bucket":"pending"}]
[{"name":"Build","state":"PENDING","bucket":"pending"}]
[{"name":"Build","state":"PENDING","bucket":"pending"}]
JSON
if OUT="$(bash "$SCRIPT" 1 --interval-sec=1 --timeout-min=0 2>/dev/null)"; then
  fail "expected non-zero exit on timeout"
else
  if echo "$OUT" | jq -e '.halt=="timeout"' >/dev/null; then
    ok "timeout produces halt envelope"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 5: missing PR arg
if bash "$SCRIPT" 2>/dev/null; then
  fail "expected non-zero exit when PR missing"
else
  ok "missing PR rejected"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: await-ci.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
