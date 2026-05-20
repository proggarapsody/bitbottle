#!/usr/bin/env bash
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/pre-merge-mechanical.sh"
FAIL=0
ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== pre-merge-mechanical_test ==="

TMPDIR=$(mktemp -d)
STUBDIR=$(mktemp -d)
trap "rm -rf '$TMPDIR' '$STUBDIR'" EXIT

# Stub gh so the squash-title check has a deterministic path
cat > "$STUBDIR/gh" <<'GH'
#!/usr/bin/env bash
exit 1
GH
chmod +x "$STUBDIR/gh"
export PATH="$STUBDIR:$PATH"

cd "$TMPDIR"
git init -q -b main
git config user.email t@t.com
git config user.name T
echo init > README.md
git add README.md
git commit -q -m "chore: init"

# Simulate origin/main = current commit
git update-ref refs/remotes/origin/main HEAD

# Case 1: clean feat branch, valid commits, no artifacts -> no blocker
git checkout -q -b feat/foo
echo new > thing.txt
git add thing.txt
git commit -q -m "feat(scope): add thing"
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.blocker==false and (.findings | length)==0' >/dev/null; then
  ok "clean feat branch -> no findings"
else
  fail "got: $OUT"
fi

# Case 2: bad conventional commits
git commit -q --allow-empty -m "not a real conventional commit"
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.blocker==true and (.findings | map(.check) | index("conventional_commits") != null)' >/dev/null; then
  ok "non-CC commit -> conventional_commits finding"
else
  fail "got: $OUT"
fi
git reset --hard HEAD~1 -q

# Case 3: dirty tree
echo dirty > scratch.txt
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.findings | map(.check) | index("dirty_tree") != null' >/dev/null; then
  ok "dirty tree -> dirty_tree finding"
else
  fail "got: $OUT"
fi
rm scratch.txt

# Case 4: on main branch -> branch_is_main blocker
git checkout -q main
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.findings | map(.check) | index("branch_is_main") != null' >/dev/null; then
  ok "main branch -> branch_is_main blocker"
else
  fail "got: $OUT"
fi

# Case 5: dist/ tracked
git checkout -q feat/foo
mkdir -p dist && echo build > dist/output
git add dist/output
git commit -q -m "chore: bad dist add"
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.findings | map(.check) | index("dist_tracked") != null' >/dev/null; then
  ok "tracked dist/ -> dist_tracked finding"
else
  fail "got: $OUT"
fi
git rm -rq dist && git commit -q -m "chore: remove dist"

# Case 6: branch name with bad prefix
git checkout -q -b weirdbranch
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.findings | map(.check) | index("branch_name") != null' >/dev/null; then
  ok "bad branch prefix -> branch_name finding"
else
  fail "got: $OUT"
fi

# Case 7: CHANGELOG.md modified on non-release branch
git checkout -q feat/foo
echo "x" > CHANGELOG.md
git add CHANGELOG.md
git commit -q -m "chore: touch changelog"
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.findings | map(.check) | index("release_please_boundary") != null' >/dev/null; then
  ok "CHANGELOG.md touched on feat branch -> boundary finding"
else
  fail "got: $OUT"
fi

# Cases 8-9: use a fresh branch to avoid CHANGELOG noise from case 7
git checkout -q origin/main
git checkout -q -b feat/backlog-test

# Case 8: BACKLOG.md flip in standalone commit -> backlog_flip_isolated BLOCKER
echo "feat code" > myfile.go
git add myfile.go
git commit -q -m "feat(x): add feature code"
echo "| PIPELINE-OBSERVABILITY | ✅ |" > BACKLOG.md
git add BACKLOG.md
git commit -q -m "chore: mark shipped in BACKLOG"  # standalone BACKLOG-only commit
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '.findings | map(.check) | index("backlog_flip_isolated") != null' >/dev/null; then
  ok "standalone BACKLOG-only commit -> backlog_flip_isolated finding"
else
  fail "expected backlog_flip_isolated, got: $OUT"
fi

# Case 9: BACKLOG.md flip in same feat commit -> no backlog_flip_isolated
git reset --hard HEAD~2 -q  # back to origin/main equivalent
echo "feat code" > myfile.go
echo "| PIPELINE-OBSERVABILITY | ✅ |" > BACKLOG.md
git add myfile.go BACKLOG.md
git commit -q -m "feat(x): ship scope and flip BACKLOG"
OUT="$(bash "$SCRIPT")"
if echo "$OUT" | jq -e '[.findings[] | select(.check=="backlog_flip_isolated")] | length == 0' >/dev/null; then
  ok "BACKLOG flip co-located in feat commit -> no backlog_flip_isolated"
else
  fail "expected no backlog_flip_isolated, got: $OUT"
fi

echo ""
[[ $FAIL -eq 0 ]] && echo "PASS: pre-merge-mechanical.sh tests OK" || { echo "FAIL: assertions failed"; exit 1; }
