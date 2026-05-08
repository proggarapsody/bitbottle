#!/usr/bin/env bash
# doc-graph.sh — detect three classes of doc drift in the repo.
#
# Exit codes:
#   0 — completed (findings, if any, written to stderr as WARN)
#   1 — script error (e.g. cannot find repo root)
#
# Classes detected:
#   1. Orphan docs    — *.md with zero inbound markdown links from outside own dir
#   2. Broken links   — [](path.md) references to non-existent files
#   3. Stale sections — literal "sections 0-N" strings (human verification)
#
# Requires: bash (any version), git, grep, python3 (for path normalisation)
set -euo pipefail

# ---------------------------------------------------------------------------
# Resolve repo root
# ---------------------------------------------------------------------------
if ! REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
  echo "doc-graph: error: cannot determine repo root (not inside a git repo)" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Temp files (cleaned up on exit)
# ---------------------------------------------------------------------------
TMPDIR_DG="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_DG"' EXIT

INBOUND_FILE="$TMPDIR_DG/inbound.txt"   # one abs-path per line (targets seen)
BROKEN_FILE="$TMPDIR_DG/broken.txt"     # findings
STALE_FILE="$TMPDIR_DG/stale.txt"       # findings
ALL_MD_FILE="$TMPDIR_DG/all_md.txt"     # relative paths of all *.md in scope
LINKS_FILE="$TMPDIR_DG/links.txt"       # raw link extraction output (src|target)

touch "$INBOUND_FILE" "$BROKEN_FILE" "$STALE_FILE" "$ALL_MD_FILE" "$LINKS_FILE"

# ---------------------------------------------------------------------------
# Collect tracked *.md files (git ls-files respects .gitignore)
# Exclude: node_modules/, reference/, dist/, conductor/
# ---------------------------------------------------------------------------
cd "$REPO_ROOT"
git ls-files '*.md' 2>/dev/null \
  | grep -vE '^(node_modules|reference|dist|conductor)/' \
  > "$ALL_MD_FILE" || true

if [[ ! -s "$ALL_MD_FILE" ]]; then
  exit 0
fi

# ---------------------------------------------------------------------------
# normalise_path: resolve a possibly non-existent path (python3 for portability)
# ---------------------------------------------------------------------------
normalise_path() {
  python3 -c "import os,sys; print(os.path.normpath(sys.argv[1]))" "$1"
}

# ---------------------------------------------------------------------------
# Phase 1: extract all markdown links across the repo into LINKS_FILE
#          Format: SRC_REL|TARGET_REL
# Use python3 for the extraction to avoid bash subshell/pipe write issues.
# ---------------------------------------------------------------------------
python3 - "$REPO_ROOT" "$ALL_MD_FILE" "$LINKS_FILE" <<'PYEOF'
import sys, os, re

repo_root = sys.argv[1]
all_md_file = sys.argv[2]
links_file = sys.argv[3]

link_re = re.compile(r'\]\(([^)]+\.md)(?:#[^)]+)?\)')
skip_schemes = ('http://', 'https://', 'mailto:', 'ftp://')

with open(all_md_file) as f:
    md_files = [line.rstrip('\n') for line in f if line.strip()]

out_lines = []
for src_rel in md_files:
    src_abs = os.path.join(repo_root, src_rel)
    src_dir = os.path.dirname(src_abs)
    try:
        content = open(src_abs, errors='replace').read()
    except OSError:
        continue
    for target in link_re.findall(content):
        if any(target.startswith(s) for s in skip_schemes):
            continue
        out_lines.append(src_rel + '|' + target)

with open(links_file, 'w') as f:
    f.write('\n'.join(out_lines))
    if out_lines:
        f.write('\n')
PYEOF

# ---------------------------------------------------------------------------
# Phase 2: process links — build inbound set, detect broken
# ---------------------------------------------------------------------------
while IFS='|' read -r src_rel raw_target; do
  [[ -z "$src_rel" ]] && continue

  src_abs="$REPO_ROOT/$src_rel"
  src_dir="$(dirname "$src_abs")"

  # Strip anchor from raw_target
  target="${raw_target%%#*}"
  [[ -z "$target" ]] && continue

  # Resolve to absolute path
  if [[ "$target" == /* ]]; then
    target_abs="$(normalise_path "$REPO_ROOT$target")"
  else
    target_abs="$(normalise_path "$src_dir/$target")"
  fi

  # Record inbound link
  echo "$target_abs" >> "$INBOUND_FILE"

  # Check existence
  if [[ ! -f "$target_abs" ]]; then
    echo "$src_rel: broken link → $target" >> "$BROKEN_FILE"
  fi
done < "$LINKS_FILE"

# ---------------------------------------------------------------------------
# Phase 3: orphan docs — files with no entry in INBOUND_FILE
# ---------------------------------------------------------------------------
ORPHAN_FILE="$TMPDIR_DG/orphans.txt"
touch "$ORPHAN_FILE"

while IFS= read -r src_rel; do
  [[ -z "$src_rel" ]] && continue
  src_abs_norm="$(normalise_path "$REPO_ROOT/$src_rel")"

  if ! grep -qF "$src_abs_norm" "$INBOUND_FILE" 2>/dev/null; then
    echo "$src_rel" >> "$ORPHAN_FILE"
  fi
done < "$ALL_MD_FILE"

# ---------------------------------------------------------------------------
# Phase 4: stale section refs
# ---------------------------------------------------------------------------
while IFS= read -r src_rel; do
  [[ -z "$src_rel" ]] && continue
  src_abs="$REPO_ROOT/$src_rel"
  grep -inE 'sections? *0[-–][0-9]+|sections? *[0-9]+' "$src_abs" 2>/dev/null \
    | while IFS= read -r line; do
        echo "$src_rel: $line"
      done >> "$STALE_FILE" || true
done < "$ALL_MD_FILE"

# ---------------------------------------------------------------------------
# Output — always to stderr; always exit 0 (WARN gate, not a blocker)
# ---------------------------------------------------------------------------
{
  echo "## Orphan docs:"
  if [[ ! -s "$ORPHAN_FILE" ]]; then
    echo "  (none)"
  else
    while IFS= read -r f; do echo "  $f"; done < "$ORPHAN_FILE"
  fi

  echo ""
  echo "## Broken links:"
  if [[ ! -s "$BROKEN_FILE" ]]; then
    echo "  (none)"
  else
    while IFS= read -r f; do echo "  $f"; done < "$BROKEN_FILE"
  fi

  echo ""
  echo "## Stale section refs:"
  if [[ ! -s "$STALE_FILE" ]]; then
    echo "  (none)"
  else
    while IFS= read -r f; do echo "  $f"; done < "$STALE_FILE"
  fi
} >&2

exit 0
