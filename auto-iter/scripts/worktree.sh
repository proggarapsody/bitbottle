#!/usr/bin/env bash
# Manage auto-iter worktrees. Convention from docs/workflows/iteration-cycle.md:
#   git worktree add -b feat/<short-slug> ../bitbottle-worktrees/<slug> main
#
# Usage:
#   auto-iter/scripts/worktree.sh create <slug> [--prefix=feat|fix|docs|chore|refactor]
#   auto-iter/scripts/worktree.sh remove <path>
#
# Output (create):
#   {"path":"../bitbottle-worktrees/<slug>","branch":"<prefix>/<slug>","created":true}
# Output (remove):
#   {"path":"<path>","removed":true}
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cd "$(repo_root)"
ROOT="$(repo_root)"
WORKTREE_ROOT="$(dirname "$ROOT")/bitbottle-worktrees"

cmd="${1:-}"
shift || true

case "$cmd" in
  create)
    slug="${1:-}"; shift || true
    [[ -z "$slug" ]] && halt "missing_arg" "worktree.sh create requires <slug>"
    prefix="feat"
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --prefix=*) prefix="${1#--prefix=}" ;;
        *) halt "bad_flag" "worktree.sh: unknown flag $1" ;;
      esac
      shift
    done
    if ! [[ "$prefix" =~ ^(feat|fix|docs|chore|refactor|test|perf|build|ci|style|revert)$ ]]; then
      halt "bad_prefix" "prefix '$prefix' not in Conventional Commits set"
    fi
    # Slug lowercased + hyphen-normalised for the branch.
    branch_slug="$(echo "$slug" | tr '[:upper:]_' '[:lower:]-')"
    branch="$prefix/$branch_slug"
    path="$WORKTREE_ROOT/$branch_slug"
    mkdir -p "$WORKTREE_ROOT"
    if [[ -d "$path" ]]; then
      halt "already_exists" "worktree path $path already exists"
    fi
    git worktree add -b "$branch" "$path" main >&2
    emit_json --path="$path" --branch="$branch" --created=true
    ;;
  remove)
    path="${1:-}"
    [[ -z "$path" ]] && halt "missing_arg" "worktree.sh remove requires <path>"
    if [[ ! -d "$path" ]]; then
      halt "not_found" "worktree path $path does not exist"
    fi
    git worktree remove "$path" --force >&2
    emit_json --path="$path" --removed=true
    ;;
  "")
    halt "missing_subcommand" "worktree.sh: expected create|remove"
    ;;
  *)
    halt "unknown_subcommand" "worktree.sh: expected create|remove, got $cmd"
    ;;
esac
