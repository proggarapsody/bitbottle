#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/pick-scope.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== pick-scope_test ==="

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT
cd "$TMPDIR"
git init -q -b main
mkdir -p docs/backlog

# Case 1: backlog with one open scope, no Scope Details section -> has_scope_details=false
cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| FOO | **Foo Scope** | `foo list` — list foos | Both | 2 | ✅ |
| BAR | **Bar Scope** | `bar list` — list bars. Cloud only. | Cloud | 2 | 🔲 |
MD
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug=="BAR" and .scope_name=="Bar Scope" and .backend=="Cloud" and .tier=="2" and .has_scope_details==false and .details_anchor==null' >/dev/null; then
  ok "newer inline scope -> has_scope_details=false, anchor=null"
else
  fail "got: $OUT"
fi
if echo "$OUT" | jq -e '.summary | contains("Cloud only")' >/dev/null; then
  ok "summary preserves Commands cell"
else
  fail "summary missing: $OUT"
fi

# Case 2: scope WITH a ## Scope Details anchor
cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| L | **Branch Create + Checkout** | `branch create`, `branch checkout` | Both | 1 | 🔲 |

## Scope Details

### L — Branch Create + Checkout
Detailed spec lives here…
MD
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug=="L" and .has_scope_details==true and (.details_anchor | startswith("#l-"))' >/dev/null; then
  ok "older scope -> has_scope_details=true with anchor"
else
  fail "got: $OUT"
fi

# Case 3: tier "DX" string survives
cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| TST | **Test scope** | `test foo` | N/A | DX | 🔲 |
MD
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.tier=="DX"' >/dev/null; then
  ok "tier=DX preserved as string"
else
  fail "got: $OUT"
fi

# Case 4: backlog has only ✅ rows -> halt backlog_empty
cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| A | **A** | x | Both | 1 | ✅ |
| B | **B** | y | Both | 2 | ✅ |
MD
if OUT="$(bash "$SCRIPT" 2>/dev/null)"; then
  fail "expected non-zero exit on empty backlog"
else
  if echo "$OUT" | jq -e '.halt=="backlog_empty"' >/dev/null; then
    ok "no open rows -> halt backlog_empty"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 5: docs/backlog/BACKLOG.md missing entirely
rm docs/backlog/BACKLOG.md
if OUT="$(bash "$SCRIPT" 2>/dev/null)"; then
  fail "expected non-zero exit when docs/backlog/BACKLOG.md missing"
else
  if echo "$OUT" | jq -e '.halt=="backlog_missing"' >/dev/null; then
    ok "missing docs/backlog/BACKLOG.md -> halt backlog_missing"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 6: 🔲 outside ## Backlog section must be ignored
cat > docs/backlog/BACKLOG.md <<'MD'
## Definition of Done
- ✅ shipped, 🔲 open, 🟡 in-progress

## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| RIGHT | **Right one** | x | Both | 1 | 🔲 |

## Notes
🔲 this should be ignored
MD
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug=="RIGHT"' >/dev/null; then
  ok "🔲 outside Backlog section is ignored"
else
  fail "picked wrong row: $OUT"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: pick-scope.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
