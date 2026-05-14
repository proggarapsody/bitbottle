#!/usr/bin/env bash
# Re-entry lock for /auto-iter. Owns .claude/auto-iter/.lock.
#
# Usage:
#   auto-iter/scripts/lock.sh acquire        # exit 0 + JSON if lock taken, exit 1 if recent (<60min) lock present
#   auto-iter/scripts/lock.sh release        # always exit 0
#   auto-iter/scripts/lock.sh status         # exit 0 + JSON describing current state
#
# Acquire semantics:
#   - No lock file       -> create, emit {"acquired":true,"age_min":0}
#   - Lock <60 min old   -> emit {"halt":"recent_lock","age_min":N}, exit 1
#   - Lock >60 min old   -> assume crashed predecessor, overwrite, emit {"acquired":true,"stale_lock":true,"age_min":N}
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

LOCK_FILE="$(auto_iter_dir)/.lock"
STALE_THRESHOLD_MIN=60

cmd="${1:-status}"

# Returns age in minutes (integer, floor). Empty string if file missing.
lock_age_min() {
  [[ -f "$LOCK_FILE" ]] || return 0
  local mtime now
  if [[ "$OSTYPE" == darwin* ]]; then
    mtime=$(stat -f %m "$LOCK_FILE")
  else
    mtime=$(stat -c %Y "$LOCK_FILE")
  fi
  now=$(date +%s)
  echo $(( (now - mtime) / 60 ))
}

case "$cmd" in
  acquire)
    ensure_auto_iter_dir
    age="$(lock_age_min || true)"
    if [[ -z "$age" ]]; then
      printf 'pid:%d\nts:%s\n' "$$" "$(now_iso)" > "$LOCK_FILE"
      emit_json --acquired=true --age_min=0
    elif (( age < STALE_THRESHOLD_MIN )); then
      emit_json --halt="recent_lock" --age_min="$age"
      exit 1
    else
      printf 'pid:%d\nts:%s\n' "$$" "$(now_iso)" > "$LOCK_FILE"
      emit_json --acquired=true --stale_lock=true --age_min="$age"
    fi
    ;;
  release)
    rm -f "$LOCK_FILE"
    emit_json --released=true
    ;;
  status)
    age="$(lock_age_min || true)"
    if [[ -z "$age" ]]; then
      emit_json --held=false
    else
      emit_json --held=true --age_min="$age"
    fi
    ;;
  *)
    halt "unknown_subcommand" "lock.sh: expected acquire|release|status, got $cmd"
    ;;
esac
