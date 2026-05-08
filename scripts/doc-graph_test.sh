#!/usr/bin/env bash
# doc-graph_test.sh — fixture-based tests for doc-graph.sh
# Exit 0 = all assertions passed. Exit 1 = failure.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/doc-graph.sh"

if [[ ! -x "$SCRIPT" ]]; then
  echo "FAIL: $SCRIPT not found or not executable" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Fixture setup
# ---------------------------------------------------------------------------
TMPDIR_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_ROOT"' EXIT

REPO="$TMPDIR_ROOT/repo"
mkdir -p "$REPO"

# Make it look like a git repo so doc-graph.sh can find root via git
cd "$REPO"
git init -q
git config user.email "test@test.com"
git config user.name "Test"

# --- orphan.md: no inbound links from outside its own directory ---
mkdir -p "$REPO/docs"
cat > "$REPO/docs/orphan.md" <<'EOF'
# Orphan
This file has no inbound links.
EOF

# --- target.md: referenced by linker.md ---
cat > "$REPO/docs/target.md" <<'EOF'
# Target
Referenced file.
EOF

# --- broken.md: references a file that does not exist ---
cat > "$REPO/docs/broken.md" <<'EOF'
# Broken
See [missing](missing-file.md) for details.
EOF

# --- stale-ref.md: contains "sections 0-4" but linked-doc has 5 sections ---
cat > "$REPO/docs/stale-ref.md" <<'EOF'
# Stale ref
See sections 0-4 in [stale-target.md](stale-target.md).
EOF

# stale-target.md: has 5 sections (## headings), so "0-4" is wrong
cat > "$REPO/docs/stale-target.md" <<'EOF'
# Doc
## Section 1
## Section 2
## Section 3
## Section 4
## Section 5
EOF

# --- linker.md at root: links to docs/target.md, so target.md is NOT orphan ---
cat > "$REPO/README.md" <<'EOF'
# Root
See [target](docs/target.md) for more.
EOF

# Commit the fixture so git ls-files works
git add -A
git commit -q -m "fixture"

# ---------------------------------------------------------------------------
# Run the script; capture output (script writes findings to stderr)
# ---------------------------------------------------------------------------
OUTPUT="$("$SCRIPT" 2>&1 || true)"

# ---------------------------------------------------------------------------
# Assertions
# ---------------------------------------------------------------------------
FAILURES=0

assert_contains() {
  local label="$1"
  local pattern="$2"
  if echo "$OUTPUT" | grep -qE "$pattern"; then
    echo "PASS: $label"
  else
    echo "FAIL: $label — pattern '$pattern' not found in output"
    FAILURES=$((FAILURES + 1))
  fi
}

assert_not_contains() {
  local label="$1"
  local pattern="$2"
  if echo "$OUTPUT" | grep -qE "$pattern"; then
    echo "FAIL: $label — pattern '$pattern' found but should NOT be"
    FAILURES=$((FAILURES + 1))
  else
    echo "PASS: $label"
  fi
}

# Orphan: docs/orphan.md has zero inbound links → must appear
assert_contains "orphan.md flagged as orphan" "orphan\.md"

# target.md is linked from README.md → must NOT appear as orphan
assert_not_contains "target.md NOT flagged as orphan" "Orphan.*target\.md|target\.md.*[Oo]rphan"

# Broken link: broken.md → missing-file.md
assert_contains "broken link missing-file.md detected" "missing-file\.md"

# Stale section ref: stale-ref.md contains "sections 0-4"
assert_contains "stale section ref in stale-ref.md detected" "stale-ref\.md"

# Section headers present in output
assert_contains "orphan section header" "## Orphan docs"
assert_contains "broken links section header" "## Broken links"
assert_contains "stale section refs header" "## Stale section refs"

# ---------------------------------------------------------------------------
# Exit
# ---------------------------------------------------------------------------
if [[ $FAILURES -gt 0 ]]; then
  echo ""
  echo "FAILED: $FAILURES assertion(s). Script output was:"
  echo "$OUTPUT"
  exit 1
fi

echo ""
echo "All assertions passed."
