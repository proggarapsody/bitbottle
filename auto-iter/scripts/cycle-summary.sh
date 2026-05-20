#!/usr/bin/env bash
# Return single-line JSON summary for a given cycle from metrics.jsonl.
#
# Usage:
#   auto-iter/scripts/cycle-summary.sh --cycle=N --metrics=/path/to/metrics.jsonl
#
# Output (always exits 0; all-zeros when no data):
#   {"tokens":<int>,"duration_active_min":<float>,"metrics_steps_count":<int>}
#
# tokens              = sum of subagent_tokens (missing rows -> 0)
# duration_active_min = sum of duration_ms / 60000, rounded to 1 decimal
# metrics_steps_count = number of rows with a .step field for this cycle
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

cycle_value=""
metrics_file=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --cycle=*)   cycle_value="${1#--cycle=}" ;;
    --metrics=*) metrics_file="${1#--metrics=}" ;;
    *) printf 'cycle-summary.sh: unsupported arg: %s\n' "$1" >&2; exit 2 ;;
  esac
  shift
done

if [[ -z "$cycle_value" ]]; then
  printf 'cycle-summary.sh: --cycle=N required\n' >&2
  exit 2
fi

if [[ -z "$metrics_file" || ! -f "$metrics_file" ]]; then
  jq -nc '{tokens:0, duration_active_min:0.0, metrics_steps_count:0}'
  exit 0
fi

jq -sc \
  --argjson c "$cycle_value" \
  'map(select(.cycle==$c and (.step // null) != null)) |
   {
     tokens: (map(.subagent_tokens // 0) | add // 0),
     duration_active_min: ((map(.duration_ms // 0) | add // 0) / 60000 * 10 | round / 10),
     metrics_steps_count: length
   }' \
  "$metrics_file"
