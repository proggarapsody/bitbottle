#!/usr/bin/env bash
# Wait for a release to land on both GitHub and npm.
#
# Usage:
#   auto-iter/scripts/await-publish.sh <version> [--timeout-min=15] [--interval-sec=30]
#
# <version> may be plain (1.61.0) or v-prefixed (v1.61.0). Internally normalises.
#
# Output (success):
#   {"version":"1.61.0","github":true,"npm":true,"elapsed_min":N}
# Halt (timeout):
#   {"halt":"timeout","version":"...","github":bool,"npm":bool,"elapsed_min":N}
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

# --- args ---
VERSION=""
TIMEOUT_MIN=15
INTERVAL_SEC=30
NPM_PKG="${AUTO_ITER_NPM_PKG:-@proggarapsody/bitbottle}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --timeout-min=*)  TIMEOUT_MIN="${1#--timeout-min=}" ;;
    --interval-sec=*) INTERVAL_SEC="${1#--interval-sec=}" ;;
    --npm-pkg=*)      NPM_PKG="${1#--npm-pkg=}" ;;
    --*)              halt "bad_flag" "await-publish.sh: unknown flag $1" ;;
    *)  [[ -z "$VERSION" ]] && VERSION="$1" || halt "bad_arg" "extra positional: $1" ;;
  esac
  shift
done
[[ -z "$VERSION" ]] && halt "missing_version" "await-publish.sh requires a version"

# Normalise. Internal canonical form is the plain version; the GitHub tag has the v-prefix.
RAW="$VERSION"
PLAIN="${RAW#v}"
TAG="v${PLAIN}"

START=$(date +%s)
TIMEOUT_SEC=$(( TIMEOUT_MIN * 60 ))
github=false
npm=false

while true; do
  if [[ "$github" != "true" ]] && command -v gh >/dev/null 2>&1; then
    if gh release view "$TAG" >/dev/null 2>&1; then github=true; fi
  fi
  if [[ "$npm" != "true" ]] && command -v npm >/dev/null 2>&1; then
    if [[ "$(npm view "$NPM_PKG" version 2>/dev/null || true)" == "$PLAIN" ]]; then npm=true; fi
  fi
  if [[ "$github" == "true" && "$npm" == "true" ]]; then break; fi

  elapsed=$(( $(date +%s) - START ))
  if (( elapsed >= TIMEOUT_SEC )); then
    emit_json --halt="timeout" --version="$PLAIN" \
      --raw github="$github" --raw npm="$npm" \
      --elapsed_min="$(( elapsed / 60 ))"
    exit 1
  fi
  sleep "$INTERVAL_SEC"
done

elapsed=$(( $(date +%s) - START ))
emit_json --version="$PLAIN" --raw github=true --raw npm=true \
  --elapsed_min="$(( elapsed / 60 ))"
