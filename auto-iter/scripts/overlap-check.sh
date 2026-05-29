#!/usr/bin/env bash
# Detect whether any open PR overlaps the candidate scope by keyword match.
#
# Usage:
#   auto-iter/scripts/overlap-check.sh <scope-slug>
#
# Output:
#   {"scope":"...","overlapping_pr":N|null,"matched_keywords":[...],
#    "all_open_prs":[{num,title,score}]}
#
# Keyword set: scope slug tokens (split on -) + lowercased words from
# docs/backlog/BACKLOG.md row Commands cell, length >= 3, excluding common stop-words.
# A PR is considered overlapping if it scores >= 2 keyword matches against
# its title+body.
#
# DETECTION ONLY — the resolution decision (resolve/skip/close) stays in the
# orchestrator's halt routing.
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cd "$(repo_root)"

SCOPE="${1:-}"
[[ -z "$SCOPE" ]] && halt "missing_arg" "overlap-check.sh requires a scope slug"

if ! command -v gh >/dev/null 2>&1; then
  emit_json --scope="$SCOPE" --raw overlapping_pr=null \
    --raw matched_keywords='[]' --raw all_open_prs='[]'
  exit 0
fi

# --- build keyword set ---
# Stop-words as space-padded string for portable lookup (bash 3 lacks
# associative arrays — macOS default bash is 3.2).
STOP=" the and for list add get set new one all via api from with both cloud server only paginated returns "
keywords=()
add_kw() {
  local raw="$1"
  local k
  k="$(echo "$raw" | tr '[:upper:]' '[:lower:]')"
  k="${k//[^a-z0-9]/ }"
  for w in $k; do
    (( ${#w} < 3 )) && continue
    case "$STOP" in *" $w "*) continue ;; esac
    keywords+=("$w")
  done
}
add_kw "${SCOPE//-/ }"

if [[ -f docs/backlog/BACKLOG.md ]]; then
  row=$(awk -v slug="$SCOPE" '
    /^## Backlog/ { inblk=1; next }
    inblk && /^## / { inblk=0 }
    inblk && $0 ~ "\\| " slug " \\|" { print; exit }
  ' docs/backlog/BACKLOG.md || true)
  if [[ -n "$row" ]]; then
    IFS='|' read -r _ _ _ commands _ _ _ <<<"$row"
    add_kw "$commands"
  fi
fi

# Dedupe keywords (portable bash 3 — no mapfile)
sorted_kw=$(printf '%s\n' "${keywords[@]:-}" | sort -u | grep -v '^$' || true)
keywords=()
while IFS= read -r w; do
  [[ -n "$w" ]] && keywords+=("$w")
done <<<"$sorted_kw"

if (( ${#keywords[@]} == 0 )); then
  emit_json --scope="$SCOPE" --raw overlapping_pr=null \
    --raw matched_keywords='[]' --raw all_open_prs='[]'
  exit 0
fi

# --- fetch open PRs ---
prs_json="$(gh pr list --state=open --limit=50 \
  --json number,title,body 2>/dev/null || echo '[]')"

# --- score each PR ---
scored=$(echo "$prs_json" | jq --arg kw "$(printf '%s\n' "${keywords[@]}")" '
  ($kw | split("\n") | map(select(length>0))) as $keys
  | map(. + {
      score: ([(.title // ""), (.body // "")] | join(" ") | ascii_downcase
              | . as $t | ($keys | map(. as $k | select($t | contains($k))) | length))
    })
  | map({num: .number, title: .title, score: .score})
')

# Pick the PR with the highest score >= 2, if any.
top=$(echo "$scored" | jq 'map(select(.score >= 2)) | sort_by(-.score) | .[0] // null')

if [[ "$top" == "null" ]]; then
  emit_json --scope="$SCOPE" --raw overlapping_pr=null \
    --raw matched_keywords="$(printf '%s\n' "${keywords[@]}" | jq -Rsc 'split("\n") | map(select(length>0))')" \
    --raw all_open_prs="$scored"
else
  top_num=$(echo "$top" | jq '.num')
  matched=$(echo "$prs_json" | jq -c --arg kw "$(printf '%s\n' "${keywords[@]}")" --argjson n "$top_num" '
    ($kw | split("\n") | map(select(length>0))) as $keys
    | .[] | select(.number == $n)
    | (.title + " " + (.body // "") | ascii_downcase) as $t
    | $keys | map(. as $k | select($t | contains($k)))
  ')
  emit_json --scope="$SCOPE" --raw overlapping_pr="$top_num" \
    --raw matched_keywords="$matched" --raw all_open_prs="$scored"
fi
