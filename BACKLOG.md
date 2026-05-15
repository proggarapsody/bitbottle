# bitbottle Backlog

## Philosophy

Follow [GitHub CLI](https://github.com/cli/cli) conventions throughout:

- **Noun-verb commands** — `bitbottle tag list`, `bitbottle tag create NAME`
- **Consistent flags** — `--limit`, `--json`, `--jq`, `--web`, `--hostname` on every applicable command
- **TTY-aware output** — aligned table on TTY; tab-separated, no header on pipes
- **Thin commands** — parse flags → call backend interface → format output; zero business logic in cmd layer
- **Interface segregation** — each capability is its own interface; composite `Client` embeds only what both backends implement; Cloud-only ops use the optional-interface pattern (type assertion, like `PipelineClient`)

## Architecture Contract (per scope)

Every scope follows the same layer pattern. No exceptions.

```
api/backend/client.go    → new capability interface(s)
api/backend/types.go     → new domain type(s)
api/cloud/<domain>.go    → Cloud implementation + _test.go
api/server/<domain>.go   → Server/DC implementation + _test.go  (skip if Cloud-only)
pkg/cmd/<domain>/        → cobra commands
pkg/cmd/mcp/tools.go     → new MCP tool registrations
pkg/cmd/mcp/handlers.go  → new MCP handler methods
README.md                → new command section
```

### Definition of Done (every scope)

- [ ] `api/backend/client.go` — new interface(s) + composite `Client` updated (or optional-interface pattern documented)
- [ ] `api/backend/types.go` — new domain type(s)
- [ ] `api/cloud/<domain>.go` — Cloud impl + unit tests
- [ ] `api/server/<domain>.go` — Server impl + unit tests (skip for Cloud-only)
- [ ] `pkg/cmd/<domain>/` — commands with `--json`, `--jq`, `--hostname`; unit + integration tests
- [ ] `pkg/cmd/mcp/` — tool registrations + handler methods + tests
- [ ] README — new section for commands
- [ ] `go test ./... -race` green

---

## Full Functionality Map

Current state of every command area against gh feature parity:

### Auth

| Command | Status | Notes |
|---|---|---|
| `auth login` | ✅ | |
| `auth logout` | ✅ | |
| `auth status` | ✅ | |
| `auth token` | ✅ | Print raw stored token (gh has this) |
| `auth refresh` | ✅ | Re-validate token + update stored user |
| `auth migrate` | ✅ | Move config-file tokens into the keyring; strip from `hosts.yml`. Non-interactive (CI-safe). — scope **SEC** |

### Repo

| Command | Status | Notes |
|---|---|---|
| `repo list` | ✅ | |
| `repo view` | ✅ | |
| `repo create` | ✅ | |
| `repo delete` | ✅ | |
| `repo clone` | ✅ | |
| `repo fork` | ✅ | Cloud only — required `--into WORKSPACE`, optional `--name`; supports `--json` / `--jq`; typed unsupported error on Server |
| `repo rename` | ✅ | Both backends; `--confirm` required on non-TTY (slug change breaks clones' `origin` URL on Cloud); supports `--json` / `--jq` |
| `repo archive` | n/a | No Bitbucket primitive (Cloud nor Server) — out of scope |
| `repo set-default` | ✅ | Writes `bitbottle.host`/`project`/`slug` to local git config; consulted by `f.BaseRepo()` |

### Pull Requests

| Command | Status | Notes |
|---|---|---|
| `pr list` | ✅ | |
| `pr view` | ✅ | |
| `pr create` | ✅ | |
| `pr merge` | ✅ | |
| `pr approve` | ✅ | |
| `pr diff` | ✅ | |
| `pr checkout` | ✅ | |
| `pr edit` | ✅ | Update title / description |
| `pr unapprove` | ✅ | Remove own approval |
| `pr decline` | ✅ | Close/decline a PR |
| `pr ready` | ✅ | Promote draft → open |
| `pr request-review` | ✅ | Add reviewers to an open PR |
| `pr request-changes` | ✅ | Cloud only |
| `pr comment list` | ✅ | List comments; surfaces inline (file:line) anchors, replies (`parentId`), `updatedAt`, `resolved` |
| `pr comment add` | ✅ | Add a general comment (inline writes — RV3) |
| `pr comment list --inline` | ✅ | Filter to only inline (file:line) review comments — RV2 of scope **RV** |
| `pr comment add --inline path:line` | ✅ | Post inline review comments. Cloud uses `inline.{path,from,to,start_*}`; Server fetches `fromHash`/`toHash` from `/pull-requests/{id}/diff/{path}` and posts `anchor.{...}`. Multi-line ranges Cloud-only. — RV3 of scope **RV** |
| `pr comment edit / delete` | ✅ | `PUT/DELETE .../comments/{id}`. Server fetches comment `version` on demand for optimistic-locking. — RV3 of scope **RV** |
| `pr comment reply / resolve` | ✅ | Reply via `--parent COMMENT_ID` on `pr comment add`. `pr comment resolve` writes Cloud's `resolution.type=resolved`; Server returns typed `host.unsupported` (resolution lives on tasks, separate scope). — RV3 of scope **RV** |
| `pr review --approve\|--request-changes\|--comment --body --inline ...` | ✅ | Compound review in one call (gh parity). Cloud sequences body → inline → action; Server mirrors but returns typed `host.unsupported` for `--request-changes`. — RV4 of scope **RV** |
| `pr activity PR_ID` | ✅ | PR event stream (`/pullrequests/{id}/activity`) — scope **RV** |
| `pr checks PR_ID` | ✅ | CI status for PR head commit — scope **GHP** |
| `pr update-branch PR_ID` | ✅ | Sync PR head with base (the action our `pr.merge.behind` hint promises) — scope **GHP** |
| `pr reopen PR_ID` | ✅ | Reverse `pr decline` (Bitbucket Server / DC only — Cloud has no reopen primitive, BCLOUD-23807) |
| `pr status` | ✅ | Cross-repo "PRs on my plate" (assigned / review-requested / mine) — scope **GHP** |

### Branch

| Command | Status | Notes |
|---|---|---|
| `branch list` | ✅ | |
| `branch delete` | ✅ | |
| `branch create` | ✅ | |
| `branch checkout` | ✅ | Thin wrapper: `git fetch origin BRANCH && git checkout BRANCH` |
| `branch protect` | ✅ | Branch restrictions; Server/DC only (`list`, `create`, `delete`) |

### Pipeline _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `pipeline list` | ✅ | |
| `pipeline view` | ✅ | |
| `pipeline run` | ✅ | |
| `pipeline steps` | ✅ | List steps in a pipeline |
| `pipeline logs` | ✅ | Stream step log |
| `pipeline watch UUID` | ✅ | Poll until terminal state, stream step transitions — scope **GHP** |

### Commits

| Command | Status | Notes |
|---|---|---|
| `commit log` | ✅ | List commits on a branch |
| `commit view` | ✅ | View a single commit |
| `commit status` | ✅ | List build statuses for a commit hash |
| `commit comment list / add / edit / delete` | ✅ | Cloud + Server/DC `/commit/{hash}/comments` — scope **RV6** |

### Tags

| Command | Status | Notes |
|---|---|---|
| `tag list` | ✅ | |
| `tag create` | ✅ | |
| `tag delete` | ✅ | |

### Webhooks

| Command | Status | Notes |
|---|---|---|
| `webhook list` | ✅ | Both backends |
| `webhook view` | ✅ | Both backends |
| `webhook create` | ✅ | `--url`, `--events` required; `--secret`, `--active`, `--secret=-` for stdin, `--secret=@PATH` for file |
| `webhook delete` | ✅ | `--confirm` required when not interactive |

### Config

| Command | Status | Notes |
|---|---|---|
| `config list` | ✅ | Lists every set key (globals, then per-host) |
| `config get KEY` | ✅ | Supports `--host` for per-host lookup |
| `config set KEY VALUE` | ✅ | Allowlisted keys: editor, pager, browser, git_protocol, prompt |

### Aliases

| Command | Status | Notes |
|---|---|---|
| `alias set NAME EXPANSION` | ✅ | Command alias; `!` prefix → shell alias with $1..$9 / $@ |
| `alias list` | ✅ | |
| `alias delete NAME` | ✅ | |
| Root expansion | ✅ | `cmd/bitbottle/main.go` resolves before cobra parsing |

### API Passthrough

| Command | Status | Notes |
|---|---|---|
| `api PATH` | ✅ | `-X/--method`, `-H/--header`, `-F/--field`, `-f/--raw-field`, `--input`, `--jq`, `--paginate` (Cloud `next` + Server `nextPageStart`), `{workspace}/{repo_slug}/{project}/{slug}` expansion |

### Output / DX

| Feature | Status | Notes |
|---|---|---|
| `--json` / `--jq` | ✅ | Implemented on all list + view commands |
| `$PAGER` support | ✅ | `IOStreams.StartPager`/`StopPager` + `cmdutil.PagerAnnotation`; opt-in on `commit log`, `commit view`, `pr diff`, `pr view`, `repo view`, `pipeline logs` |
| Color output | ✅ | State columns colourised in formatters; `--no-color` global flag + `NO_COLOR` env honoured |

### Workspace / Projects _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `workspace list` | ✅ | Cloud only; surfaces `host.unsupported` on Server via `AsWorkspaceClient` |
| `project list WORKSPACE` | ✅ | Cloud only |

### Issues _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `issue list` | ✅ | Cloud only; `--state`, `--limit`, `--json`, `--jq` |
| `issue view` | ✅ | Cloud only |
| `issue create` | ✅ | Cloud only; `--title`, `--body`, `--kind`, `--priority` |
| `issue close` | ✅ | Cloud only |
| `issue edit` | ✅ | `PUT /issues/{id}` — `--title`, `--body`, `--kind`, `--priority`, `--state`, `--assignee` |
| `issue reopen` | ✅ | Reverse of `issue close` (state transition back to open) |
| `issue assign` | ✅ | Set assignee (`--assignee` or positional USER arg) |
| `issue comment list / add / edit / delete` | ✅ | `/issues/{id}/comments` CRUD with `--json`/`--jq` on list |

### Source / Files at ref _(missing)_

| Command | Status | Notes |
|---|---|---|
| `repo file get PATH --ref REF` | ✅ | Read file content at any ref. Cloud: `GET /repositories/{ws}/{slug}/src/{commit}/{path}`. Server: `GET /projects/{k}/repos/{s}/raw/{path}?at=REF`. Binary-safe; `--out FILE` for download. (RV1 of scope **RV**.) |
| `repo tree --ref REF [--path P]` | ✅ | List files at a ref. Cloud: same `/src/` endpoint returns directory metadata when path is a dir. Server: `GET /projects/{k}/repos/{s}/browse/{path}?at=REF`. Type normalised to `file`/`dir` across both backends. (RV1 of scope **RV**.) |

### Search

| Command | Status | Notes |
|---|---|---|
| `bitbottle search code QUERY [--workspace W]` | ✅ | Cloud `GET /workspaces/{ws}/search/code?search_query=...` — content + path matches. — scope **SR** |
| `bitbottle search code QUERY` _(Server)_ | 🔲 | Optional. Server has no first-class code-search REST API; defer. |

### Deployments _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `deployment list` | ✅ | `GET /repositories/{ws}/{slug}/deployments` — scope **DEP** |
| `deployment view UUID` | ✅ | `GET .../deployments/{uuid}` — scope **DEP** |
| `environment list` | ✅ | `GET .../environments` — scope **DEP** |
| `environment create` | ✅ | `POST .../environments` (body: type/uuid/name) — scope **DEP** |
| `environment delete UUID` | ✅ | `DELETE .../environments/{uuid}` — scope **DEP** |
| `environment variable list / set / delete` | ✅ | `/deployments_config/environments/{uuid}/variables` (CRUD) — scope **DEP** |

### Permissions _(Server / DC only — missing)_

| Command | Status | Notes |
|---|---|---|
| `perms project list PROJECT` | ✅ | List all permission grants (users + groups) for a project — scope **PERMS** |
| `perms project grant PROJECT [--user SLUG \| --group NAME] PERM` | ✅ | Grant `PROJECT_READ`/`WRITE`/`ADMIN` — scope **PERMS** |
| `perms project revoke PROJECT [--user SLUG \| --group NAME]` | ✅ | Revoke a project permission grant — scope **PERMS** |
| `perms repo list PROJECT/REPO` | ✅ | List all permission grants (users + groups) for a repo — scope **PERMS** |
| `perms repo grant PROJECT/REPO [--user SLUG \| --group NAME] PERM` | ✅ | Grant `REPO_READ`/`WRITE`/`ADMIN` — scope **PERMS** |
| `perms repo revoke PROJECT/REPO [--user SLUG \| --group NAME]` | ✅ | Revoke a repo permission grant — scope **PERMS** |

### Admin _(Server / DC only — missing)_

| Command | Status | Notes |
|---|---|---|
| `admin secrets rotate` | ✅ | Rotate application secrets (HTTP Strict Transport, etc.) — scope **ADMIN** |
| `admin logging get` | ✅ | Show current log level + async flag — scope **ADMIN** |
| `admin logging set` | ✅ | `--level DEBUG\|INFO\|WARN\|ERROR`, `--async` — scope **ADMIN** |

### PR Auto-Merge _(both backends — missing)_

| Command | Status | Notes |
|---|---|---|
| `pr merge --auto PR_ID` | ✅ | Queue PR for auto-merge once all checks pass. DC: stable API. Cloud: beta endpoint, gated. — scope **AUTOMERGE** |
| `pr merge --auto-off PR_ID` | ✅ | Cancel a queued auto-merge — scope **AUTOMERGE** |
| `pr view PR_ID` | ✅ (extended) | Command shows `Auto-merge: enabled (squash)` line in its output — scope **AUTOMERGE** |

### PR Tasks _(Server / DC only — missing)_

| Command | Status | Notes |
|---|---|---|
| `pr task list PR_ID` | ✅ | List blocker-comments on a PR (modern API: comments with `severity=BLOCKER`) — scope **TASK** |
| `pr task create PR_ID` | ✅ | Post a blocker-comment (anchors to a comment or stands alone) — scope **TASK** |
| `pr task resolve PR_ID TASK_ID` | ✅ | Set comment `state=RESOLVED` — scope **TASK** |
| `pr task reopen PR_ID TASK_ID` | ✅ | Set comment `state=OPEN` — scope **TASK** |

### Comment Reactions _(Server / DC only — missing)_

| Command | Status | Notes |
|---|---|---|
| `pr comment react PR_ID COMMENT_ID --emoji E` | ✅ | Add an emoji reaction to a PR comment — scope **REACT-PR** |
| `pr comment unreact PR_ID COMMENT_ID --emoji E` | ✅ | Remove own reaction from a comment — scope **REACT-PR** |
| `pr comment list --reactions` | ✅ | Existing command grows a reactions column when flag is set — scope **REACT-PR** |
| `commit comment react / unreact` | ✅ | Same pattern for commit comments — scope **REACT-COMMIT** |

### Variable _(standalone promotion — missing)_

| Command | Status | Notes |
|---|---|---|
| `variable list PROJECT/REPO` | ✅ | List variables; `--scope repository\|workspace\|deployment` — scope **VAR** |
| `variable set PROJECT/REPO KEY VALUE` | ✅ | Upsert; `--secured`, `--scope` — scope **VAR** |
| `variable delete PROJECT/REPO KEY` | ✅ | Delete; `--scope` — scope **VAR** |

### Extensions _(missing)_

| Command | Status | Notes |
|---|---|---|
| `extension install REPO` | ✅ | Install a third-party bitbottle extension from a GitHub/Bitbucket repo — scope **EXT-CORE** |
| `extension install --local PATH` | ✅ | Symlink a local directory as an extension (for extension authors) — scope **EXT-CORE** |
| `extension list` | ✅ | List installed extensions — scope **EXT-CORE** |
| `extension upgrade [NAME\|--all]` | ✅ | Check each installed extension for a new release and upgrade — scope **EXT-MGMT** |
| `extension remove NAME` | ✅ | Remove an installed extension — scope **EXT-MGMT** |
| `extension exec NAME [args...]` | ✅ | Run an installed extension (BITBOTTLE_KEYRING_PASSPHRASE stripped, BITBOTTLE_TOKEN injected fresh) — scope **EXT-RUNTIME** |

### Named Profiles

| Command | Status | Notes |
|---|---|---|
| `profile create NAME --hostname HOST --token TOKEN` | ✅ | Create a named credential profile (kubectl-context-like) — scope **PROF** |
| `profile use NAME` | ✅ | Switch the active profile — scope **PROF** |
| `profile list` | ✅ | List all defined profiles — scope **PROF** |
| `profile delete NAME` | ✅ | Delete a profile — scope **PROF** |

### Code Insights _(Server / DC only)_

| Command | Status | Notes |
|---|---|---|
| `code-insights report list / view / set / delete` | ✅ | `/rest/insights/1.0/projects/{k}/repos/{s}/commits/{hash}/reports/{key}` — PASS/FAIL summary attached to commit/PR — scope **CI** |
| `code-insights annotation add / list / delete` | ✅ | Bulk-attach line-level findings to a report — scope **CI** |
| `code-insights merge-check set / get / delete` | ✅ | Configure required reports as merge gates (`/rest/insights/latest/.../merge-check/{key}`; partly undocumented; marked experimental) — scope **CI** |

### Context / Orientation _(missing)_

| Command | Status | Notes |
|---|---|---|
| `bitbottle context [--json]` | ✅ | One-call orientation: host + repo + branch + user + scopes + default-branch + ahead/behind. Replaces 3-4 calls for agents. — scope **CTX** |
| `bitbottle status` | ✅ | Cross-repo "what's on my plate": review requests, mentions, assigned issues. (gh `status` analogue, workspace-scoped on Cloud.) — scope **GHP** |

### Top-level / Web

| Command | Status | Notes |
|---|---|---|
| `bitbottle browse [PATH\|NUMBER]` | ✅ | Unified web shortcut (commit/PR/issue/path) — scope **GHP** |

---

## Backlog

| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| L | **Branch Create + Checkout** | `branch create`, `branch checkout` | Both | 1 | ✅ |
| E | **Tags** | `tag list`, `tag create`, `tag delete` | Both | 1 | ✅ |
| G | **PR Lifecycle** | `pr decline`, `pr unapprove`, `pr edit`, `pr ready`, `pr request-review`, `pr request-changes` | Both / Cloud | 1 | ✅ |
| M | **Shell Completion** | `completion bash\|zsh\|fish\|powershell` | N/A | DX | ✅ |
| P | **Auth Extras** | `auth token`, `auth refresh` | N/A | DX | ✅ |
| Q | **Repo Extras** | `repo rename`, `repo fork`, `repo set-default` _(`archive` dropped — no Bitbucket primitive)_ | Both / Cloud | 2 | ✅ |
| F | **Commits** | `commit log`, `commit view` | Both | 1 | ✅ |
| H | **Pipeline Depth** | `pipeline steps`, `pipeline logs` | Cloud | 1 | ✅ |
| I | **Webhooks** | `webhook list`, `webhook view`, `webhook create`, `webhook delete` | Both | 2 | ✅ |
| J | **PR Comments** | `pr comment list`, `pr comment add` | Both | 2 | ✅ |
| K | **Commit Statuses** | `commit status` | Both | 2 | ✅ |
| T | **Output DX** | pager (`$PAGER`), color output | N/A | DX | ✅ |
| U | **Config** | `config list`, `config get`, `config set` | N/A | 2 | ✅ |
| V | **API Passthrough** | `api PATH` | Both | 2 | ✅ |
| N | **Workspace / Projects** | `workspace list`, `project list` | Cloud | 3 | ✅ |
| O | **Issues** | `issue list`, `issue view`, `issue create`, `issue close` | Cloud | 3 | ✅ |
| BP | **Branch Protect** | `branch protect list`, `branch protect create`, `branch protect delete` | Server/DC | 2 | ✅ |
| EX | **Error UX** | Centralised, human-readable errors with actionable hints across every command | N/A | DX | ✅ |
| RV | **Code-Review Primitives** | `repo file get`, `repo tree`, `pr review`, `pr comment {add\|edit\|delete\|reply\|resolve} --inline`, `pr comment list --inline`, `pr activity`, `commit comment *` | Both | 1 | ✅ RV1+RV2+RV3+RV4+RV5+RV6 ✅ |
| SR | **Code Search** | `search code QUERY [--workspace W]` | Cloud | 2 | ✅ |
| CTX | **Context Primitive** | `context --json` (one-call orientation: host + repo + branch + user + scopes + default-branch + ahead/behind) | N/A | DX | ✅ |
| GHP | **gh-Parity Gaps** | `pr checks`, `pr update-branch`, `pr reopen`, `pr status`, `status`, `browse`, `pipeline watch` | Both | 2 | ✅ |
| OF | **Issues Finish** | `issue edit`, `issue reopen`, `issue assign`, `issue comment {list\|add\|edit\|delete}` | Cloud | 3 | ✅ |
| CI | **Code Insights** | `code-insights report *`, `code-insights annotation *`, `code-insights merge-check *` | Server/DC | 2 | ✅ |
| DEP | **Deployments** | `deployment list/view`, `environment list/create/delete`, `environment variable {list\|set\|delete}` | Cloud | 3 | ✅ |
| SEC | **Secret Store & Config Security** | (infrastructure) + `auth migrate` | N/A | DX | ✅ |
| HTTPH | **HTTP Client Hardening** | (infrastructure) | N/A | DX | ✅ |
| OUT2 | **Extended Output Formats** | `--yaml` / `--template` global flags + validation | N/A | DX | ✅ |
| CIS | **CI Supply Chain Hardening** | (GitHub Actions) | N/A | DX | ✅ |
| VAR | **Variable Command Promotion** | `variable list/set/delete --scope repository\|workspace\|deployment` | Cloud | 2 | ✅ |
| PERMS | **Permissions Management** | `perms project list/grant/revoke`, `perms repo list/grant/revoke` | Server/DC | 3 | ✅ |
| ADMIN | **Admin Commands** | `admin secrets rotate`, `admin logging get/set` | Server/DC | 3 | ✅ |
| AUTOMERGE | **PR Auto-Merge** | `pr merge --auto[=off]` flag + `pr view` extension | Both (Cloud beta) | 2 | ✅ |
| TASK | **PR Tasks** | `pr task list/create/resolve/reopen` (Server severity-BLOCKER comments) | Server/DC | 3 | ✅ |
| REACT-PR | **PR Comment Reactions** | `pr comment react/unreact`, `pr comment list --reactions`; `CommentReactor` interface + Server impl + Cloud stub | Server/DC | 3 | ✅ |
| REACT-COMMIT | **Commit Comment Reactions** | `commit comment react/unreact`, `commit comment list --reactions`; `CommitCommentReactor` interface + Server impl + MCP tools | Server/DC | 3 | ✅ |
| PROF | **Named Profiles** | `profile create/use/list/delete` | N/A | 3 | ✅ |
| EXT-CORE | **Extension Install + List** | `extension install USER/REPO`, `extension install --local PATH`, `extension list`; core package + SHA lockfile | N/A | 4 | ✅ |
| EXT-RUNTIME | **Extension Exec** | `extension exec NAME [args...]`; SHA verification, env sanitise/inject, root-command dispatch hook | N/A | 4 | ✅ |
| EXT-MGMT | **Extension Upgrade + Remove** | `extension upgrade [NAME\|--all]`, `extension remove NAME` | N/A | 4 | ✅ |
| VAROPS | **Variable scope-ops strategy** | Collapse the three near-identical `As<X>Client` scope switches in `pkg/cmd/variable/{list,set,delete}/*.go` (and the MCP handler) into one `resolveVariableOps(scope)` helper returning a `VariableOps` interface. Pre-empts OCP debt before the 4th scope lands. Move deployment delete-by-key lookup from cmd/MCP into the cloud adapter. Cite v1.31.0 design-judge findings. | N/A | DX | ✅ |
| CMDTEST | **Shared cmdtest helper** | Already done — `pkg/cmd/internal/cmdtest` exists and is used widely. | N/A | DX | ✅ |
| ENVVAR-DEPREC | **Deprecate `environment variable` tree** | `environment variable {list,set,delete}` (shipped v1.29.0) is structurally a subset of `variable --scope deployment` (shipped v1.31.0). Mark the old tree as deprecated in `pkg/cmd/environment/variable/` with a `Deprecated:` cobra field, route to the new commands, and document the migration in README + skills/SKILL.md. Remove after one minor release. | N/A | DX | ✅ |
| PIPEVAR-DEPREC | **Remove deprecated `pipeline variable` tree** | `pipeline variable {list,set,delete}` was marked deprecated in favour of `variable {list,set,delete} --scope repository` (scope **VAR**, shipped v1.31.0). The commands carry Cobra `Deprecated:` notices and route through `variable/shared.ResolveVariableOps`. Remove `pkg/cmd/pipeline/variable/` entirely after one minor release. | N/A | DX | ✅ |
| SRVVER | **Server version detection helper** | `api/server/version.go` parsing `ServerCapabilities.GetApplicationProperties()` into a `semver.Version` + `(v Version) AtLeast(major, minor int) bool` helper. Cached per-host for the process lifetime. Required by **TASK** (comments-with-severity dispatch for Server >= 7.2) and any future Server-version-conditional behaviour. Ship as part of TASK's first PR or as a standalone precursor. | Server/DC | DX | ✅ |
| DEPLOY-KEY | **Deploy Key Management** | `deploy-key list PROJECT/REPO`, `deploy-key add PROJECT/REPO --key "..." [--label "..."]`, `deploy-key delete PROJECT/REPO ID`; list with `--json`/`--jq`. Server: `GET/POST/DELETE /rest/api/1.0/projects/{ns}/repos/{slug}/ssh`. Cloud: `GET/POST/DELETE /repositories/{ws}/{slug}/deploy-keys`. Both backends. | Both | 2 | ✅ |
| PIPE-TRIGGER | **Pipeline Trigger** | `pipeline trigger [PROJECT/REPO] [--branch BRANCH] [--variables key=val,...]` — manually trigger a Bitbucket Cloud pipeline. `POST /repositories/{ws}/{slug}/pipelines/`. Cloud only; typed `UnsupportedOnHost` on Server. Useful for CI/CD automation scripts. | Cloud | 2 | ✅ |
| DIFF | **Diff Between Refs** | `diff REF1..REF2 [--stat]` — compare two refs (branches, tags, commits); outputs unified diff or `--stat` summary. Cloud: `GET /repositories/{ws}/{slug}/diff/{spec}`. Server: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/diff/{path}?since=REF1&until=REF2`. Both backends. | Both | 2 | ✅ |
| PR-TEMPLATE | **PR Description Templates** | File-based templates (`.bitbucket/pull-request-template.md`) — no dedicated REST API; already readable via `repo file get`. Deferred: not a distinct API scope. | N/A | — | ✅ |
| DEFAULT-REVIEWERS | **PR Default Reviewers** | `pr default-reviewer list`, `pr default-reviewer add USER`, `pr default-reviewer remove USER` — manage per-repo default reviewers. Cloud: `GET/POST/DELETE /repositories/{ws}/{slug}/effective-default-reviewers`. Server: `GET/PUT/DELETE /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/reviewers/{userSlug}`. Both backends. | Both | 2 | ✅ |
| SSH-KEYS | **User SSH Key Management** | `ssh-key list`, `ssh-key add --key "..." [--label "..."]`, `ssh-key delete ID` — user-level SSH keys (not repo deploy keys). Cloud: `GET/POST/DELETE /users/{username}/ssh-keys`. Cloud only initially. | Cloud | 2 | ✅ |
| REPO-TRANSFER | **Repository Transfer** | `repo transfer PROJECT/REPO --to TARGET-PROJECT [--hostname H]` — move a repository to a different project/workspace. Cloud: `POST /repositories/{ws}/{slug}/transfer`. Server: `PUT /rest/api/1.0/projects/{ns}/repos/{slug}` updating `project.key`. Both backends. | Both | 3 | ✅ |
| BRANCH-RULE | **Cloud Branch Restriction Rules** | `branch-rule list [PROJECT/REPO]`, `branch-rule add [PROJECT/REPO] --kind KIND --pattern PATTERN`, `branch-rule delete [PROJECT/REPO] ID` — manage Cloud branch restrictions (require PR, prevent force-push, require approvals). Cloud: `GET/POST/DELETE /repositories/{ws}/{slug}/branch-restrictions`. Cloud only (Server has `branch protect` ✅). | Cloud | 2 | ✅ |
| PIPELINE-SCHEDULE | **Pipeline Schedules** | `pipeline schedule list [PROJECT/REPO]`, `pipeline schedule create [PROJECT/REPO] --cron EXPR --branch BRANCH [--enabled]`, `pipeline schedule delete [PROJECT/REPO] ID` — manage scheduled pipeline triggers. Cloud: `GET/POST/DELETE /repositories/{ws}/{slug}/pipelines_config/schedules`. Cloud only. | Cloud | 2 | ✅ |
| COMMIT-FILE | **Files Changed in a Commit** | `commit files HASH [PROJECT/REPO]` — list files added/modified/deleted in a specific commit. Cloud: `GET /repositories/{ws}/{slug}/diffstat/{node}~1..{node}`. Server: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/commits/{commitId}/changes`. Both backends. Useful for agents inspecting specific commits. | Both | 2 | ✅ |
| PR-COMMITS | **PR Commit List** | `pr commits PR_ID [PROJECT/REPO]` — list commits included in a pull request. Cloud: `GET /repositories/{ws}/{slug}/pullrequests/{id}/commits` (paginated). Server: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{id}/commits` (paginated). Both backends. Useful for automation and code review bots. | Both | 2 | ✅ |
| PR-FILES | **PR Changed Files** | `pr files PR_ID [PROJECT/REPO]` — list files changed in a pull request. Cloud: `GET /repositories/{ws}/{slug}/pullrequests/{id}/diffstat` (paginated). Server: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{id}/changes` (paginated). Reuses `DiffStatEntry` domain type. Both backends. | Both | 2 | ✅ |
| REPO-WATCHER | **Repository Watchers** | `repo watcher list [PROJECT/REPO]` — list users watching a repository. Cloud: `GET /repositories/{ws}/{slug}/watchers` (paginated). Server: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/watchers` (paginated). Both backends. | Both | 2 | ✅ |
| COMMIT-STATUS-REPORT | **Report Commit Build Status** | `commit status report HASH --key KEY --state PASSED\|FAILED\|INPROGRESS [--url URL] [--name NAME] [--description DESC]` — post a build status against a commit hash (the write side of the existing `commit status list`). Cloud: `POST /repositories/{ws}/{slug}/commit/{hash}/statuses/build`. Server: `POST /rest/build-status/1.0/commits/{hash}` (uses existing `buildStatusHTTP` transport). Both backends. | Both | 2 | ✅ |
| PR-PARTICIPANTS | **PR Participants** | `pr participant list PR_ID [PROJECT/REPO]` — list all participants in a PR (author, reviewers, observers) with their role and approval status. Cloud: `GET /repositories/{ws}/{slug}/pullrequests/{id}/participants` (paginated). Server: `GET /rest/api/1.0/projects/{k}/repos/{s}/pull-requests/{id}/participants` (paginated). Both backends. | Both | 2 | ✅ |
| WORKSPACE-MEMBERS | **Workspace Members** | `workspace member list [WORKSPACE]` — list members of a Cloud workspace. Cloud: `GET /workspaces/{ws}/members` (paginated). Cloud only (Server has no workspace concept). | Cloud | 2 | ✅ |
| USER-VIEW | **User Profile** | `user view [USERNAME]` — view profile for current user (no arg) or by slug. Cloud: `GET /user` (current) or `GET /users/{account_id}`. Server: `GET /rest/api/1.0/users/{userSlug}`. Both backends. Agent-useful orientation primitive. | Both | 1 | ✅ |
| REPO-FORKS | **Repository Forks** | `repo fork list [PROJECT/REPO]` — list forks of a repository. Cloud: `GET /repositories/{ws}/{slug}/forks` (paginated). Server: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/forks` (paginated). Both backends. | Both | 2 | ✅ |
| REPO-VISIBILITY | **Repository Visibility** | `repo visibility [PROJECT/REPO] [public\|private]` — get or toggle repository visibility. Cloud: `PUT /repositories/{ws}/{slug}` `{"is_private": bool}`. Server: `PUT /rest/api/1.0/projects/{ns}/repos/{slug}` `{"public": bool}`. Both backends. | Both | 2 | ✅ |
| WORKSPACE-HOOKS | **Workspace Webhooks** | `workspace hook list WORKSPACE`, `workspace hook create WORKSPACE --url URL --events E1,E2`, `workspace hook delete WORKSPACE ID` — workspace-level webhooks (distinct from repo-level). Cloud: `GET/POST/DELETE /workspaces/{ws}/hooks`. Cloud only. | Cloud | 2 | ✅ |
| PIPELINE-CACHE | **Pipeline Cache Management** | `pipeline cache list [PROJECT/REPO]`, `pipeline cache delete [PROJECT/REPO] UUID` — list and delete Cloud pipeline caches. Cloud: `GET /repositories/{ws}/{slug}/pipelines_config/caches/`, `DELETE /repositories/{ws}/{slug}/pipelines_config/caches/{uuid}`. Cloud only. | Cloud | 2 | ✅ |
| PR-SUGGESTION | **PR Code Suggestions** | `pr suggestion apply PR_ID COMMENT_ID SUGGESTION_ID [--preview]` — apply a DC suggested-change block via the native endpoint `POST /rest/api/1.0/projects/{k}/repos/{s}/pull-requests/{id}/comments/{cid}/suggestions/{sid}/apply` (server commits to the PR source branch; no local file edits). `--preview` reads the suggestion body without applying. Cloud has no equivalent primitive — return typed `host.unsupported`. Gap vs `bkt pr suggestion` (verified single API call in `bbdc/suggestions.go:18-29`). | Server/DC | 3 | ✅ |
| PR-REVIEWER-GROUP | **PR Reviewer Groups** | `pr reviewer-group list [PROJECT/REPO]`, `pr reviewer-group add [PROJECT/REPO] --name NAME --users u1,u2`, `pr reviewer-group remove [PROJECT/REPO] NAME` — manage named reviewer groups (distinct from per-user default reviewers shipped in DEFAULT-REVIEWERS). Server: `/rest/default-reviewers/1.0/projects/{k}/repos/{s}/conditions` with `reviewers[]` groups. Cloud: project-level reviewer-groups API (gated). Both backends where API supports it. Gap vs `bkt pr reviewer-group`. | Both | 3 | ✅ |
| AUTH-DOCTOR | **Auth Diagnostics** | `auth doctor [--hostname H]` — diagnose credential/keychain issues without printing the secret. Reports: keyring backend (macOS/Secret Service/Win Credential Manager), whether a token is stored for the host, token format heuristic (BBDC- vs ATATT- vs app-password), reachability of API base URL, and `auth status` round-trip. Never echoes the token. Gap vs `bkt auth doctor` — high-leverage UX for the corporate DC user (TLS/keyring/proxy interplay). | N/A | DX | ✅ |
| ~~GOMOD-FIX~~ | **Pin `go.mod` to released Go version** | `go.mod:3` declares `go 1.25.5` (unreleased); pin to the latest released minor (e.g. `1.23.x`) to avoid GOTOOLCHAIN auto-download surprises on contributor and CI machines. One-line change + CI re-run. | N/A | DX | ❌ (mcp-go requires go 1.25.5) |
| HTTPX-CTX | **Thread context.Context through `httpx.Transport`** | `api/internal/httpx/httpx.go:227,341` (`newRequest`, `fetchBodyAt`) hardcode `context.Background()`. Result: Ctrl-C during a paginated list cannot cancel in-flight HTTP; MCP server (long-lived) cannot impose per-request deadlines. Add `ctx context.Context` to `Transport.GetJSON/PostJSON/PutJSON/DeleteJSON/GetAllJSON/Raw*`, thread `cmd.Context()` from Cobra commands and the MCP handler context through every adapter. Big diff, mechanical. Should land before any further long-running command (e.g. pipeline log tailing). | N/A | DX | 🔲 |
| ~~DEPGUARD-CI~~ | **CI-enforce the layer boundary (with composition-root exception)** | `docs/ARCHITECTURE.md` mandates that `pkg/cmd/**` must not import `api/cloud` or `api/server` (commands talk to `api/backend` interfaces only). Currently enforcement is documentation-only; `.golangci.yml` has no `depguard` stanza. Add a depguard rule that fails CI on the forbidden imports, with two principled exceptions encoded in the rule itself: (a) **composition root** — `pkg/cmd/factory/**` is the DI assembly point and must know concrete adapters by definition (canonical hex/clean architecture exemption); (b) **integration tests** — `pkg/cmd/**/*_integration_test.go` legitimately construct real adapter clients against `httptest` servers to verify wire compatibility. Document both exceptions in `docs/ARCHITECTURE.md` so the rule's intent stays clear. Pairs with the existing smell-scan as belt-and-suspenders. | N/A | DX | ✅ |
| REGISTRY-FINISH | **Finish `cmdregistry` migration** | `pkg/cmd/root/root.go:8–119` mixes a 30-line hardcoded `AddCommand` block (legacy) with `cmdregistry.All(f)` (new). Migrate the remaining ~17 hardcoded commands to self-registered `init()` blocks under their own package, leaving `root.go` as a tiny shell. Removes a chronic merge-conflict surface on `root.go`. | N/A | DX | 🔲 |
| TYPES-SPLIT | **Split `api/backend/types.go`** | Single 877-line file holds every domain type (PR, pipeline, code-insights, deployment, reactions, …). Split per feature (`types_pr.go`, `types_pipeline.go`, `types_codeinsights.go`, …) — pure mechanical refactor (the file is leaf-imported by everything, no behavioural change). Reduces merge contention as new scopes land. | N/A | DX | 🔲 |
| FAKECLIENT-SAFETY | **Compile-time safety for `FakeClient`** | `test/testhelpers/client.go` is a 150+-field struct that implements every optional capability of `backend.Client` via function fields; missing methods explode at runtime, not at compile time. Either (a) split into per-interface fakes assembled via composition, or (b) add a `var _ backend.PRClient = (*FakeClient)(nil)` block for every optional interface so the compiler enforces method-set completeness. Option (b) is the one-PR fix. | N/A | DX | 🔲 |
| ~~CONCURRENT-ERRORS~~ | **Surface errors from concurrent reaction fetchers** | `pkg/cmd/pr/comment.go:270-285` (and two near-clones in `pkg/cmd/mcp/handlers.go:1015,2052`) silently swallow worker errors: `if err == nil && len(rxns) > 0 { results[j.idx] = rxns }` — no else, no log, no aggregate. A 500 on one comment's reactions is indistinguishable from "no reactions exist". Fix: collect errors into an aggregate, return `errors.Join` (or count + first error) so the user knows N reactions failed to load. Also DRY the three copies into one helper in `pkg/cmd/internal/`. | N/A | DX | ✅ |
| PAGING-PARTIAL | **Surface partial results on mid-pagination failure** | `api/internal/paging/paging.go:32-56` already returns accumulated items alongside the transport error (correct), but callers like `pkg/cmd/pr/list.go:48-52` do `if err != nil { return err }` and discard them. Result: a 429 on page 3 of 5 produces zero output even though 60% was fetched. Fix: in command callers, render partial results to stderr/stdout with a final warning line, or add a `--partial-on-error` flag that opts in. Touches every `*list*.go` that uses `paging.Collect`. | N/A | DX | 🔲 |
| LIMIT-CLAMP | **Validate `--limit` in command layer** | `paging.Collect[T]` treats `cap <= 0` as unbounded — by design. But `pkg/cmd/pr/list.go:48,61` (and every other `*list*` command) passes the user `--limit` flag straight through, so `--limit 0` silently means "fetch the world". Fix: in `pkg/cmdutil`, add a `ValidateLimit(limit int) error` that rejects `< 1` with a typed error, and call it from every list command. Or clamp to a documented max (e.g. 1000). | N/A | DX | 🔲 |
| FEATURE-REGISTRY | **Central registry for optional `As<X>Client` interfaces** | 35 distinct optional-capability interfaces exist (`api/backend/client_*.go`) with no central list. A new backend implementation can quietly omit one — only signal is a runtime `ErrUnsupportedOnHost`. Add `api/backend/features.go` with `var AllFeatures = []Feature{...}` enumerated explicitly + a `capability_contract_test.go` that, for each known backend (cloud, server), asserts `client.SupportsFeature(f)` returns a stable bool for every feature in `AllFeatures`. Forces new features to declare support on both adapters at compile- or test-time. | N/A | DX | 🔲 |
| JSON-STABILITY | **Golden-file tests for `--json` field sets** | Since `OUT2` (v1.39+) every lister emits the full serialized struct; field names are json tags from `api/backend/types.go` and there is no test that fails when those tags change. Scripted consumers can break silently across releases. Fix: add `testdata/json/<command>.golden.json` for 10+ representative commands (pr list, pr view, pipeline list, commit log, …) and a test that diffs the actual `--json` output against the golden under a `cmdtest` harness. Field additions are fine (golden update); renames/removals fail the build and force a release-note. | N/A | DX | 🔲 |
| MCP-VALIDATION | **MCP input validation beyond nonempty check** | `pkg/cmd/mcp/handlers.go:124` (`requireString`) only rejects `""`. No enum validation (`state ∈ {open,closed,merged}`), no bounds check on numerics (`limit`), no path/URL shape check, no schema-level enforcement post-decode. Bad agent inputs reach the wire and produce server-side 400s with worse messaging than we could produce locally. Fix: add `validateEnum(field, value, allowed...)`, `validateRange(field, n, min, max)`, and use them at the top of every handler that takes constrained inputs. | N/A | DX | 🔲 |
| BACKEND-TYPE-STRICT | **Validate `backend_type` at write time** | `internal/bbinstance/bbinstance.go:25-34` `IsCloud`'s `default → false` is intentional (per docstring: "server", "datacenter" both legitimately map to Server). But typos like `backend_type: clud` silently become Server at read time, surfacing as confusing API path errors. Real fix is **input validation**: add `BackendType.IsValid()` covering `{cloud, server, datacenter, ""}`, and call it in the `auth login` / `profile create` flows so typos are rejected at config write. Don't change `IsCloud` read semantics. | N/A | DX | 🔲 |
| ERR-EMPTY-400 | **Stamp `Code` on bare 400/422 responses** | `api/internal/httpx/httpx.go:364-385` → `backend.ClassifyHTTPError` at `api/backend/errors.go:206-243` already maps 401/403/404/409/5xx to a `Kind`, but only 401 gets an auto-`Code` (`CodeAuthInvalidToken`). A 400 / 422 with an empty response body therefore has `Kind: ""`, `Code: ""` and falls through to the catch-all errfmt fallback — agents and scripts see "HTTP 400: Bad Request" with no actionable hint. Fix: add `Kind: ErrInvalidRequest` + `Code: CodeInvalidRequest` for 400/422 in `ClassifyHTTPError`, register a catalogue entry, and let adapters override via `StampCode` when they have a more specific reason. | N/A | DX | 🔲 |
| PR-UNREADY | **Promote PR back to draft** | `pr unready PR_ID` (alias `pr ready --undo`) — convert an open PR back to draft state. Cloud: `PUT /repositories/{ws}/{slug}/pullrequests/{id}` with `{"draft": true}`. Server: marks the PR as draft via `PUT /rest/api/1.0/.../pull-requests/{id}` body `{"draft": true}` (Server 8.0+). Companion to existing `pr ready`. Both backends, with version probe on Server. Gap vs `bkt pr publish --undo`. | Both | 3 | 🔲 |
| PR-EDIT-REVIEWER-RM | **Remove reviewers via `pr edit`** | `pr edit PR_ID --remove-reviewer USER` (repeatable) — drop reviewers from an open PR without rebuilding the full list. Today `pr request-review` only adds; removal requires raw API. Cloud: read current reviewers, `PUT /pullrequests/{id}` with filtered `reviewers` array. Server: `DELETE /rest/api/1.0/.../pull-requests/{id}/participants/{userSlug}` per removed user. Both backends. Gap vs `bkt pr edit --remove-reviewer`. | Both | 3 | 🔲 |
| CMD-SUBPKG | **Subpackage-per-verb command layout** | `pkg/cmd/pr/` is a flat package with **78 files** (`list.go`, `view.go`, `comment.go` at 15 KB, `comment_test.go` at 19 KB, …). gh's canonical layout is `pkg/cmd/pr/<verb>/<verb>.go` (gh's `pkg/cmd/pr/` has 17 subdirs, each its own package). The flat layout invites accidental coupling between sibling commands (private helpers leak across verbs), bloats per-command test files, and produces merge conflicts on package-level state. Migrate the largest cmd groups (`pr/`, `repo/`, `pipeline/`) to `<group>/<verb>/<verb>.go` with shared helpers in `<group>/shared/`. Mechanical refactor; one PR per group. Pairs with the `REGISTRY-FINISH` cmdregistry migration. | N/A | DX | 🔲 |
| JSON-WHITELIST | **Per-command `--json` field whitelist** | gh's `pkg/cmdutil/json_flags.go` exposes `AddJSONFlags(cmd, &out, fields)`: each command declares exactly which fields are exposed on `--json`, validates `--json a,b,c` against that list, and rejects unknown names. bitbottle's OUT2 ships **all** serialized struct fields from `api/backend/types.go` — every json tag is a de-facto public contract. Result: refactors of domain types silently break scripted consumers (already flagged in `JSON-STABILITY`). Fix: add `pkg/cmdutil/json_flags.go` with `AddJSONFlags`, declare per-command field lists, and have `--json` validate the requested subset. Pair `JSON-STABILITY` golden-file tests with this. | N/A | DX | 🔲 |
| DOMAIN-DTO-SPLIT | **Move output-DTOs out of `api/backend/types.go`** | The domain types package has **62 `json:` tags**, including output-shaped DTOs like `ContextSnapshot` (`api/backend/types.go:555-560`, fields `default_branch`/`ahead`/`behind` with snake_case JSON tags consumed only by `bitbottle context --json`). Hexagonal: domain types should be transport-free; JSON-tagged renderers belong outside the inward boundary. Move `Context*`, `*Status*`, and any output-only DTOs to `pkg/cmd/<feature>/output.go` (or a new `pkg/output/`). Core domain (`Repository`, `PullRequest`, `Commit`, …) stays tag-free as it largely already is. Pure mechanical refactor — no behaviour change. | N/A | DX | 🔲 |

---

## Scope Details

### L — Branch Create + Checkout

**New interfaces** (`api/backend/client.go`):
```go
type BranchCreator interface {
    CreateBranch(ns, slug string, in CreateBranchInput) (Branch, error)
}
```
Add `BranchCreator` to composite `Client`.
`branch checkout` requires no backend call — thin git wrapper only.

**New type** (`api/backend/types.go`):
```go
type CreateBranchInput struct {
    Name    string
    StartAt string // branch name or commit hash
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `branch create PROJECT/REPO NAME` | 2 | `--start-at` | `--hostname` |
| `branch checkout NAME` | 1 | — | (uses current repo from `.git/config` or `--hostname`) |

`branch checkout` fetches the branch from origin and checks it out locally — same pattern as `pr checkout`.

**MCP tools**: `create_branch`

---

### E — Tags

**New interfaces**:
```go
type TagLister  interface { ListTags(ns, slug string, limit int) ([]Tag, error) }
type TagCreator interface { CreateTag(ns, slug string, in CreateTagInput) (Tag, error) }
type TagDeleter interface { DeleteTag(ns, slug, name string) error }
```
All in composite `Client`.

**New types**:
```go
type Tag struct {
    Name   string
    Hash   string  // target commit hash
    WebURL string
}

type CreateTagInput struct {
    Name    string
    StartAt string // branch name or commit hash
    Message string // empty = lightweight; non-empty = annotated
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `tag list PROJECT/REPO` | 1 | — | `--limit`, `--json`, `--jq`, `--hostname` |
| `tag create PROJECT/REPO NAME` | 2 | `--start-at` | `--message`, `--hostname` |
| `tag delete PROJECT/REPO NAME` | 2 | — | `--hostname` |

**MCP tools**: `list_tags`, `create_tag`, `delete_tag`

---

### G — PR Lifecycle

No new domain types. New interfaces:

```go
type PREditor           interface { UpdatePR(ns, slug string, id int, in UpdatePRInput) (PullRequest, error) }
type PRDecliner         interface { DeclinePR(ns, slug string, id int) error }
type PRUnapprover       interface { UnapprovePR(ns, slug string, id int) error }
type PRReadier          interface { ReadyPR(ns, slug string, id int) error }         // draft → open
type PRReviewRequester  interface { RequestReview(ns, slug string, id int, users []string) error }
type PRChangesRequester interface { RequestChangesPR(ns, slug string, id int) error } // Cloud only
```

`PREditor`, `PRDecliner`, `PRUnapprover`, `PRReadier`, `PRReviewRequester` → composite `Client`.
`PRChangesRequester` → Cloud-only optional interface.

**New type**:
```go
type UpdatePRInput struct {
    Title       string // empty = no change
    Description string // empty = no change
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `pr edit PR_ID` | 1 | — | `--title`, `--body`, `--hostname` |
| `pr decline PR_ID` | 1 | — | `--hostname` |
| `pr unapprove PR_ID` | 1 | — | `--hostname` |
| `pr ready PR_ID` | 1 | — | `--hostname` |
| `pr request-review PR_ID` | 1 | `--reviewer` (repeatable) | `--hostname` |
| `pr request-changes PR_ID` | 1 | — | `--hostname` _(Cloud only)_ |

**MCP tools**: `update_pr`, `decline_pr`, `unapprove_pr`, `ready_pr`, `request_review`

---

### M — Shell Completion

No backend changes. Single file `pkg/cmd/completion/completion.go`.
Cobra provides built-in completion generation.

```go
rootCmd.AddCommand(&cobra.Command{
    Use:       "completion [bash|zsh|fish|powershell]",
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    Args:      cobra.ExactValidArgs(1),
    RunE:      func(...) { /* dispatch to cobra gen method */ },
})
```

No MCP tool needed.

---

### P — Auth Extras

No backend interface changes. No new types.

**`auth token`** — reads `HostConfig.OAuthToken` from config and prints to stdout. One-liner; matches `gh auth token`.

**`auth refresh`** — calls `GetCurrentUser()`, updates `HostConfig.User` if changed, calls `cfg.Save()`. Optionally re-stores in keyring.

**Commands**:

| Command | Args | Flags |
|---|---|---|
| `auth token` | 0 | `--hostname` |
| `auth refresh` | 0 | `--hostname` |

No MCP tools needed.

---

### Q — Repo Extras

**New interfaces**:
```go
type RepoRenamer  interface { RenameRepo(ns, slug, newSlug string) (Repository, error) }
type RepoArchiver interface { ArchiveRepo(ns, slug string) error }  // Cloud only optional
type RepoForker   interface { ForkRepo(ns, slug string, in ForkRepoInput) (Repository, error) } // Cloud only optional
```

`RepoRenamer` → composite `Client` (both backends support rename).
`RepoArchiver`, `RepoForker` → Cloud-only optional interfaces.

**New type**:
```go
type ForkRepoInput struct {
    Workspace string // destination workspace (empty = user's default)
    Name      string // new slug (empty = same as source)
}
```

**`repo set-default`** — writes `PROJECT/REPO` to `.git/config` under `[bitbottle]`, reads it in `f.ResolveRef` when no arg is given. Enables arg-free commands like `bitbottle pr list`.

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `repo fork PROJECT/REPO` | 1 | — | `--workspace`, `--name`, `--hostname` _(Cloud only)_ |
| `repo rename PROJECT/REPO NEW-NAME` | 2 | — | `--hostname` |
| `repo archive PROJECT/REPO` | 1 | — | `--confirm`, `--hostname` _(Cloud only)_ |
| `repo set-default PROJECT/REPO` | 1 | — | `--hostname` |

**MCP tools**: `fork_repo`, `rename_repo`

---

### F — Commits

**New interfaces**:
```go
type CommitLister interface { ListCommits(ns, slug, branch string, limit int) ([]Commit, error) }
type CommitReader interface { GetCommit(ns, slug, hash string) (Commit, error) }
```
Both in composite `Client`.

**New types**:
```go
type Commit struct {
    Hash      string
    Message   string
    Author    User
    Timestamp time.Time
    WebURL    string
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `commit log PROJECT/REPO` | 1 | — | `--branch`, `--limit`, `--json`, `--jq`, `--hostname` |
| `commit view PROJECT/REPO HASH` | 2 | — | `--web`, `--json`, `--jq`, `--hostname` |

**MCP tools**: `list_commits`, `get_commit`

---

### H — Pipeline Depth _(Cloud only)_

Extend `PipelineClient` (already Cloud-only optional interface):

```go
type PipelineClient interface {
    // existing:
    ListPipelines(ns, slug string, limit int) ([]Pipeline, error)
    GetPipeline(ns, slug, uuid string) (Pipeline, error)
    RunPipeline(ns, slug string, in RunPipelineInput) (Pipeline, error)
    // new:
    ListPipelineSteps(ns, slug, uuid string) ([]PipelineStep, error)
    GetPipelineStepLog(ns, slug, pipelineUUID, stepUUID string) (string, error)
    ListPipelineVariables(ns, slug string) ([]PipelineVariable, error)
    SetPipelineVariable(ns, slug string, in PipelineVariableInput) (PipelineVariable, error)
    DeletePipelineVariable(ns, slug, uuid string) error
}
```

**New types**:
```go
type PipelineStep struct {
    UUID     string
    Name     string
    State    string  // PENDING | RUNNING | SUCCESSFUL | FAILED
    Result   string
    Duration int     // seconds
}

type PipelineVariable struct {
    UUID    string
    Key     string
    Value   string  // empty if Secured
    Secured bool
}

type PipelineVariableInput struct {
    Key     string
    Value   string
    Secured bool
}
```

**Commands**:

| Command | Args | Optional flags |
|---|---|---|
| `pipeline steps PROJECT/REPO UUID` | 2 | `--json`, `--jq`, `--hostname` |
| `pipeline logs PROJECT/REPO PIPELINE-UUID STEP-UUID` | 3 | `--hostname` |
**MCP tools**: `list_pipeline_steps`, `get_pipeline_step_log`, `list_pipeline_variables`, `set_pipeline_variable`, `delete_pipeline_variable`

---

### I — Webhooks

**New interfaces**:
```go
type WebhookLister  interface { ListWebhooks(ns, slug string) ([]Webhook, error) }
type WebhookReader  interface { GetWebhook(ns, slug, id string) (Webhook, error) }
type WebhookCreator interface { CreateWebhook(ns, slug string, in CreateWebhookInput) (Webhook, error) }
type WebhookDeleter interface { DeleteWebhook(ns, slug, id string) error }
```
All in composite `Client`.

**New types**:
```go
type Webhook struct {
    ID     string
    URL    string
    Events []string
    Active bool
}

type CreateWebhookInput struct {
    URL    string
    Events []string
    Active bool
    Secret string // write-only; not returned by API
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `webhook list PROJECT/REPO` | 1 | — | `--json`, `--jq`, `--hostname` |
| `webhook view PROJECT/REPO ID` | 2 | — | `--json`, `--hostname` |
| `webhook create PROJECT/REPO` | 1 | `--url`, `--events` | `--secret`, `--active`, `--hostname` |
| `webhook delete PROJECT/REPO ID` | 2 | — | `--hostname` |

**MCP tools**: `list_webhooks`, `get_webhook`, `create_webhook`, `delete_webhook`

---

### J — PR Comments

**New interfaces**:
```go
type PRCommentLister interface { ListPRComments(ns, slug string, id int) ([]PRComment, error) }
type PRCommentAdder  interface { AddPRComment(ns, slug string, id int, in AddPRCommentInput) (PRComment, error) }
```
Both in composite `Client`.

**New types**:
```go
type PRComment struct {
    ID        int
    Author    User
    Text      string
    CreatedAt time.Time
}

type AddPRCommentInput struct {
    Text string
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `pr comment list PR_ID` | 1 | — | `--json`, `--jq`, `--hostname` |
| `pr comment add PR_ID` | 1 | `--body` | `--hostname` |

**MCP tools**: `list_pr_comments`, `add_pr_comment`

---

### K — Commit Statuses

**New interface**:
```go
type CommitStatusLister interface {
    ListCommitStatuses(ns, slug, hash string) ([]CommitStatus, error)
}
```
In composite `Client`. Implement after Scope F (commits).

**New type**:
```go
type CommitStatus struct {
    Key         string
    State       string // SUCCESSFUL | FAILED | INPROGRESS | STOPPED
    Name        string
    Description string
    URL         string
}
```

**Commands**:

| Command | Args | Optional flags |
|---|---|---|
| `commit status PROJECT/REPO HASH` | 2 | `--json`, `--jq`, `--hostname` |

**MCP tools**: `list_commit_statuses`

---

### T — Output DX

No backend changes. Two sub-tasks:

**Pager** — wire `$PAGER` in `IOStreams.StartPager()`. When stdout is a TTY and output exceeds terminal height, pipe through `$PAGER` (default `less -FRX`). Pattern: open pager subprocess, replace `IOStreams.Out` with pager stdin, call `IOStreams.StopPager()` in defer. Apply to `pr diff` and `commit log` first.

**Color** — implement ANSI coloring in `internal/tableprinter` and `format.Printer`. States like `SUCCESSFUL`/`OPEN` should render green; `FAILED`/`DECLINED` red; `MERGED` magenta. Respect `NO_COLOR` env var and `--no-color` flag. `IOStreams.ColorEnabled()` is already plumbed.

---

### U — Config

No backend changes. New command group `pkg/cmd/config/`.

Reads/writes `~/.config/bitbottle/hosts.yml` fields that are not credentials (credentials stay in `auth` commands). Targets: `git_protocol`, `backend_type`, `skip_tls_verify`.

```go
bitbottle config list                          // print all key=value
bitbottle config get git_protocol              // print single value
bitbottle config set git_protocol https        // write value
```

**Commands** (all accept optional `--hostname`):

| Command | Args | Notes |
|---|---|---|
| `config list` | 0 | Shows all non-secret config fields |
| `config get KEY` | 1 | Exits 1 if key not set |
| `config set KEY VALUE` | 2 | Validates known keys; rejects unknown |

No MCP tools needed.

---

### V — API Passthrough

Single command `pkg/cmd/api/api.go`. Matches `gh api`.

Makes an authenticated HTTP request to any Bitbucket API path, streams the response to stdout. Useful for long-tail operations not covered by dedicated commands.

```
bitbottle api /2.0/repositories/myws/myrepo
bitbottle api /2.0/user
bitbottle api --method POST /2.0/repositories/myws/myrepo/hooks \
  --field url=https://example.com --field events='["repo:push"]'
```

No backend interface needed — calls `f.Backend(host)` internal HTTP client directly.

**Flags**: `--method` (default GET), `--field key=value` (JSON body), `--hostname`, `--jq`.

No MCP tool needed (MCP tools cover specific operations; raw passthrough is a CLI-only escape hatch).

---

### N — Workspace / Projects _(Cloud only)_

Optional interface (not in composite `Client`):
```go
type WorkspaceClient interface {
    ListWorkspaces(limit int) ([]Workspace, error)
    ListProjects(workspace string, limit int) ([]Project, error)
}
```

**New types**:
```go
type Workspace struct { Slug string; Name string }
type Project   struct { Key  string; Name string }
```

**Commands**: `workspace list`, `project list WORKSPACE`
**MCP tools**: `list_workspaces`, `list_projects`

---

### O — Issues _(Cloud only)_

Optional interface (not in composite `Client`):
```go
type IssueClient interface {
    ListIssues(ns, slug, status string, limit int) ([]Issue, error)
    GetIssue(ns, slug string, id int) (Issue, error)
    CreateIssue(ns, slug string, in CreateIssueInput) (Issue, error)
    UpdateIssue(ns, slug string, id int, in UpdateIssueInput) (Issue, error)
}
```

**Commands**: `issue list`, `issue view`, `issue create`, `issue close`
**MCP tools**: `list_issues`, `get_issue`, `create_issue`, `close_issue`

---

### BP — Branch Protect (Server/DC only)

Bitbucket Server/DC exposes branch restrictions via the `branch-permissions/2.0` REST namespace. Cloud has a separate concept (`branch-restrictions`) that is differently shaped — out of scope for this iteration; surface `ErrUnsupportedOnHost` on Cloud.

**New optional interface** (`api/backend/client.go`):
```go
type BranchProtector interface {
    ListBranchProtections(ns, slug string, limit int) ([]BranchProtection, error)
    CreateBranchProtection(ns, slug string, in CreateBranchProtectionInput) (BranchProtection, error)
    DeleteBranchProtection(ns, slug string, id int) error
}
```
Optional-interface pattern, with `AsBranchProtector` returning `ErrUnsupportedOnHost` on Cloud.

**New types** (`api/backend/types.go`):
```go
type BranchProtection struct {
    ID         int
    Type       string   // no-deletes, fast-forward-only, pull-request-only, read-only
    MatcherID  string   // pattern, branch name, model id, or branch type
    MatcherKind string  // PATTERN, BRANCH, MODEL_BRANCH, MODEL_CATEGORY
    Users      []string // exempted user slugs
    Groups     []string // exempted group slugs
}

type CreateBranchProtectionInput struct {
    Type        string
    MatcherID   string
    MatcherKind string
    Users       []string
    Groups      []string
}
```

Uses dedicated transport at `/rest/branch-permissions/2.0` (mirrors the `default-reviewers` pattern).

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `branch protect list PROJECT/REPO` | 1 | — | `--limit`, `--json`, `--jq`, `--hostname` |
| `branch protect create PROJECT/REPO` | 1 | `--type`, `--pattern` \| `--branch` | `--user` (repeatable), `--group` (repeatable), `--hostname` |
| `branch protect delete PROJECT/REPO ID` | 2 | — | `--hostname` |

**MCP tools**: `list_branch_protections`, `create_branch_protection`, `delete_branch_protection`

---

### EX — Error UX (centralised, human-readable, actionable)

Today the CLI surfaces backend errors largely as raw `fmt.Errorf` strings — fine for engineers, hostile for everyone else. This scope introduces a single error UX layer so users see the same shape of message regardless of which command hit which API failure, with hints that tell them what to do next.

**Goals**

- Every command exits with a *humanized* error: title line + cause + hint(s).
- Hints are deterministic and based on a tokenised error code, not heuristic substring matching.
- Backend adapters classify errors once (HTTP status + endpoint + payload) → typed `DomainError` with a code; the cmd layer prints them.
- TTY-aware: colour + secondary lines on TTY, single-line on pipes.
- `--debug` keeps the raw transport error for engineers.

**New types** (`api/backend/errors.go`):
```go
type ErrorCode string // e.g. "auth.invalid_token", "pr.merge.conflict", "perm.write_required"

type DomainError struct {
    Code    ErrorCode
    Title   string   // short summary
    Cause   string   // server's message, sanitised
    Hints   []string // 0..N actionable next steps
    HTTP    int      // optional, debug only
    URL     string   // request URL, debug only
    Wrapped error    // original
}
```

**New package** (`pkg/errfmt/`):
- Catalogue of error codes → templated hints (Markdown-free, plain text).
- `Render(io *iostreams.IOStreams, err error)` — knows how to format a `DomainError` with TTY colour, fallback for plain `error`.
- `errfmt.Wrap(err, code, hint...)` — adapters use this to attach a code + hint at the boundary.

**Adapter changes**:
- `httpx.Transport.UseDomainErrors(host)` already wraps to `ErrAuth` / `ErrPermission` / `ErrNotFound` / `ErrConflict` — extend it to attach error codes (e.g. `auth.invalid_token`, `perm.write_required`, `repo.not_found`, `pr.merge.conflict`).
- Cloud and Server adapters add operation-specific codes where the generic mapping is too coarse (e.g. `pr.create.duplicate_branch`, `pr.reviewer.unknown`, `branch.protected`).

**Cmd layer changes**:
- `cmd/root` wraps `cmd.Execute()` and routes any error through `errfmt.Render` before exit.
- Each command keeps using regular `error` returns — no per-command boilerplate.

**Catalogue (initial)**:

| Code | Trigger | Hint |
|---|---|---|
| `auth.no_token` | empty token in config | `bitbottle auth login --hostname HOST` |
| `auth.invalid_token` | 401 | Token expired or revoked. Run `bitbottle auth refresh`. |
| `perm.write_required` | 403 on a write op | Your token lacks write scope on this repo. |
| `repo.not_found` | 404 on repo path | Check `PROJECT/REPO` casing; on Server use the project key, not the project name. |
| `pr.not_found` | 404 on pr path | The PR may have been deleted; try `bitbottle pr list`. |
| `pr.merge.conflict` | 409 on merge | Resolve conflicts locally and push, then retry. |
| `pr.merge.behind` | 409 + "behind" | Update branch from base, then retry (`gh pr update-branch` analogue). |
| `pr.create.duplicate_branch` | 409 on create | A PR for this source/target already exists. |
| `pr.reviewer.unknown` | 400 on create with reviewer | One or more `--reviewer` slugs is not a member. |
| `branch.protected` | 403 on push/delete | This branch is protected; ask an admin or use `branch protect list`. |
| `host.unsupported` | `ErrUnsupportedOnHost` | This command is not available on Bitbucket Cloud / Server. |
| `network.tls_unknown_authority` | TLS error | Add `-k` (or `skip_tls_verify: true` in config) for self-signed CAs. |
| `transport.timeout` | request timeout | Network slow or VPN down. Retry with `--debug` for details. |

**New codes needed by upcoming scopes** (add when the originating scope lands):

| Code | Originating scope | Trigger | Hint |
|---|---|---|---|
| `pr.automerge.beta_disabled` | **AUTOMERGE** | Cloud-specific 404 body when workspace beta is off | Ask your workspace admin to enable auto-merge in workspace settings. |
| `perms.admin_required` | **PERMS** | 403 on `perms project\|repo grant/revoke` | You need PROJECT_ADMIN on this project to manage permissions. |
| `admin.sys_admin_required` | **ADMIN** | 403 on `admin secrets rotate` / `admin logging set` | Standard admin tokens do not include SYS_ADMIN; ask a system administrator to perform this action. |
| `variable.system_managed` | **VAR** | 400 on writing a `system=true` variable | This variable is managed by Bitbucket and cannot be modified. |
| `task.unsupported_server_version` | **TASK** | Server < 7.2 detected via `SRVVER` | Severity-BLOCKER comments require Bitbucket Server 7.2 or newer. |
| `keyring.unavailable` | **SEC** | Keyring `Set` times out / no backend | Keyring not available in this environment. Set BITBOTTLE_ALLOW_INSECURE_STORE=1 to use the file fallback, or run `bitbottle auth migrate` on a desktop. |
| `keyring.token_too_large` | **SEC** | Windows 2048-byte limit hit | Token exceeds Credential Manager's size limit. Falling back to encrypted file store. |
| `extension.binary_changed` | **EXT** | Installed binary SHA differs from lockfile | The extension binary has changed since install. Run `bitbottle extension upgrade NAME` to refresh. |
| `extension.no_arch_binary` | **EXT** | No matching `<os>-<arch>` asset in release | No `bitbottle-NAME` binary available for your OS/arch. Check the extension's release assets. |

**Migration**:

- One PR per cluster of codes (auth, repo, pr, branch, network) — each is a thin error-mapping change with snapshot-style tests.
- Final PR wires `errfmt.Render` into root and adds golden-file tests.

**Status**:

- ✅ EX1 — foundation + `auth` cluster: `pkg/errfmt/` with code catalogue, `Render` delegated from `cmdutil.ExplainError`, `DomainError.Code` field, `auth.no_token` / `auth.invalid_token` / `perm.write_required` codes, classifier auto-attaches `auth.invalid_token` on 401.
- ✅ EX2-EX6 — all remaining clusters: `repo.not_found`, `pr.not_found`, `pr.merge.conflict`, `pr.merge.behind`, `pr.create.duplicate_branch`, `pr.reviewer.unknown`, `branch.protected`, `host.unsupported`, `network.tls_unknown_authority`, `transport.timeout`. Transport-layer classification at `httpx.Transport.do()` via `backend.ClassifyTransportError`. MCP envelope carries dotted codes + hints from the catalogue (breaking change to `errorEnvelope.code` shape: was kind-based strings, now matches `backend.ErrorCode`).

**Definition of Done**:

- `pkg/errfmt/` with table-driven tests against the full code catalogue.
- Every existing typed error in `api/backend/errors.go` carries a code.
- All TTY golden tests for top-level commands updated to assert the new format.
- README has a "When errors happen" section with example output.

---

### RV — Code-Review Primitives

The strategic prize for coding-agent UX. Today an agent reviewing a PR cannot:
read a file at the PR head without cloning, list inline review comments,
post inline review comments at file:line, edit/delete/resolve comments, or
reconstruct the PR's review history. RV closes all of that in one epic.

**Why**: coding agents (Claude / Codex / Cursor in pair-programming mode) spend
the majority of their PR-review time in three loops — read the diff + read
surrounding code, read existing review feedback, post new review feedback.
Bitbucket Cloud + Server expose all the necessary endpoints; bitbottle wraps
none of them today.

**New interfaces** (`api/backend/client.go`):
```go
type SourceReader interface {
    GetFileContent(ns, slug, ref, path string) ([]byte, error)
    ListTree(ns, slug, ref, path string) ([]TreeEntry, error)
}

type PRInlineCommentClient interface {
    ListPRComments(ns, slug string, id int) ([]PRComment, error)        // already exists
    AddPRComment(ns, slug string, id int, in AddPRCommentInput) (PRComment, error)
    EditPRComment(ns, slug string, id, commentID int, body string) (PRComment, error)
    DeletePRComment(ns, slug string, id, commentID int) error
    ReplyPRComment(ns, slug string, id, parentID int, body string) (PRComment, error)
    ResolvePRComment(ns, slug string, id, commentID int) error          // Cloud `resolution`; Server: task or thread close
}

type PRReviewer interface {
    SubmitReview(ns, slug string, id int, in SubmitReviewInput) error
}

type PRActivityReader interface {
    GetPRActivity(ns, slug string, id int, limit int) ([]PRActivityEvent, error)
}

type CommitCommenter interface {
    ListCommitComments(ns, slug, hash string) ([]Comment, error)
    AddCommitComment(ns, slug, hash string, in AddCommentInput) (Comment, error)
    EditCommitComment(ns, slug, hash string, commentID int, body string) (Comment, error)
    DeleteCommitComment(ns, slug, hash string, commentID int) error
}
```

`SourceReader`, `PRInlineCommentClient` (extend the existing PRComment* interfaces),
`PRReviewer`, `PRActivityReader`, `CommitCommenter` → composite `Client` (both
backends support all of these).

**New types**:
```go
type TreeEntry struct {
    Path string
    Type string // "commit_file" / "commit_directory" (Cloud) → mapped to "file"/"dir"
    Size int64
    Hash string
}

// AddPRCommentInput grows an optional Inline field; absent = general comment.
type AddPRCommentInput struct {
    Body   string
    Inline *PRCommentInline // nil = general comment
    Parent *int             // nil = top-level; set = reply
}

type PRCommentInline struct {
    Path      string  // required when Inline != nil
    Side      string  // "old" or "new" — translates to from/to (Cloud) or fileType FROM/TO (Server)
    Line      int     // single-line target
    StartLine int     // for multi-line; 0 = single-line
    // Server-only — populated internally from PR diff endpoint:
    fromHash, toHash string
}

type PRComment struct {
    ID         int
    Author     User
    Body       string
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Inline     *PRCommentInline // nil = general
    ParentID   int              // 0 = top-level
    Resolved   bool
}

type SubmitReviewInput struct {
    Action string         // "approve" | "request_changes" | "comment"
    Body   string         // top-level review body
    Inline []PRCommentInline // optional inline comments to attach atomically
}

type PRActivityEvent struct {
    Type      string    // "approval" | "comment" | "update" | "merge" | ...
    Actor     User
    CreatedAt time.Time
    Detail    map[string]any // type-specific payload
}
```

**Adapter notes**:
- **Cloud inline comments**: `inline.{path,from,to,start_from,start_to}`. `from` = old-side line, `to` = new-side line; set the other to nil. Multi-line via `start_*`.
- **Server inline comments**: requires `anchor.{diffType,line,lineType,fileType,fromHash,toHash,path,srcPath}`. The `fromHash`/`toHash` are obtained from `GET /pull-requests/{id}/diff/{path}` and cached internally so the user only supplies `path:line:body`.
- **Source/raw**: Cloud `GET /repositories/{ws}/{slug}/src/{ref}/{path}` returns file content for a file-path and JSON tree metadata for a directory-path. Server uses `GET /raw/{path}?at=` for content and `GET /browse/{path}?at=` for tree.
- **`pr review` is one HTTP call on Cloud** (POST review endpoint with optional inline-comments array) but on Server is multiple calls (post each comment + then approve/decline). Wrap the difference; surface a single CLI verb.

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `repo file get PROJECT/REPO PATH` | 2 | `--ref` (branch / tag / hash) | `--out FILE`, `--hostname` |
| `repo tree PROJECT/REPO [PATH]` | 1-2 | `--ref` | `--json`, `--jq`, `--hostname` |
| `pr review PR_ID` | 1 | one of `--approve` / `--request-changes` / `--comment` | `--body`, `--inline path:line:body` (repeatable), `--inline-from-stdin` (JSON array), `--hostname` |
| `pr comment list PR_ID` | 1 | — | `--inline`, `--json`, `--jq`, `--hostname` |
| `pr comment add PR_ID` | 1 | `--body` | `--inline path:line` (single), `--side new\|old`, `--parent COMMENT_ID`, `--hostname` |
| `pr comment edit PR_ID COMMENT_ID` | 2 | `--body` | `--hostname` |
| `pr comment delete PR_ID COMMENT_ID` | 2 | — | `--hostname` |
| `pr comment resolve PR_ID COMMENT_ID` | 2 | — | `--hostname` |
| `pr activity PR_ID` | 1 | — | `--limit`, `--json`, `--jq`, `--hostname` |
| `commit comment list PROJECT/REPO HASH` | 2 | — | `--json`, `--jq`, `--hostname` |
| `commit comment add PROJECT/REPO HASH` | 2 | `--body` | `--inline path:line`, `--hostname` |
| `commit comment edit PROJECT/REPO HASH COMMENT_ID` | 3 | `--body` | `--hostname` |
| `commit comment delete PROJECT/REPO HASH COMMENT_ID` | 3 | — | `--hostname` |

**MCP tools** (new): `get_file_content`, `list_tree`, `submit_pr_review`,
`edit_pr_comment`, `delete_pr_comment`, `reply_pr_comment`, `resolve_pr_comment`,
`get_pr_activity`, `list_commit_comments`, `add_commit_comment`,
`edit_commit_comment`, `delete_commit_comment`. Existing `add_pr_comment` /
`list_pr_comments` extend their schemas with `inline` / `parent` parameters.

**Sub-PRs (suggested split)**:
- **RV1** Source/file primitives (`repo file get` + `repo tree`)
- **RV2** Inline PR comments — read path (list + filter)
- **RV3** Inline PR comments — write path (add + edit + delete + reply + resolve)
- **RV4** `pr review` compound command (orchestrates RV3 + approve/request-changes)
- **RV5** `pr activity`
- **RV6** `commit comment *` (CRUD + inline)

---

### SR — Code Search _(Cloud only)_

**Optional interface**:
```go
type CodeSearcher interface {
    SearchCode(workspace, query string, limit int) ([]CodeSearchHit, error)
}
```
Cloud-only optional; surfaces `host.unsupported` on Server (Server has no
first-class code-search REST API in DC stock; Sourcegraph or a Server plugin
fills the gap there).

**New types**:
```go
type CodeSearchHit struct {
    Repository      string  // "workspace/slug"
    Path            string
    PathMatches     []SearchSegment // segments of the path with match flags
    ContentMatches  []ContentMatch  // matched lines with surrounding context
    ContentMatchCount int
    FileURL         string
}

type ContentMatch struct {
    Line     int
    Segments []SearchSegment // alternating non-match / match runs
}

type SearchSegment struct {
    Text  string
    Match bool
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `bitbottle search code QUERY` | 1 | `--workspace W` (or pinned default) | `--limit`, `--json`, `--jq`, `--hostname` |

**MCP tools**: `search_code`

---

### CTX — Context Primitive

No new backend interface — wraps existing `f.BaseRepo()`, `f.Backend(host)`,
config + git resolution. Single high-leverage command for agents.

```
bitbottle context [--json]
```

Returns: host, project, slug, current branch, default branch, ahead/behind vs
default, authenticated user (slug + display name), scopes / token kind.

**Why**: today an agent calls `auth status` + `repo view` + `git status` to
orient. One call replaces three. Especially valuable for MCP — `get_context`
becomes the standard first call in any agent flow.

**Commands**:

| Command | Args | Optional flags |
|---|---|---|
| `bitbottle context` | 0 | `--json`, `--jq`, `--hostname` |

**MCP tools**: `get_context`

---

### GHP — gh-Parity Gaps

Small, mostly-reads, high-frequency UX wins from the gh CLI surface. Each is
roughly a one-PR change.

| Command | Backends | Notes |
|---|---|---|
| `pr checks PR_ID` | Both | Resolve PR head → call `commit status HASH` under the hood. Optional `--watch` polls until terminal state. |
| `pr update-branch PR_ID` | Both | Sync PR head with base. Cloud: `POST /pullrequests/{id}/update-branch`-style endpoint or merge-target-into-source. Server: equivalent merge action. The verb our `pr.merge.behind` hint already promises. |
| `pr reopen PR_ID` | Both | Reverse `pr decline`. Cloud + Server both expose state transition back to OPEN. |
| `pr status` | Both | Cross-repo "what's on my plate". Cloud: `/dashboard/pullrequests` (your-PRs / review-requested). Server: `/inbox/pull-requests`. Workspace-scoped on Cloud. |
| `bitbottle status` | Both | Combined inbox: review requests + assigned issues + recent activity. Composite of `pr status` + `issue list --assigned-to-me`. |
| `bitbottle browse [PATH\|NUMBER]` | Both | Top-level web shortcut. Resolves: `123` → PR, `abc1234` → commit, path → `src/branch/path`. |
| `pipeline watch PROJECT/REPO UUID` | Cloud | Poll until terminal state, optionally stream step transitions. |

**MCP tools**: `pr_checks`, `pr_update_branch`, `pr_reopen`, `pr_status`,
`status`, `pipeline_watch`. (`browse` is interactive-only — no MCP tool.)

---

### OF — Issues Finish _(Cloud only)_ ✅ DONE

Closes the gap left after scope **O**. Today users can list/view/create/close
issues but cannot *edit*, *comment on*, *reopen*, or *assign* them.

> **Implemented in `feat/of-issues-finish`**. All commands, MCP tools, and
> unit tests shipped. See `docs/manual-tests/cloud/issue-lifecycle.md`.

**Extend the existing `IssueClient` optional interface**:
```go
type IssueClient interface {
    // existing:
    ListIssues(ns, slug, status string, limit int) ([]Issue, error)
    GetIssue(ns, slug string, id int) (Issue, error)
    CreateIssue(ns, slug string, in CreateIssueInput) (Issue, error)
    CloseIssue(ns, slug string, id int) error
    // new:
    UpdateIssue(ns, slug string, id int, in UpdateIssueInput) (Issue, error)
    ReopenIssue(ns, slug string, id int) error
    AssignIssue(ns, slug string, id int, assignee string) error

    ListIssueComments(ns, slug string, id int) ([]Comment, error)
    AddIssueComment(ns, slug string, id int, body string) (Comment, error)
    EditIssueComment(ns, slug string, id, commentID int, body string) (Comment, error)
    DeleteIssueComment(ns, slug string, id, commentID int) error
}
```

**New type**:
```go
type UpdateIssueInput struct {
    Title       string // empty = no change
    Body        string // empty = no change
    Kind        string // empty = no change ("bug" | "enhancement" | "proposal" | "task")
    Priority    string // empty = no change ("trivial" | "minor" | "major" | "critical" | "blocker")
    Assignee    string // empty = no change
    State       string // empty = no change ("new" | "open" | "resolved" | "on hold" | "invalid" | "duplicate" | "wontfix" | "closed")
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `issue edit ISSUE_ID` | 1 | — | `--title`, `--body`, `--kind`, `--priority`, `--assignee`, `--state`, `--hostname` |
| `issue reopen ISSUE_ID` | 1 | — | `--hostname` |
| `issue assign ISSUE_ID USER` | 2 | — | `--hostname` |
| `issue comment list ISSUE_ID` | 1 | — | `--json`, `--jq`, `--hostname` |
| `issue comment add ISSUE_ID` | 1 | `--body` | `--hostname` |
| `issue comment edit ISSUE_ID COMMENT_ID` | 2 | `--body` | `--hostname` |
| `issue comment delete ISSUE_ID COMMENT_ID` | 2 | — | `--hostname` |

**MCP tools**: `update_issue`, `reopen_issue`, `assign_issue`,
`list_issue_comments`, `add_issue_comment`, `edit_issue_comment`,
`delete_issue_comment`.

---

### CI — Code Insights _(Server / DC only)_

Bitbucket Server's first-class concept for posting build / quality / security
analysis as PR annotations. Different REST namespace
(`/rest/insights/1.0/...`) — separate transport setup like the
`branch-permissions/2.0` and `default-reviewers` patterns. Cloud has no
equivalent — surface `host.unsupported` there.

**New optional interface**:
```go
type CodeInsightsClient interface {
    ListReports(project, slug, hash string) ([]CodeInsightsReport, error)
    GetReport(project, slug, hash, key string) (CodeInsightsReport, error)
    SetReport(project, slug, hash, key string, in CodeInsightsReportInput) (CodeInsightsReport, error) // PUT (upsert)
    DeleteReport(project, slug, hash, key string) error

    ListAnnotations(project, slug, hash, key string) ([]CodeInsightsAnnotation, error)
    AddAnnotations(project, slug, hash, key string, in []CodeInsightsAnnotationInput) error           // bulk POST
    DeleteAnnotations(project, slug, hash, key string) error

    // Merge-check (partly undocumented endpoint — flag as experimental in CLI):
    SetMergeCheck(project, slug, key string, in MergeCheckInput) error
    GetMergeCheck(project, slug, key string) (MergeCheck, error)
    DeleteMergeCheck(project, slug, key string) error
}
```

**New types**:
```go
type CodeInsightsReport struct {
    Key      string
    Title    string
    Details  string
    Result   string  // PASS | FAIL
    Reporter string
    Link     string
    LogoURL  string
    Data     []CodeInsightsReportDatum
}

type CodeInsightsReportDatum struct {
    Title string
    Type  string  // BOOLEAN | DATE | DURATION | LINK | NUMBER | PERCENTAGE | TEXT
    Value any
}

type CodeInsightsReportInput struct {
    Title, Details, Reporter, Link, LogoURL string
    Result string                  // PASS | FAIL
    Data   []CodeInsightsReportDatum
}

type CodeInsightsAnnotation struct {
    ExternalID string
    Path       string
    Line       int
    Severity   string  // LOW | MEDIUM | HIGH
    Type       string  // VULNERABILITY | CODE_SMELL | BUG
    Message    string
    Link       string
}

type CodeInsightsAnnotationInput = CodeInsightsAnnotation // input shape matches output

type MergeCheckInput struct {
    ReportKey            string
    MustPass             bool
    MinProhibitedSeverity string // LOW | MEDIUM | HIGH | "" (any)
}

type MergeCheck = MergeCheckInput
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `code-insights report list PROJECT/REPO HASH` | 2 | — | `--json`, `--jq`, `--hostname` |
| `code-insights report view PROJECT/REPO HASH KEY` | 3 | — | `--json`, `--hostname` |
| `code-insights report set PROJECT/REPO HASH KEY` | 3 | `--title`, `--result PASS\|FAIL` | `--details`, `--reporter`, `--link`, `--logo-url`, `--data K=V:TYPE` (repeatable), `--hostname` |
| `code-insights report delete PROJECT/REPO HASH KEY` | 3 | — | `--hostname` |
| `code-insights annotation list PROJECT/REPO HASH KEY` | 3 | — | `--json`, `--jq`, `--hostname` |
| `code-insights annotation add PROJECT/REPO HASH KEY` | 3 | `--from-json @PATH\|-` (bulk) OR `--path`, `--line`, `--severity`, `--type`, `--message` for single | `--external-id`, `--link`, `--hostname` |
| `code-insights annotation delete PROJECT/REPO HASH KEY` | 3 | — | `--hostname` |
| `code-insights merge-check set PROJECT/REPO KEY` | 2 | `--report-key`, `--must-pass` | `--min-severity LOW\|MEDIUM\|HIGH`, `--hostname` |
| `code-insights merge-check get PROJECT/REPO KEY` | 2 | — | `--json`, `--hostname` |
| `code-insights merge-check delete PROJECT/REPO KEY` | 2 | — | `--hostname` |

**MCP tools**: `list_code_insights_reports`, `get_code_insights_report`,
`set_code_insights_report`, `delete_code_insights_report`,
`list_code_insights_annotations`, `add_code_insights_annotations`,
`delete_code_insights_annotations`, `set_merge_check`, `get_merge_check`,
`delete_merge_check`.

---

### DEP — Deployments _(Cloud only)_

Bitbucket Cloud's deployment-environments primitive plus the per-environment
variable bag (separate API from repo-level pipeline variables shipped under
scope **H**). Operational scope, lower priority for coding agents but high
priority for CI integrations.

**New optional interface**:
```go
type DeploymentClient interface {
    ListDeployments(ns, slug string, limit int) ([]Deployment, error)
    GetDeployment(ns, slug, uuid string) (Deployment, error)

    ListEnvironments(ns, slug string) ([]Environment, error)
    CreateEnvironment(ns, slug string, in CreateEnvironmentInput) (Environment, error)
    DeleteEnvironment(ns, slug, uuid string) error

    ListEnvVariables(ns, slug, envUUID string) ([]EnvVariable, error)
    SetEnvVariable(ns, slug, envUUID string, in EnvVariableInput) (EnvVariable, error)
    DeleteEnvVariable(ns, slug, envUUID, varUUID string) error
}
```

**New types**:
```go
type Deployment struct {
    UUID        string
    State       string  // PENDING | IN_PROGRESS | COMPLETED | STOPPED | FAILED
    Environment Environment
    Release     struct{ Name string; URL string; CommitHash string }
}

type Environment struct {
    UUID string
    Name string
    Type string  // Test | Staging | Production
    Rank int
}

type CreateEnvironmentInput struct {
    Name string
    Type string
    Rank int
}

type EnvVariable struct {
    UUID    string
    Key     string
    Value   string  // empty if Secured
    Secured bool
}

type EnvVariableInput struct {
    Key     string
    Value   string
    Secured bool
}
```

> **Note**: if scope **VAR** ships before DEP, drop `EnvVariable` /
> `EnvVariableInput` from this scope and reuse `PipelineVariable` /
> `PipelineVariableInput` instead. The two types are byte-for-byte identical; do
> not duplicate. See VAR's "Implementation notes" point 4.

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `deployment list PROJECT/REPO` | 1 | — | `--limit`, `--json`, `--jq`, `--hostname` |
| `deployment view PROJECT/REPO UUID` | 2 | — | `--json`, `--hostname` |
| `environment list PROJECT/REPO` | 1 | — | `--json`, `--jq`, `--hostname` |
| `environment create PROJECT/REPO` | 1 | `--name`, `--type Test\|Staging\|Production` | `--rank`, `--hostname` |
| `environment delete PROJECT/REPO UUID` | 2 | — | `--confirm`, `--hostname` |
| `environment variable list PROJECT/REPO ENV-UUID` | 2 | — | `--json`, `--jq`, `--hostname` |
| `environment variable set PROJECT/REPO ENV-UUID KEY VALUE` | 4 | — | `--secured`, `--hostname` |
| `environment variable delete PROJECT/REPO ENV-UUID KEY` | 3 | — | `--hostname` |

**MCP tools**: `list_deployments`, `get_deployment`, `list_environments`,
`create_environment`, `delete_environment`, `list_env_variables`,
`set_env_variable`, `delete_env_variable`.

---

### SEC — Secret Store & Config Security

Two closely related hardening items that should ship together in one PR.

**Problem**: `HostConfig.OAuthToken` is written to `hosts.yml` on disk today. In
CI / containers the keyring is unavailable and the CLI can hang indefinitely
waiting for a keyring timeout that never fires.

**Sub-task 1 — Token-never-in-config** (`internal/config/`):

Add `MarshalYAML()` to `HostConfig` that zeroes the token field before serialisation:
```go
func (h HostConfig) MarshalYAML() (any, error) {
    type plain HostConfig
    safe := plain(h)
    safe.OAuthToken = ""
    return safe, nil
}
```
Add a load-time warning: if a token is found in the deserialized struct, print
`! warning: token found in config file — run bitbottle auth login to migrate` and
continue (do not error; existing users must not be broken on upgrade).

**Sub-task 2 — Keyring hardening** (`internal/keyring/`):

Replace the current `OSKeyring` thin wrapper with a production-grade implementation
modelled on bkt's `internal/secret/store.go`:

- `IsHeadless() bool` — returns true when `SSH_TTY`/`DISPLAY`/`WAYLAND_DISPLAY`
  are all unset, or `CI`/`GITHUB_ACTIONS`/`DOCKER` env vars are set.
- `keyringTimeout() time.Duration` — 3s when headless, 60s when interactive.
- Wrap every `Get`/`Set`/`Delete` call with `context.WithTimeout`.
- On macOS, use an advisory file lock (`~/.config/bitbottle/.keyring.lock`) to
  prevent concurrent Keychain access (race between parallel CLI invocations).
- File-based fallback (`~/.config/bitbottle/token.enc`, AES-256-GCM) only when
  `BITBOTTLE_ALLOW_INSECURE_STORE=1` or `--allow-insecure-store` is passed to
  `auth login`.

**No new backend interfaces. No new types. No new commands.**

**Implementation notes**:
1. **macOS Keychain "trust application after upgrade" footgun.** Every time
   `bitbottle` is rebuilt (i.e. every release), macOS treats the new binary as
   untrusted and re-prompts the user for the Keychain password. Set
   `KeychainTrustApplication: true` in the `99designs/keyring` config to avoid
   this — it tells the Keychain to trust the binary by code-signing identity
   rather than path+inode. Forget this and every release breaks Keychain UX for
   every macOS user.
2. **Linux Secret Service needs DBus + an active session.** Headless Linux
   servers (most CI runners, most SSH-into-VPS workflows) have neither.
   `IsHeadless()` must return true when `DBUS_SESSION_BUS_ADDRESS` is unset,
   even if `DISPLAY` is set.
3. **Windows Credential Manager 2048-byte limit.** OAuth refresh tokens can
   exceed this. If `Set` returns `ERROR_INVALID_PARAMETER` on Windows, fall back
   to the file-based store with a clear user-facing message — don't silently
   lose data.
4. **Migration command, not just a warning.** Users in CI pipelines who
   authenticate via config-file token cannot run `auth login` interactively.
   Ship `bitbottle auth migrate` that reads the file token, writes it to the
   keyring, and rewrites the config file with the token field stripped. The
   load-time warning points users at this command, not at interactive `auth login`.
5. **Test infrastructure.** Real keyring backends can't run in CI. Use
   `keyring.NewArrayKeyring` (in-memory) for unit tests; gate real-backend tests
   behind `// +build keyring_integration`.

**Definition of Done**:
- [ ] `HostConfig.MarshalYAML()` strips token; `cfg.Save()` verified by test
- [ ] Load-time warning test for pre-migration config with token in file
- [ ] `bitbottle auth migrate` command + integration test against fake keyring
- [ ] `IsHeadless()` + timeout wrapping; verified in CI environment simulation
- [ ] Darwin advisory lock; verified with parallel test
- [ ] `KeychainTrustApplication: true` set on Darwin (regression test against re-prompt)
- [ ] Windows: handle 2048-byte limit with graceful fallback (test with synthetic large token)
- [ ] File-fallback gated behind env var; integration test for headless flow
- [ ] Existing `auth login` golden tests pass unchanged

---

### HTTPH — HTTP Client Hardening

**Problem**: `api/internal/httpx/Transport` has no retry, no rate-limit awareness,
no ETag caching. A transient 5xx or a momentary rate-limit spike causes an
immediate error surfaced to the user. Under heavy usage against Bitbucket Server
the CLI hammers the host unnecessarily.

Extend `api/internal/httpx/Transport` (or wrap it in a new `ResilientTransport`):

**Retry + backoff** (`httpx/retry.go`):
```go
type RetryPolicy struct {
    MaxAttempts    int           // default 3
    InitialBackoff time.Duration // default 200ms
    MaxBackoff     time.Duration // default 2s
    Jitter         bool          // default true
}
```

**Retry is opt-in per call-site, not method-based**. HTTP method is not a
reliable signal for "is this safe to retry" — Bitbucket exposes some idempotent
operations under POST (search, deploy-key-by-fingerprint) and some genuinely
mutating operations under PUT. A method-based retry policy would either
duplicate `POST /pullrequests` (creating duplicate PRs) or fail to retry safe
POST queries.

The call-site signals retry-safety via context:
```go
ctx := httpx.WithRetry(ctx, retryPolicy) // opt-in
resp, err := tr.Do(req.WithContext(ctx))
```
`Transport.Do` checks for the retry policy on the request context. If absent,
it executes once. If present, it wraps in `retryRoundTripper` with exponential
backoff + full jitter on 5xx and 429 responses (4xx other than 429 are never
retried, regardless of opt-in).

Call-sites that are unambiguously safe to retry (all `GET *`, list operations,
diff fetches, source reads) attach the retry policy in the adapter layer. Write
operations (create/update/delete/merge) do not attach retry and must succeed
or fail on a single attempt.

**Rate-limit tracking** (`httpx/ratelimit.go`):
```go
type RateLimitState struct {
    Limit     int
    Remaining int
    Reset     time.Time
}
// ParseRateLimit reads X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
// (and X-Attempt-RateLimit-* where present on Cloud).
```
When `Remaining == 0`, block until `Reset` before the next request
(`applyAdaptiveThrottle`).

**ETag cache** (`httpx/etag.go`):
- Maintain `map[string]*etagEntry` behind `sync.RWMutex`.
- On GET: if a cached ETag exists for the URL, send `If-None-Match: <etag>`.
- On 304: return cached body + 200 status.
- On 200: store `ETag` header + body.
- Cache is in-process only (no persistence); max 256 entries, LRU eviction.

**429 + Retry-After** (`httpx/retry.go`):
- If response is 429, read `Retry-After` header (seconds or HTTP-date).
- Sleep for that duration (capped at `MaxBackoff`) before retrying.

**No new backend interfaces. No new command changes. No new types.**

**Implementation notes**:
1. **Middleware order matters.** The wrapping must be:
   `Caller → Retry → ETag → RateLimit → ContentTypePolicy → DomainError-Classify → http.Transport`.
   Retry sits *outside* domain-error classification, otherwise a successful retry
   after a 500 gets reclassified by the wrap-around. ContentTypePolicy must sit
   *inside* ETag (the cached response was for a request that already had its
   content-type adjusted). Document this stack order in `httpx/transport.go`.
2. **ETag cache must invalidate on writes to the same resource path.** Caching
   `GET /repositories/X` and then `PUT /repositories/X` (rename) without
   flushing returns stale data on the next GET. On any non-GET response with
   2xx status, flush all cache entries whose URL path starts with the request
   path. Document this is best-effort — Bitbucket sometimes mutates parent
   resources from child URLs.
3. **Cache bound: total bytes, not entry count.** 256 entries × 50KB diff bodies
   = 12MB; that's fine, but one cached `pr diff` of a giant PR can be 5MB
   alone. Use a byte-budget LRU (default 16MB total) rather than entry count.
4. **Retry body replay needs `req.GetBody`.** Go's `http.Request` consumes the
   body on first send. Retry requires `req.GetBody` to be set so the body can
   be re-read. Document the contract: any call-site that calls
   `httpx.WithRetry(ctx, ...)` AND passes a non-nil body MUST set
   `req.GetBody`. Add a runtime check that panics in dev builds if violated.
5. **Adaptive throttle thundering herd.** When `Remaining == 0` and many
   goroutines hit the throttle, they all sleep until the same reset time and
   wake together. Add per-goroutine jitter (e.g., `+rand(0..500ms)`) to the
   reset deadline to spread the wake.
6. **Rate-limit headers are optional.** Bitbucket Cloud doesn't emit them on
   every endpoint; Server emits them only when rate limiting is enabled in
   admin config. Missing headers must be treated as "no limit info"
   (no-op), not as a parse error.

**Definition of Done**:
- [ ] `retryRoundTripper` with table-driven tests covering 5xx retry, 429+Retry-After, non-retriable 4xx
- [ ] Per-call retry opt-in via `httpx.WithRetry(ctx, policy)` context value
- [ ] `RateLimitState` parser with test against Bitbucket header shapes (incl. missing headers)
- [ ] `applyAdaptiveThrottle` test (mock clock + jitter verification)
- [ ] ETag cache unit tests (hit, miss, 304, write-invalidates-prefix, byte-budget eviction)
- [ ] Middleware-order regression test (verify a 500-then-200 retry doesn't re-classify)
- [ ] `req.GetBody` runtime check in dev builds
- [ ] `go test ./... -race` green — cache map is safe under concurrent access
- [ ] Integration test against a mock server simulating 500→500→200 sequence

---

### OUT2 — Extended Output Formats

**Problem**: bitbottle only supports `--json` + `--jq` today. Power users and
script authors often need YAML (readable diffs) or Go templates (custom
one-liners). More importantly, every command must declare its own `--json`/`--jq`
flags today — a global declaration would simplify every command and enable uniform
validation.

**Sub-task 1 — YAML + template support** (`internal/format/`):

Extend `Printer[T].Print` to honour two new formats:

```go
type OutputFormat string
const (
    FormatTable    OutputFormat = ""
    FormatJSON     OutputFormat = "json"
    FormatYAML     OutputFormat = "yaml"
    FormatTemplate OutputFormat = "template"
)
```

Add `format.WriteYAML(w, v)` (uses `gopkg.in/yaml.v3`) and
`format.WriteTemplate(w, tmpl string, v any)` (uses `text/template`).

**Sub-task 2 — Global root-level flags** (`pkg/cmd/root/root.go`):

Move `--json`, `--jq` from per-command flags to persistent flags on the root
command. Add `--yaml` and `--template`:

```go
rootCmd.PersistentFlags().Bool("json", false, "Output as JSON")
rootCmd.PersistentFlags().Bool("yaml", false, "Output as YAML")
rootCmd.PersistentFlags().String("format", "", "Output format: json, yaml, table")
rootCmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
rootCmd.PersistentFlags().String("template", "", "Format output with a Go template")
```

**Sub-task 3 — Validation in `PersistentPreRunE`**:

```
--json and --yaml are mutually exclusive
--jq requires --json (or --format json)
--template requires neither --json nor --yaml
```

Surface a clean user error (not a panic) when combinations are invalid.

**Migration**: cobra's persistent flags are inherited by every subcommand
automatically — there is no shim layer to write. Existing per-command `--json` /
`--jq` declarations should be **deleted** in the same PR that adds the root
declarations, because once a flag is declared persistently on root, redeclaring
it on a child causes a "flag redefined" panic at command construction time.

Audit pass before the PR lands:
1. List every `pflag` declaration for `--json`, `--jq` across `pkg/cmd/**`.
2. Confirm none of them depend on per-command defaults or required-on-this-command-only behaviour. (Spot-check `pr list --json` — does it validate against a per-command field list? If yes, that validation moves to the formatter, not the flag.)
3. Delete the per-command declarations in the same PR as the root declarations.

**No new backend interfaces. No new types. No new commands.**

**Implementation notes**:
1. **Template function set must match gh.** Users moving from gh expect
   `color`, `truncate`, `timeago`, `pluck`, `join`, `tablerender`, `autocolor`,
   `hyperlink`. Without these, ported `gh ... --template` snippets break.
   Implement the same names + signatures as `cli/cli/pkg/cmd/factory/template.go`.
2. **`--jq` and `--yaml` are incompatible.** jq is a JSON query language.
   Surface a clean error: `--jq requires --json (or --format json)`.
3. **Color and structured output don't mix.** When `--format yaml/json/template`
   is active, force `IOStreams.SetColorEnabled(false)` for that invocation.
   ANSI escapes in piped YAML break downstream parsers.
4. **Field selection (gh's `--json field1,field2,...`).** gh allows
   `--json title,state,author` to subselect. Decide whether to support this
   now or defer. Recommendation: defer, ship full-object JSON first, add
   field selection in a follow-up PR (it's a Printer extension, not a flag change).
5. **Pager auto-disable for structured output.** `$PAGER` makes no sense for
   JSON/YAML output destined for `jq` or a script. Skip `IOStreams.StartPager`
   when format is not `table`.
6. **Default field stability.** Once we emit `--json` for a command, the field
   set becomes an API contract. Snapshot golden tests against the JSON output
   so any unintended schema change fails CI.

**Definition of Done**:
- [ ] `format.WriteYAML` + `format.WriteTemplate` with table-driven tests
- [ ] Global flags on root; `PersistentPreRunE` validates combos
- [ ] Golden tests for invalid-combo error messages
- [ ] At least `pr list`, `pr view`, `repo list`, `commit log` verified with `--yaml` and `--template`
- [ ] `go test ./... -race` green

---

### CIS — CI Supply Chain Hardening

**Problem**: bitbottle's GitHub Actions workflows use tag-pinned actions
(`uses: actions/checkout@v4`), have no secret scanning, no SBOM, and no OpenSSF
Scorecard badge. These are table-stakes supply-chain controls for a published CLI.

**Changes** (all in `.github/workflows/`):

1. **Pin all action SHAs** — replace every `@vN` action ref with `@<full SHA>`.
   Use `step-security/harden-runner` as the first step in each job.

2. **gitleaks** — add `.github/workflows/secret-scan.yml`:
   ```yaml
   - uses: gitleaks/gitleaks-action@<SHA>
     env:
       GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   ```
   Run on every push and PR.

3. **SBOM** — add SBOM generation to the release workflow using
   `anchore/sbom-action@<SHA>`. Attach the SBOM as a release asset and
   attest it with `actions/attest-sbom@<SHA>`.

4. **OpenSSF Scorecard** — add `.github/workflows/scorecard.yml`:
   ```yaml
   - uses: ossf/scorecard-action@<SHA>
     with:
       results_file: results.sarif
       publish_results: true
   ```
   Publish results to GitHub's dependency graph. Add Scorecard badge to README.

5. **Codecov** — add `codecov/codecov-action@<SHA>` to `ci.yml` after the
   test step. Add a coverage badge to README.

6. **Skill sync check** — add a job to `ci.yml` that runs `make check-skills`
   (or equivalent) to verify `skills/SKILL.md` is consistent with implemented
   commands.

**No backend changes. No new Go code beyond Makefile targets.**

**Implementation notes**:
1. **Dependabot for SHA-pinned actions.** Tag-pinned actions get auto-bumped by
   Dependabot's default config. SHA-pinned actions do NOT — Dependabot needs
   explicit configuration in `.github/dependabot.yml`:
   ```yaml
   updates:
     - package-ecosystem: "github-actions"
       directory: "/"
       schedule: { interval: "weekly" }
   ```
   Without this, our SHA pins go stale and accumulate CVEs faster than we
   notice. Ship the Dependabot config in the same PR.
2. **SBOM timing.** The SBOM must be generated AFTER release-please creates the
   release tag but BEFORE Goreleaser uploads artifacts, so it attaches to the
   GitHub release and to npm. Add the SBOM step inside `release.yml` after
   `goreleaser release` runs, not as a separate workflow.
3. **Scorecard requires `id-token: write`.** The workflow needs
   `permissions: { id-token: write, contents: read }` at the job level.
   Forget this and Scorecard silently runs with degraded checks (no
   keyless signing verification).
4. **gitleaks licence.** gitleaks-action wraps the gitleaks binary; the binary
   is BSD-3 with a 2023 commercial-use clarification. Our public-OSS use is
   fine, but document the choice in `docs/dev/ci.md` so future contributors
   don't need to re-verify.
5. **Codecov free tier covers public repos.** Confirm bitbottle's repo is
   `Public` in Codecov's dashboard — private repos require a paid tier and
   uploads silently 401.
6. **`make check-skills` source of truth.** The skill-sync check needs a
   definition of "in sync." Recommend: parse `skills/SKILL.md` for the
   command list, compare against `pkg/cmdregistry.All(f)` output. Fail if
   any command exists in code but not skill (or vice versa).

**Definition of Done**:
- [ ] All `uses:` in all workflows reference pinned SHAs (no tag refs)
- [ ] `.github/dependabot.yml` configured for weekly action-version bumps
- [ ] `gitleaks` workflow runs and passes on main
- [ ] SBOM attached to a test release run (verified in release artifact)
- [ ] Scorecard workflow publishes with full check coverage; README badge added
- [ ] Codecov upload confirmed in a CI run; badge added to README
- [ ] `make check-skills` target exists, has a clear pass/fail contract, and is called in CI

---

### VAR — Variable Command Promotion

**Problem** _(historical, shipped)_: `pipeline variable` was nested under `pipeline` and operated only
on **repository-scoped** pipeline variables. Bitbucket Cloud has the same primitive
at three scopes: repository, workspace, and deployment-environment (the third lives
under scope **DEP**). The CLI exposes all three through one consistent
verb-noun pair via `variable {list,set,delete} --scope`, matching `gh variable`.

**Reuse the existing `PipelineClient` interface, do NOT duplicate `PipelineVariable`.**

The shipped `PipelineClient` (`api/backend/client_pipeline.go`) already has
`ListPipelineVariables` / `SetPipelineVariable` / `DeletePipelineVariable` plus
the `PipelineVariable` + `PipelineVariableInput` types. VAR extends that interface
with workspace + deployment scopes; the existing repo-scoped methods stay as-is.

**Extend `PipelineClient`** (`api/backend/client_pipeline.go`):
```go
type PipelineClient interface {
    // existing (unchanged):
    ListPipelineVariables(ns, slug string) ([]PipelineVariable, error)
    SetPipelineVariable(ns, slug string, in PipelineVariableInput) (PipelineVariable, error)
    DeletePipelineVariable(ns, slug, uuid string) error
    // ... other pipeline methods unchanged ...

    // new:
    ListWorkspaceVariables(workspace string) ([]PipelineVariable, error)
    SetWorkspaceVariable(workspace string, in PipelineVariableInput) (PipelineVariable, error)
    DeleteWorkspaceVariable(workspace, uuid string) error

    ListDeploymentVariables(ns, slug, envUUID string) ([]PipelineVariable, error)
    SetDeploymentVariable(ns, slug, envUUID string, in PipelineVariableInput) (PipelineVariable, error)
    DeleteDeploymentVariable(ns, slug, envUUID, uuid string) error
}
```

The deployment methods overlap with scope **DEP**'s `EnvVariable` / `EnvVariableInput`.
**Resolve before shipping**: DEP should use `PipelineVariable` too — collapse the
duplicate types as part of VAR. Update DEP's interface to call into the same methods.

**No new domain types.** `PipelineVariable` already has `UUID`, `Key`, `Value`,
`Secured`. (If a `Scope` field is wanted on read responses, add it as an
optional string — does not affect existing call-sites.)

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `variable list PROJECT/REPO` | 1 | — | `--scope repository\|workspace\|deployment` (default `repository`), `--env UUID` (required for deployment), `--json`, `--jq`, `--hostname` |
| `variable set PROJECT/REPO KEY VALUE` | 3 | — | `--scope`, `--env UUID`, `--secured`, `--hostname` |
| `variable delete PROJECT/REPO KEY` | 2 | — | `--scope`, `--env UUID`, `--hostname` |

For workspace scope, `PROJECT/REPO` becomes `WORKSPACE/-` (the slug is ignored).
For deployment scope, `--env UUID` is required.

**Migration**: `pipeline variable *` was deprecated (v1.31.0) and removed (v1.40.x+).
Use `variable {list,set,delete} --scope repository` instead.

**MCP tools**: extend existing `set_pipeline_variable` / `list_pipeline_variables`
/ `delete_pipeline_variable` schemas with optional `scope` + `env_uuid` fields.
No new MCP tools needed.

**Implementation notes**:
1. **Server has no workspace variables.** Workspaces are a Cloud-only primitive.
   `ListWorkspaceVariables` and friends MUST return `host.unsupported` from the
   Server adapter — don't try to back-fill from project-level variables (different
   primitive, different semantics).
2. **System-managed variables are read-only.** Bitbucket Cloud marks some
   variables `system=true` (e.g., `BITBUCKET_REPO_SLUG`, `BITBUCKET_BUILD_NUMBER`).
   These appear in `ListPipelineVariables` but `SetPipelineVariable` returns
   400 on attempts to write. Surface this cleanly: add `System bool` to
   `PipelineVariable`, and have the CLI's `variable set` print a friendly error
   if the user targets a system-managed key.
3. **Workspace path shape.** For workspace scope, the CLI takes `WORKSPACE/-`
   (the slug `-` is a placeholder that means "ignore"). Pick this over a
   `--workspace W` flag for two reasons: matches the existing pattern
   (`PROJECT/REPO` for all other commands), and means the same `variable set`
   command works on any scope by varying the positional.
4. **DEP scope dependency.** Scope **DEP** must use `PipelineVariable` /
   `PipelineVariableInput` (not a new `EnvVariable` type). Update DEP's scope
   detail when VAR lands to point at the consolidated type. Add a note in DEP
   right now.
5. **Secret values write-only.** `Value` is empty on read when `Secured=true`.
   `variable list --scope` output must NOT show empty value columns
   misleadingly; print `(secured)` in the value column for secured variables in
   table output. JSON keeps the empty string for schema stability.

---

### PERMS — Permissions Management _(Server / DC only)_

Bitbucket Server/DC exposes permission management via separate
user-grant and group-grant endpoints:
- `/rest/api/1.0/projects/{key}/permissions/users` and `.../groups`
- `/rest/api/1.0/projects/{key}/repos/{slug}/permissions/users` and `.../groups`

Cloud has no equivalent REST API (managed via workspace membership) — surface
`ErrUnsupportedOnHost`.

**New optional interface** (`api/backend/client.go`):
```go
type PermissionsClient interface {
    ListProjectPermissions(project string) ([]PermissionGrant, error)
    GrantProjectPermission(project string, subject PermissionSubject, perm string) error
    RevokeProjectPermission(project string, subject PermissionSubject) error

    ListRepoPermissions(project, slug string) ([]PermissionGrant, error)
    GrantRepoPermission(project, slug string, subject PermissionSubject, perm string) error
    RevokeRepoPermission(project, slug string, subject PermissionSubject) error
}
```

The list methods union user-grants and group-grants from the two separate
endpoints. Grant/revoke dispatch to `/users` or `/groups` based on
`PermissionSubject.Kind`.

**New types**:
```go
type PermissionSubject struct {
    Kind        string // "user" | "group"
    Slug        string // user slug (Kind=user) — empty for groups
    Name        string // group name (Kind=group) — empty for users
    DisplayName string // populated on read; ignored on write
}

type PermissionGrant struct {
    Subject    PermissionSubject
    Permission string // PROJECT_READ | PROJECT_WRITE | PROJECT_ADMIN
                      // REPO_READ   | REPO_WRITE   | REPO_ADMIN
}
```

Revoke does not take `perm` — Bitbucket Server allows only one permission level
per subject, so revoke is "remove whatever grant they have."

**Commands** (mutually exclusive `--user` / `--group`):

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `perms project list PROJECT` | 1 | — | `--json`, `--jq`, `--hostname` |
| `perms project grant PROJECT PERM` | 2 | `--user SLUG` \| `--group NAME` | `--hostname` |
| `perms project revoke PROJECT` | 1 | `--user SLUG` \| `--group NAME` | `--hostname` |
| `perms repo list PROJECT/REPO` | 1 | — | `--json`, `--jq`, `--hostname` |
| `perms repo grant PROJECT/REPO PERM` | 2 | `--user SLUG` \| `--group NAME` | `--hostname` |
| `perms repo revoke PROJECT/REPO` | 1 | `--user SLUG` \| `--group NAME` | `--hostname` |

`PERM` values: `PROJECT_READ` / `PROJECT_WRITE` / `PROJECT_ADMIN` for project;
`REPO_READ` / `REPO_WRITE` / `REPO_ADMIN` for repo.

**MCP tools**: `list_project_permissions`, `grant_project_permission`,
`revoke_project_permission`, `list_repo_permissions`, `grant_repo_permission`,
`revoke_repo_permission` — each takes `{subject_kind, subject_slug_or_name, permission}`.

**Implementation notes**:
1. **Grant is really "set."** The PUT endpoint accepts any permission level
   for an existing subject — granting READ to someone who has WRITE
   downgrades them silently. The CLI should warn before downgrading:
   "User X already has REPO_WRITE; this will change to REPO_READ. Continue?"
   (`--force` or non-TTY skips the prompt.)
2. **List unions two endpoints.** `ListRepoPermissions` calls both
   `/permissions/users` and `/permissions/groups`, paginates both via
   `paging.Collect[T]`, and returns the merged slice. Don't expose the
   split — users don't care about the URL shape.
3. **Default project permission is a separate API.** `/projects/{key}/permissions/{perm}/all`
   toggles "all users get this permission by default." Not exposed in this
   scope — document it as out of scope and link to the API for users who need
   it via `api` passthrough.
4. **Username slugs are not display names.** LDAP-integrated Server instances
   often have user slugs like `j.smith` and display names `John Smith`. The
   CLI takes slugs (`--user j.smith`) and the list output shows both. Don't
   accept display names — they're not unique.
5. **Group names may contain spaces.** URL-encode group names in the request
   path: `--group "Senior Developers"` → `%20`. Test against a group with a
   space character.
6. **403 vs 404 disambiguation.** Server returns 401 if not authenticated,
   403 if authenticated but lacking PROJECT_ADMIN, 404 if the project itself
   doesn't exist. All three currently map to similar typed errors; the
   `perms` commands should map to distinct hints:
   - 403 → "you need PROJECT_ADMIN on this project to manage permissions"
   - 404 → standard `repo.not_found` / `project.not_found`

---

### ADMIN — Admin Commands _(Server / DC only)_

Server/DC admin endpoints for secrets rotation and log-level management.
Cloud has no equivalent — surface `ErrUnsupportedOnHost`.

**New optional interface** (`api/backend/client.go`):
```go
type AdminClient interface {
    RotateSecrets() error
    GetLoggingConfig() (LoggingConfig, error)
    SetLoggingConfig(in LoggingConfigInput) error
}
```

**New types**:
```go
type LoggingConfig struct {
    Level string // DEBUG | INFO | WARN | ERROR
    Async bool
}

type LoggingConfigInput = LoggingConfig
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `admin secrets rotate` | 0 | — | `--hostname` |
| `admin logging get` | 0 | — | `--json`, `--hostname` |
| `admin logging set` | 0 | one of `--level`, `--async` | `--hostname` |

**MCP tools**: `rotate_secrets`, `get_logging_config`, `set_logging_config`

**Implementation notes**:
1. **`admin secrets rotate` is NOT user/credential rotation.** It rotates the
   cluster's internal HTTP Strict Transport Security secret used for
   inter-node authentication in DC deployments. **A rolling restart of all
   cluster nodes is required afterwards.** The CLI must print a banner before
   confirming the action:
   `"This rotates the cluster's internal HTTPS secret. ALL nodes must be
   restarted for the new secret to take effect. Continue? [y/N]"`
   `--confirm` (or non-TTY) skips the prompt.
2. **Requires SYS_ADMIN, not just PROJECT_ADMIN.** Most admin tokens lack this.
   Map 403 to a specific hint:
   `"This requires SYS_ADMIN permission. Standard admin tokens do not include
   it; the action must be performed by a system administrator."`
3. **`logging set` is non-persistent by default.** Changes apply to the running
   instance and reset on next restart. Add `--persistent` to also write to
   `log4j.properties` (separate endpoint:
   `PUT /rest/api/1.0/admin/logging/properties`). Document both modes clearly
   — users hitting "my log level reset on restart" will assume a bug.
4. **Log levels are case-sensitive.** Server only accepts `DEBUG`, `INFO`,
   `WARN`, `ERROR` (uppercase). Reject mixed-case input at the CLI layer.
5. **No equivalent on Cloud.** Both methods surface `host.unsupported` on
   Cloud. Document this in `--help` text so users don't get confused.

---

### AUTOMERGE — PR Auto-Merge _(both backends)_

Queues a PR to merge automatically once all merge checks (builds, approvals,
required reviewers) pass.

**Backend coverage**:
- **Bitbucket Server / DC**: stable `/rest/api/1.0/.../pull-requests/{id}/auto-merge` endpoint.
- **Bitbucket Cloud**: beta `/2.0/repositories/{ws}/{slug}/pullrequests/{id}/auto-merge` endpoint (currently behind workspace beta flag). Treat as available; surface a clean error if the workspace hasn't opted in.

**Extend `PRMerger`** (do not add a new optional interface):
```go
type PRMerger interface {
    // existing:
    MergePR(ns, slug string, id int, in MergePRInput) (PullRequest, error)
    // new:
    EnableAutoMerge(ns, slug string, id int, strategy string) error // merge|squash|rebase
    DisableAutoMerge(ns, slug string, id int) error
}
```
`PRMerger` is already in the composite `Client`, so both backends gain these
methods. Adding to an existing interface keeps the optional-interface count flat
(see review note 6).

**No new domain types.** Auto-merge state is surfaced via an additional field on
the existing `PullRequest` type:
```go
type PullRequest struct {
    // ... existing fields ...
    AutoMerge *AutoMergeState // nil when not enabled
}

type AutoMergeState struct {
    Enabled  bool
    Strategy string // merge | squash | rebase
}
```

**Command surface**: extend the existing `pr merge` command with a single flag,
matching `gh pr merge --auto` ergonomics. No new subcommands.

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `pr merge PR_ID --auto` | 1 | — | `--squash` \| `--rebase` (strategy; default `merge`), `--hostname` |
| `pr merge PR_ID --auto=off` | 1 | — | `--hostname` (cancels a queued auto-merge) |

Auto-merge **status** is read from the existing `pr view` output — when
`AutoMerge != nil` the view formatter prints `Auto-merge: enabled (squash)`.
No new `status` command needed; status is data, not an action.

**MCP tools**: extend existing `merge_pr` schema with `auto bool` + `auto_strategy string`.
No new MCP tools.

**Implementation notes**:
1. **Strategy name translation.** Three vocabularies in play — keep a single
   translation table in `api/backend/types.go`:
   - CLI flag: `--merge` (default) | `--squash` | `--rebase`
   - Cloud body: `merge_commit` | `squash` | `fast_forward`
   - Server body: `merge-commit` | `squash` | `fast-forward`
   Note `--rebase` maps to `fast_forward` on Cloud — that's not strictly a
   rebase (it's fast-forward only when possible). Document the semantic gap in
   `--help`.
2. **Cloud beta detection.** When Cloud's auto-merge is unavailable (workspace
   hasn't opted into beta), the endpoint returns 404 *with a specific error
   body*. Don't conflate with "PR not found." Map this specific shape to a new
   `pr.automerge.beta_disabled` error code in scope **EX**'s catalogue with
   hint: "Ask your workspace admin to enable auto-merge in workspace settings."
3. **Race with manual merge.** If auto-merge is queued and the user then runs
   `pr merge` without `--auto`, the server-side behaviour differs by backend
   (Cloud: rejects with 409; Server: cancels the auto-queue and merges
   immediately). Surface both behaviours consistently — pre-check current
   auto-merge state and prompt: "PR is queued for auto-merge. Cancel and merge
   immediately? [y/N]".
4. **Status from `pr view`.** The `AutoMerge` field on `PullRequest` must be
   populated by both adapters' `GetPR` implementation, not lazily fetched.
   Adding a second roundtrip from `pr view` to display auto-merge state is a
   regression in command latency.
5. **`pr view --json` schema.** Once `AutoMerge` is a JSON field, it's an API
   contract. Use `*AutoMergeState` (pointer) so JSON omits it when nil, not an
   empty object — keeps the contract minimal for older PRs without auto-merge state.

---

### TASK — PR Tasks _(Server / DC only)_

Server/DC PR tasks are actionable items reviewers leave on a PR. **The old
`/pull-requests/{id}/tasks` API was deprecated in Server 7.2 (2020)**. The
modern shape is "comments with `severity: BLOCKER`":

- A task is a comment with `severity=BLOCKER` and `state=OPEN|RESOLVED`.
- Resolving a task = `PUT /comments/{id}` with `{state: "RESOLVED", version: N}`.
- Tasks can be top-level on the PR (standalone) or anchored to a parent comment.

Cloud has no equivalent — surface `host.unsupported`.

**Implementation notes**:
1. **Version detection required.** Server < 7.2 still uses the legacy tasks
   endpoint and ignores `severity` on comments. Use the existing
   `ServerCapabilities.GetApplicationProperties()` to read the Server version
   string; dispatch to legacy `/tasks` API for < 7.2, modern comments API otherwise.
   This is the first scope to need version-conditional dispatch — establish the
   pattern (helper in `api/server/version.go`?) cleanly.
2. **Optimistic locking.** Server comments carry a `version` integer; mutations
   require it. The shipped `EditPRComment` already handles this — reuse, don't
   reimplement.
3. **Reuse `PRCommentClient` interfaces.** Don't add `PRTaskClient` as a separate
   interface — extend the existing `PRCommentAdder` / `PRCommentEditor` with
   `severity` and `state` fields on the input. This keeps the optional-interface
   count flat and acknowledges that tasks ARE comments at the API level.

**Extend existing types** (`api/backend/types.go`):
```go
type PRComment struct {
    // ... existing fields ...
    Severity string  // "" | "BLOCKER"  (BLOCKER comments are tasks)
    State    string  // "" | "OPEN" | "RESOLVED"  (only meaningful when Severity=BLOCKER)
    Version  int     // Server optimistic-lock token; 0 on Cloud
}

type AddPRCommentInput struct {
    // ... existing fields ...
    Severity string  // "BLOCKER" to create a task; "" for normal comment
}
```

**Extend existing interface**:
```go
type PRCommentEditor interface {
    EditPRComment(ns, slug string, id, commentID int, body string) (PRComment, error)
    // new — Server-only; clean error on Cloud:
    SetPRCommentState(ns, slug string, id, commentID int, state string) error
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `pr task list PR_ID` | 1 | — | `--state open\|resolved\|all`, `--json`, `--jq`, `--hostname` |
| `pr task create PR_ID` | 1 | `--body` | `--parent COMMENT_ID` (anchor), `--hostname` |
| `pr task resolve PR_ID TASK_ID` | 2 | — | `--hostname` |
| `pr task reopen PR_ID TASK_ID` | 2 | — | `--hostname` |

`pr task list` is a filtered view of `pr comment list` (only `Severity=BLOCKER`).
`pr task create` is a thin wrapper around `pr comment add --severity BLOCKER`.

**MCP tools**: `list_pr_tasks`, `create_pr_task`, `resolve_pr_task`, `reopen_pr_task` —
each implemented as a thin filter over the existing comment tools.

---

### REACT-PR — PR Comment Reactions _(Server / DC only)_

**Important**: Bitbucket reactions exist at the **comment** level, not the PR
level. The endpoint is `/comments/{commentId}/reactions` — there is no
PR-level reactions primitive. Cloud has no documented REST API for reactions —
surface `host.unsupported`.

**New optional interface** (`api/backend/client.go`):
```go
type CommentReactor interface {
    ListCommentReactions(commentID int) ([]CommentReaction, error)
    AddCommentReaction(commentID int, emoji string) error
    RemoveCommentReaction(commentID int, emoji string) error
}
```
Server/DC reactions are keyed only by `commentID` — the PR / commit context is
not part of the URL. The same interface is reused by REACT-COMMIT.

**New types** (`api/backend/types.go`):
```go
type CommentReaction struct {
    Emoji string  // canonical shortcode: "thumbs_up" | "thumbs_down" | "heart" | etc.
    Users []User  // all users who reacted with this emoji
}
```
List returns one row per emoji with the deduplicated user list (the API returns
one row per (emoji, user) pair; group in the adapter for ergonomic output).

**Implementation notes**:
1. **Emoji shortcode normalisation.** Server accepts `:thumbsup:` and `thumbs_up`
   inconsistently across versions. Normalise to underscore form on input;
   normalise responses back to the same form. Document the canonical set in
   README so users know what to pass.
2. **No batch API.** Listing reactions for many comments requires N calls. For
   `pr comment list --reactions`, batch-resolve reactions concurrently with a
   small worker pool (4-8); don't serialise.
3. **Permission semantics.** Adding/removing requires write access to the
   comment's repo. Anonymous users can read reactions but not write.

**Commands** (extend existing `pr comment` surface only):

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `pr comment react PR_ID COMMENT_ID` | 2 | `--emoji E` | `--hostname` |
| `pr comment unreact PR_ID COMMENT_ID` | 2 | `--emoji E` | `--hostname` |
| `pr comment list PR_ID` | 1 | — | `--reactions` (existing command grows a column), `--json`, `--jq`, `--hostname` |

**MCP tools**: `add_comment_reaction`, `remove_comment_reaction`,
`list_comment_reactions`. Extend the existing `list_pr_comments` schema with an
optional `include_reactions bool` field.

---

### REACT-COMMIT — Commit Comment Reactions _(Server / DC only)_

**Depends on**: REACT-PR (reuses `CommentReactor` interface and Server/Cloud
implementations — do not re-implement).

**Commands** (extend existing `commit comment` surface only):

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `commit comment react PROJECT/REPO HASH COMMENT_ID` | 3 | `--emoji E` | `--hostname` |
| `commit comment unreact PROJECT/REPO HASH COMMENT_ID` | 3 | `--emoji E` | `--hostname` |

**MCP tools**: extend `list_commit_comments` schema with `include_reactions bool`.
No new standalone tools needed — `add_comment_reaction` / `remove_comment_reaction`
from REACT-PR work for any `commentID` regardless of context.

---

### PROF — Named Profiles

**Problem**: users with access to multiple Bitbucket instances (e.g. company
Server + personal Cloud) must pass `--hostname` on every command. Named profiles
(à la kubectl contexts) let them switch the active credential profile globally.

**Why `profile` and not `context`**: scope **CTX** already owns the `bitbottle context`
command — it prints orientation JSON (host + repo + branch + user). Reusing the
same top-level verb for credential selection would force a default-when-bare
subcommand and confuse `bitbottle context` (info) with `bitbottle context use`
(action). `profile` is the more common term outside k8s anyway (`aws --profile`,
`gh auth switch` works on profiles internally).

**Config change** (`internal/config/`):

Add `ActiveProfile string` + `Profiles map[string]*Profile`. Each profile points
at a host; the token lives in the keyring under `bitbottle/<profile-name>`
(not in the config file — aligns with scope **SEC**).

```go
type Profile struct {
    Hostname      string
    Username      string  // stored in config; token goes to keyring
    BackendType   string  // cloud | server
    SkipTLSVerify bool
}
```

Backward compat: if `ActiveProfile` is empty, fall back to the existing
flat `Hosts` map so existing configs continue to work unchanged.

**Factory change** (`pkg/cmd/factory/`):

`f.Backend(hostname)` reads `ActiveProfile` when `hostname` is empty, resolves
the `Profile` to a `HostConfig`. Explicit `--hostname` always wins.

**Commands** (`pkg/cmd/profile/`):

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `profile create NAME` | 1 | `--hostname HOST` | `--token`, `--username`, `--skip-tls-verify` |
| `profile use NAME` | 1 | — | — |
| `profile list` | 0 | — | `--json`, `--jq` |
| `profile delete NAME` | 1 | — | — |

**MCP tools**: none (profile selection is CLI-session state, not suitable for
stateless MCP calls).

**Implementation notes**:
1. **Migration from flat `Hosts`.** Existing users have
   `hosts: {git.example.com: HostConfig}` in `hosts.yml`. On first
   `profile create`, auto-import existing hosts as profiles named after
   hostnames (`profile create git_example_com --from-host git.example.com`).
   Don't drop the legacy `Hosts` map — keep both shapes in the YAML for one
   release cycle, then deprecate.
2. **YAML schema versioning.** Add `version: 2` to `hosts.yml` on first write
   that includes `Profiles`. The loader switches on version: version 0/1
   reads the flat shape, version 2 reads both. This is the first scope to
   introduce a config schema break — establish the pattern cleanly.
3. **`--hostname` always wins.** When `--hostname` is set, the active profile
   is ignored. Document in `profile use` help text: "Sets the default for
   subsequent commands; explicit --hostname still overrides."
4. **Token storage keying.** Profile tokens go to keyring under
   `bitbottle/profile/<name>` (not `bitbottle/<hostname>`). Two profiles can
   point at the same host (e.g., two accounts on the same Server instance) —
   keying by hostname would collide.
5. **No token rotation on switch.** Switching profiles doesn't revoke the
   previous profile's token. The keyring entries for inactive profiles
   persist. Document this — users wanting to "log out" must use
   `profile delete` (which DOES remove the keyring entry).
6. **`auth login` behaviour.** Decide whether `auth login` operates on the
   active profile or creates a new one. Recommendation: `auth login --hostname H`
   updates the matching profile if one exists, else creates an unnamed
   default profile. Document this in `auth login --help`.

---

### EXT-CORE — Extension Install + List

A plugin mechanism letting the community add `bitbottle <ext-name>` subcommands
without forking core. Modelled on `gh extension`.

**Core design**:
- Extensions are single-file executables named `bitbottle-<name>` stored in
  `~/.config/bitbottle/extensions/`.
- `extension install USER/REPO` downloads the OS/arch-matched binary from GitHub
  Releases. Binary naming convention: `bitbottle-<name>-<os>-<arch>[.exe]`.
- Records install-time SHA256 in `~/.config/bitbottle/extensions/<name>.lock`
  for change detection (not a trust anchor — attacker who can publish a binary
  can publish a matching checksum; see EXT-RUNTIME for enforcement).

**New package** (`pkg/cmd/extension/`):
```go
type Extension struct {
    Name    string
    Path    string
    Version string
}

func List() ([]Extension, error)
func Install(repo, version string) error        // version="" → latest
func InstallLocal(path string) error            // symlink for extension authors
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `extension install USER/REPO` | 1 | — | `--hostname` (private repos), `--pin VERSION` |
| `extension install --local PATH` | 0 | `--local PATH` | — |
| `extension list` | 0 | — | `--json` |

**Security model**: print extension author + source URL + "this will run as you"
warning on install. Require `--confirm` or interactive y/N to proceed.

**macOS Gatekeeper**: strip `com.apple.quarantine` xattr during install with a
printed warning. Better UX than right-click→Open; documents the bypass.

**Implementation notes**:
- If no asset matches `runtime.GOOS`+`runtime.GOARCH`, surface a clean error:
  `"no binary for darwin/arm64 in USER/REPO@VERSION"` — don't silently download wrong-arch.
- `--local PATH` symlinks instead of copying; SHA lockfile records the symlink
  target's hash at link time.

**MCP tools**: none (extension management is CLI-only).

**Definition of Done**:
- [x] `extension install` downloads correct OS/arch binary, confirms with user, records SHA
- [ ] macOS quarantine xattr stripped at install with visible warning
- [x] `extension list` shows installed extensions with version + path
- [x] `extension install --local PATH` symlinks for development
- [x] README section: trust model, install flow, binary naming convention

---

### EXT-RUNTIME — Extension Exec _(depends on EXT-CORE)_

**`extension exec`** and the root-level dispatch hook that makes installed
extensions first-class subcommands.

**Root dispatch**: before Cobra parses args, if `os.Args[1]` is not a known
command and `~/.config/bitbottle/extensions/bitbottle-<name>` exists, exec that
binary. This makes `bitbottle <extname> [args...]` work without `extension exec`.

**Security enforcement**:
- Verify binary SHA256 matches the lockfile from install. If mismatch: refuse to
  run, print `"extension binary has changed since install — re-install to continue"`.
- Strip `BITBOTTLE_KEYRING_PASSPHRASE` and `GITHUB_TOKEN` from subprocess env.
- Inject `BITBOTTLE_HOST`, `BITBOTTLE_USER`, `BITBOTTLE_TOKEN` (read fresh from
  keyring for the active host).

**Argv passthrough**: extensions receive `os.Args[2:]`. Global flags
(`--debug`, `--hostname`, `--no-color`) are NOT auto-injected — extensions read
them from env vars (`BITBOTTLE_DEBUG=1`).

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `extension exec NAME [args...]` | 1+ | — | — |

**Definition of Done**:
- [x] `extension exec` forks with sanitised+injected env (KEYRING_PASSPHRASE/PASSWORD stripped; BB_TOKEN + BITBOTTLE_VERSION injected)
- [ ] Root dispatch hook wired into cobra `PersistentPreRunE` or `RunE` fallback
- [ ] Non-token auth (basic/app-password) still injects via `BITBOTTLE_TOKEN`
- [ ] Test: SHA mismatch returns clear error, not a panic

---

### EXT-MGMT — Extension Upgrade + Remove _(depends on EXT-CORE)_

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `extension upgrade NAME` | 1 | — | — |
| `extension upgrade --all` | 0 | `--all` | — |
| `extension remove NAME` | 1 | — | — |

**Upgrade flow**: check installed extension's recorded version against the latest
GitHub Release tag. If newer: download, replace binary, update lockfile SHA.
Print `"already up to date"` if version matches.

**Remove flow**: delete binary + lockfile. Print confirmation.

**Definition of Done**:
- [x] `extension upgrade NAME` upgrades single extension
- [x] `extension upgrade --all` upgrades all installed extensions
- [x] `extension remove` deletes binary + lockfile cleanly
- [x] `extension upgrade` on a `--local` extension prints `"local install — skipping"`

---

## Implementation Order

| Order | Scope | Rationale |
|---|---|---|
| 1 | **L** Branch Create + Checkout | Extends existing package; trivial |
| 2 | **E** Tags | New domain template; both backends |
| 3 | **G** PR Lifecycle | Extends existing pr; no new types |
| 4 | **M** Completion | Zero backend work; high DX value |
| 5 | **P** Auth Extras | Small; high parity with gh |
| 6 | **F** Commits | New domain; both backends |
| 7 | **H** Pipeline Depth | Cloud-only; extends pipeline |
| 8 | **T** Output DX | Pager + color; cross-cutting |
| 9 | **I** Webhooks | New domain; both backends |
| 10 | **J** PR Comments | New domain; both backends |
| 11 | **K** Commit Statuses | Extends F |
| 12 | **Q** Repo Extras | Fork/rename/archive/set-default |
| 13 | **U** Config | Config subcommand |
| 14 | **V** API Passthrough | Raw escape hatch |
| 15 | **N** Workspace/Projects | Cloud only; lower priority |
| 16 | **O** Issues | Cloud only; many teams use Jira |
| 17 | **BP** Branch Protect | Closes the last `branch` gap; Server/DC only |
| 18 | **EX** Error UX | Cross-cutting; do once and every command benefits |
| 19 | **RV** Code-Review Primitives | Highest agent leverage; unlocks code-review bots; split into RV1-RV6 sub-PRs |
| 20 | **SR** Code Search | Cloud-only; agent "find references" primitive; small interface, single endpoint |
| 21 | **CTX** Context Primitive | One-PR DX win; replaces 3-call orientation with 1; high MCP value |
| 22 | **GHP** gh-Parity Gaps | Bundle of one-PR wins (`pr checks`, `pr update-branch`, `pr reopen`, `pr status`, `status`, `browse`, `pipeline watch`) |
| 23 | **OF** Issues Finish | Closes the gap left by scope O; Cloud-only; APIs all exist |
| 24 | **CI** Code Insights | Server/DC only; separate REST namespace; required for CI-integration story on Server |
| 25 | **DEP** Deployments | Cloud-only operational scope; lowest priority unless requested |
| 26 | **OUT2** Extended Output Formats | gh-parity: `--template`, global `--json`/`--jq`/`--yaml` on every command. Scripts written against gh muscle memory port directly. |
| 27 | **AUTOMERGE** PR Auto-Merge | gh-parity: `gh pr merge --auto` is core gh muscle memory. Bitbucket DC has the primitive — expose it cleanly. |
| 28 | **VAR** Variable Command Promotion | gh-parity: `gh variable` is top-level. Mirror with `--scope repository\|workspace\|deployment`. |
| 29 | **PROF** Named Profiles | gh-parity: `gh auth switch` + multi-account. Real user pain for anyone running work Server + personal Cloud. |
| 30 | **EXT** Extension System | gh-parity: signature gh feature — `gh-dash`, `gh-copilot` show what an ecosystem looks like. Offloads the long tail of features to the community. |
| 31 | **REACT** Comment Reactions | gh-parity: `gh pr comment --reaction`. Reactions live on comments (not on PRs) per Bitbucket's API — completes the comment surface. |
| 32 | **TASK** PR Tasks | gh-adjacent: gh has no tasks (GitHub lacks the primitive) but tasks fit gh philosophy of exposing platform-native verbs cleanly. DC only. |
| 33 | **SEC** Secret Store & Config Security | Hygiene: token-never-in-file + keyring hardening. gh ships this; you should too. |
| 34 | **HTTPH** HTTP Client Hardening | Hygiene: retry + rate limiting + ETag cache. gh has all three. |
| 35 | **CIS** CI Supply Chain Hardening | Hygiene: SHA-pin actions, SBOM, Scorecard, gitleaks, Codecov. gh ships SBOMs and has a Scorecard badge. |
| 37 | **PERMS** Permissions Management | DC-only extra (gh has no equivalent — covered by `gh api`). Bitbucket-native, not parity-driven. |
| 38 | **ADMIN** Admin Commands | DC-only ops extra (no gh analogue). Ship only if a real ops user asks. |
| 39 | **DEPLOY-KEY** Deploy Key Management | CI/CD primitive: deploy-key list/add/delete per-repo. Both backends. Small, self-contained. |
| 40 | **PIPE-TRIGGER** Pipeline Trigger | Automation: manually trigger a Cloud pipeline with optional variables. Complements pipeline watch/view. |
| 41 | **DIFF** Diff Between Refs | Script automation: compare two branches/commits/tags without checkout. Both backends. |
| 42 | **PR-TEMPLATE** PR Description Templates | DX: manage PR templates per-repo. Cloud only initially. |
| 43 | **DEFAULT-REVIEWERS** PR Default Reviewers | DevOps automation: set/list per-repo default PR reviewers. Both backends. |
| 44 | **SSH-KEYS** User SSH Key Management | User onboarding: manage SSH keys for the current user. Cloud only. |
| 45 | **REPO-TRANSFER** Repository Transfer | Admin: move a repo to another project/workspace. Both backends. |
| 46 | **BRANCH-RULE** Cloud Branch Restriction Rules | DevOps: automate branch protection rules. Cloud only (Server uses `branch protect`). |
| 47 | **PIPELINE-SCHEDULE** Pipeline Schedules | CI automation: manage scheduled pipeline runs. Cloud only. |
| 48 | **COMMIT-FILE** Files Changed in a Commit | Agent primitive: list changed files per commit. Both backends. |
| 49 | **PR-COMMITS** PR Commit List | Agent primitive: list commits in a PR. Both backends. Complements `pr view` and diff tools. |
| 50 | **PR-FILES** PR Changed Files | Agent primitive: list files changed in a PR. Both backends. Enables agent code-review workflows. |
| 51 | **REPO-WATCHER** Repository Watchers | Discovery: list who's watching a repo. Both backends. |
