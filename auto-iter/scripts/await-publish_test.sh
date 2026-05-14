#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/await-publish.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== await-publish_test ==="

TMPDIR=$(mktemp -d)
STUBDIR=$(mktemp -d)
trap "rm -rf '$TMPDIR' '$STUBDIR'" EXIT
cd "$TMPDIR"
git init -q -b main

# Programmable stubs for gh + npm. Each reads exit code + output from env vars.
cat > "$STUBDIR/gh" <<'GH'
#!/usr/bin/env bash
# Only handle `gh release view <tag>`. Success iff STUB_GH_OK=1.
[[ "${STUB_GH_OK:-0}" == "1" ]] && { echo '{"tagName":"v1.0.0"}'; exit 0; }
exit 1
GH
chmod +x "$STUBDIR/gh"

cat > "$STUBDIR/npm" <<'NPM'
#!/usr/bin/env bash
# Only handle `npm view <pkg> version`. Echoes STUB_NPM_VERSION (or empty).
echo "${STUB_NPM_VERSION:-}"
NPM
chmod +x "$STUBDIR/npm"

export PATH="$STUBDIR:$PATH"

# Case 1: both already published -> immediate success
export STUB_GH_OK=1
export STUB_NPM_VERSION=1.0.0
OUT="$(bash "$SCRIPT" 1.0.0 --interval-sec=1 --timeout-min=1)"
if echo "$OUT" | jq -e '.version=="1.0.0" and .github==true and .npm==true' >/dev/null; then
  ok "both already published -> success"
else
  fail "got: $OUT"
fi

# Case 2: v-prefix accepted
OUT="$(bash "$SCRIPT" v1.0.0 --interval-sec=1 --timeout-min=1)"
if echo "$OUT" | jq -e '.version=="1.0.0"' >/dev/null; then
  ok "v-prefix normalised"
else
  fail "got: $OUT"
fi

# Case 3: npm version mismatch -> timeout, github=true, npm=false
export STUB_GH_OK=1
export STUB_NPM_VERSION=0.9.0
if OUT="$(bash "$SCRIPT" 1.0.0 --interval-sec=1 --timeout-min=0 2>/dev/null)"; then
  fail "expected timeout exit"
else
  if echo "$OUT" | jq -e '.halt=="timeout" and .github==true and .npm==false' >/dev/null; then
    ok "github ready, npm lagging -> timeout halt with partial state"
  else
    fail "got: $OUT"
  fi
fi

# Case 4: missing version arg
if bash "$SCRIPT" 2>/dev/null; then
  fail "expected non-zero on missing version"
else
  ok "missing version rejected"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: await-publish.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
