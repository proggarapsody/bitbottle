#!/usr/bin/env bash
# Append a single cycle outcome line to .claude/auto-iter/cycles.jsonl.
#
# Usage:
#   auto-iter/scripts/log-cycle.sh --cycle=42 --mode=iteration \
#     --scope=DEPLOY-KEY --outcome=shipped --pr=234 --release=v1.46.0
#
# Required: --cycle (or --stream=started|completed), --outcome (for cycle entries).
# Optional: any number of --key=value pairs.
# Adds: ts=<now>.
#
# Canonical schema: auto-iter/quickref.md § Cycle log schema.
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

# Count step entries this cycle wrote to metrics.jsonl so post-hoc analysis can
# spot telemetry gaps. Cycle 93 (the 75-min release-cascade) wrote 4 step lines
# total when a healthy shipped cycle writes 8-12. Surfacing the count alongside
# the outcome makes drift visible in a single jq query instead of cross-file
# diffing two JSONLs.
#
# Only computed for cycle entries (--cycle=N). Stream rows (--stream=...) skip.
extra_args=()
if [[ $seen_cycle -eq 1 ]]; then
  cycle_value=""
  for arg in "$@"; do
    [[ "$arg" == --cycle=* ]] && cycle_value="${arg#--cycle=}"
  done
  metrics_file="$(auto_iter_dir)/metrics.jsonl"
  steps_count=0
  if [[ -n "$cycle_value" ]] && [[ -f "$metrics_file" ]]; then
    steps_count="$(jq -s --argjson c "$cycle_value" \
      'map(select(.cycle==$c and (.step // null) != null)) | length' \
      "$metrics_file" 2>/dev/null || echo 0)"
    [[ -z "$steps_count" || "$steps_count" == "null" ]] && steps_count=0
  fi
  extra_args+=(--metrics_steps_count="$steps_count")
fi

LINE="$(emit_json --ts="$(now_iso)" "$@" ${extra_args[@]+"${extra_args[@]}"})"
printf '%s\n' "$LINE" >> "$(auto_iter_dir)/cycles.jsonl"
printf '%s\n' "$LINE"
