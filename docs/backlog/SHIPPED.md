# bitbottle — Shipped Scopes

> **Append-only record of shipped backlog scopes.** When a scope's `feat:` commit lands on `main`, its row is **moved** from [`BACKLOG.md`](BACKLOG.md) into this file (not flipped in place). See [`docs/workflows/iteration-cycle/quickref.md`](../workflows/iteration-cycle/quickref.md) §"Definition of Done" for the convention and [`docs/workflows/iteration-cycle/README.md`](../workflows/iteration-cycle/README.md) §4 for the iteration-cycle step.

**Layout:** newest scopes first. Each entry carries the ship date, the feat commit, and (if assigned by then) the release-please version. Detail subsections are preserved verbatim from BACKLOG.md so search history stays intact.

**Why a separate file:**
- BACKLOG.md is for **upcoming work** — keeping shipped scopes inline made it grow past 4000 lines and buried the queue under archaeology.
- This file is **not** CHANGELOG.md (release-please owns that, formatted per Conventional Commits). SHIPPED.md tracks **scope-level intent**: what was queued, what shipped, why. CHANGELOG.md tracks **per-PR semver-bumping notes** for end users.
- Append-only means git history is the audit log; no edits-in-place.

---

## 2026-06-02 — REPO-HOOK-SCRIPTS — Server/DC Repo Hook Settings

`repo hook list/view/enable/disable/settings get/set` — manage plugin hook
scripts on Bitbucket Server/DC repositories. Cloud returns `host.unsupported`.
Closes #626.

---

## 2026-06-02 — BACKLOG-MIGRATION — swept shipped scope details

Batch migration of shipped scope detail sections from BACKLOG.md to SHIPPED.md.
Scopes moved: TESTSCRIPT, REPO-EDIT, PIPE-STOP, SNIPPETS, BRANCH-MODEL, PIPE-ARTIFACTS, GROUP-MGMT, EXT-SCAFFOLD, NIGHTLY-E2E, SOURCE-WRITE, ADMIN-MAIL, ADMIN-BANNER, PIPE-RUNNERS, WORKSPACE-PIPELINE-VARS, ISSUE-ACTIVITY, CLOUD-PROJECT-REVIEWERS, PR-MERGE-PREVIEW.
Closes #624.

---

### TESTSCRIPT — Whole-binary script tests

**Status:** ✅ — see issue [#399](https://github.com/proggarapsody/bitbottle/issues/399) for the full PRD.

**Why P0:** Cycles 90–99 shipped four behaviour bugs ([#387](https://github.com/proggarapsody/bitbottle/pull/387), [#394](https://github.com/proggarapsody/bitbottle/pull/394), [#390](https://github.com/proggarapsody/bitbottle/pull/390), [`08b3d9a`](https://github.com/proggarapsody/bitbottle/commit/08b3d9a)) that every single existing unit/adapter test was green for. The shared trait: the failure only manifests when real argv, real env, real cobra dispatch, real factory wiring, and real errfmt rendering are composed end-to-end. No tier below tier 3 catches this class.

**Shape (summarised; full PRD in #399):**

1. New package `test/script/` exposes `Run(t *testing.T, dir string)` — single deep-module entry. Command-area test files become one-liners.
2. `cmd/bitbottle/main.go` refactored so the real entry point is `Main() int`; `testscript.RunMain` dispatches in-process for parallel-safe per-script execution.
3. Custom `bb-fake server|cloud` testscript command boots a fixture-backed `httptest.Server` (reuses `test/testhelpers.BitbucketCloudServer` / Server builders) and exports `BITBOTTLE_TEST_BASE_URL` + seeded credentials.
4. Hermetic per script: temp `HOME`, scrubbed env (`BITBOTTLE_*`, `HTTPS_PROXY`, `GIT_*`, `XDG_*`), ephemeral ports rewritten in golden output.
5. Seed corpus (8–10 `.txtar`): `auth_status.txtar`, `repo_list_cloud.txtar`, `repo_list_server.txtar`, `pr_list_cloud.txtar`, `pr_list_server.txtar`, `pr_view.txtar`, `errfmt_catalogue.txtar`, `bitbottle_host_defaulting.txtar`, `capability_gap.txtar`, `content_type_policy.txtar`.
6. `Makefile` gets `make test-script`. Pre-merge gate runs `go test ./test/script -race`.
7. Nightly GHA `BITBOTTLE_E2E=1` runs the same corpus against real Bitbucket Server (Alfa) + Cloud sandbox; failure opens a GitHub issue with the failing script name + diff.

**Definition of Done:**

- [ ] `test/script/script.go` — `Run(t, dir)` harness.
- [ ] `cmd/bitbottle/main.go` — `Main() int` entry point.
- [ ] `test/script/bbfake/` — `bb-fake` testscript command.
- [ ] `test/script/testdata/*.txtar` — seed corpus (10 scripts).
- [ ] `Makefile` — `test-script` target.
- [ ] `.github/workflows/ci.yml` — script tier wired into CI.
- [ ] `.github/workflows/e2e-nightly.yml` — opt-in real-host job.
- [ ] `docs/testing.md` — testing-strategy doc (one source of truth, mirrored from BACKLOG.md § Testing Strategy).
- [ ] `go test ./test/script -race` green; suite under 30 s wall.

---

### REPO-EDIT — Repository metadata edit

**Status:** ✅

Update mutable repository fields — description, website, language, fork policy, issues/wiki toggles — without performing a rename or visibility toggle. Complements the existing `repo rename`, `repo visibility`, and `repo set-default-branch` commands to cover full parity with `gh repo edit`.

**Interface:**
```go
type RepoEditor interface {
    EditRepo(ns, slug string, in EditRepoInput) (Repository, error)
}
type EditRepoInput struct {
    Description *string
    Website     *string
    Language    *string
    ForkPolicy  *string
    HasIssues   *bool
    HasWiki     *bool
}
```

**Commands:** `repo edit [PROJECT/REPO] --description STR --website URL --language LANG --fork-policy POLICY --enable-issues --disable-issues --enable-wiki --disable-wiki`

**Backends:** Cloud (`PUT /repositories/{ws}/{slug}` — all fields); Server (`PUT /rest/api/1.0/projects/{ns}/repos/{slug}` — `description` only; other fields return typed `host.unsupported` with `field=website` etc.).

**MCP tools:** `edit_repo`

**Definition of Done:**
- [ ] `api/backend/client_repo_edit.go` — `RepoEditor` interface + `EditRepoInput`
- [ ] `api/cloud/repo_edit.go` + `api/cloud/repo_edit_test.go`
- [ ] `api/server/repo_edit.go` + `api/server/repo_edit_test.go` (description only; other fields → `host.unsupported`)
- [ ] `pkg/cmd/repo/edit/edit.go` + `pkg/cmd/repo/edit/edit_test.go`
- [ ] `pkg/cmd/mcp/tools_repo.go` entry + `pkg/cmd/mcp/handlers.go` method + `pkg/cmd/mcp/handlers_test.go` test
- [ ] `skills/SKILL.md` + `skills/references/repo.md` updated
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### PIPE-STOP — Pipeline stop

**Status:** ✅

Stop a running or pending Cloud pipeline. Useful for agents that trigger pipelines via `pipeline trigger`, watch them, and need to abort on undesired behaviour. `--confirm` required on non-TTY to prevent accidental stops.

**Interface method** (extension of existing `PipelineClient`):
```go
StopPipeline(ws, slug, pipelineUUID string) error
```

**Commands:** `pipeline stop PIPELINE_UUID [PROJECT/REPO] [--confirm]`

**Backends:** Cloud only (`POST /repositories/{ws}/{slug}/pipelines/{uuid}/stopPipeline`, empty body, 204). Server returns typed `host.unsupported`.

**MCP tools:** `stop_pipeline`

**Definition of Done:**
- [x] `StopPipeline` added to `PipelineClient` interface in `api/backend/client_pipeline.go`
- [x] `api/cloud/pipelines.go` — `StopPipeline` impl + test
- [x] `api/server/pipelines.go` — `StopPipeline` → `host.unsupported` + test
- [x] `pkg/cmd/pipeline/stop/stop.go` + test
- [x] MCP triplet
- [x] `skills/SKILL.md` + `skills/references/pipeline.md` updated
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### SNIPPETS — Cloud Snippets (gist parity)

**Status:** ✅

Bitbucket Cloud's analogue of GitHub Gists. Snippets let users share short code or text files with optional privacy. Server has no primitive.

**Interface:**
```go
type SnippetClient interface {
    ListSnippets(workspace string, limit int) ([]Snippet, error)
    GetSnippet(workspace, id string) (Snippet, error)
    CreateSnippet(workspace string, in CreateSnippetInput) (Snippet, error)
    DeleteSnippet(workspace, id string) error
}
type Snippet struct {
    ID, Title, Owner string
    IsPrivate        bool
    CreatedOn        time.Time
    Files            []SnippetFile
    WebURL           string
}
```

**Commands:** `snippet list [--workspace W] [--json]`, `snippet view ID`, `snippet create --title T --file PATH[,PATH...] [--private]`, `snippet delete ID [--confirm]`

**Backends:** Cloud (`GET/POST/DELETE /snippets/{workspace}`, `GET /snippets/{workspace}/{id}`). Server → `host.unsupported`.

**MCP tools:** `list_snippets`, `view_snippet`, `create_snippet`, `delete_snippet`

**Definition of Done:**
- [x] `api/backend/client_snippet.go` — `SnippetClient`, `Snippet`, `SnippetFile`, `CreateSnippetInput`
- [x] `api/cloud/snippets.go` + `api/cloud/snippets_test.go`
- [x] `pkg/cmd/snippet/` (list, view, create, delete) + tests
- [x] MCP quadruplet
- [x] `skills/SKILL.md` + `skills/references/snippet.md`
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### BRANCH-MODEL — Cloud branching model

**Status:** ✅

Read and update a repository's branching model — the development/production branch configuration and branch-type naming prefixes used by Bitbucket's in-UI "Create branch" wizard and by `pipelines.yml` `branches:` triggers. Distinct from BRANCH-RULE (which controls restrictions/enforcement policies).

**Interface:**
```go
type BranchModelClient interface {
    GetBranchModel(ws, slug string) (BranchModel, error)
    GetBranchModelSettings(ws, slug string) (BranchModelSettings, error)
    UpdateBranchModelSettings(ws, slug string, in BranchModelSettingsInput) (BranchModelSettings, error)
}
```

**Commands:** `branch-model get [PROJECT/REPO] [--json]`, `branch-model set [PROJECT/REPO] --dev-branch NAME --prod-branch NAME [--prod-enabled] [--branch-type-prefix feature=feat/,hotfix=hf/]`

**Backends:** Cloud (`GET /repositories/{ws}/{slug}/branching-model`, `GET/PUT /repositories/{ws}/{slug}/branching-model/settings`). Server → `host.unsupported`.

**MCP tools:** `get_branch_model`, `set_branch_model`

**Definition of Done:**
- [ ] `api/backend/client_branch_model.go`
- [ ] `api/cloud/branch_model.go` + `api/cloud/branch_model_test.go`
- [ ] `pkg/cmd/branch-model/` (get, set) + tests
- [ ] MCP pair
- [ ] `skills/SKILL.md` + `skills/references/branch.md`
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### PIPE-ARTIFACTS — Pipeline artifacts

**Status:** ✅

List and download per-step build artifacts declared via `artifacts:` in `bitbucket-pipelines.yml`. Today agents that trigger pipelines via `pipeline trigger` cannot retrieve their outputs without raw `api` calls — this closes that gap.

**Interface:**
```go
type PipelineArtifactClient interface {
    ListPipelineArtifacts(ws, slug, pipelineUUID, stepUUID string, limit int) ([]PipelineArtifact, error)
    DownloadPipelineArtifact(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error
}
type PipelineArtifact struct {
    Name      string
    SizeBytes int64
    URL       string
}
```

**Commands:** `pipeline artifact list PIPELINE_UUID --step STEP_UUID [PROJECT/REPO] [--json]`, `pipeline artifact download PIPELINE_UUID --step STEP_UUID --name FILE [--out PATH]` (defaults to filename in cwd; `--out -` for stdout)

**Backends:** Cloud (`GET /repositories/{ws}/{slug}/pipelines/{pipeline_uuid}/steps/{step_uuid}/artifacts`, download via `links.self.href`). Server → `host.unsupported`.

**MCP tools:** `list_pipeline_artifacts`, `download_pipeline_artifact` (base64 envelope for small files)

**Definition of Done:**
- [ ] `api/backend/client_pipeline_artifacts.go`
- [ ] `api/cloud/pipeline_artifacts.go` + `api/cloud/pipeline_artifacts_test.go`
- [ ] `pkg/cmd/pipeline/artifact/` (list, download) + tests
- [ ] MCP pair
- [ ] `skills/SKILL.md` + `skills/references/pipeline.md`
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### GROUP-MGMT — Server/DC Group Management

**Status:** ✅

Bitbucket Server/DC has a first-class internal group primitive used everywhere — `perms project grant --group ENG`, `pr reviewer-group add --users …`, branch-restriction grants — but bitbottle has no surface for managing the groups themselves. Today admins must drop into the Bitbucket web UI (or raw `bitbottle api`) to create a new group or add a user to it. This scope closes the gap and lands the standard `gh org member`-shaped CRUD on the existing peer concept. Cloud's authorization model is workspace-permission-shaped rather than group-shaped, so Cloud returns typed `host.unsupported` from every method on this interface (consistent with PERMS, BRANCH-PROTECT, TASK).

**Interface:**
```go
// api/backend/client_group.go
type GroupClient interface {
    ListGroups(filter string, limit int) ([]Group, error)
    CreateGroup(name string) (Group, error)
    DeleteGroup(name string) error
}

type GroupMemberClient interface {
    ListGroupMembers(group string, limit int) ([]User, error)
    AddGroupMember(group, userSlug string) error
    RemoveGroupMember(group, userSlug string) error
}

type Group struct {
    Name      string
    Deletable bool
}
```

**Commands:** `bitbottle group list [--filter PREFIX] [--limit N] [--json]`, `bitbottle group create NAME`, `bitbottle group delete NAME [--confirm]`, `bitbottle group member list NAME [--limit N] [--json]`, `bitbottle group member add NAME USER`, `bitbottle group member remove NAME USER`

**Backends:** Server (`GET /rest/api/1.0/admin/groups`, `POST /rest/api/1.0/admin/groups`, `DELETE /rest/api/1.0/admin/groups`, `GET /rest/api/1.0/admin/groups/more-members?context=NAME`, `POST /rest/api/1.0/admin/users/add-group`, `POST /rest/api/1.0/admin/users/remove-group`). Cloud → `host.unsupported`.

**MCP tools:** `list_groups`, `create_group`, `delete_group`, `list_group_members`, `add_group_member`, `remove_group_member`

**Definition of Done:**
- [ ] `api/backend/client_group.go` — `GroupClient` + `GroupMemberClient` + `Group` type
- [ ] `api/server/groups.go` + `api/server/groups_test.go` (uses `paging.Collect[Group]` and `paging.Collect[User]`)
- [ ] `api/cloud/groups.go` returning `host.unsupported` (`unsupported_groups_test.go`)
- [ ] `pkg/cmd/group/` (list, create, delete, member/list, member/add, member/remove) + tests
- [ ] MCP triplet (`tools.go` + `handlers.go` + `handlers_test.go`) — six tools
- [ ] `test/script/testdata/group_*.txtar` covering list + create-then-delete + member add+remove
- [ ] `skills/SKILL.md` + `skills/references/admin.md` (or new `skills/references/group.md`)
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### EXT-SCAFFOLD — Extension Scaffold

**Status:** ✅

EXT-CORE / EXT-RUNTIME / EXT-MGMT shipped the install + exec + upgrade + remove half of the extension ecosystem, but there's still no `bitbottle extension scaffold NAME` to seed a new extension project — equivalent to `gh extension create`. Today an author has to read the EXT-CORE spec, hand-write the binary-naming convention (`bitbottle-NAME-<os>-<arch>`), and hand-write the GitHub release workflow that EXT-MGMT's auto-upgrade lockfile expects. This scope closes the on-ramp. Pure-local: no Bitbucket API, no backend interface, no MCP exposure (template generation is a one-shot per-developer act, not an agent loop).

**Interface:** No new backend interface. New scaffold package `pkg/cmd/extension/scaffold/` with embedded templates under `pkg/cmd/extension/scaffold/templates/` consumed via `embed.FS`.

**Commands:** `bitbottle extension scaffold NAME [--lang go|bash] [--dir DIR]` (default `--lang go`, default `--dir .`)

**Backends:** N/A — pure-local file generation. No `host.unsupported`.

**MCP tools:** none

**Definition of Done:**
- [ ] `pkg/cmd/extension/scaffold/scaffold.go` + `scaffold_test.go` (golden-file assertions on emitted tree)
- [ ] `pkg/cmd/extension/scaffold/templates/go/{main.go.tmpl,go.mod.tmpl,README.md.tmpl,release.yml.tmpl,LICENSE.tmpl}`
- [ ] `pkg/cmd/extension/scaffold/templates/bash/{bitbottle-NAME.tmpl,README.md.tmpl,release.yml.tmpl,LICENSE.tmpl}`
- [ ] Generated `release.yml` produces an asset named `bitbottle-NAME-<os>-<arch>` recognised by EXT-CORE's installer (round-trip-tested with a fake release)
- [ ] `test/script/testdata/extension_scaffold_*.txtar` for both `--lang` values, plus a round-trip script that scaffolds → builds → installs locally → execs
- [ ] `skills/SKILL.md` + `skills/references/extension.md` updated
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### NIGHTLY-E2E — Nightly live-wire E2E workflow

**Status:** ✅

`docs/ARCHITECTURE.md` ("Test tiers") and `docs/TESTING-PYRAMID.md` both call out tier 6 — nightly live-wire E2E against real Bitbucket Cloud + real Bitbucket Server — as an explicit gap. The TESTSCRIPT corpus (✅) and TESTSCRIPT-BACKFILL (✅) already carry `BITBOTTLE_E2E=1` opt-in, so the harness is wired; the missing piece is a scheduled workflow that flips the flag and runs against real backends with secrets-based credentials. This catches the class of bug that's invisible to hermetic `bb-fake` runs: real auth flows, real rate-limit shaping, real Bitbucket version drift, real CSRF/content-type policy regressions.

**Interface:** No new code interface. New GHA workflow + secrets contract.

**Commands:** none (CI-only)

**Backends:** N/A — exercises both via the existing testscript corpus.

**MCP tools:** none

**Definition of Done:**
- [ ] `.github/workflows/nightly-e2e.yml` — `schedule: '0 2 * * *'` + `workflow_dispatch`, matrix `{cloud, server}`, runs `go test ./test/script/... -run TestScript -tags e2e` with `BITBOTTLE_E2E=1` and the per-backend secrets exported
- [ ] Secrets contract documented: `BB_CLOUD_WORKSPACE`, `BB_CLOUD_REPO`, `BB_CLOUD_TOKEN`, `BB_SERVER_URL`, `BB_SERVER_PROJECT`, `BB_SERVER_REPO`, `BB_SERVER_TOKEN` (documented in `docs/RELEASE.md` or a new `docs/CI-SECRETS.md`)
- [ ] On failure: a single rolling GitHub issue is opened or updated (title `nightly-e2e: failing as of <date>`) with the failing script names + run URL; closes on next green run
- [ ] Failures never block `main` (workflow is not a required check)
- [ ] `test/script/scripts.go` (or the testscript bootstrap) reads the env vars and skips Cloud-only scripts on the server matrix leg and vice versa
- [ ] `docs/TESTING-PYRAMID.md` tier-6 status flipped from "not shipped" to "shipped"
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### SOURCE-WRITE — Write file content via API

**Status:** ✅

Complement to `repo file get` (✅ RV1). Allows agents and scripts to create or update a file on a branch without a local git clone. Useful for programmatically updating config files, changelogs, READMEs, etc.

**Interface:**
```go
// api/backend/client_source_write.go
type SourceWriter interface {
    PutFile(ns, slug, path string, in PutFileInput) error
}

type PutFileInput struct {
    Content      string // raw file content (UTF-8 text or base64 for binary)
    Branch       string // target branch
    Message      string // commit message
    SourceCommit string // optional: expected HEAD SHA for conflict detection
}
```

**Commands:** `bitbottle repo file put PATH --branch BRANCH --message MSG [--content TEXT | --content-file FILE] [--source-commit SHA] [PROJECT/REPO]`

**Backends:**
- Server: `PUT /rest/api/1.0/projects/{k}/repos/{s}/browse/{path}` with multipart form body: `content=<bytes>`, `branch=<name>`, `message=<text>`, `sourceCommitId=<sha>` (optional, for conflict detection)
- Cloud: `POST /repositories/{ws}/{slug}/src` with multipart form: file part named as the path, plus `message=<text>`, `branch=<name>`, `parents=<sha>` (optional)

**MCP tools:** `put_file`

**Definition of Done:**
- [x] `api/backend/client_source_write.go` — `SourceWriter` interface + `PutFileInput` type + `FeatureSourceWrite` + `AsSourceWriter`
- [x] `api/server/source_write.go` + `api/server/source_write_test.go`
- [x] `api/cloud/source_write.go` + `api/cloud/source_write_test.go`
- [x] `pkg/cmd/repo/file/put/put.go` + `pkg/cmd/repo/file/put/put_test.go`
- [x] `pkg/cmd/mcp/tools_repo.go` entry + `pkg/cmd/mcp/handlers_repo.go` method + `pkg/cmd/mcp/handlers_repo_test.go` test
- [x] `test/testhelpers/client.go` — `PutFileFn` field + compile assertion
- [x] `api/backend/features.go` — `FeatureSourceWrite` + `AllFeatureSpecs` entry (count 50 → 51)
- [x] `api/backend/capability_contract_test.go` — count + name updated
- [x] `skills/SKILL.md` + `skills/references/repos.md` updated
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### ADMIN-MAIL — Admin mail server configuration

**Status:** ✅

Bitbucket Server/DC exposes SMTP configuration via a dedicated admin REST endpoint. This is separate from general `admin logging` and `admin secrets` — it controls how Bitbucket sends email notifications. Cloud → `ErrUnsupportedOnHost` (Cloud uses Atlassian's managed email infrastructure).

**Interface** (extends `AdminClient` in `api/backend/client_admin.go`):
```go
GetMailServerConfig() (MailServerConfig, error)
SetMailServerConfig(in MailServerConfig) error

type MailServerConfig struct {
    Hostname        string
    Port            int
    Protocol        string // "smtp" | "smtps"
    UseStartTLS     bool
    RequireStartTLS bool
    Username        string
    SenderAddress   string
    // Password is write-only; never returned in GET response
}
```

**Commands:**
- `bitbottle admin mail get [--hostname HOST] [--json]`
- `bitbottle admin mail set [--hostname HOST] --host HOSTNAME --port N [--protocol smtp|smtps] [--username U] [--sender EMAIL] [--use-start-tls] [--require-start-tls]`

**Backends:** Server (`GET /rest/api/1.0/admin/mail-server`, `PUT /rest/api/1.0/admin/mail-server`). Cloud → `ErrUnsupportedOnHost`.

**MCP tools:** `get_mail_server_config`, `set_mail_server_config`

**Definition of Done:**
- [x] `api/backend/client_admin.go` — add `GetMailServerConfig` + `SetMailServerConfig` methods + `MailServerConfig` type
- [x] `api/server/admin.go` — implement both methods
- [x] `pkg/cmd/admin/mail/mail.go` + `pkg/cmd/admin/mail/get/get.go` + `pkg/cmd/admin/mail/set/set.go` + matching `_test.go`
- [x] `pkg/cmd/admin/admin.go` — wire `mail.NewCmdMail(f)`
- [x] `pkg/cmd/mcp/tools_admin.go` entries + `pkg/cmd/mcp/handlers_admin.go` methods + `pkg/cmd/mcp/handlers_admin_test.go` tests
- [x] `test/testhelpers/client.go` — `GetMailServerConfigFn` + `SetMailServerConfigFn` fields + implementations
- [x] `skills/references/admin.md` updated
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### ADMIN-BANNER — Admin site-wide announcement banner

**Status:** ✅

Bitbucket Server/DC exposes a site-wide announcement banner endpoint. Admins use this to post maintenance windows, upgrade notices, and scheduled downtime messages to users logging in. Cloud has no equivalent → `ErrUnsupportedOnHost`.

**Interface** (extends `AdminClient` in `api/backend/client_admin.go`):
```go
GetBanner() (BannerConfig, error)
SetBanner(in BannerConfig) error
ClearBanner() error

type BannerConfig struct {
    Message  string
    Audience string // "ALL" | "AUTHENTICATED" | "UNAUTHENTICATED"
    Enabled  bool
}
```

**Commands:**
- `bitbottle admin banner get [--json]`
- `bitbottle admin banner set MESSAGE [--audience all|authenticated|unauthenticated]`
- `bitbottle admin banner clear [--confirm]`

**Backends:** Server (`GET /rest/api/1.0/admin/banner`, `PUT /rest/api/1.0/admin/banner`, `DELETE /rest/api/1.0/admin/banner`). Cloud → `ErrUnsupportedOnHost`.

**MCP tools:** `get_banner`, `set_banner`, `clear_banner`

**Definition of Done:**
- [x] `api/backend/client_admin.go` — add `GetBanner`, `SetBanner`, `ClearBanner` methods + `BannerConfig` type
- [x] `api/server/admin.go` — implement all 3 methods
- [x] `api/server/admin_test.go` — tests for all 3 methods
- [x] `pkg/cmd/admin/banner/banner.go` + `get/get.go` + `set/set.go` + `clear/clear.go` + matching `_test.go`
- [x] `pkg/cmd/admin/admin.go` — wire `banner.NewCmdBanner(f)`
- [x] `pkg/cmd/mcp/tools_admin.go` entries + `pkg/cmd/mcp/handlers_admin.go` methods + `pkg/cmd/mcp/handlers_admin_test.go` tests
- [x] `test/testhelpers/client.go` — `GetBannerFn`, `SetBannerFn`, `ClearBannerFn` + compile assertions
- [x] `skills/references/admin.md` updated
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### PIPE-RUNNERS — Pipeline self-hosted runner management

**Status:** ✅

Bitbucket Cloud Pipelines supports self-hosted runners that execute pipeline steps on customer-owned infrastructure. Runners are registered at the workspace level and identified by a UUID. Cloud only — Server/DC has no equivalent → `ErrUnsupportedOnHost`.

**Interface** (`api/backend/client_runner.go`):
```go
type RunnerClient interface {
    ListRunners(workspace string) ([]Runner, error)
    CreateRunner(workspace string, in CreateRunnerInput) (Runner, error)
    DeleteRunner(workspace, runnerUUID string) error
}

type Runner struct {
    UUID     string
    Name     string
    State    string
    Platform RunnerPlatform
    Labels   []string
}

type RunnerPlatform struct {
    Operating string // LINUX | WINDOWS | MACOS
    Arch      string // AMD64 | ARM64
}

type CreateRunnerInput struct {
    Name     string
    Labels   []string
    Platform RunnerPlatform
}
```

**Commands:**
- `bitbottle runner list [WORKSPACE]`
- `bitbottle runner create [WORKSPACE] --name NAME [--platform linux_amd64|linux_arm64|windows_amd64|macos_arm64] [--label LABEL...]`
- `bitbottle runner delete [WORKSPACE] UUID`

**Backends:** Cloud (`GET /2.0/workspaces/{ws}/pipelines-config/runners`, `POST /2.0/workspaces/{ws}/pipelines-config/runners`, `DELETE /2.0/workspaces/{ws}/pipelines-config/runners/{uuid}`). Server/DC → `ErrUnsupportedOnHost`.

**Note:** API uses `X86_64` for arch; internal uses `AMD64`. `normalizeArch` converts on read; `apiArch` converts on write.

**MCP tools:** `list_runners`, `create_runner`, `delete_runner`

**Definition of Done:**
- [x] `api/backend/client_runner.go` — `RunnerClient` interface, `Runner`/`RunnerPlatform`/`CreateRunnerInput` types, `FeatureRunner` constant, `AsRunnerClient` helper
- [x] `api/backend/features.go` — `FeatureRunner` spec added (Cloud: true, Server: false); count 51 → 52
- [x] `api/backend/capability_contract_test.go` — `RunnerClient` added to expected names
- [x] `api/cloud/runner.go` — wire types + `ListRunners`, `CreateRunner`, `DeleteRunner` implementations with `normalizeArch`/`apiArch`
- [x] `api/cloud/runner_test.go` — httptest server tests for all 3 methods
- [x] `pkg/cmd/runner/runner.go` — umbrella command + `init()` self-registration
- [x] `pkg/cmd/runner/list/list.go` + `list_test.go`
- [x] `pkg/cmd/runner/create/create.go` + `create_test.go`
- [x] `pkg/cmd/runner/delete/delete.go` + `delete_test.go`
- [x] `pkg/cmd/root/root.go` — blank import for `pkg/cmd/runner`
- [x] `pkg/cmd/mcp/tools_runner.go` + `handlers_runner.go` + `handlers_runner_test.go`
- [x] `test/testhelpers/client.go` — `ListRunnersFn`, `CreateRunnerFn`, `DeleteRunnerFn` + compile assertion
- [x] `skills/references/runner.md` created
- [x] `skills/SKILL.md` updated (reference table + Cloud-only list)
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### WORKSPACE-PIPELINE-VARS — Workspace-scoped Pipeline Variables

**Status:** ✅

Cloud workspace pipeline variables are shared across all repos in the workspace and are injected into every pipeline step. Unlike repo-level pipeline variables (`variable --scope repository`) and deployment env vars, these live at the workspace level. The API uses UUIDs rather than plain keys for addressing.

**Interface:**
```go
type WorkspacePipelineVariableClient interface {
    ListWorkspacePipelineVariables(workspace string) ([]WorkspacePipelineVariable, error)
    GetWorkspacePipelineVariable(workspace, uuid string) (WorkspacePipelineVariable, error)
    SetWorkspacePipelineVariable(workspace string, in WorkspacePipelineVariableInput) (WorkspacePipelineVariable, error)
    DeleteWorkspacePipelineVariable(workspace, uuid string) error
}
type WorkspacePipelineVariable struct {
    UUID    string
    Key     string
    Value   string
    Secured bool
}
type WorkspacePipelineVariableInput struct {
    Key     string
    Value   string
    Secured bool
    UUID    string // non-empty = update existing
}
```

**Commands:** `workspace pipeline-variable list [WORKSPACE]`, `workspace pipeline-variable get [WORKSPACE] KEY`, `workspace pipeline-variable set [WORKSPACE] KEY VALUE [--secured]`, `workspace pipeline-variable delete [WORKSPACE] KEY [--confirm]`

**Backends:** Cloud (`GET/POST /workspaces/{ws}/pipelines-config/variables`, `GET/PUT/DELETE /workspaces/{ws}/pipelines-config/variables/{uuid}`). Server → typed `host.unsupported`.

**MCP tools:** `list_workspace_pipeline_vars`, `get_workspace_pipeline_var`, `set_workspace_pipeline_var`, `delete_workspace_pipeline_var`

**Definition of Done:**
- [ ] `api/backend/client_workspace_pipeline_vars.go` — interface + types + `FeatureWorkspacePipelineVars` + helper
- [ ] `api/backend/features.go` — spec entry; count bumped
- [ ] `api/cloud/workspace_pipeline_vars.go` + `_test.go` — `get` resolves key→uuid via list then fetches by uuid
- [ ] `pkg/cmd/workspace/pipelinevar/` — list, get, set, delete (under `workspace pipeline-variable`)
- [ ] `pkg/cmd/workspace/workspace.go` — add `cmdPipelineVar.NewCmdWorkspacePipelineVariable(f)`
- [ ] `pkg/cmd/mcp/tools_workspace_pipeline_vars.go` + `handlers_workspace_pipeline_vars.go` + `_test.go`
- [ ] `test/testhelpers/client.go` — FakeClient extensions + compile assertions
- [ ] `test/script/testdata/workspace_pipeline_vars.txtar`
- [ ] `skills/SKILL.md` + `skills/references/workspace.md` updated
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### ISSUE-ACTIVITY — Cloud Issue Activity Log

**Status:** ✅

The Cloud issue tracker records every state change as an immutable event (state/priority/assignee/component/milestone/title transitions plus comment events). Today `issue view` ✅ shows current state only. This scope exposes the full audit trail — useful for agents tracing issue lifecycle or generating changelogs.

**Interface:**
```go
type IssueActivityClient interface {
    ListIssueActivity(ns, slug string, issueID int, limit int) ([]IssueChange, error)
}
type IssueChange struct {
    ID        int
    Kind      string // "status" | "priority" | "assignee" | "component" | "milestone" | "title" | "content" | "comment"
    OldVal    string
    NewVal    string
    CreatedOn time.Time
    User      User
}
```

**Commands:** `bitbottle issue activity ISSUE_ID [PROJECT/REPO] [--limit N] [--json]`

**Backends:** Cloud (`GET /repositories/{ws}/{slug}/issues/{id}/changes`, paginated). Server → typed `host.unsupported`.

**MCP tools:** `list_issue_activity`

**Definition of Done:**
- [x] `api/backend/client_issue_activity.go` — `IssueActivityClient` interface + `IssueChange` type + `FeatureIssueActivity` + helper
- [x] `api/backend/features.go` — spec entry; count bumped
- [x] `api/cloud/issue_activity.go` + `_test.go`
- [x] `pkg/cmd/issue/activity/activity.go` + `_test.go`; add to `pkg/cmd/issue/issue.go`
- [x] `pkg/cmd/mcp/tools_issue_activity.go` + `handlers_issue_activity.go` + `_test.go`
- [x] `test/testhelpers/client.go` — FakeClient + compile assertion
- [x] `test/script/testdata/issue_activity.txtar`
- [x] `skills/SKILL.md` + `skills/references/issue.md` updated
- [x] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### CLOUD-PROJECT-REVIEWERS — Cloud Project Default Reviewers

**Status:** ✅

DEFAULT-REVIEWERS (✅) shipped per-repo default reviewers for both Cloud and Server. Cloud additionally exposes a project-level default reviewer collection at `/workspaces/{ws}/projects/{key}/default-reviewers` whose entries cascade to every repo inside that project — the canonical way to set team-wide PR review policy without touching N repos. This scope adds the missing CRUD surface as a separate command tree under `workspace project default-reviewer ...`.

**Interface:** New optional interface `CloudProjectDefaultReviewerClient` with `ListProjectDefaultReviewers(ws, projectKey string, limit int) ([]User, error)`, `AddProjectDefaultReviewer(ws, projectKey, accountID string) error`, `RemoveProjectDefaultReviewer(ws, projectKey, accountID string) error`.

**Commands:** `workspace project default-reviewer list WORKSPACE PROJECT_KEY [--limit N] [--json]`, `add WORKSPACE PROJECT_KEY --user ACCOUNT_ID`, `remove WORKSPACE PROJECT_KEY --user ACCOUNT_ID [--confirm]`

**Backends:** Cloud only. Server/DC → typed `host.unsupported`.

**MCP tools:** `list_project_default_reviewers`, `add_project_default_reviewer`, `remove_project_default_reviewer`

**Definition of Done:**
- [ ] `api/backend/client_cloud_project_reviewers.go` — interface + feature const + `AsCloudProjectDefaultReviewerClient`
- [ ] `api/cloud/workspace_project_reviewers.go` — impl
- [ ] `pkg/cmd/workspace/project/defaultreviewer/` — list/add/remove commands
- [ ] `pkg/cmd/mcp/` triplet
- [ ] `test/testhelpers/client.go` — FakeClient updated
- [ ] `test/script/testdata/workspace_project_default_reviewer.txtar`
- [ ] `skills/SKILL.md` + `skills/references/workspace.md` updated
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---

### PR-MERGE-PREVIEW — PR Merge Preview / Dry-Run

**Status:** ✅

The shipped `pr merge` command is fire-and-forget. Agents wanting to know whether a merge will succeed (and which files conflict) have to trust the lossy `mergeable: bool` field on `pr view`, or clone locally. Both Cloud and Server expose a true dry-run endpoint returning conflicted paths and plugin vetoes. This scope adds `--dry-run` as a top-level flag on the existing `pr merge` command (no state change, no `--confirm` required).

**Interface:** Extends `PRMerger` interface with `DryRunMergePR(ns, slug string, prID int, in DryRunMergePRInput) (MergeDryRunResult, error)`. Domain type `MergeDryRunResult{CanMerge bool, MessagePreview string, ConflictedPaths []string, Vetoes []MergeVeto}`. `MergeVeto{SummaryMessage, DetailedMessage string}`.

**Commands:** `pr merge PR_ID --dry-run [--strategy ff|squash|merge-commit] [PROJECT/REPO] [--json]`

**Backends:** Cloud (`POST /repositories/{ws}/{slug}/pullrequests/{id}/merge?dry_run=true`). Server 7.0+ (`POST /rest/api/1.0/projects/{k}/repos/{s}/pull-requests/{id}/merge/dry-run`). Both backends.

**MCP tools:** `dry_run_merge_pr`

**Definition of Done:**
- [ ] `api/backend/client_pr.go` — `DryRunMergePR` added to `PRMerger` interface + domain types
- [ ] `api/cloud/pr_merge.go` — `DryRunMergePR` impl
- [ ] `api/server/pr_merge.go` — `DryRunMergePR` impl (Server 7.0+)
- [ ] `pkg/cmd/pr/merge/merge.go` — `--dry-run` flag wired
- [ ] `pkg/cmd/mcp/` triplet entry + handler + test
- [ ] `test/testhelpers/client.go` — FakeClient updated
- [ ] `test/script/testdata/pr_merge_dry_run.txtar`
- [ ] `skills/SKILL.md` + `skills/references/pr.md` updated
- [ ] BACKLOG.md row flipped 🔲 → ✅ in the same `feat:` commit

---


## 2026-06-02 — REF-UX — branch/tag create --start-at positional (Option A)

`branch create` and `tag create` now accept START_AT as a 3rd positional argument
(`[PROJECT/REPO] NAME [START_AT]`). The `--start-at` flag remains for backward
compatibility. Closes #621.

---

## 2026-05-31 — CLOUD-WIRE

### CLOUD-WIRE — Cloud API path + response-struct drift (BB-08, BB-09, BB-10)

- **Fix commit:** `fix(cloud): CLOUD-WIRE — /permissions-config/ path, /commit/ singular, pipeline trigger response struct` (2026-05-31)
- **Backends:** Cloud only
- **Estimate (planned):** 1.5 days

**Status:** ✅ — sourced from the 2026-05-27 CLI-comparison audit.

**Why P1:** Three Cloud-only wire-level defects. (1) `workspace project perms list` calls `/permissions/users` but the correct path is `/permissions-config/users` — 404. (2) `commit comment {list,add,edit,delete}` all hardcode `/commits/{hash}/comments` (plural) but Bitbucket Cloud uses `/commit/{hash}/comments` (singular) — 404 on all four operations. (3) `pipeline trigger`'s response struct types `CloudTriggerResponseLinks.links.self` as `[]CloudTriggerResponseSelfLink` (slice) but the API returns a single object — JSON unmarshal failure, no useful output (and exit 0, see BB-12).

**Shape:**

1. **Endpoint fixes** — `api/cloud/workspace_project_perms.go`: change `/permissions/users` → `/permissions-config/users`. `api/cloud/commit_comment.go` lines 32, 52, 65, 74: change `/commits/{hash}/comments` → `/commit/{hash}/comments`. Both are ~4 lines + regression tests against captured fixtures.
2. **Pipeline trigger struct** — fix the generator (or hand-correct the type) so `links.self` is `*CloudTriggerResponseSelfLink` (single, nullable). Add a unit test loading a captured pipeline-trigger response fixture and asserting unmarshal success.
3. **Add live-wire (tier 6) tests** for each of the three fixed endpoints to prevent silent re-drift if Atlassian moves the paths again.

**Definition of Done:**

- [x] `api/cloud/workspace_project_perms.go` — path corrected to `/permissions-config/`.
- [x] `api/cloud/commit_comment.go` — all four operations use `/commit/` (singular).
- [x] `api/cloud/gen/openapi.yaml` + `types.go` regenerated — `links.self` typed as `*CloudTriggerResponseSelfLink` + unmarshal regression test green.
- [x] `test/script/script_test.go` — bb-fake Cloud stubs updated to `/permissions-config/` paths.
- [ ] `test/testdata/fixtures/pipeline_trigger_response.json` — captured live response (deferred: requires live Cloud sandbox).
- [ ] `.txtar` scripts: `commit_comment_lifecycle.txtar`, `pipeline_trigger.txtar` (deferred: no bb-fake stubs for these endpoints yet).
- [ ] Nightly `BITBOTTLE_E2E=1` runs against real Cloud sandbox (deferred: infra not yet wired).

---

## 2026-05-31 — PR-GUARDS

### PR-GUARDS — PR state-machine + `--state` enum validation (BB-07, BB-11, BB-20, BB-21)

- **Fix commit:** `fix(pr): PR-GUARDS — state-machine guard + --state enum validation` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 1.5 days

**Status:** ✅ — sourced from the 2026-05-27 CLI-comparison audit.

**Why P1:** Two related defect classes. (1) `pr approve` succeeds with exit 0 + "Approved pull request #N" on a DECLINED PR — because the Bitbucket Cloud API returns 200 for participant approval regardless of PR state. `pr decline` and `pr merge` correctly reject DECLINED/MERGED PRs (because the API returns 400 there). So bitbottle is "right by accident" three times, and wrong on the one path the API is permissive on. (2) `pr list --state INVALID_STATE` (and `--state ""`) silently returns all 11 PRs across MERGED/OPEN/DECLINED — no client-side validation.

**Shape:**

1. **`ValidateMutablePRState(pr) error`** in `api/backend/pr_state.go` — returns typed `*DomainError{Kind: ErrConflict}` if state ∈ {DECLINED, MERGED, SUPERSEDED}. Called by `pr approve`, `pr unapprove`, `pr request-changes`, `pr edit`, before the mutation request.
2. **`--state` enum validation** via `pkg/cmd/internal/enumflag/` `pflag.Value` helper. Rejects `""`, `INVALID_STATE`, and any off-enum value at parse time.

**Definition of Done (as shipped):**

- [x] `api/backend/pr_state.go` — `ValidateMutablePRState` helper + tests.
- [x] `pkg/cmd/pr/{approve,unapprove,request-changes,edit}` — call the guard before the API request.
- [x] `pkg/cmd/internal/enumflag/enumflag.go` — generic enum flag helper + tests.
- [x] `pkg/cmd/pr/list.go` — `--state` uses enum flag, rejects invalid values.
- [x] `.txtar` scripts: `pr_approve_on_declined.txtar`, `pr_list_state_invalid.txtar`.
- [x] `pr approve N` on DECLINED PR exits 1 with typed error.

---

## 2026-05-31 — MCP-TAXONOMY

### MCP-TAXONOMY — Unify tool catalog (MCP-04, MCP-05, MCP-16)

- **Fix commit:** `fix(mcp): unify MCP tool catalog (MCP-04, MCP-05, MCP-16)` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 1.5 days

Collapsed three structural inconsistencies in the 254-tool MCP catalog that
forced AI clients to special-case bitbottle:

- **MCP-04 — canonical `{project, slug}` repo-arg shape.** Migrated the five
  BACKLOG-named alternates — `compare_refs`, `list_pr_commits`,
  `list_pr_files` (were `{repo}` = `WORKSPACE/REPO`), and
  `get_repo_pr_settings` / `set_repo_pr_settings` (were `{project, repo}`) —
  to `{project, slug}`. The old shape still works for one release; the tool
  result prepends a `DEPRECATION` text block. Shared resolver helpers live in
  `pkg/cmd/mcp/repo_arg.go` (`repoFromProjectSlugOrRepo`,
  `repoFromProjectSlugOrProjectRepo`, `withDeprecation`). Many other tools
  intentionally keep `{repo}` — the regression test is an allowlist of the
  five migrated tools, not a global ban.
- **MCP-05 — reject unknown hostnames.** `handlers.resolveBackend` now calls
  `requireConfiguredHost`, which returns a typed
  `*DomainError{Kind: ErrUnknownHost, Code: host.unknown}` listing the
  configured hosts when a non-empty `hostname` isn't in `hosts.yml` — before
  any HTTP or Server-vs-Cloud URL inference. Added `ErrUnknownHost` kind +
  `CodeHostUnknown` code (wired through `AllCodes` + the errfmt catalogue).
- **MCP-16 — `_meta.backends` + pre-HTTP gating.** A small registration
  wrapper (`addGatedTool` in `pkg/cmd/mcp/tools_backends.go`) stamps
  `_meta.backends` (`["server"]` / `["cloud"]` / both) onto a tool's
  `tools/list` entry and wraps its handler with a pre-HTTP gate: a
  single-backend tool invoked against the wrong flavour returns
  `host.unsupported` (with the allowed list) before the backend is dialed.
  Scoped to backend-specific tools — the five migrated tools were rewired
  through the wrapper; the broader catalog defaults to both backends (no
  gating), avoiding 250+ hand-edits. The optional Phase-2 host-filtered
  `tools/list` (`BITBOTTLE_MCP_HOST_FILTER=1`) was skipped as out-of-scope
  for the diff budget.

## 2026-05-31 — MCP-INPUT-VALIDATION

### MCP-INPUT-VALIDATION — Tighten client-side validators across MCP tools (MCP-06 through MCP-14)

- **Fix commit:** `fix(mcp): tighten client-side arg validators (MCP-06..MCP-14)` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 2.5 days

Shipped a shared `pkg/cmd/mcp/argval` package of typed extractors —
`Int` (with `Required`/`Min`/`Max`, treats explicit 0 as present, rejects
wrong-type and out-of-range), `Hash(minLen=7)`, `RefName`, `EnumOneOf`
(panics on an empty member so `""` can never re-enter an enum),
`MutuallyRequired`, and `OneOfRequired` — each returning a structured
`{code, field, got, message}` envelope (`arg.missing` / `arg.invalid_type`
/ `arg.out_of_range` / `arg.invalid_value`). All numeric-id handlers across
the MCP surface were migrated to a shared `requireIntArg` helper, and the
six affected tools (`merge_pr`, `add_pr_comment`, `add_commit_comment`,
`create_branch`, `update_pr`, `compare_refs`) now reject malformed input
client-side instead of forwarding it to a generic upstream 404. The
`merge_pr` `strategy` schema gained an explicit `Enum(merge, squash,
rebase)`. A 14-case acceptance matrix (`handlers_input_validation_test.go`)
replays each of MCP-06 … MCP-14 against the structured envelope, alongside
the `argval` package's own unit tests.

Note vs. the original DoD: the scope was delivered as a `fix(mcp):` commit
(behavioural correctness fixes, not a new feature). The structured
envelope and 14-case Go matrix cover the acceptance criteria in lieu of a
`.txtar` corpus; migration touched the affected + mechanically-safe
numeric-id handlers rather than all 254 (the PRD explicitly de-scoped a
full sweep to avoid ballooning the diff).

### MCP-INPUT-VALIDATION — original backlog detail

**Status:** ✅ — sourced from the 2026-05-27 MCP sweep (Phase 3).

**Why P1:** MCP arg validation is inconsistent — some fields validate beautifully (`inline_side` enum, 1-segment repo on `compare_refs`, missing-required-string), but adjacent fields on the same tools forward garbage to the HTTP layer and return a generic upstream 404. AI agents reading the error can't tell whether their input was malformed or the resource doesn't exist. Concretely, the negative-input matrix surfaced **nine** distinct gaps:

- **MCP-06** Wrong-type `id` → reported as "missing required parameter: id" instead of "id must be integer".
- **MCP-07** `id: 0` → falsely "missing" (Go zero-value issue in MCP unmarshal). Hits every numeric-id tool.
- **MCP-08** Negative `id` → passes through, generic 404.
- **MCP-09** `merge_pr.strategy` enum lists `""` as a valid value; error messages show "must be one of , merge, squash, …" (note the bare comma).
- **MCP-10** `add_pr_comment` inline-anchor asymmetry: `inline_path` w/o `inline_line` is caught client-side, but `inline_line` w/o `inline_path` is not.
- **MCP-11** `add_commit_comment.hash` not validated for length/hex; "a" or "NOT_HEX_!@#" reaches Cloud and returns generic 404.
- **MCP-12** `create_branch.name` accepts `"/"`, leading/trailing slashes, and other refs that Git refuses by spec.
- **MCP-13** `update_pr` with neither `title` nor `body` hits the API instead of returning a clean "nothing to update".
- **MCP-14** `compare_refs.repo` rejects 1-segment input cleanly but silently accepts 3-segment (`bitbucket.org/proj/repo`).

These are individually small but collectively they prevent any safe automated retry policy on the MCP surface.

**Shape:**

1. **Shared `argval` helper** in `pkg/cmd/mcp/argval/` — typed extractors: `Int(name, required, min=)`, `Hash(name, minLen=7)`, `RefName(name)`, `EnumOneOf(name, allowed)`, `MutuallyRequired(field, dependency)`, `OneOfRequired(fields)`. Each extractor returns a structured error mapped to MCP's tool-error envelope so the client gets `{code: "arg.invalid_type", field: "id", got: "string"}`-style payloads.
2. **Migrate every handler** in `pkg/cmd/mcp/handlers_*.go` to call the new helpers in their first 3–5 lines.
3. **Fix the strategy enum** for `merge_pr` — drop `""` from the canonical list, treat empty as "use default" via a separate branch.
4. **Sweep all existing tool registrations** for similar zero-value pitfalls — anywhere a number can legally be 0 (limits, pagination), confirm `_, ok := args[name]` is used not `args[name].(int) > 0`.
5. **Tabulate per-handler arg specs** in a generated table — one row per `{tool, field, type, required, validators}`. Becomes the spec source for both runtime check and JSONSchema export.

**Definition of Done (as shipped):**

- [x] `pkg/cmd/mcp/argval/argval.go` lands with typed extractors + tests.
- [x] Affected + mechanically-safe numeric-id handlers migrated to the shared helpers (full 254-handler sweep de-scoped per PRD).
- [x] Test matrix replaying MCP-06 … MCP-14 — every one returns a structured error with the right `code` and `field`.
- [x] `merge_pr` strategy enum no longer lists `""`.
- [x] Go acceptance matrix covers each of MCP-06 through MCP-14 with the corrected error envelope (`.txtar` not required given Go coverage).
- [ ] CHANGELOG line per bug class — deferred to release-please (owns CHANGELOG.md).

---

## 2026-05-31 — FMT-CONTRACT

### FMT-CONTRACT — `--jq` / `--template` / hint consistency (BB-17, BB-18, BB-19, MCP-15)

- **Fix commit:** `fix(fmt): unify --jq guard, template trailing newline, neutral not-found hint, user-view Cloud IDs` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 1.25 days

Output-format middleware silently misbehaved, breaking scripts. Four fixes, one PR:

1. **`--jq` consistency (BB-17)** — root's `PersistentPreRunE` carried the `--jq requires --json` guard plus the json/yaml/template mutual-exclusion, but cobra runs only the single *deepest* `PersistentPreRunE` in a command chain. `factory.EnableRepoOverride` (the `-R` flag wiring) installs its own hook on every command group, which shadowed the root guard — so `pr view --jq .x` / `repo view --jq .x` silently ignored `--jq` and returned the full text view, while `status`/`pr list` correctly errored. Extracted the contract into `cmdutil.ValidateOutputFlags`, called from **both** root's hook and the repo-override hook. Now `--jq` without `--json` errors identically on `pr view`, `repo view`, `pr list`, `branch list`; `--jq` with `--json` applies on all of them.
2. **`--template` trailing newline (BB-18)** — `format.WriteTemplate` now newline-terminates rendered output (text/template appends none), matching `--json`/`--yaml` encoders.
3. **Hint accuracy (BB-19)** — the `pr.not_found` errfmt catalogue hint no longer claims the PR "may have been deleted" (misleading when `pr view 99999` never existed); reworded to the neutral "No pull request #N exists in this repo."
4. **MCP-15** — `backend.User` gained `AccountID`/`UUID`/`CreatedOn`/`Links.HTML.Href` (Cloud-stable identifiers, populated from `GET /user`). `user view --json` exposes them (JSONOnly so the TTY table stays slug+name) and the MCP `get_current_user` tool marshals them, giving AI clients durable machine-readable handles.

A `.txtar` corpus (`fmt_contract_jq_requires_json`, `fmt_contract_jq_applies`, `fmt_contract_template_newline`, `fmt_contract_user_view_fields`) proves all four behaviours end-to-end.

---

## 2026-05-31 — SCRIPT-TRUST

### SCRIPT-TRUST — Exit codes + `-R` flag + ref parser (BB-12, BB-13, BB-14)

- **Fix commits:** `fix(cli): unify repo-ref parser; accept HOST/PROJECT/REPO everywhere` + `fix(cli): wire -R/--repo on every command; add exit-code contract test` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 3 days

Scripts could not trust bitbottle's contract surface. Three fixes, one PR:

1. **Exit-code contract** — added `pkg/cmd/contract_test.go`, which walks every leaf cobra command, runs it with intentionally-bad input, and asserts that any command writing to stderr exits non-zero (guards against the `print-then-return-nil` anti-pattern). Leaves that shell out to external processes (`skill`, `extension`, `mcp`, `completion`, `alias`) are skipped — their exit-code behaviour is covered by their own package tests.
2. **`-R`/`--repo` unification** — `-R` was advertised in every command's INHERITED FLAGS but silently ignored on most non-PR commands because their `cobra.ExactArgs(N)` forced a positional repo ref. Loosened the leading repo positional to optional (`ExactArgs(N)` → `MaximumNArgs(1)` / `RangeArgs(N-1, N)`) on ~30 commands across branch, tag, commit, pipeline, codeinsights, webhook, deployment, environment and variable. Each splits args via the new `repoarg.SplitLeadingRepo` and falls back to `factory.ResolveTarget`, so `-R` / `BB_REPO` / pinned defaults now resolve the repo everywhere.
3. **Ref-parser unification** — added `pkg/cmd/internal/repoarg/ref.go` with the single `ParseRef` helper accepting `PROJECT/REPO` and `HOST/PROJECT/REPO`. `factory.ResolveTarget` routes through it, and `repo view` migrated off the 2-part-only `bbrepo.Parse`. Regression: `repo view bitbucket.org/ws/repo` (3-part) is now accepted everywhere.

A `.txtar` corpus (`script_trust_repo_flag.txtar`, `script_trust_error_exit.txtar`) proves `-R` resolves the positional, 3-part refs are accepted, and error paths exit non-zero.

---

## 2026-05-29 — CLOUD-DISCOVERY

### CLOUD-DISCOVERY — Migrate off deprecated `/2.0/workspaces` (MCP-01, MCP-02, MCP-03)

- **Feat commit:** `feat(workspace): migrate workspace list/search to /user/permissions/workspaces` (2026-05-29)
- **Backends:** Cloud
- **Estimate (planned):** 1 day

Atlassian deprecated `GET /2.0/workspaces` under CHANGE-2770; the endpoint returns HTTP 410 Gone. Both `workspace list` and `workspace search` were broken, leaving fresh users with no in-tool way to find their workspace slug. Migration target: `GET /2.0/user/permissions/workspaces` (workspaces now wrapped under `value[].workspace`). Added `cloudWorkspacePermission` DTO, updated both list/search unmarshal paths, fixed all test fixtures. Also added `ErrEndpointDeprecated` / `CodeEndpointDeprecated` to the errors catalogue so future 410s get a clean typed message with an upgrade link.

---

## 2026-05-29 — HOST-INFO

### HOST-INFO — Host Capabilities & Version Info

- **Feat commit:** `feat(host): add host info command (Cloud + Server)` (2026-05-29)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 2 days

Today an agent starting work against a new Bitbucket host has no single endpoint to discover (a) which backend type it is, (b) what version (matters for Server feature gates like `pr unready` 8.0+ requirement), (c) which optional `AsXxxClient` capabilities the host implements. The pieces exist internally — `ServerCapabilities.GetApplicationProperties()`, `api/server/version.go`'s parsed `Version`, and `AllFeatureSpecs` — but no command surfaces them. `context` (✅) answers "where am I?"; `host info` answers "what can I do here?".

**Interface (shipped):**
```go
// api/backend/client_host_info.go
type HostInfoClient interface {
    GetHostInfo() (HostInfo, error)
}

type HostInfo struct {
    BackendType       string   `json:"backend_type"`
    BaseURL           string   `json:"base_url"`
    Version           string   `json:"version,omitempty"`
    BuildNumber       string   `json:"build_number,omitempty"`
    DisplayName       string   `json:"display_name"`
    SupportedFeatures []string `json:"supported_features"`
}
```

**Commands:** `host info [--hostname H] [--json]`

**MCP tools:** `get_host_info`

---

## 2026-05-25 — Bootstrap migration

Three scopes pulled from BACKLOG.md "Up Next" that had shipped before the file split. Each has a corresponding `feat:` commit; ✅ markers in the Full Functionality Map cross-reference these.

### PR-SETTINGS — Per-Repo PR Settings

- **Feat commit:** `0be675b feat(repo): add repo pr-settings get/set commands (Cloud + Server/DC)` (2026-05-25)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 2 days

Bitbucket exposes per-repo pull-request configuration that's distinct from branch protections and from the branching model: required number of approvers, "all approvers must approve", "all tasks must be resolved", required green builds, and which merge strategies are allowed. Before this scope shipped, these could only be set in the web UI (Server) or by hand-rolling a `PUT /repositories/{ws}/{slug}` payload (Cloud). The scope shipped the read/write CLI surface that completes the trio: BRANCH-RULE (Cloud ref restrictions ✅), BRANCH-PROTECT (Server ref restrictions ✅), PR-SETTINGS (PR merge gate, both backends).

**Interface (shipped):**
```go
// api/backend/client_repo_pr_settings.go
type RepoPRSettingsClient interface {
    GetRepoPRSettings(ns, slug string) (RepoPRSettings, error)
    UpdateRepoPRSettings(ns, slug string, in RepoPRSettingsInput) (RepoPRSettings, error)
}

type RepoPRSettings struct {
    RequiredApprovers           int
    RequiredAllApprovers        bool
    RequiredAllTasksComplete    bool
    RequiredSuccessfulBuilds    int
    MergeStrategy               string   // default
    AllowedStrategies           []string // subset of {ff, ff-only, rebase, squash, merge-commit}
}

// Pointer fields → only set the ones the user passed.
type RepoPRSettingsInput struct {
    RequiredApprovers           *int
    RequiredAllApprovers        *bool
    RequiredAllTasksComplete    *bool
    RequiredSuccessfulBuilds    *int
    MergeStrategy               *string
    AllowedStrategies           *[]string
}
```

**Commands shipped:** `bitbottle repo pr-settings get [PROJECT/REPO] [--json]`, `bitbottle repo pr-settings set [PROJECT/REPO] [--required-approvers N] [--required-all-approvers[=false]] [--required-all-tasks-complete[=false]] [--required-successful-builds N] [--merge-strategy STRAT] [--allowed-strategies s1,s2,...]`

**Backends:** Server (`GET /rest/api/1.0/projects/{ns}/repos/{slug}/settings/pull-requests`, `POST` same path with merge config body). Cloud (`GET /repositories/{ws}/{slug}` for read, `PUT /repositories/{ws}/{slug}` for the subset Cloud supports — `fork_policy` + the project-level merge defaults; fields not honoured by Cloud return `ErrUnsupportedOnHost` per-field rather than failing the whole call).

**MCP tools shipped:** `get_repo_pr_settings`, `set_repo_pr_settings`

**Definition of Done (verified):**
- [x] `api/backend/client_repo_pr_settings.go` — interface + `RepoPRSettings` + `RepoPRSettingsInput`
- [x] `api/cloud/repo_pr_settings.go` + `_test.go` (subset support + typed errors on Cloud-unsupported fields)
- [x] `api/server/repo_pr_settings.go` + `_test.go`
- [x] `pkg/cmd/repo/pr-settings/` (get, set) + tests
- [x] MCP pair (`tools.go` + `handlers.go` + `handlers_test.go`)
- [x] `test/script/testdata/repo_pr_settings_*.txtar` for both backends
- [x] `skills/SKILL.md` + `skills/references/repos.md`

### PIPE-RERUN — Pipeline rerun

- **Feat commit:** `cd9cc7c feat(pipeline): add pipeline rerun command (Cloud)` (2026-05-25)
- **Backend:** Cloud only
- **Estimate (planned):** 1 day

Re-trigger a pipeline at the same commit. Cloud-only because Server/DC has no equivalent rerun primitive. Command shipped: `pipeline rerun UUID`. Functionality map row at `docs/backlog/BACKLOG.md` cross-references this scope.

### CHERRY-PICK — Cherry-pick commit onto a branch

- **Feat commit:** `979db8a feat(commit): add commit cherry-pick command (Server/DC)` (2026-05-25)
- **Release tag:** `v1.108.0`
- **Backend:** Server/DC only (branch-utils plugin)
- **Estimate (planned):** 1 day

Cherry-pick an arbitrary commit onto a named branch. Server/DC only — Cloud has no equivalent endpoint. Command shipped: `commit cherry-pick HASH BRANCH`. Functionality map row at `docs/backlog/BACKLOG.md` cross-references this scope.

---

## How to add an entry (for /auto-iter and humans)

When a scope's `feat:` commit lands on `main`:

1. **Same commit** that ships the code: cut its Up-Next table row and its `### SCOPE — Title` detail section from `docs/backlog/BACKLOG.md`.
2. Prepend a new entry to SHIPPED.md under a dated subsection (`## YYYY-MM-DD — <SCOPE>`). Include:
   - `Feat commit:` short SHA + subject + date
   - `Release tag:` if assigned by release-please at write time (else add later when known)
   - `Backend:` Cloud / Server-DC / Both
   - `Estimate (planned):` what was budgeted vs reality (optional callout)
   - The scope's detail section verbatim, with all checkboxes in the Definition of Done flipped to `[x]`.
3. Leave the Functionality Map `✅` markers in `BACKLOG.md` in place — they're cross-references, not status duplicates.
4. **Never** rewrite history in SHIPPED.md. If a shipped scope later breaks, file a fresh BACKLOG scope to fix it; do not edit the SHIPPED entry.
