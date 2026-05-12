#!/usr/bin/env bash
#
# smell-scan.sh — the reproducible (non-LLM) half of design-judge.
#
# Encodes the architecture-smell rules from
# docs/workflows/pre-merge-check.md §6a as a single executable check.
# Failures here are BLOCKERs — taste-level review still runs locally
# via the agent-driven design-judge.
#
# Exit codes:
#   0 — all thresholds satisfied
#   1 — at least one rule fired
#
# Update thresholds when shipping a refactor that legitimately lowers
# the floor (e.g. CMDTEST consolidation dropped clones from 5 → 1).

set -euo pipefail

# 0. Quick `cd` to repo root so the script works from anywhere.
cd "$(dirname "$0")/.."

fail=0

# Rule 1 — duplicate cmdtest packages. After CMDTEST consolidation
# there should be exactly one shared cmdtest package at
# pkg/cmd/internal/cmdtest. Per-group clones are a regression.
clones=$(find pkg/cmd -path '*/internal/cmdtest/cmdtest.go' | wc -l | tr -d ' ')
if [ "$clones" -gt 1 ]; then
  echo "BLOCKER: $clones cmdtest.go files found; expected exactly 1 (pkg/cmd/internal/cmdtest/cmdtest.go)." >&2
  find pkg/cmd -path '*/internal/cmdtest/cmdtest.go' >&2
  fail=1
fi

# Rule 2 — N-way capability switch density. Three or more files each
# stacking As<X>Client triplets in pkg/cmd is the resolveOps signal.
switch_hits=$( { grep -rEo 'As[A-Z][a-zA-Z]+Client\(\)' pkg/cmd 2>/dev/null || true; } | wc -l | tr -d ' ')
if [ "$switch_hits" -ge 9 ]; then
  echo "WARN: $switch_hits As<X>Client() hits in pkg/cmd. Three sibling files of triplets is the resolveOps trigger." >&2
fi

# Rule 3 — translation-table sprawl in api/backend/types.go. Each To*/
# *ToCLI pair is a candidate for collapsing onto a typed enum with
# Marshal/Unmarshal methods.
pairs=$(grep -cE '^func (To|.*ToCLI)' api/backend/types.go 2>/dev/null || echo 0)
if [ "$pairs" -ge 6 ]; then
  echo "BLOCKER: $pairs translation-table funcs in api/backend/types.go (>=6). Collapse to typed enum + Marshal/Unmarshal." >&2
  fail=1
fi

# Rule 4 — comment density on the current diff vs origin/main (only when
# origin/main is reachable, i.e. CI on a PR). Skipped on push to main.
if git rev-parse --verify origin/main >/dev/null 2>&1 && \
   [ "$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then
  added_go=$( { git diff origin/main...HEAD -- '*.go' 2>/dev/null | grep -E '^\+[^+]' || true; } | wc -l | tr -d ' ')
  added_comments=$( { git diff origin/main...HEAD -- '*.go' 2>/dev/null | grep -E '^\+\s*//' || true; } | wc -l | tr -d ' ')
  if [ "${added_go:-0}" -ge 200 ] && [ "${added_comments:-0}" -gt 0 ]; then
    # Integer math: bail to bash arithmetic, avoid bc dependency.
    pct=$(( added_comments * 100 / added_go ))
    if [ "$pct" -gt 5 ]; then
      echo "WARN: comment density ${pct}% on this diff (${added_comments}/${added_go}); CLAUDE.md prefers minimal comments." >&2
    fi
  fi
fi

if [ "$fail" -eq 0 ]; then
  echo "smell-scan: ok (cmdtest clones=$clones, AsXClient hits=$switch_hits, translation funcs=$pairs)"
fi

exit "$fail"
