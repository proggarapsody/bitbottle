#!/usr/bin/env bash
# Pick the next open scope from BACKLOG.md's ## Backlog table.
#
# Reads:
#   BACKLOG.md
#
# Emits:
#   {"slug":"WORKSPACE-MEMBERS","scope_name":"Workspace Members",
#    "summary":"...","backend":"Cloud","tier":"2",
#    "has_scope_details":bool,"details_anchor":"#workspace-members"|null}
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
BACKLOG_FILE="BACKLOG.md"

if [[ ! -f "$BACKLOG_FILE" ]]; then
  halt "backlog_missing" "$BACKLOG_FILE not found"
fi

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
    --has_scope_details=true --details_anchor="$anchor"
else
  emit_json \
    --slug="$slug" --scope_name="$scope_name" --summary="$summary" \
    --backend="$backend" --str tier="$tier" \
    --has_scope_details=false --raw details_anchor=null
fi
