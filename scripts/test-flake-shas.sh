#!/usr/bin/env bash
# Validate that flake.nix is structurally well-formed:
# - Contains the expected version
# - Contains all 4 platform entries
# - All sha256 fields are non-empty 64-char hex strings
# - Has the expected URL pattern
set -euo pipefail

FLAKE="${1:-flake.nix}"
FAIL=0

ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== test-flake-shas: validating $FLAKE ==="

# 1. File exists
[[ -f "$FLAKE" ]] || { echo "FAIL: $FLAKE not found"; exit 1; }

# 2. Contains a version line
if grep -qE 'version = "[0-9]+\.[0-9]+\.[0-9]+";' "$FLAKE"; then
    VERSION=$(grep -oE 'version = "[0-9]+\.[0-9]+\.[0-9]+"' "$FLAKE" | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
    ok "version = $VERSION"
else
    fail "no valid version line found"
fi

# 3. All 4 platforms present
for combo in "linux.*amd64" "linux.*arm64" "darwin.*amd64" "darwin.*arm64"; do
    if grep -qE "$combo" "$FLAKE"; then
        ok "platform $combo"
    else
        fail "platform $combo missing"
    fi
done

# 4. All sha256 values are 64-char hex
SHA_COUNT=$(grep -oE 'sha256 = "[0-9a-f]{64}"' "$FLAKE" | wc -l | tr -d ' ')
if [[ "$SHA_COUNT" -ge 4 ]]; then
    ok "$SHA_COUNT sha256 entries (64-char hex)"
else
    fail "expected >=4 sha256 entries with 64-char hex; got $SHA_COUNT"
    echo "  sha256 lines found:"
    grep 'sha256 = ' "$FLAKE" || true
fi

# 5. URL pattern present
if grep -q 'github.com/proggarapsody/bitbottle/releases/download' "$FLAKE"; then
    ok "release URL pattern present"
else
    fail "release URL pattern missing"
fi

# 6. MIT license referenced
if grep -q 'licenses.mit' "$FLAKE"; then
    ok "MIT license"
else
    fail "MIT license not referenced"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
    echo "PASS: flake.nix validates OK"
else
    echo "FAIL: one or more checks failed"
    exit 1
fi
