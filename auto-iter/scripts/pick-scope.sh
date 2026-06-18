#!/usr/bin/env bash
# Pick the next open scope from docs/backlog/BACKLOG.md's ## Backlog table.
#
# Priority order (highest first):
#   1. Open GitHub issues labelled "nightly-e2e" or "bug" (real-backend
#      failures jump the queue so the loop addresses live breakage first).
#   2. First 🔲 row in the ## Backlog table.
#
# Reads:
#   docs/backlog/BACKLOG.md  (queue — unshipped scopes only after 2026-05-29 reorg)
#   gh issue list --label nightly-e2e|bug (when gh is available and online)
#
# Emits:
#   {"slug":"WORKSPACE-MEMBERS","scope_name":"Workspace Members",
#    "summary":"...","backend":"Cloud","tier":"2",
#    "has_scope_details":bool,"details_anchor":"#workspace-members"|null,
#    "source":"backlog"|"gh_issue"}
#
# The "source" field distinguishes how the scope was selected:
#   "gh_issue"  — selected from an open GitHub issue (nightly-e2e or bug)
#   "backlog"   — selected from the ## Backlog table
#
# Halt:
#   {"halt":"backlog_empty","details":"no 🔲 rows in ## Backlog"}
#
# Two BACKLOG shapes are supported:
#   - Older scopes with a `## Scope Details` -> `### <slug> — Name` anchor
#   - Newer (brainstorm-added) scopes that put the full description inline
#     in the Commands column and have no anchor
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cd "$(repo_root)"
BACKLOG_FILE="docs/backlog/BACKLOG.md"

if [[ ! -f "$BACKLOG_FILE" ]]; then
  halt "backlog_missing" "$BACKLOG_FILE not found"
fi

# ---------------------------------------------------------------------------
# Phase 1: Check open GitHub issues labelled "nightly-e2e" or "bug".
# If gh is unavailable or returns non-zero (offline, no auth), fall back
# silently to the BACKLOG. We query both labels, merge, deduplicate by
# number, and pick the highest-numbered (newest) issue.
# ---------------------------------------------------------------------------
issue_json=""
if command -v gh >/dev/null 2>&1; then
  # Capture output; if either call fails, we get an empty string or partial
  # output — both are handled gracefully below.
  e2e_issues="$(gh issue list --label nightly-e2e --state open --json number,title,labels --limit 5 2>/dev/null || true)"
  bug_issues="$(gh issue list --label bug --state open --json number,title,labels --limit 5 2>/dev/null || true)"

  # Merge the two arrays (may be empty strings or "[]"), deduplicate by
  # number, and pick the one with the highest number (newest).
  # Note: jq unique_by re-sorts, so we sort_by descending AFTER deduplication.
  if [[ -n "$e2e_issues" || -n "$bug_issues" ]]; then
    e2e_safe="${e2e_issues:-[]}"
    bug_safe="${bug_issues:-[]}"
    issue_json="$(printf '%s\n%s\n' "$e2e_safe" "$bug_safe" \
      | jq -sc 'add // [] | unique_by(.number) | sort_by(.number) | reverse | .[0] // empty' 2>/dev/null || true)"
  fi
fi

if [[ -n "$issue_json" ]] && echo "$issue_json" | jq -e '.number' >/dev/null 2>&1; then
  # Emit the issue as the next scope.
  issue_number="$(echo "$issue_json" | jq '.number')"
  issue_title="$(echo "$issue_json" | jq -r '.title')"
  emit_json \
    --raw "slug=$issue_number" \
    --scope_name="$issue_title" \
    --summary="$issue_title" \
    --backend="n/a" \
    --str tier="1" \
    --has_scope_details=false \
    --raw details_anchor=null \
    --source="gh_issue"
  exit 0
fi

# ---------------------------------------------------------------------------
# Phase 2: Fall back to the BACKLOG table.
# ---------------------------------------------------------------------------

# Find the first 🔲 row inside the ## Backlog section.
# Skip rows marked 🔲📝 — those are manual-only (e.g. external form
# submission, third-party signup) and not eligible for the auto-iter loop.
# The picker would otherwise pick them, the cycle would skip them, and the
# row would block the picker on every subsequent cycle until a human
# resolves it. Treat 🔲📝 as "not 🔲" for picker purposes.
row=$(awk '
  /^## Backlog/         { inblk=1; next }
  inblk && /^## /       { inblk=0 }
  inblk && /🔲📝/      { next }
  inblk && /🔲/         { print; exit }
' "$BACKLOG_FILE")

if [[ -z "$row" ]]; then
  halt "backlog_empty" "no 🔲 rows in ## Backlog (manual-only 🔲📝 rows skipped)"
fi

# Strip leading/trailing | and whitespace from each pipe-separated cell.
# Expected shape: | ID | Scope | Commands | Backends | Tier | Status |
IFS='|' read -r _ slug scope summary backend tier _status _ <<<"$row"
trim() { local s="${1#"${1%%[![:space:]]*}"}"; echo "${s%"${s##*[![:space:]]}"}"; }
slug=$(trim "$slug")
scope_name=$(trim "$scope")
# Strip leading **bold** markers if present: **Foo** -> Foo
scope_name="${scope_name#\*\*}"; scope_name="${scope_name%\*\*}"
summary=$(trim "$summary")
backend=$(trim "$backend")
tier=$(trim "$tier")

# Look for a Scope Details anchor: `### <slug> ` or `### <slug> — `
has_details=false
anchor=""
if grep -qE "^### ${slug}( |$|—)" "$BACKLOG_FILE"; then
  has_details=true
  # Build the GitHub-flavored anchor from the heading text.
  heading=$(grep -E "^### ${slug}( |$|—)" "$BACKLOG_FILE" | head -1 | sed 's/^### //')
  anchor="#$(echo "$heading" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9 -]//g; s/ +/-/g')"
fi

if [[ "$has_details" == "true" ]]; then
  emit_json \
    --slug="$slug" --scope_name="$scope_name" --summary="$summary" \
    --backend="$backend" --str tier="$tier" \
    --has_scope_details=true --details_anchor="$anchor" \
    --source="backlog"
else
  emit_json \
    --slug="$slug" --scope_name="$scope_name" --summary="$summary" \
    --backend="$backend" --str tier="$tier" \
    --has_scope_details=false --raw details_anchor=null \
    --source="backlog"
fi
