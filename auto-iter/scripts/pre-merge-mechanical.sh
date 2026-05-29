#!/usr/bin/env bash
# Run the mechanical sections of the pre-merge gate.
# Canonical source: docs/workflows/pre-merge-check.md §1, §2, §3, §7.
# Judgment sections (§4 lint/tests via CI, §5 doc-sync, §6 design-judge,
# §8 secrets via gitleaks) stay in the orchestrator / CI.
#
# Output:
#   {"findings":[{"section":"...","check":"...","message":"...","severity":"BLOCKER"}],
#    "blocker":bool}
#
# Exit 0 always (caller reads .blocker). Halt envelopes only for unrecoverable
# script-level errors (no git repo, etc.).
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cd "$(repo_root)"

findings=()
add_finding() {
  local section="$1" check="$2" message="$3"
  findings+=("$(jq -nc --arg s "$section" --arg c "$check" --arg m "$message" \
    '{section:$s,check:$c,message:$m,severity:"BLOCKER"}')")
}

branch="$(git rev-parse --abbrev-ref HEAD)"

# === §1 Branch & tree hygiene ===
if [[ "$branch" == "main" ]]; then
  add_finding "1" "branch_is_main" "cannot merge from main into main"
elif ! [[ "$branch" =~ ^(feature|feat|fix|docs|chore|refactor|test|perf|build|ci|style|revert)/ ]]; then
  add_finding "1" "branch_name" "branch '$branch' doesn't match ^(feature|feat|fix|docs|chore|...)/ — see pre-merge-check.md §1"
fi
if [[ -n "$(git status --porcelain)" ]]; then
  add_finding "1" "dirty_tree" "git status --porcelain is non-empty"
fi
if git rev-parse --verify --quiet origin/main >/dev/null; then
  behind=$(git rev-list --count HEAD..origin/main 2>/dev/null || echo 0)
  if (( behind > 0 )); then
    add_finding "1" "behind_origin" "branch is $behind commit(s) behind origin/main"
  fi
fi

# === §2 Conventional Commits + squash-merge title gotcha ===
# Pattern from pre-merge-check.md.
CC_RE='^(feat|fix|docs|chore|refactor|test|perf|build|ci|style|revert)(\(.+\))?!?: '
if git rev-parse --verify --quiet origin/main >/dev/null && [[ "$branch" != "main" ]]; then
  commits=$(git log origin/main..HEAD --format='%s' 2>/dev/null || true)
  while IFS= read -r subj; do
    [[ -z "$subj" ]] && continue
    if ! [[ "$subj" =~ $CC_RE ]]; then
      add_finding "2" "conventional_commits" "commit subject doesn't match Conventional Commits: $subj"
    fi
  done <<<"$commits"

  # Squash-merge gotcha: if any commit is feat/fix, the PR title must be too.
  has_release_commit="false"
  while IFS= read -r subj; do
    [[ -z "$subj" ]] && continue
    if [[ "$subj" =~ ^(feat|fix)(\(.+\))?!?: ]]; then
      has_release_commit="true"; break
    fi
  done <<<"$commits"

  if [[ "$has_release_commit" == "true" ]] && command -v gh >/dev/null 2>&1; then
    pr_title=$(gh pr view --json title --jq .title 2>/dev/null || true)
    if [[ -n "$pr_title" ]] && ! [[ "$pr_title" =~ ^(feat|fix)(\(.+\))?!?: ]]; then
      add_finding "2" "squash_title_mismatch" \
        "branch has feat/fix commits but PR title '$pr_title' is not feat:/fix: — release-please won't bump"
    fi
  fi
fi

# === §3 Build artifacts ===
if [[ -n "$(git ls-files dist/ 2>/dev/null)" ]]; then
  add_finding "3" "dist_tracked" "files under dist/ are tracked (must be gitignored)"
fi
big_files=$(git ls-files | while read -r f; do
  [[ -f "$f" ]] || continue
  sz=$(du -k "$f" 2>/dev/null | awk '{print $1}')
  [[ -n "$sz" && "$sz" -gt 1024 ]] && echo "$f ($sz KB)"
done || true)
if [[ -n "$big_files" ]]; then
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    add_finding "3" "file_too_large" "$line — exceeds 1 MB"
  done <<<"$big_files"
fi
if git ls-files | grep -qE '^bitbottle$'; then
  add_finding "3" "compiled_binary_tracked" "compiled bitbottle binary is tracked (should be gitignored)"
fi
if git ls-files | grep -qE '(\.DS_Store|\.log$|coverage\.out)'; then
  add_finding "3" "artifact_tracked" "tracked .DS_Store / *.log / coverage.out files present"
fi

# === §4 BACKLOG/SHIPPED isolation check ===
# Reorg (2026-05-29): shipping a scope is a MOVE from docs/backlog/BACKLOG.md to
# docs/backlog/SHIPPED.md, performed in the same commit as the feat work. So a
# commit that touches ONLY one of those files (and nothing else) is the
# "separate chore PR" anti-pattern. The move must land in the same commit as the
# feat work. See iteration-cycle §4 and quickref anti-patterns.
if git rev-parse --verify --quiet origin/main >/dev/null && [[ "$branch" != "main" ]]; then
  while IFS= read -r sha; do
    [[ -z "$sha" ]] && continue
    files_in_commit=$(git diff-tree --no-commit-id -r --name-only "$sha" 2>/dev/null || true)
    file_count=$(printf '%s\n' "$files_in_commit" | grep -c '[^[:space:]]' || true)
    if [[ "$file_count" -eq 1 ]]; then
      if printf '%s\n' "$files_in_commit" | grep -qxE "(docs/backlog/BACKLOG\.md|docs/backlog/SHIPPED\.md|BACKLOG\.md)"; then
        only=$(printf '%s\n' "$files_in_commit" | head -1)
        add_finding "4" "backlog_flip_isolated" \
          "commit $sha touches only $only — the BACKLOG→SHIPPED move must land in the same commit as the feat work, not a standalone chore commit"
      fi
    fi
  done < <(git log origin/main..HEAD --format='%H' 2>/dev/null || true)
fi

# === §7 Release-please boundaries ===
# On non-release branches, must not touch the manifest, CHANGELOG, or
# x-release-please-version markers.
if [[ ! "$branch" =~ ^release-please-- ]]; then
  if git rev-parse --verify --quiet origin/main >/dev/null; then
    changed=$(git diff --name-only origin/main..HEAD 2>/dev/null || true)
    for f in CHANGELOG.md .release-please-manifest.json; do
      if echo "$changed" | grep -qxF "$f"; then
        add_finding "7" "release_please_boundary" "$f modified on non-release branch (release-please owns this file)"
      fi
    done
    # Don't grep through commits — that's slow; the doc says the markers
    # live in skills/SKILL.md + README.md as `<!-- x-release-please-version -->`
    # comments that release-please rewrites. Manual edits to those lines
    # are also boundary violations.
    if echo "$changed" | grep -qxF "skills/SKILL.md"; then
      if git diff origin/main..HEAD -- skills/SKILL.md 2>/dev/null | grep -qE '^\+.*x-release-please-version'; then
        add_finding "7" "release_please_boundary" "skills/SKILL.md x-release-please-version marker touched manually"
      fi
    fi
  fi
fi

# === Emit ===
findings_json="[]"
if (( ${#findings[@]} > 0 )); then
  findings_json=$(printf '%s\n' "${findings[@]}" | jq -sc '.')
fi
blocker=$(jq -nc --argjson f "$findings_json" '$f | length > 0')
jq -nc --argjson findings "$findings_json" --argjson blocker "$blocker" \
  '{findings:$findings, blocker:$blocker}'
