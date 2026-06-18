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

# ---------------------------------------------------------------------------
# Helper: create a stub `gh` binary in a temp bin dir and prepend it to PATH.
# Call with the JSON array it should return for both issue list invocations.
# ---------------------------------------------------------------------------
GH_BIN_DIR="$TMPDIR/stub-bin"
mkdir -p "$GH_BIN_DIR"

install_gh_stub() {
  local nightly_json="$1"
  local bug_json="${2:-[]}"
  # Write a stub that returns the appropriate JSON based on the --label flag.
  cat > "$GH_BIN_DIR/gh" <<STUB
#!/usr/bin/env bash
# Stub gh for pick-scope tests.
for arg in "\$@"; do
  case "\$arg" in
    nightly-e2e) printf '%s\n' '$nightly_json'; exit 0 ;;
    bug)         printf '%s\n' '$bug_json'; exit 0 ;;
  esac
done
printf '[]\n'
exit 0
STUB
  chmod +x "$GH_BIN_DIR/gh"
  export PATH="$GH_BIN_DIR:$PATH"
}

remove_gh_stub() {
  # Replace with a broken stub so pick-scope falls back to BACKLOG.
  cat > "$GH_BIN_DIR/gh" <<'STUB'
#!/usr/bin/env bash
exit 1
STUB
  chmod +x "$GH_BIN_DIR/gh"
}

# ============================================================
# BACKLOG PATH (Cases 1-6): stub gh returns empty arrays so
# pick-scope falls through to BACKLOG every time.
# ============================================================
install_gh_stub "[]" "[]"

# Case 1: backlog with one open scope, no Scope Details section -> has_scope_details=false
cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| FOO | **Foo Scope** | `foo list` — list foos | Both | 2 | ✅ |
| BAR | **Bar Scope** | `bar list` — list bars. Cloud only. | Cloud | 2 | 🔲 |
MD
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug=="BAR" and .scope_name=="Bar Scope" and .backend=="Cloud" and .tier=="2" and .has_scope_details==false and .details_anchor==null and .source=="backlog"' >/dev/null; then
  ok "newer inline scope -> has_scope_details=false, anchor=null, source=backlog"
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
if echo "$OUT" | jq -e '.slug=="L" and .has_scope_details==true and (.details_anchor | startswith("#l-")) and .source=="backlog"' >/dev/null; then
  ok "older scope -> has_scope_details=true with anchor, source=backlog"
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

# ============================================================
# GH ISSUE PRIORITY PATH (Cases 7-11)
# ============================================================

# Re-create a basic BACKLOG for fallback cases
cat > docs/backlog/BACKLOG.md <<'MD'
## Backlog
| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| BACKLOG-ITEM | **Backlog Item** | `x` | Both | 2 | 🔲 |
MD

# Case 7: open nightly-e2e issue -> emits it as next scope (source=gh_issue)
install_gh_stub '[{"number":42,"title":"nightly-e2e: pr edit wipes reviewers","labels":[{"name":"nightly-e2e"}]}]' '[]'
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug==42 and .scope_name=="nightly-e2e: pr edit wipes reviewers" and .source=="gh_issue"' >/dev/null; then
  ok "open nightly-e2e issue -> emitted as scope with source=gh_issue"
else
  fail "got: $OUT"
fi

# Case 8: open bug issue -> emitted as next scope (source=gh_issue)
install_gh_stub '[]' '[{"number":99,"title":"bug: pr list crashes on empty repo","labels":[{"name":"bug"}]}]'
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug==99 and .scope_name=="bug: pr list crashes on empty repo" and .source=="gh_issue"' >/dev/null; then
  ok "open bug issue -> emitted as scope with source=gh_issue"
else
  fail "got: $OUT"
fi

# Case 9: both nightly-e2e and bug issues -> newest (highest number) wins
install_gh_stub \
  '[{"number":55,"title":"nightly-e2e: timeout","labels":[{"name":"nightly-e2e"}]}]' \
  '[{"number":77,"title":"bug: crash on empty","labels":[{"name":"bug"}]}]'
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug==77' >/dev/null; then
  ok "when both labels have issues, highest number (newest) wins"
else
  fail "expected slug 77, got: $OUT"
fi

# Case 10: gh returns non-zero (offline/no auth) -> falls back to BACKLOG
remove_gh_stub
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug=="BACKLOG-ITEM" and .source=="backlog"' >/dev/null; then
  ok "gh failure falls back to BACKLOG (source=backlog)"
else
  fail "expected BACKLOG-ITEM fallback, got: $OUT"
fi

# Case 11: gh returns empty arrays (no matching issues) -> falls back to BACKLOG
install_gh_stub "[]" "[]"
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.slug=="BACKLOG-ITEM" and .source=="backlog"' >/dev/null; then
  ok "no open issues -> falls back to BACKLOG (source=backlog)"
else
  fail "expected BACKLOG-ITEM fallback, got: $OUT"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: pick-scope.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
