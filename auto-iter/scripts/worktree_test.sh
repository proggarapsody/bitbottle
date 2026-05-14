#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/worktree.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== worktree_test ==="

PARENT=$(mktemp -d)
trap "rm -rf '$PARENT'" EXIT

# Create a "bitbottle" repo under PARENT — worktree paths land in PARENT/bitbottle-worktrees
REPO="$PARENT/bitbottle"
mkdir -p "$REPO"
cd "$REPO"
git init -q -b main
git config user.email t@t.com
git config user.name T
echo init > README.md
git add README.md
git commit -q -m "chore: init"

# Case 1: create
OUT="$(bash "$SCRIPT" create WORKSPACE-MEMBERS)"
if echo "$OUT" | jq -e '.created==true and (.branch=="feat/workspace-members")' >/dev/null; then
  ok "create lowercases + hyphen-normalises slug into branch"
else
  fail "got: $OUT"
fi
EXPECTED_PATH="$PARENT/bitbottle-worktrees/workspace-members"
if [[ -d "$EXPECTED_PATH" ]]; then
  ok "worktree directory created at sibling path"
else
  fail "expected $EXPECTED_PATH to exist"
fi

# Case 2: duplicate create -> halt already_exists
if OUT="$(bash "$SCRIPT" create WORKSPACE-MEMBERS 2>/dev/null)"; then
  fail "expected non-zero exit on duplicate create"
else
  if echo "$OUT" | jq -e '.halt=="already_exists"' >/dev/null; then
    ok "duplicate create -> already_exists halt"
  else
    fail "wrong halt: $OUT"
  fi
fi

# Case 3: remove
OUT="$(bash "$SCRIPT" remove "$EXPECTED_PATH")"
if echo "$OUT" | jq -e '.removed==true' >/dev/null; then
  ok "remove returns removed=true"
else
  fail "got: $OUT"
fi
if [[ ! -d "$EXPECTED_PATH" ]]; then
  ok "worktree directory removed"
else
  fail "expected $EXPECTED_PATH to be gone"
fi

# Case 4: custom prefix
OUT="$(bash "$SCRIPT" create some-fix --prefix=fix)"
if echo "$OUT" | jq -e '.branch=="fix/some-fix"' >/dev/null; then
  ok "--prefix=fix honored"
else
  fail "got: $OUT"
fi

# Case 5: bad prefix
if bash "$SCRIPT" create another --prefix=banana 2>/dev/null; then
  fail "expected non-zero on bad prefix"
else
  ok "bad prefix rejected"
fi

# Case 6: missing subcommand
if bash "$SCRIPT" 2>/dev/null; then
  fail "expected non-zero on missing subcommand"
else
  ok "missing subcommand rejected"
fi

# Case 7: remove non-existent
if bash "$SCRIPT" remove "$PARENT/nope" 2>/dev/null; then
  fail "expected non-zero on remove non-existent"
else
  ok "remove non-existent rejected"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: worktree.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
