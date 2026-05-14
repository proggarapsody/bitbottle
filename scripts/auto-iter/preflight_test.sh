#!/usr/bin/env bash
# Unit test for scripts/auto-iter/preflight.sh.
# Doesn't require gh — script handles its absence gracefully.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/preflight.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== preflight_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Build a self-contained fixture repo.
cd "$TMPDIR"
git init -q -b main
git config user.email "test@test.com"
git config user.name "Test"
echo "hello" > README.md
git add README.md
git commit -q -m "init"

# Block gh from running so we exercise the no-gh branch deterministically.
# Stash the stub outside the repo so it doesn't show up as a dirty file.
STUBDIR="$TMPDIR/../stub-$$"
mkdir -p "$STUBDIR"
cat > "$STUBDIR/gh" <<'GH'
#!/usr/bin/env bash
exit 1
GH
chmod +x "$STUBDIR/gh"
export PATH="$STUBDIR:$PATH"
trap "rm -rf '$TMPDIR' '$STUBDIR'" EXIT

# Case 1: clean tree, on main, no remote → exit 0
if OUT="$(bash "$SCRIPT")"; then
  ok "exits 0 on clean tree"
else
  fail "expected exit 0 on clean tree"
fi
if echo "$OUT" | jq -e '.clean==true and .on_main==true and (.findings | length)==0' >/dev/null; then
  ok "reports clean=true, on_main=true, no findings"
else
  fail "unexpected JSON: $OUT"
fi
if echo "$OUT" | jq -e '.open_prs | type=="array"' >/dev/null; then
  ok "open_prs is array even without gh"
else
  fail "open_prs not array: $OUT"
fi

# Case 2: dirty tree → exit 1, finding present
echo "scratch" > scratch.txt
if OUT="$(bash "$SCRIPT" 2>/dev/null)"; then
  fail "expected non-zero exit on dirty tree"
else
  ok "exits non-zero on dirty tree"
fi
if echo "$OUT" | jq -e '.clean==false and (.findings | map(select(startswith("workspace_dirty"))) | length > 0)' >/dev/null; then
  ok "reports clean=false with workspace_dirty finding"
else
  fail "expected workspace_dirty finding, got: $OUT"
fi
rm scratch.txt

# Case 3: not on main → finding
git checkout -q -b feature/x
if OUT="$(bash "$SCRIPT" 2>/dev/null)"; then
  fail "expected non-zero exit when not on main"
else
  ok "exits non-zero when off main"
fi
if echo "$OUT" | jq -e '.on_main==false and .branch=="feature/x"' >/dev/null; then
  ok "reports branch name correctly"
else
  fail "branch reporting failed: $OUT"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: preflight.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
