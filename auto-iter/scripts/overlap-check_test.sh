#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/overlap-check.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== overlap-check_test ==="

TMPDIR=$(mktemp -d)
STUBDIR=$(mktemp -d)
trap "rm -rf '$TMPDIR' '$STUBDIR'" EXIT
cd "$TMPDIR"
git init -q -b main

# Programmable gh stub: reads from STUB_PRS env var
cat > "$STUBDIR/gh" <<'GH'
#!/usr/bin/env bash
# Always emit whatever STUB_PRS holds (default empty array)
echo "${STUB_PRS:-[]}"
GH
chmod +x "$STUBDIR/gh"
export PATH="$STUBDIR:$PATH"

cat > BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| WORKSPACE-MEMBERS | **Workspace Members** | `workspace member list [WORKSPACE]` — list members of a Cloud workspace | Cloud | 2 | 🔲 |
MD

# Case 1: no open PRs -> no overlap
export STUB_PRS='[]'
OUT="$(bash "$SCRIPT" WORKSPACE-MEMBERS)"
if echo "$OUT" | jq -e '.overlapping_pr==null' >/dev/null; then
  ok "no open PRs -> overlapping_pr=null"
else
  fail "got: $OUT"
fi

# Case 2: matching PR (workspace + members in title)
export STUB_PRS='[{"number":42,"title":"feat(workspace): list members","body":"workspace member list command"}]'
OUT="$(bash "$SCRIPT" WORKSPACE-MEMBERS)"
if echo "$OUT" | jq -e '.overlapping_pr==42 and (.matched_keywords | length >= 2)' >/dev/null; then
  ok "matching PR detected as overlap"
else
  fail "got: $OUT"
fi

# Case 3: unrelated PR (no keyword overlap)
export STUB_PRS='[{"number":99,"title":"fix(deploy): pipeline status","body":"unrelated"}]'
OUT="$(bash "$SCRIPT" WORKSPACE-MEMBERS)"
if echo "$OUT" | jq -e '.overlapping_pr==null' >/dev/null; then
  ok "unrelated PR not flagged as overlap"
else
  fail "got: $OUT"
fi

# Case 4: only one keyword match (under threshold)
export STUB_PRS='[{"number":1,"title":"workspace doc fix","body":""}]'
OUT="$(bash "$SCRIPT" WORKSPACE-MEMBERS)"
if echo "$OUT" | jq -e '.overlapping_pr==null' >/dev/null; then
  ok "single-keyword match below threshold"
else
  fail "got: $OUT"
fi

# Case 5: missing scope arg
if bash "$SCRIPT" 2>/dev/null; then
  fail "expected non-zero exit on missing scope arg"
else
  ok "missing scope rejected"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: overlap-check.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
