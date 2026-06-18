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

```sh
BITBOTTLE_E2E=1 \
BB_HOST=bb.example.com \
BB_TOKEN=<your-token> \
BB_E2E_REPO=TEST/acceptance-sandbox \
BB_E2E_BRANCH=acceptance/test-branch \
BB_REVIEWER=someuser \
BITBOTTLE_BIN=$(go build -o /tmp/bb ./cmd/bitbottle && echo /tmp/bb) \
go test ./acceptance/... -v -timeout 20m -run TestAcceptance
```

## Safety guard

The test runner enforces three `t.Skip` (never `t.Fatal`) guards so offline
runs and misconfigured environments stay green:

| Guard | Trigger |
|---|---|
| `BITBOTTLE_E2E != "1"` | All acceptance tests skipped |
| `BB_E2E_REPO` absent | All acceptance tests skipped |
| `BB_E2E_REPO` contains `"prod"` (case-insensitive) | All acceptance tests skipped with a clear message |

The third guard prevents accidentally running destructive operations against a
production repository even if someone mis-sets the env var.

## BB_ACCEPTANCE_SKIP_DEFER

Set `BB_ACCEPTANCE_SKIP_DEFER=1` to skip PR cleanup (the final `pr decline`
call) when you want to inspect the created PR manually. This env var is
documented in the txtar script comments but **not** enforced at the Go level;
scripts that want to honour it should check it explicitly.

## CI wiring

`nightly-e2e.yml` runs `go test ./acceptance/... -v -timeout 20m -run TestAcceptance`
as a second step after the existing testscript run. Required secrets:

| Secret | Used for |
|---|---|
| `BB_E2E_REPO` | `PROJECT/REPO` slug of the throwaway sandbox repo |
| `BB_E2E_BRANCH` | Source branch that already exists in the sandbox repo |
| `BB_E2E_REVIEWER` | Username to add as reviewer in reviewer-safety scripts |

The acceptance step inherits the same `BB_TOKEN` / `BB_HOST` as the testscript
step (already wired per matrix backend).

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
