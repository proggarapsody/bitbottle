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
| `pr review --approve\|--request-changes\|--comment --body --inline ...` | 🔲 | Compound review in one call (gh parity) — scope **RV** |
| `pr activity PR_ID` | 🔲 | PR event stream (`/pullrequests/{id}/activity`) — scope **RV** |
| `pr checks PR_ID` | 🔲 | CI status for PR head commit — scope **GHP** |
| `pr update-branch PR_ID` | 🔲 | Sync PR head with base (the action our `pr.merge.behind` hint promises) — scope **GHP** |
| `pr reopen PR_ID` | ✅ | Reverse `pr decline` (Bitbucket Server / DC only — Cloud has no reopen primitive, BCLOUD-23807) |
| `pr status` | 🔲 | Cross-repo "PRs on my plate" (assigned / review-requested / mine) — scope **GHP** |

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
| `pipeline variable list` | ✅ | |
| `pipeline variable set` | ✅ | Upsert by KEY; `--body=-` reads stdin |
| `pipeline variable delete` | ✅ | By KEY (UUID lookup is internal) |
| `pipeline watch UUID` | 🔲 | Poll until terminal state, stream step transitions — scope **GHP** |

### Commits

| Command | Status | Notes |
|---|---|---|
| `commit log` | ✅ | List commits on a branch |
| `commit view` | ✅ | View a single commit |
| `commit status` | ✅ | List build statuses for a commit hash |
| `commit comment list / add / edit / delete` | 🔲 | Cloud `/commit/{hash}/comments` (CRUD + inline) — scope **RV** |

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
| `issue edit` | 🔲 | `PUT /issues/{id}` — title/body/kind/priority/state — scope **OF** |
| `issue reopen` | 🔲 | Reverse of `issue close` (state transition) — scope **OF** |
| `issue assign` | 🔲 | Set assignee (PUT with `assignee.username`) — scope **OF** |
| `issue comment list / add / edit / delete` | 🔲 | `/issues/{id}/comments` CRUD — scope **OF** |

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

### Deployments _(Cloud only — missing)_

| Command | Status | Notes |
|---|---|---|
| `deployment list` | 🔲 | `GET /repositories/{ws}/{slug}/deployments` — scope **DEP** |
| `deployment view UUID` | 🔲 | `GET .../deployments/{uuid}` — scope **DEP** |
| `environment list` | 🔲 | `GET .../environments` — scope **DEP** |
| `environment create` | 🔲 | `POST .../environments` (body: type/uuid/name) — scope **DEP** |
| `environment delete UUID` | 🔲 | `DELETE .../environments/{uuid}` — scope **DEP** |
| `environment variable list / set / delete` | 🔲 | `/deployments_config/environments/{uuid}/variables` (CRUD) — scope **DEP** |

### Code Insights _(Server / DC only — missing)_

| Command | Status | Notes |
|---|---|---|
| `code-insights report list / view / set / delete` | 🔲 | `/rest/insights/1.0/projects/{k}/repos/{s}/commits/{hash}/reports/{key}` — PASS/FAIL summary attached to commit/PR — scope **CI** |
| `code-insights annotation add / list / delete` | 🔲 | Bulk-attach line-level findings to a report — scope **CI** |
| `code-insights merge-check set / get / delete` | 🔲 | Configure required reports as merge gates (`/rest/insights/latest/.../merge-check/{key}`; partly undocumented) — scope **CI** |

### Context / Orientation _(missing)_

| Command | Status | Notes |
|---|---|---|
| `bitbottle context [--json]` | ✅ | One-call orientation: host + repo + branch + user + scopes + default-branch + ahead/behind. Replaces 3-4 calls for agents. — scope **CTX** |
| `bitbottle status` | 🔲 | Cross-repo "what's on my plate": review requests, mentions, assigned issues. (gh `status` analogue, workspace-scoped on Cloud.) — scope **GHP** |

### Top-level / Web

| Command | Status | Notes |
|---|---|---|
| `bitbottle browse [PATH\|NUMBER]` | 🔲 | Unified web shortcut (commit/PR/issue/path) — scope **GHP** |

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
| H | **Pipeline Depth** | `pipeline steps`, `pipeline logs`, `pipeline variable *` | Cloud | 1 | ✅ |
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
| RV | **Code-Review Primitives** | `repo file get`, `repo tree`, `pr review`, `pr comment {add\|edit\|delete\|reply\|resolve} --inline`, `pr comment list --inline`, `pr activity`, `commit comment *` | Both | 1 | 🟡 RV1+RV2+RV3 ✅; RV4-RV6 🔲 |
| SR | **Code Search** | `search code QUERY [--workspace W]` | Cloud | 2 | ✅ |
| CTX | **Context Primitive** | `context --json` (one-call orientation: host + repo + branch + user + scopes + default-branch + ahead/behind) | N/A | DX | ✅ |
| GHP | **gh-Parity Gaps** | `pr checks`, `pr update-branch`, `pr reopen`, `pr status`, `status`, `browse`, `pipeline watch` | Both | 2 | 🔲 |
| OF | **Issues Finish** | `issue edit`, `issue reopen`, `issue assign`, `issue comment {list\|add\|edit\|delete}` | Cloud | 3 | 🔲 |
| CI | **Code Insights** | `code-insights report *`, `code-insights annotation *`, `code-insights merge-check *` | Server/DC | 2 | 🔲 |
| DEP | **Deployments** | `deployment list/view`, `environment list/create/delete`, `environment variable {list\|set\|delete}` | Cloud | 3 | 🔲 |

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
| `pipeline variable list PROJECT/REPO` | 1 | `--json`, `--jq`, `--hostname` |
| `pipeline variable set PROJECT/REPO KEY VALUE` | 3 | `--secured`, `--hostname` |
| `pipeline variable delete PROJECT/REPO KEY` | 2 | `--hostname` |

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

### OF — Issues Finish _(Cloud only)_

Closes the gap left after scope **O**. Today users can list/view/create/close
issues but cannot *edit*, *comment on*, *reopen*, or *assign* them.

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
