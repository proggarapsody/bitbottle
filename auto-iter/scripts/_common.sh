#!/usr/bin/env bash
# Shared helpers for auto-iter/scripts/*.sh.
# Source from any auto-iter script: . "$(dirname "$0")/_common.sh"
#
# All scripts in this directory follow one convention:
#   - Exit 0 on success, write a single JSON object to stdout.
#   - Exit non-zero on failure, write {"halt":"<reason>","details":"..."} to stdout.
#   - Never write progress/log lines to stdout (use stderr for that).
# Canonical contract: auto-iter/scripts.md.

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
# else is string-quoted.
#
# Variants for explicit typing:
#   --str key=value   force STRING typing (use for enum fields like tier="2")
#   --raw key=expr    inject a pre-formatted JSON value verbatim
#   --arr key=value   accumulate repeated occurrences into a JSON array.
#                     Each --arr key=value call appends value to the array for
#                     that key. Integers and booleans are typed; other values
#                     are string-quoted. Repeated --arr key=val calls build the
#                     array in order. Use this for scopes/prs fields.
#                     Plural output key: scope->scopes, pr->prs, else key+"s".
#
# Example:
#   emit_json --cycle=42 --step=step2_tdd --str tier=2 --raw findings='["a","b"]'
#   emit_json --arr scope=FOO --arr scope=BAR   # -> {"scopes":["FOO","BAR"]}
emit_json() {
  local jq_args=() filter='.'
  # Accumulate --arr entries as "out_key\tvalue" lines in a single variable.
  # bash 3.2-compatible (no associative arrays).
  local arr_entries=""

  while [[ $# -gt 0 ]]; do
    local arg="$1"
    if [[ "$arg" == "--raw" ]]; then
      shift
      local key="${1%%=*}" expr="${1#*=}"
      jq_args+=(--argjson "$key" "$expr")
      filter="$filter | .${key}=\$${key}"
    elif [[ "$arg" == "--str" ]]; then
      shift
      local key="${1%%=*}" val="${1#*=}"
      jq_args+=(--arg "$key" "$val")
      filter="$filter | .${key}=\$${key}"
    elif [[ "$arg" == "--arr" ]]; then
      shift
      local key="${1%%=*}" val="${1#*=}"
      # Derive plural output key
      local out_key
      case "$key" in
        scope) out_key="scopes" ;;
        pr)    out_key="prs" ;;
        *)     out_key="${key}s" ;;
      esac
      # Encode value as a jq-ready literal
      local jq_val
      if [[ "$val" =~ ^(true|false|null)$ ]] || [[ "$val" =~ ^-?[0-9]+$ ]]; then
        jq_val="$val"
      else
        jq_val="$(printf '%s' "$val" | jq -Rs '.')"
      fi
      # Append "out_key<TAB>jq_val" to accumulator
      if [[ -n "$arr_entries" ]]; then
        arr_entries="${arr_entries}"$'\n'"${out_key}"$'\t'"${jq_val}"
      else
        arr_entries="${out_key}"$'\t'"${jq_val}"
      fi
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

  # Process accumulated --arr entries: group by out_key, build JSON array per key.
  if [[ -n "$arr_entries" ]]; then
    # Extract distinct out_keys (preserve order of first occurrence)
    local seen_keys=""
    while IFS=$'\t' read -r out_key jq_val; do
      if ! printf '%s\n' "$seen_keys" | grep -qxF "$out_key"; then
        seen_keys="${seen_keys:+$seen_keys$'\n'}${out_key}"
      fi
    done <<<"$arr_entries"

    while IFS= read -r out_key; do
      [[ -z "$out_key" ]] && continue
      # Collect all values for this key (tab-separated field 2)
      local arr_json
      arr_json="$(printf '%s\n' "$arr_entries" | awk -F'\t' -v k="$out_key" '$1==k{print $2}' | jq -sc '.')"
      jq_args+=(--argjson "$out_key" "$arr_json")
      filter="$filter | .${out_key}=\$${out_key}"
    done <<<"$seen_keys"
  fi

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
