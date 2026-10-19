#!/usr/bin/env bash
set -euo pipefail

if [[ "${BITBOTTLE_E2E:-}" != "1" ]]; then
  echo "live acceptance requires explicit opt-in: set BITBOTTLE_E2E=1" >&2
  exit 2
fi

result_file="$(mktemp)"
stderr_file="$(mktemp)"
trap 'rm -f "$result_file" "$stderr_file"' EXIT

set +e
go test -count=1 -json ./acceptance/... -timeout 20m -run '^TestAcceptance$' >"$result_file" 2>"$stderr_file"
go_status=$?
set -e

python3 - "$result_file" "$go_status" <<'PY'
import json
import sys

path, raw_status = sys.argv[1:]
go_status = int(raw_status)
root_pass = False
descendant_passes = 0
descendant_events = 0
skips = 0
failures = 0
parse_errors = 0
configuration_error = False
safe_config_messages = set()
required_names = {
    "BB_E2E_BACKEND", "BB_TOKEN", "BB_HOST", "BB_E2E_REPO",
    "BITBOTTLE_BIN", "BB_E2E_BRANCH", "BB_REVIEWER",
    "BB_CLOUD_BOOTSTRAP_TOKEN", "BB_CLOUD_EMAIL", "BB_CLOUD_WORKSPACE",
    "BB_CLOUD_PROJECT", "BB_ACCEPTANCE_REPORT_DIR", "BB_ACCEPTANCE_SHA",
}
safe_messages = {
    "BB_E2E_BACKEND must be cloud or server",
    "BB_E2E_BACKEND does not match the adapter selected by BB_HOST",
    "BB_E2E_REPO must be WORKSPACE/REPO or PROJECT/REPO",
    "BB_E2E_REPO is blocked by the production safety guard",
    "BITBOTTLE_BIN must point to an executable",
    "no .txt or .txtar scenarios found",
}

with open(path, encoding="utf-8") as result:
    for line in result:
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            parse_errors += 1
            continue

        test = event.get("Test", "")
        action = event.get("Action", "")
        in_suite = test == "TestAcceptance" or test.startswith("TestAcceptance/")
        if not in_suite:
            continue
        if test.startswith("TestAcceptance/"):
            descendant_events += 1
        if action == "pass":
            if test == "TestAcceptance":
                root_pass = True
            else:
                descendant_passes += 1
        elif action == "skip":
            skips += 1
        elif action == "fail":
            failures += 1
        output = event.get("Output", "")
        if "acceptance configuration:" in output:
            configuration_error = True
            message = output.split("acceptance configuration:", 1)[1].splitlines()[0].strip()
            if message in safe_messages:
                safe_config_messages.add(message)
            elif message.startswith("required values are empty: "):
                names = [name.strip() for name in message.removeprefix("required values are empty: ").split(",")]
                if names and all(name in required_names for name in names):
                    safe_config_messages.add("required values are empty: " + ", ".join(names))
                else:
                    safe_config_messages.add("configuration validation failed")
            elif message.startswith("cannot inspect scenario set:"):
                safe_config_messages.add("scenario set could not be inspected")
            else:
                safe_config_messages.add("configuration validation failed")

def report_config():
    for message in sorted(safe_config_messages):
        print(f"acceptance configuration: {message}", file=sys.stderr)

if parse_errors:
    if go_status != 0:
        print(f"live acceptance failed before reporting test results; go test exited with status {go_status}", file=sys.stderr)
        raise SystemExit(go_status)
    print("acceptance result was not valid go test JSON", file=sys.stderr)
    raise SystemExit(1)

if configuration_error and descendant_events == 0:
    report_config()
    print("live acceptance unavailable: required configuration is missing", file=sys.stderr)
    raise SystemExit(go_status or 1)

if go_status != 0 or failures:
    print(f"live acceptance failed: go test exit status {go_status}; failing test events {failures}", file=sys.stderr)
    raise SystemExit(go_status or 1)

if skips:
    print(f"live acceptance incomplete: {skips} selected test(s) skipped", file=sys.stderr)
    raise SystemExit(1)
if not root_pass:
    print("live acceptance unavailable: no passing TestAcceptance result", file=sys.stderr)
    raise SystemExit(1)
if descendant_passes == 0:
    print("live acceptance unavailable: no scenario subtest passed", file=sys.stderr)
    raise SystemExit(1)

print(f"live acceptance passed: {descendant_passes} scenario subtest(s)")
PY
