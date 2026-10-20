# Acceptance test suite

> **Tier 6 — live wire.** The `acceptance/` package hits a real Bitbucket
> Server or Cloud sandbox so it can detect wire-level drift that hermetic
> fakes cannot. It complements `test/script/` (tier 3); it does not replace it.

## What it is

`acceptance/` is a [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
suite that runs against a real Bitbucket backend. Unlike `test/script/`, it
does **not** scrub `BB_*` / `BITBOTTLE_*` env vars — the whole point is to
pass real credentials through so the binary exercises real API paths.

The first seed script (`testdata/pr/pr-edit-reviewer-safety.txtar`) is wired
to reproduce the BQ-1 / BQ-2 class of bugs: `pr edit --title` wiping reviewers
and `pr request-review` failing with a 400 because `version` was omitted.

## Running locally

Build the binary first with `go build -o /tmp/bb ./cmd/bitbottle`. Server/DC
uses the existing sandbox configuration below. Cloud creates a unique
temporary repository and does not require a pre-existing repository, branch,
or reviewer:

```sh
BITBOTTLE_E2E=1 \
BB_E2E_BACKEND=cloud \
BB_TOKEN=<actor-token> \
BB_CLOUD_BOOTSTRAP_TOKEN=<bootstrap-token> \
BB_CLOUD_EMAIL=<actor-email> \
BB_CLOUD_WORKSPACE=bitbottle-qa-lab \
BB_CLOUD_PROJECT=BBQA \
BB_ACCEPTANCE_SHA="$(git rev-parse HEAD)" \
BB_ACCEPTANCE_REPORT_DIR=/tmp/bitbottle-acceptance \
BITBOTTLE_BIN=/tmp/bb \
bash scripts/run-acceptance.sh
```

The bootstrap token provisions and removes the temporary repository. The
separate `BB_TOKEN` actor token exercises the public CLI. Keep both scoped to
the dedicated QA workspace. The distinct-token setup does not prove that the
tokens belong to different Atlassian accounts; the run report records that
identity separation as unverified until it is independently confirmed.

## Configuration and safety

Ordinary test runs skip the acceptance test unless `BITBOTTLE_E2E=1` is set.
Once opted in, incomplete or unsafe configuration fails before any scenario
starts. Required values are:

| Variable | Requirement |
|---|---|
| `BB_E2E_BACKEND` | Exactly `cloud` or `server` |
| `BB_TOKEN` | Nonempty actor token with the rights required by the scenarios |
| `BITBOTTLE_BIN` | Path or command name resolving to an executable Bitbottle binary |
| Cloud: `BB_CLOUD_BOOTSTRAP_TOKEN` | Separate bootstrap token used to create and remove only the temporary Cloud repository |
| Cloud: `BB_CLOUD_EMAIL` | Email for the configured actor token |
| Cloud: `BB_CLOUD_WORKSPACE` | Must be the dedicated `bitbottle-qa-lab` workspace |
| Cloud: `BB_CLOUD_PROJECT` | Must be `BBQA` |
| Cloud: `BB_ACCEPTANCE_REPORT_DIR` | Writable, per-run directory for `manifest.json` and `report.json` |
| Server/DC: `BB_HOST` | Server/DC hostname matching the adapter selected by the CLI |
| Server/DC: `BB_E2E_REPO` | Nonempty `PROJECT/REPO`; values containing `prod` case-insensitively are rejected |
| Server/DC: `BB_E2E_BRANCH` | Existing source branch in the sandbox repository |
| Server/DC: `BB_REVIEWER` | Existing reviewer username |

The suite also fails if it discovers no `.txt` or `.txtar` scenario. The CI
runner, `scripts/run-acceptance.sh`, checks `go test` JSON output and fails if
the opted-in run skips tests or completes with zero scenario subtests. Use that
runner for mandatory live acceptance; a direct `go test` invocation does not
enforce the runner-level skip and scenario-count check.

## BB_ACCEPTANCE_SKIP_DEFER

Set `BB_ACCEPTANCE_SKIP_DEFER=1` to skip PR cleanup (the final `pr decline`
call) when you want to inspect the created PR manually. This env var is
documented in the txtar script comments but **not** enforced at the Go level;
scripts that want to honour it should check it explicitly.

## CI wiring

`.github/workflows/nightly-e2e.yml` builds the binary from the workflow SHA,
runs the hermetic testscript suite, then runs the selected live backend. The
Cloud job reads secrets from the `cloud-acceptance` GitHub Environment and
always invokes `go run ./cmd/qa-cloud-cleanup` with the run's persisted
manifest, even after a product failure. It uploads only the report and
manifest; neither file may contain credentials. Configure the environment
secrets `BB_CLOUD_TOKEN`, `BB_CLOUD_BOOTSTRAP_TOKEN`, and `BB_CLOUD_EMAIL`,
plus the repository variable `CLOUD_ACCEPTANCE_ENABLED=true` to enable the
scheduled Cloud lane. The Cloud workspace and project are fixed to
`bitbottle-qa-lab` and `BBQA`.

The workflow dispatch input selects `cloud`, `server`, or `both`. An explicit
Server/DC selection with missing credentials fails configuration validation.
The scheduled run executes Cloud only when `CLOUD_ACCEPTANCE_ENABLED=true`;
otherwise its status summary says the lane is unavailable. Live jobs run only
from `main`. Dispatches from other refs fail in a secret-free trust check.
The Cloud environment must also restrict deployment branches to `main`.

Server/DC credentials remain repository secrets:

| Secret | Purpose |
|---|---|
| `BB_SERVER_TOKEN` | Actor token |
| `BB_SERVER_HOST` | Server/DC hostname |
| `BB_SERVER_E2E_REPO` | `PROJECT/REPO` sandbox |
| `BB_SERVER_E2E_BRANCH` | Existing source branch |
| `BB_SERVER_E2E_REVIEWER` | Reviewer username |

Cloud evidence does not establish Server/DC compatibility. The workflow
reports `DC не проверен live` until a licensed, configured Data Center live
lane is separately implemented and run.

## Cloud auto-merge diagnostic

`.github/workflows/cloud-auto-merge-probe.yml` is a manual, main-branch-only
diagnostic for Bitbucket Cloud auto-merge behavior. Dispatch it from `main`
after the workflow is merged. It uses the same `cloud-acceptance` Environment,
fixed QA workspace and project, separate bootstrap and actor credentials,
and manifest-owned cleanup as the Cloud acceptance lane. The workflow builds
the CLI and probe from the selected commit and uploads sanitized probe evidence
after cleanup. Inspect `probe.json` and the workflow summary for the observed
provider result and cleanup state.

The probe tries `https://api.bitbucket.org/internal/repositories/.../post-build-merge`
with GET and POST on the same manifest-owned, conflicting PR. It tests actor
and bootstrap tokens with Bearer and email/token Basic authentication, then
reads the PR after each POST to verify it remains open. The internal origin is
fixed in the diagnostic; redirects are not followed and cookies are not sent.
The report stores only HTTP status, bounded sanitized `error.message`, and the
boolean `can_create_post_build_merge` field when present. This internal route
may have changed or disappeared, so its responses are evidence about this run,
not a supported API contract or proof of auto-merge readiness.

This probe is an observation, not a release pass or a replacement for the
frozen critical profile. Keep the release gate and nightly acceptance results
separate. A failed, skipped, or cleanup-incomplete probe cannot establish a
provider pass.

## Adding new scripts

1. Create `acceptance/testdata/<domain>/<name>.txtar`.
2. Begin with a comment explaining what BQ-* rows or bug class the script
   targets, and what env vars it needs.
3. Use `exec $BITBOTTLE_BIN ...` for all binary invocations.
4. Use `stdout2env KEY` to capture the last stdout line into a script env var
   (e.g. to capture a PR number from `pr create` output).
5. Always clean up created resources at the end of the script.

## stdout2env command

`stdout2env KEY` is a custom testscript command registered by the acceptance
test runner. It reads the stdout from the most recent `exec` call, trims
trailing whitespace, takes the last non-empty line, and exports it as a
testscript env var with the given key name.

Example:

```
exec $BITBOTTLE_BIN pr create -R $BB_E2E_REPO ...
stdout2env PR_NUMBER
exec $BITBOTTLE_BIN pr view $PR_NUMBER -R $BB_E2E_REPO ...
```
