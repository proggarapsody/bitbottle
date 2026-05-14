#!/usr/bin/env bash
# Shared helpers for scripts/auto-iter/*.sh.
# Source from any auto-iter script: . "$(dirname "$0")/_common.sh"
#
# All scripts in this directory follow one convention:
#   - Exit 0 on success, write a single JSON object to stdout.
#   - Exit non-zero on failure, write {"halt":"<reason>","details":"..."} to stdout.
#   - Never write progress/log lines to stdout (use stderr for that).
# Canonical contract: docs/auto-iter/scripts.md.

set -euo pipefail

# Resolve repo root from the caller's current working directory.
# auto-iter scripts are always invoked from inside the repo they operate on
# (the orchestrator cd's first); tests rely on this to sandbox state in a
# temporary git repo.
repo_root() {
  git rev-parse --show-toplevel
}

# Path to .claude/auto-iter/ (gitignored sibling of cycles.jsonl).
auto_iter_dir() {
  printf '%s/.claude/auto-iter\n' "$(repo_root)"
}

# Ensure the auto-iter state dir exists. Idempotent.
ensure_auto_iter_dir() {
  mkdir -p "$(auto_iter_dir)"
}

# Emit a JSON object to stdout from --key=value args.
# Booleans (`true`/`false`), integers, and `null` are passed through; everything
# else is string-quoted. Use --raw key=expr to inject a pre-formatted JSON value.
#
# Example:
#   emit_json --cycle=42 --step=step2_tdd --raw findings='["a","b"]'
emit_json() {
  local jq_args=() filter='.'
  while [[ $# -gt 0 ]]; do
    local arg="$1"
    if [[ "$arg" == "--raw" ]]; then
      shift
      local key="${1%%=*}" expr="${1#*=}"
      jq_args+=(--argjson "$key" "$expr")
      filter="$filter | .${key}=\$${key}"
    elif [[ "$arg" == --*=* ]]; then
      local key="${arg#--}"; key="${key%%=*}"
      local val="${arg#*=}"
      if [[ "$val" =~ ^(true|false|null)$ ]] || [[ "$val" =~ ^-?[0-9]+$ ]]; then
        jq_args+=(--argjson "$key" "$val")
      else
        jq_args+=(--arg "$key" "$val")
      fi
      filter="$filter | .${key}=\$${key}"
    else
      printf 'emit_json: unsupported arg: %s\n' "$arg" >&2
      return 2
    fi
    shift
  done
  jq -nc "${jq_args[@]}" "$filter"
}

# Emit a halt envelope and exit non-zero.
halt() {
  local reason="$1" details="${2:-}"
  emit_json --halt="$reason" --details="$details"
  exit 1
}

# ISO-8601 UTC timestamp suitable for jsonl logs.
now_iso() {
  date -u +'%Y-%m-%dT%H:%M:%SZ'
}
