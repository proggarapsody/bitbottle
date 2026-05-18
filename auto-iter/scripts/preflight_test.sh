#!/usr/bin/env bash
# Unit test for auto-iter/scripts/preflight.sh.
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
trap "rm -rf '$TMPDIR' '$STUBDIR' '${EXTRA_ROOT:-}'" EXIT

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

# Case 4: behind origin on clean main → auto-reset, exit 0 (no halt).
# Branch protection blocks direct pushes; local main is read-only at cycle
# boundaries, so being behind origin is the EXPECTED state after the previous
# cycle's PR merged. Preflight resolves by hard-reset, not by halting.
git checkout -q main
# Set up the bare remote + helper worktree OUTSIDE $TMPDIR so they don't
# show up as untracked dirs in `git status --porcelain` of the fixture.
EXTRA_ROOT="$(mktemp -d)"
BARE="$EXTRA_ROOT/origin.git"
git init -q --bare -b main "$BARE" 2>/dev/null || git init -q --bare "$BARE"
git remote add origin "$BARE"
git push -q origin main
# Set HEAD on the bare to main so subsequent clones get the right default.
git --git-dir="$BARE" symbolic-ref HEAD refs/heads/main
# Add a commit on origin that local doesn't have.
WORK="$EXTRA_ROOT/work-extra"
git clone -q "$BARE" "$WORK"
( cd "$WORK" && git config user.email "x@x.com" && git config user.name x \
    && echo "extra" > extra.txt && git add extra.txt \
    && git commit -q -m "extra commit on origin" \
    && git push -q origin main )
# Local main is now 1 behind, 0 ahead, clean.
if OUT="$(bash "$SCRIPT" 2>/dev/null)"; then
  ok "exits 0 when behind origin on clean main (auto-reset path)"
else
  fail "expected exit 0 (auto-reset), got non-zero: $OUT"
fi
if echo "$OUT" | jq -e '.behind==0 and (.findings | map(select(startswith("behind_origin_reset"))) | length > 0)' >/dev/null; then
  ok "reports behind_origin_reset finding and behind=0 after reset"
else
  fail "expected behind_origin_reset finding, got: $OUT"
fi
# Confirm the working tree actually picked up the origin commit.
if [[ -f extra.txt ]]; then
  ok "hard-reset pulled in origin commit"
else
  fail "expected extra.txt after reset"
fi

# Case 5: behind AND ahead (diverged) → halt, do NOT auto-reset.
echo "local change" > local.txt
git add local.txt
git commit -q -m "local commit not on origin"
# Now: ahead=1, behind=0 — fast-forward origin first to recreate divergence.
( cd "$WORK" && echo "another" > another.txt && git add another.txt \
    && git commit -q -m "another on origin" && git push -q origin main )
git fetch -q origin
# Now: ahead=1, behind=1 → diverged.
if OUT="$(bash "$SCRIPT" 2>/dev/null)"; then
  fail "expected non-zero exit on diverged state"
else
  ok "exits non-zero when diverged (ahead + behind)"
fi
if echo "$OUT" | jq -e '(.findings | map(select(startswith("behind_origin:"))) | length > 0)' >/dev/null; then
  ok "reports behind_origin (halt-class) when also ahead — does NOT auto-reset"
else
  fail "expected behind_origin halt finding on diverged state, got: $OUT"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "PASS: preflight.sh tests OK"
else
  echo "FAIL: one or more assertions failed"
  exit 1
fi
