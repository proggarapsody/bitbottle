# bitbottle — manual test cases

Curated end-to-end scenarios for verifying `bitbottle` against real Bitbucket
instances. Each scenario is a coherent user flow (log in → do something → clean
up), not a per-command checklist.

These are **manual** tests by design. Run them before a release, or after
touching a backend-specific code path. They are documentation; there is no
runner.

## Layout

```
docs/manual-tests/
├── README.md            # this file
├── shared/              # flows identical on both backends — run once per backend
├── cloud/               # Bitbucket Cloud (bitbucket.org) only
└── server/              # Bitbucket Server / Data Center only
```

A scenario is duplicated across `cloud/` and `server/` only when the **failure
modes differ** under the hood (e.g. PR comments, commit build status, API path
shapes). Otherwise it lives in `shared/` and the prereqs say "run once per
backend".

## Prerequisites

Set these before running anything.

| Variable                    | Example                          | Notes                                  |
|-----------------------------|----------------------------------|----------------------------------------|
| `BB_TEST_CLOUD_HOST`        | `bitbucket.org`                  | Cloud host                             |
| `BB_TEST_CLOUD_REPO`        | `myws/bitbottle-qa`              | `<workspace>/<repo>` — disposable      |
| `BB_TEST_CLOUD_TOKEN`       | (PAT)                            | Scopes: account/repo/pullrequest r+w   |
| `BB_TEST_SERVER_HOST`       | `bitbucket.example.com`          | Server / DC host                       |
| `BB_TEST_SERVER_REPO`       | `MYPROJ/bitbottle-qa`            | `<projectKey>/<repo>` — disposable     |
| `BB_TEST_SERVER_TOKEN`      | (PAT)                            | Scopes: PROJECT_READ + PROJECT_WRITE   |
| `BB_TEST_SERVER_SKIP_TLS`   | `true` or `false`                | Self-signed cert?                      |

Provision the two scratch repos by hand once. They will accumulate branches,
tags, and PRs over time — each scenario has a Cleanup section, but expect drift.

Build a fresh CLI before testing:

```bash
make build
export PATH="$PWD/dist:$PATH"
bitbottle --version
```

## Scenario template

Every scenario file follows this shape:

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

## Index

### Shared (run once per backend)

- [`shared/output-modes.md`](shared/output-modes.md) — TTY vs pipe, `--json`, `--jq`, `--web`
- [`shared/config-and-alias.md`](shared/config-and-alias.md) — `config`, `alias`
- [`shared/completion.md`](shared/completion.md) — `completion bash|zsh|fish|powershell`
- [`shared/mcp-serve-smoke.md`](shared/mcp-serve-smoke.md) — `mcp serve` handshake
- [`shared/multi-host.md`](shared/multi-host.md) — Cloud + Server/DC both logged in

### Cloud only

- [`cloud/auth-lifecycle.md`](cloud/auth-lifecycle.md) — login → status → refresh → token → logout
- [`cloud/repo-lifecycle.md`](cloud/repo-lifecycle.md) — create → view → clone → delete
- [`cloud/branch-lifecycle.md`](cloud/branch-lifecycle.md) — create → list → checkout → delete
- [`cloud/tag-lifecycle.md`](cloud/tag-lifecycle.md) — lightweight + annotated
- [`cloud/commit-inspection.md`](cloud/commit-inspection.md) — log → view → status
- [`cloud/pr-happy-path.md`](cloud/pr-happy-path.md) — full PR lifecycle, squash-merge
- [`cloud/pr-decline.md`](cloud/pr-decline.md)
- [`cloud/pr-checkout.md`](cloud/pr-checkout.md) — check out someone else's PR
- [`cloud/pr-comments.md`](cloud/pr-comments.md) — `pr comment add` / `list`
- [`cloud/pr-request-changes.md`](cloud/pr-request-changes.md) — `pr request-changes` (Cloud only)
- [`cloud/pipelines.md`](cloud/pipelines.md) — run → list → view
- [`cloud/api-passthrough.md`](cloud/api-passthrough.md) — `api`, `--paginate`, `--jq`, `-F`

### Server / DC only

- [`server/auth-lifecycle.md`](server/auth-lifecycle.md) — including `--skip-tls-verify`
- [`server/repo-lifecycle.md`](server/repo-lifecycle.md)
- [`server/branch-lifecycle.md`](server/branch-lifecycle.md)
- [`server/tag-lifecycle.md`](server/tag-lifecycle.md)
- [`server/commit-inspection.md`](server/commit-inspection.md) — uses `build-status/1.0`
- [`server/pr-happy-path.md`](server/pr-happy-path.md)
- [`server/pr-decline.md`](server/pr-decline.md)
- [`server/pr-checkout.md`](server/pr-checkout.md)
- [`server/pr-comments.md`](server/pr-comments.md) — uses `activities` feed
- [`server/pipelines-rejected.md`](server/pipelines-rejected.md) — negative test
- [`server/api-passthrough.md`](server/api-passthrough.md) — `rest/api/1.0/...` paths
