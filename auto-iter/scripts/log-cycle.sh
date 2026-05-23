#!/usr/bin/env bash
# Append a single cycle outcome line to .claude/auto-iter/cycles.jsonl.
#
# Usage:
#   auto-iter/scripts/log-cycle.sh --cycle=42 --mode=iteration \
#     --scope=DEPLOY-KEY --outcome=shipped --pr=234 --release=v1.46.0
#
# --scope and --pr may be repeated to build arrays:
#   --scope=FOO --scope=BAR   emits "scopes":["FOO","BAR"]
#   --pr=111 --pr=222         emits "prs":[111,222]
#
# Required: --cycle (or --stream=started|completed), --outcome (for cycle entries).
# Optional: --duration_min=N (wall-clock, becomes both duration_min and duration_wall_min).
# Adds: ts=<now>, pipeline_version, tokens, duration_active_min (from cycle-summary.sh),
#       metrics_steps_count.
#
# Canonical schema: docs/workflows/iteration-cycle/quickref.md § Cycle log schema.
set -euo pipefail

PIPELINE_VERSION="2026.05.20"

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

# Re-parse args: convert bare --scope=X and --pr=X to --arr scope=X / --arr pr=X.
# All other args pass through. Also extract cycle_value and duration_min.
translated_args=()
cycle_value=""
duration_min_value=""
for arg in "$@"; do
  case "$arg" in
    --scope=*)
      translated_args+=(--arr "scope=${arg#--scope=}")
      ;;
    --pr=*)
      translated_args+=(--arr "pr=${arg#--pr=}")
      ;;
    --cycle=*)
      cycle_value="${arg#--cycle=}"
      translated_args+=("$arg")
      ;;
    --duration_min=*)
      duration_min_value="${arg#--duration_min=}"
      translated_args+=("$arg")
      ;;
    *)
      translated_args+=("$arg")
      ;;
  esac
done

# Compute summary fields (tokens, duration_active_min, metrics_steps_count)
# via cycle-summary.sh. Only for cycle entries.
extra_args=()
if [[ $seen_cycle -eq 1 && -n "$cycle_value" ]]; then
  metrics_file="$(auto_iter_dir)/metrics.jsonl"
  summary="$(bash "$DIR/cycle-summary.sh" --cycle="$cycle_value" --metrics="$metrics_file")"
  tokens="$(printf '%s' "$summary" | jq '.tokens')"
  duration_active_min="$(printf '%s' "$summary" | jq '.duration_active_min')"
  steps_count="$(printf '%s' "$summary" | jq '.metrics_steps_count')"

  extra_args+=(
    --raw "tokens=$tokens"
    --raw "duration_active_min=$duration_active_min"
    --raw "metrics_steps_count=$steps_count"
    --str "pipeline_version=$PIPELINE_VERSION"
  )

  # duration_wall_min: alias for duration_min when provided
  if [[ -n "$duration_min_value" ]]; then
    extra_args+=(--raw "duration_wall_min=$duration_min_value")
  fi
fi

LINE="$(emit_json --ts="$(now_iso)" "${translated_args[@]}" ${extra_args[@]+"${extra_args[@]}"})"
printf '%s\n' "$LINE" >> "$(auto_iter_dir)/cycles.jsonl"
printf '%s\n' "$LINE"
