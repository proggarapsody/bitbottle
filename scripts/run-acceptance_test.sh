#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
stub_bin="$tmp_dir/bin"
mkdir -p "$stub_bin"

cat >"$stub_bin/go" <<'GO_STUB'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$#" -lt 1 || "$1" != "test" || "$2" != "-count=1" || "$3" != "-json" ]]; then
  exit 90
fi
case "${ACCEPTANCE_RUNNER_CASE:-success}" in
  success)
    cat <<JSON
{"Action":"run","Test":"TestAcceptance"}
{"Action":"run","Test":"TestAcceptance/pr-edit-reviewer-safety"}
{"Action":"output","Test":"TestAcceptance/pr-edit-reviewer-safety","Output":"token=${BB_TOKEN:-test-secret}"}
{"Action":"pass","Test":"TestAcceptance/pr-edit-reviewer-safety"}
{"Action":"pass","Test":"TestAcceptance"}
JSON
    ;;
  skip)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"run","Test":"TestAcceptance/pr-edit-reviewer-safety"}
{"Action":"skip","Test":"TestAcceptance/pr-edit-reviewer-safety"}
{"Action":"pass","Test":"TestAcceptance"}
JSON
    ;;
  empty)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"pass","Test":"TestAcceptance"}
JSON
    ;;
  failed)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"fail","Test":"TestAcceptance","Output":"token=private-value"}
JSON
    exit 7
    ;;
  failed-zero)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"fail","Test":"TestAcceptance","Output":"token=private-value"}
JSON
    ;;
  missing)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"output","Test":"TestAcceptance","Output":"    acceptance_test.go:50: acceptance configuration: required values are empty: BB_E2E_REPO\n"}
{"Action":"fail","Test":"TestAcceptance"}
JSON
    exit 1
    ;;
  cloud-missing)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"output","Test":"TestAcceptance","Output":"    acceptance_test.go:50: acceptance configuration: required values are empty: BB_CLOUD_BOOTSTRAP_TOKEN\n"}
{"Action":"fail","Test":"TestAcceptance"}
JSON
    exit 1
    ;;
  unsafe-config)
    cat <<'JSON'
{"Action":"run","Test":"TestAcceptance"}
{"Action":"output","Test":"TestAcceptance","Output":"acceptance configuration: rejected value private-value"}
{"Action":"fail","Test":"TestAcceptance"}
JSON
    exit 1
    ;;
  *) exit 91 ;;
esac
GO_STUB
chmod +x "$stub_bin/go"

run_case() {
  local scenario="$1" expected_status="$2" output="$tmp_dir/output"
  local status=0
  ACCEPTANCE_RUNNER_CASE="$scenario" BITBOTTLE_E2E=1 BB_TOKEN=private-value PATH="$stub_bin:$PATH" \
    bash "$repo_root/scripts/run-acceptance.sh" >"$output" 2>&1 || status=$?
  if [[ "$status" != "$expected_status" ]]; then
    echo "runner case $scenario returned $status; expected $expected_status" >&2
    exit 1
  fi
  if grep -Eq 'private-value|test-secret' "$output"; then
    echo "runner case $scenario exposed credential-like output" >&2
    exit 1
  fi
}

opt_in_output="$tmp_dir/opt-in-output"
opt_in_status=0
env -u BITBOTTLE_E2E PATH="$stub_bin:$PATH" bash "$repo_root/scripts/run-acceptance.sh" >"$opt_in_output" 2>&1 || opt_in_status=$?
if [[ "$opt_in_status" != 2 ]] || ! grep -q 'explicit opt-in' "$opt_in_output"; then
  echo "runner did not require explicit opt-in" >&2
  exit 1
fi

run_case success 0
run_case skip 1
run_case empty 1
run_case failed 7
run_case failed-zero 1
grep -q 'live acceptance failed' "$tmp_dir/output"
run_case missing 1
grep -q 'acceptance configuration: required values are empty: BB_E2E_REPO' "$tmp_dir/output"
run_case cloud-missing 1
grep -q 'acceptance configuration: required values are empty: BB_CLOUD_BOOTSTRAP_TOKEN' "$tmp_dir/output"
run_case unsafe-config 1
grep -q 'acceptance configuration: configuration validation failed' "$tmp_dir/output"
echo "acceptance runner guard tests passed"
