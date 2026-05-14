#!/usr/bin/env bash
# Append a single metric row to .claude/auto-iter/metrics.jsonl.
#
# Usage:
#   auto-iter/scripts/metric.sh --cycle=42 --step=step2_design_judge \
#     --duration_ms=85000 --subagent_tokens=72000 --findings_count=2
#
# Required: --cycle, --step.
# Optional: any number of --key=value pairs (typed per emit_json rules).
# Adds: ts=<now>.
#
# Canonical schema list: auto-iter/quickref.md § Metrics log schema.
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

# Pre-flight: require --cycle and --step.
seen_cycle=0; seen_step=0
for arg in "$@"; do
  case "$arg" in
    --cycle=*) seen_cycle=1 ;;
    --step=*)  seen_step=1 ;;
  esac
done
if [[ $seen_cycle -eq 0 || $seen_step -eq 0 ]]; then
  halt "missing_required_field" "metric.sh requires --cycle=N and --step=NAME"
fi

ensure_auto_iter_dir
LINE="$(emit_json --ts="$(now_iso)" "$@")"
printf '%s\n' "$LINE" >> "$(auto_iter_dir)/metrics.jsonl"
printf '%s\n' "$LINE"
