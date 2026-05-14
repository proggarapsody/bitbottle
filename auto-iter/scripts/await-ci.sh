#!/usr/bin/env bash
# Wait for CI to finish on a PR, emit a structured result.
#
# Usage:
#   auto-iter/scripts/await-ci.sh <pr-number> [--timeout-min=30] [--interval-sec=30]
#
# Output (success):
#   {"pr":N,"all_passed":bool,"failed":["check","..."],
#    "skipped":["check","..."],"elapsed_min":N}
#
# Halt:
#   {"halt":"timeout","details":"...","pr":N,"elapsed_min":N}
#   {"halt":"no_gh","details":"gh CLI not available"}
#
# Exit 0 when CI has settled (regardless of pass/fail — caller reads
# .all_passed). Exit 1 only on halt (timeout, no gh).
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

PR=""
TIMEOUT_MIN=30
INTERVAL_SEC=30

while [[ $# -gt 0 ]]; do
  case "$1" in
    --timeout-min=*) TIMEOUT_MIN="${1#--timeout-min=}" ;;
    --interval-sec=*) INTERVAL_SEC="${1#--interval-sec=}" ;;
    --*) halt "bad_flag" "await-ci.sh: unknown flag $1" ;;
    *)  [[ -z "$PR" ]] && PR="$1" || halt "bad_arg" "await-ci.sh: extra positional arg $1" ;;
  esac
  shift
done

[[ -z "$PR" ]] && halt "missing_pr" "await-ci.sh requires a PR number"
command -v gh >/dev/null 2>&1 || halt "no_gh" "gh CLI not available"

START=$(date +%s)
TIMEOUT_SEC=$(( TIMEOUT_MIN * 60 ))

while true; do
  output=$(gh pr checks "$PR" --json name,state,bucket 2>/dev/null || echo '[]')
  pending=$(echo "$output" | jq '[.[] | select(.bucket=="pending")] | length')
  if [[ "$pending" == "0" ]]; then
    break
  fi
  elapsed=$(( $(date +%s) - START ))
  if (( elapsed >= TIMEOUT_SEC )); then
    emit_json --halt="timeout" \
      --details="CI still pending after ${TIMEOUT_MIN} min" \
      --pr="$PR" --elapsed_min="$(( elapsed / 60 ))"
    exit 1
  fi
  sleep "$INTERVAL_SEC"
done

elapsed=$(( $(date +%s) - START ))
elapsed_min=$(( elapsed / 60 ))

failed_json=$(echo "$output" | jq -c '[.[] | select(.bucket=="fail") | .name]')
skipped_json=$(echo "$output" | jq -c '[.[] | select(.bucket=="skipping") | .name]')
all_passed=$(echo "$output" | jq 'all(.[]; .bucket == "pass" or .bucket == "skipping")')

emit_json --pr="$PR" \
  --raw all_passed="$all_passed" \
  --raw failed="$failed_json" \
  --raw skipped="$skipped_json" \
  --elapsed_min="$elapsed_min"
