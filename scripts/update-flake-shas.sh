#!/usr/bin/env bash
# Usage: scripts/update-flake-shas.sh vX.Y.Z
# Downloads checksums.txt for the given release tag and rewrites flake.nix
# with the new version and SHA256 values for the four platform tarballs.
set -euo pipefail

TAG="${1:?usage: $0 vX.Y.Z}"
VERSION="${TAG#v}"
CHECKSUMS_URL="https://github.com/proggarapsody/bitbottle/releases/download/${TAG}/checksums.txt"
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

echo "Fetching $CHECKSUMS_URL ..."
curl -fsSL "$CHECKSUMS_URL" -o "$TMPDIR/checksums.txt"

sha_for() {
    grep "bitbottle_${1}_${2}.tar.gz" "$TMPDIR/checksums.txt" | awk '{print $1}'
}

SHA_LINUX_AMD64=$(sha_for linux amd64)
SHA_LINUX_ARM64=$(sha_for linux arm64)
SHA_DARWIN_AMD64=$(sha_for darwin amd64)
SHA_DARWIN_ARM64=$(sha_for darwin arm64)

echo "linux/amd64:  $SHA_LINUX_AMD64"
echo "linux/arm64:  $SHA_LINUX_ARM64"
echo "darwin/amd64: $SHA_DARWIN_AMD64"
echo "darwin/arm64: $SHA_DARWIN_ARM64"

# Update version
sed -i.bak "s|version = \"[^\"]*\";|version = \"${VERSION}\";|" flake.nix && rm flake.nix.bak

# Update SHAs — each platform has a unique os+arch key on the same line as sha256
# Pattern: lines like: "x86_64-linux"   = { os = "linux";  arch = "amd64"; sha256 = "HEXHEX"; };
# We use awk to rewrite each matching line.
awk -v la="$SHA_LINUX_AMD64" -v lr="$SHA_LINUX_ARM64" \
    -v da="$SHA_DARWIN_AMD64" -v dr="$SHA_DARWIN_ARM64" '
/arch = "amd64"/ && /os = "linux"/  { sub(/sha256 = "[^"]*"/, "sha256 = \"" la "\"") }
/arch = "arm64"/ && /os = "linux"/  { sub(/sha256 = "[^"]*"/, "sha256 = \"" lr "\"") }
/arch = "amd64"/ && /os = "darwin"/ { sub(/sha256 = "[^"]*"/, "sha256 = \"" da "\"") }
/arch = "arm64"/ && /os = "darwin"/ { sub(/sha256 = "[^"]*"/, "sha256 = \"" dr "\"") }
{ print }
' flake.nix > flake.nix.new && mv flake.nix.new flake.nix

echo "flake.nix updated for $TAG"

git add flake.nix
if git diff --cached --quiet; then
    echo "No changes to commit (already up to date)"
else
    git commit -m "chore(nix): update flake SHAs for ${TAG}"
fi
