#!/usr/bin/env bash
# Workspace inventory for §0 of /auto-iter.
#
# Emits a single JSON object describing the working tree:
#   {
#     "clean": true|false,
#     "branch": "main",
#     "on_main": true|false,
#     "ahead": <int>,        # commits ahead of origin/main
#     "behind": <int>,       # commits behind origin/main
#     "open_prs": [{"num":N,"title":"...","head":"branch","author":"login"}],
#     "findings": ["..."]    # human-readable issues that may need halt
#   }
#
# Exit codes:
#   0 - clean enough to proceed (no findings, or findings are informational)
#   1 - findings include a halt-class issue (dirty tree, divergence, etc.)
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cd "$(repo_root)"

branch="$(git rev-parse --abbrev-ref HEAD)"
on_main="false"
[[ "$branch" == "main" ]] && on_main="true"

# Refresh remote refs quietly (best-effort, skip if offline).
git fetch --quiet origin main 2>/dev/null || true

ahead=0; behind=0
if git rev-parse --verify --quiet origin/main >/dev/null; then
  ahead=$(git rev-list --count origin/main..HEAD 2>/dev/null || echo 0)
  behind=$(git rev-list --count HEAD..origin/main 2>/dev/null || echo 0)
fi

dirty_count=$(git status --porcelain | wc -l | tr -d ' ')
clean="true"
[[ "$dirty_count" != "0" ]] && clean="false"

# Open PRs (best-effort, skip gracefully if no gh / no auth).
open_prs_json="[]"
if command -v gh >/dev/null 2>&1; then
  open_prs_json="$(gh pr list --state=open --limit=20 \
      --json number,title,headRefName,author \
      --jq 'map({num:.number, title:.title, head:.headRefName, author:.author.login})' \
      2>/dev/null || echo '[]')"
fi

findings=()
halt_class=0
if [[ "$clean" == "false" ]]; then
  findings+=("workspace_dirty: $dirty_count uncommitted change(s)")
  halt_class=1
fi
if [[ "$on_main" == "false" ]]; then
  findings+=("not_on_main: currently on $branch")
  halt_class=1
fi
if (( behind > 0 )); then
  findings+=("behind_origin: $behind commit(s) behind origin/main")
  halt_class=1
fi
if (( ahead > 0 )); then
  findings+=("ahead_origin: $ahead unpushed commit(s)")
fi

findings_json=$(printf '%s\n' "${findings[@]:-}" \
  | jq -Rsc 'split("\n") | map(select(length>0))')

jq -nc \
  --argjson clean "$clean" \
  --arg branch "$branch" \
  --argjson on_main "$on_main" \
  --argjson ahead "$ahead" \
  --argjson behind "$behind" \
  --argjson open_prs "$open_prs_json" \
  --argjson findings "$findings_json" \
  '{clean:$clean, branch:$branch, on_main:$on_main, ahead:$ahead, behind:$behind, open_prs:$open_prs, findings:$findings}'

exit $halt_class
