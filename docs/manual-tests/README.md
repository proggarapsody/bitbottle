# bitbottle — manual smoke tests

A small set of end-to-end smoke scenarios for verifying `bitbottle` against
real Bitbucket instances before a release. Per-command coverage is the job
of automated tests (`go test ./... -race`); this directory exists for the
flows that automated tests cannot exercise — anything touching real auth,
real HTTP, or real git remotes.

These are **manual** tests by design. There is no runner. Run the scenarios
relevant to your change before cutting a release, or after touching a
backend-specific code path.

## Scenarios

### Cloud

- [`cloud/pr-happy-path.md`](cloud/pr-happy-path.md) — full PR lifecycle
  on Bitbucket Cloud (login → repo → branch → PR → squash-merge →
  cleanup). Exercises the largest surface in one flow plus GHP shortcuts
  (`pr checks`, `pr update-branch`, `pr status`, root `status` / `browse`),
  PR-COMMITS/PR-FILES/PR-PARTICIPANTS, and AUTOMERGE (`--auto`/`--auto-off`).
- [`cloud/auth-lifecycle.md`](cloud/auth-lifecycle.md) — full `auth`
  surface (login → status → token → refresh → migrate → doctor → logout).
  Exercises the OS keyring path and plaintext-token migration.
- [`cloud/issue-lifecycle.md`](cloud/issue-lifecycle.md) — full issue
  lifecycle (create → view → edit → assign → comment CRUD → close →
  reopen). Cloud only; also verifies Server returns `host.unsupported`.
- [`cloud/pipeline-lifecycle.md`](cloud/pipeline-lifecycle.md) — full
  pipeline surface (trigger → list → watch → view → steps → logs → schedule
  + cache CRUD). Cloud only; exercises long-polling and real CI state.
- [`cloud/deployment-lifecycle.md`](cloud/deployment-lifecycle.md) — DEP
  scope + all three VAR scopes (`--scope repository|workspace|deployment`,
  including `--secured` masking). Cloud only.
- [`cloud/workspace-admin.md`](cloud/workspace-admin.md) — `workspace
  list/member list`, `workspace hook list/create/delete`, `project list`,
  `search code`, and Cloud user-account `ssh-key` CRUD. Cloud only.

### Server / Data Center

- [`server/pr-happy-path.md`](server/pr-happy-path.md) — full PR lifecycle
  on Bitbucket Server / DC.
- [`server/code-insights.md`](server/code-insights.md) — `code-insights
  report` / `annotation` / `merge-check` (experimental). Server only.
- [`server/review-extras.md`](server/review-extras.md) — `pr task`,
  `pr suggestion apply` (+ `--preview`), and `pr comment react/unreact`
  / `--reactions` listing. Server only.

### Shared (both backends)

- [`shared/multi-host.md`](shared/multi-host.md) — Cloud + Server logged
  in simultaneously; `--hostname` routing.
- [`shared/source-primitives.md`](shared/source-primitives.md) — read-only
  `repo file get` and `repo tree`; binary-safety + `--ref` branch/tag/hash.
- [`shared/refs-and-diff.md`](shared/refs-and-diff.md) — branch create /
  list / checkout / delete, tag (lightweight + annotated) CRUD, `commit
  log/view/files/status/status report`, `commit comment` (incl. Server
  reactions), and `diff REF1..REF2` (full + `--stat`).
- [`shared/repo-lifecycle.md`](shared/repo-lifecycle.md) — `repo create`
  → clone → `set-default` → rename → visibility toggle → fork list
  (Cloud `fork create`) → `watcher list` → transfer → delete.
- [`shared/repo-settings.md`](shared/repo-settings.md) — webhook CRUD,
  deploy-key CRUD, default-reviewer CRUD, reviewer-group CRUD (Server),
  `branch-rule` CRUD (Cloud), `branch protect` CRUD (Server).
- [`shared/profile-config.md`](shared/profile-config.md) — `profile
  create/use/list/delete` (incl. keyring entries), local `config get/set/
  list`, `user view`, and `context` / `context --json`.
- [`shared/api-passthrough.md`](shared/api-passthrough.md) — raw `api`
  passthrough: `{workspace}`/`{repo_slug}` (Cloud) and `{project}`/`{slug}`
  (Server) expansion, `-F`/`-f` typed fields, `--input -` stdin body,
  `-H` headers, `--paginate`.
- [`shared/extensions.md`](shared/extensions.md) — `extension install`
  (remote + `--local`), `list`, `exec`, `upgrade [--all|--force]`,
  `remove`. SHA tamper detection and env sanitisation.

If a scope changes a flow not covered by these smokes (e.g. a new
top-level command group, a new backend-specific failure mode), add a fresh
scenario file rather than expanding one of these. Keep each scenario
coherent (login → action → cleanup), not exhaustive — per-command coverage
belongs in automated tests, not here.

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
| `BB_TEST_CLOUD_WORKSPACE`   | `myws`                           | Workspace slug for `workspace`/`project` scenarios |
| `BB_TEST_CLOUD_WORKSPACE_ALT` | `myws-alt`                     | Second workspace for `repo transfer` (optional) |
| `BB_TEST_SERVER_PROJECT`    | `MYPROJ`                         | Project key for `repo create` on Server |
| `BB_TEST_SERVER_PROJECT_ALT`| `MYPROJ_ALT`                     | Second project for Server `repo transfer` (optional) |
| `BB_TEST_CLOUD_REVIEWER`    | `other-username`                 | Second user for review-related scenarios |
| `BB_TEST_SERVER_REVIEWER`   | `other.user`                     | Second user (Server slug) for reviews   |
| `BB_TEST_CLOUD_TOKEN_ALT`   | (PAT)                            | Second Cloud token for `profile` smoke (optional; may equal `BB_TEST_CLOUD_TOKEN`) |

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
