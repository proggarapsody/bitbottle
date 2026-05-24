#!/usr/bin/env bash
# Workspace inventory for §0 of /auto-iter.
#
# nightly-e2e PRs: workflow-only changes; secrets configured separately in repo UI — not gated by normal preflight
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

# Wire .githooks/ into core.hooksPath if a tracked .githooks/ directory exists.
# Idempotent: only writes when the current value differs.
#
# Why this lives in preflight: cycle 96 pushed an ST1023 lint error that CI
# caught after the fact because `.githooks/pre-push` (which runs golangci-lint)
# was inactive — core.hooksPath defaulted to .git/hooks/. The hook existed and
# was executable, just not invoked. Wiring it here means every cycle's §0
# preflight repairs this for any agent or contributor whose clone never set it.
hooks_path_finding=""
if [[ -d .githooks ]]; then
  current_hooks="$(git config core.hooksPath 2>/dev/null || true)"
  if [[ "$current_hooks" != ".githooks" ]]; then
    git config core.hooksPath .githooks
    hooks_path_finding="hooks_path_wired: set core.hooksPath=.githooks (was '${current_hooks:-unset}')"
  fi
fi

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
[[ -n "$hooks_path_finding" ]] && findings+=("$hooks_path_finding")
if [[ "$clean" == "false" ]]; then
  findings+=("workspace_dirty: $dirty_count uncommitted change(s)")
  halt_class=1
fi
if [[ "$on_main" == "false" ]]; then
  findings+=("not_on_main: currently on $branch")
  halt_class=1
fi

# Behind-origin handling: branch protection blocks direct pushes to main, so
# local `main` is purely read-only at cycle boundaries. Being behind origin
# is the EXPECTED state after the previous cycle's PR squash-merged. Resolve
# by hard-reset, not by `git merge origin/main` (which the orchestrator
# previously did inline, littering main with 7 merge commits across the
# May-17 stream cycles 81–86).
#
# Halt-class only when local has unpushed commits (diverged state — could
# represent in-progress work that hard-reset would destroy).
if (( behind > 0 )); then
  if [[ "$on_main" == "true" ]] && [[ "$clean" == "true" ]] && (( ahead == 0 )); then
    git reset --hard "origin/main" >&2
    behind=0
    findings+=("behind_origin_reset: synced local main → origin/main (was $behind behind)")
  else
    findings+=("behind_origin: $behind commit(s) behind origin/main")
    halt_class=1
  fi
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
