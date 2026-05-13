#!/usr/bin/env bash
# Unit test for scripts/update-flake-shas.sh
# Uses a fixture checksums.txt and a copy of flake.nix; verifies SHAs and version are rewritten.
set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/update-flake-shas.sh"
FLAKE="$(cd "$(dirname "$0")/.." && pwd)/flake.nix"
FAIL=0

ok()   { echo "  OK  $*"; }
fail() { echo "  FAIL $*"; FAIL=1; }

echo "=== update-flake-shas_test ==="

# Setup temp workspace
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Create a minimal fixture flake.nix with placeholder SHAs
cat > "$TMPDIR/flake.nix" << 'NIX'
{
  outputs = { self, nixpkgs }:
    let
      version = "1.41.0";
      assetFor = system:
        let
          map = {
            "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"; };
            "aarch64-linux"  = { os = "linux";  arch = "arm64"; sha256 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"; };
            "x86_64-darwin"  = { os = "darwin"; arch = "amd64"; sha256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"; };
            "aarch64-darwin" = { os = "darwin"; arch = "arm64"; sha256 = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"; };
          };
        in map.${system};
    in {};
}
NIX

# Create fixture checksums.txt
cat > "$TMPDIR/checksums.txt" << 'SUMS'
1111111111111111111111111111111111111111111111111111111111111111  bitbottle_linux_amd64.tar.gz
2222222222222222222222222222222222222222222222222222222222222222  bitbottle_linux_arm64.tar.gz
3333333333333333333333333333333333333333333333333333333333333333  bitbottle_darwin_amd64.tar.gz
4444444444444444444444444444444444444444444444444444444444444444  bitbottle_darwin_arm64.tar.gz
SUMS

# Run a patched version of the script against the fixture.
# We patch the curl call by pre-seeding the checksums file in $TMPDIR.
# The script uses mktemp internally, so we replicate its logic inline.
(
    cd "$TMPDIR"
    # Init a bare git repo so the script can commit
    git init -q
    git config user.email "test@test.com"
    git config user.name "Test"
    git add flake.nix
    git commit -q -m "init"

    # Override curl by providing a wrapper on PATH
    mkdir -p bin
    # shellcheck disable=SC2016
    FIXTURE="$TMPDIR/checksums.txt"
    cat > bin/curl << CURL
#!/usr/bin/env bash
# Fake curl: parse -o <dest> and copy fixture checksums.txt there
while [[ \$# -gt 0 ]]; do
    if [[ "\$1" == "-o" ]]; then
        cp "${FIXTURE}" "\$2"
        exit 0
    fi
    shift
done
CURL
    chmod +x bin/curl
    export PATH="$TMPDIR/bin:$PATH"

    bash "$SCRIPT" "v9.9.9"
)

UPDATED="$TMPDIR/flake.nix"

# Check version updated
if grep -q 'version = "9.9.9"' "$UPDATED"; then
    ok "version rewritten to 9.9.9"
else
    fail "version not rewritten"
    grep 'version' "$UPDATED" || true
fi

# Check SHAs rewritten
if grep -q '1111111111111111111111111111111111111111111111111111111111111111' "$UPDATED"; then
    ok "linux/amd64 SHA rewritten"
else
    fail "linux/amd64 SHA not rewritten"
fi
if grep -q '2222222222222222222222222222222222222222222222222222222222222222' "$UPDATED"; then
    ok "linux/arm64 SHA rewritten"
else
    fail "linux/arm64 SHA not rewritten"
fi
if grep -q '3333333333333333333333333333333333333333333333333333333333333333' "$UPDATED"; then
    ok "darwin/amd64 SHA rewritten"
else
    fail "darwin/amd64 SHA not rewritten"
fi
if grep -q '4444444444444444444444444444444444444444444444444444444444444444' "$UPDATED"; then
    ok "darwin/arm64 SHA rewritten"
else
    fail "darwin/arm64 SHA not rewritten"
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
    echo "PASS: update-flake-shas.sh test OK"
else
    echo "FAIL: one or more assertions failed"
    exit 1
fi
