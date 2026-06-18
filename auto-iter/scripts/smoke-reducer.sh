#!/usr/bin/env bash
# Stamp smoke=passed|failed on all cycles.jsonl rows whose release field
# matches the given version.
#
# Usage:
#   auto-iter/scripts/smoke-reducer.sh --version=v1.X.Y --result=passed|failed
#
# Finds all rows in .claude/auto-iter/cycles.jsonl where .release == "v1.X.Y"
# and rewrites the smoke field in-place (atomic tmp-file swap to prevent
# corruption on partial writes).
#
# Emits on stdout:
#   {"updated":<N>,"version":"v1.X.Y","result":"passed|failed"}
#
# where updated = number of rows that were stamped.
#
# Exit non-zero + halt envelope if required args are missing or the result
# value is not a valid enum member.
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/_common.sh"

# ---------------------------------------------------------------------------
# Parse args
# ---------------------------------------------------------------------------
version_value=""
result_value=""

for arg in "$@"; do
  case "$arg" in
    --version=*) version_value="${arg#--version=}" ;;
    --result=*)  result_value="${arg#--result=}"  ;;
    *) halt "unknown_arg" "smoke-reducer.sh: unrecognised argument: $arg" ;;
  esac
done

if [[ -z "$version_value" ]]; then
  halt "missing_required_field" "smoke-reducer.sh requires --version=v<X.Y.Z>"
fi
if [[ -z "$result_value" ]]; then
  halt "missing_required_field" "smoke-reducer.sh requires --result=passed|failed"
fi

case "$result_value" in
  passed|failed) ;;
  *) halt "invalid_result" "smoke-reducer.sh --result must be passed|failed, got: $result_value" ;;
esac

# ---------------------------------------------------------------------------
# Locate cycles.jsonl
# ---------------------------------------------------------------------------
ensure_auto_iter_dir
cycles_file="$(auto_iter_dir)/cycles.jsonl"

if [[ ! -f "$cycles_file" ]]; then
  # Nothing to stamp — emit zero updated.
  emit_json --updated=0 --version="$version_value" --result="$result_value"
  exit 0
fi

# ---------------------------------------------------------------------------
# Atomic rewrite: write to a tmp file, then mv over the original.
# jq reads each line, stamps smoke on matching rows, passes others through.
# ---------------------------------------------------------------------------
tmp_file="$(auto_iter_dir)/cycles.jsonl.tmp.$$"

updated=0
while IFS= read -r line || [[ -n "$line" ]]; do
  [[ -z "$line" ]] && continue
  row_release="$(printf '%s' "$line" | jq -r '.release // empty' 2>/dev/null || true)"
  if [[ "$row_release" == "$version_value" ]]; then
    line="$(printf '%s' "$line" | jq -c --arg s "$result_value" '. + {smoke: $s}')"
    updated=$((updated + 1))
  fi
  printf '%s\n' "$line"
done < "$cycles_file" > "$tmp_file"

# Swap atomically (mv is atomic within the same filesystem).
mv "$tmp_file" "$cycles_file"

emit_json --updated="$updated" --version="$version_value" --result="$result_value"
