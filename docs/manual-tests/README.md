# bitbottle — manual smoke tests

A small set of end-to-end smoke scenarios for verifying `bitbottle` against
real Bitbucket instances before a release. Per-command coverage is the job
of automated tests (`go test ./... -race`); this directory exists for the
flows that automated tests cannot exercise — anything touching real auth,
real HTTP, or real git remotes.

These are **manual** tests by design. There is no runner. Run all three
before cutting a release, or after touching a backend-specific code path.

## Scenarios

- [`cloud/pr-happy-path.md`](cloud/pr-happy-path.md) — full PR lifecycle
  against Bitbucket Cloud (login → repo → branch → PR → squash-merge →
  cleanup). Exercises the largest surface in one flow.
- [`cloud/issue-lifecycle.md`](cloud/issue-lifecycle.md) — full issue lifecycle
  against Bitbucket Cloud (create → view → edit → assign → comment CRUD →
  close → reopen). Cloud only; also verifies Server returns host.unsupported.
- [`server/pr-happy-path.md`](server/pr-happy-path.md) — same flow against
  Bitbucket Server / Data Center.
- [`shared/multi-host.md`](shared/multi-host.md) — both backends configured
  simultaneously, verifies host routing and `--hostname` selection.
- [`shared/source-primitives.md`](shared/source-primitives.md) — read-only
  smoke for `repo file get` and `repo tree` (both backends; verifies
  binary-safety, type normalisation, and `--ref` accepts branches /
  tags / hashes).

If a scope changes a flow not covered by these three smokes (e.g. a new
top-level command group, a new backend-specific failure mode), add a fresh
scenario file rather than expanding one of these. Keep each scenario
coherent (login → action → cleanup), not exhaustive.

## Prerequisites

| Variable                    | Example                          | Notes                                  |
|-----------------------------|----------------------------------|----------------------------------------|
| `BB_TEST_CLOUD_HOST`        | `bitbucket.org`                  | Cloud host                             |
| `BB_TEST_CLOUD_REPO`        | `myws/bitbottle-qa`              | `<workspace>/<repo>` — disposable      |
| `BB_TEST_CLOUD_TOKEN`       | (PAT)                            | Scopes: account/repo/pullrequest r+w   |
| `BB_TEST_SERVER_HOST`       | `bitbucket.example.com`          | Server / DC host                       |
| `BB_TEST_SERVER_REPO`       | `MYPROJ/bitbottle-qa`            | `<projectKey>/<repo>` — disposable     |
| `BB_TEST_SERVER_TOKEN`      | (PAT)                            | Scopes: PROJECT_READ + PROJECT_WRITE   |
| `BB_TEST_SERVER_SKIP_TLS`   | `true` or `false`                | Self-signed cert?                      |

Provision the two scratch repos by hand once. They will accumulate state
over time — each scenario has a Cleanup section, but expect drift.

Build a fresh CLI before testing:

```bash
make build
export PATH="$PWD/dist:$PATH"
bitbottle --version
```

## Scenario file shape

Each scenario follows:

1. **Title** — `# Scenario: …`
2. **Backend** — Cloud / Server-DC / both
3. **Prerequisites** — env vars, repo state, local checkout state
4. **Setup** — copy-pasteable shell to reach a known state
5. **Steps** — numbered; each has command, expected stdout shape, expected
   stderr, expected exit code, and a "Verify in UI" note where applicable
6. **Cleanup** — copy-pasteable shell to remove created resources

Volatile values (hashes, timestamps, PR numbers, UUIDs) are masked as `…` or
`<placeholder>`. Stable structure (column headers, exit codes, error wording)
is exact.
