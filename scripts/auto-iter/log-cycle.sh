#!/usr/bin/env bash
# Append a single cycle outcome line to .claude/auto-iter/cycles.jsonl.
#
# Usage:
#   scripts/auto-iter/log-cycle.sh --cycle=42 --mode=iteration \
#     --scope=DEPLOY-KEY --outcome=shipped --pr=234 --release=v1.46.0
#
# Required: --cycle (or --stream=started|completed), --outcome (for cycle entries).
# Optional: any number of --key=value pairs.
# Adds: ts=<now>.
#
# Canonical schema: docs/auto-iter/quickref.md § Cycle log schema.
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

# Pre-flight: must have either --cycle or --stream.
seen_cycle=0; seen_stream=0; seen_outcome=0
for arg in "$@"; do
  case "$arg" in
    --cycle=*)   seen_cycle=1 ;;
    --stream=*)  seen_stream=1 ;;
    --outcome=*) seen_outcome=1 ;;
  esac
done
if [[ $seen_cycle -eq 0 && $seen_stream -eq 0 ]]; then
  halt "missing_required_field" "log-cycle.sh requires --cycle=N or --stream=started|completed"
fi
if [[ $seen_cycle -eq 1 && $seen_outcome -eq 0 ]]; then
  halt "missing_required_field" "cycle rows require --outcome=<enum> (see quickref Outcome enum)"
fi

ensure_auto_iter_dir
LINE="$(emit_json --ts="$(now_iso)" "$@")"
printf '%s\n' "$LINE" >> "$(auto_iter_dir)/cycles.jsonl"
printf '%s\n' "$LINE"
