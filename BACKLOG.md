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
| `repo fork` | ❌ | Cloud only |
| `repo rename` | ❌ | Both backends |
| `repo archive` | ❌ | Cloud only |
| `repo set-default` | ❌ | Write `PROJECT/REPO` to `.git/config`; enables arg-free commands |

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
| `pr comment list` | ✅ | List general comments |
| `pr comment add` | ✅ | Add a general comment |

### Branch

| Command | Status | Notes |
|---|---|---|
| `branch list` | ✅ | |
| `branch delete` | ✅ | |
| `branch create` | ✅ | |
| `branch checkout` | ✅ | Thin wrapper: `git fetch origin BRANCH && git checkout BRANCH` |
| `branch protect` | ❌ | Branch restrictions; Server/DC only |

### Pipeline _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `pipeline list` | ✅ | |
| `pipeline view` | ✅ | |
| `pipeline run` | ✅ | |
| `pipeline steps` | ❌ | List steps in a pipeline |
| `pipeline logs` | ❌ | Stream step log |
| `pipeline variable list` | ❌ | |
| `pipeline variable set` | ❌ | |
| `pipeline variable delete` | ❌ | |

### Commits

| Command | Status | Notes |
|---|---|---|
| `commit log` | ✅ | List commits on a branch |
| `commit view` | ✅ | View a single commit |
| `commit status` | ✅ | List build statuses for a commit hash |

### Tags

| Command | Status | Notes |
|---|---|---|
| `tag list` | ✅ | |
| `tag create` | ✅ | |
| `tag delete` | ✅ | |

### Webhooks

| Command | Status | Notes |
|---|---|---|
| `webhook list` | ❌ | |
| `webhook view` | ❌ | |
| `webhook create` | ❌ | |
| `webhook delete` | ❌ | |

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
| `$PAGER` support | ❌ | `StartPager` is a no-op stub in iostreams |
| Color output | ❌ | `colorEnabled` set but never used in formatters |

### Workspace / Projects _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `workspace list` | ❌ | |
| `project list WORKSPACE` | ❌ | |

### Issues _(Cloud only)_

| Command | Status | Notes |
|---|---|---|
| `issue list` | ❌ | |
| `issue view` | ❌ | |
| `issue create` | ❌ | |
| `issue close` | ❌ | |

---

## Backlog

| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| L | **Branch Create + Checkout** | `branch create`, `branch checkout` | Both | 1 | ✅ |
| E | **Tags** | `tag list`, `tag create`, `tag delete` | Both | 1 | ✅ |
| G | **PR Lifecycle** | `pr decline`, `pr unapprove`, `pr edit`, `pr ready`, `pr request-review`, `pr request-changes` | Both / Cloud | 1 | ✅ |
| M | **Shell Completion** | `completion bash\|zsh\|fish\|powershell` | N/A | DX | ✅ |
| P | **Auth Extras** | `auth token`, `auth refresh` | N/A | DX | ✅ |
| Q | **Repo Extras** | `repo fork`, `repo rename`, `repo archive`, `repo set-default` | Both / Cloud | 2 | 🔲 |
| F | **Commits** | `commit log`, `commit view` | Both | 1 | ✅ |
| H | **Pipeline Depth** | `pipeline steps`, `pipeline logs`, `pipeline variable *` | Cloud | 1 | 🔲 |
| I | **Webhooks** | `webhook list`, `webhook view`, `webhook create`, `webhook delete` | Both | 2 | 🔲 |
| J | **PR Comments** | `pr comment list`, `pr comment add` | Both | 2 | ✅ |
| K | **Commit Statuses** | `commit status` | Both | 2 | ✅ |
| T | **Output DX** | pager (`$PAGER`), color output | N/A | DX | 🔲 |
| U | **Config** | `config list`, `config get`, `config set` | N/A | 2 | ✅ |
| V | **API Passthrough** | `api PATH` | Both | 2 | ✅ |
| N | **Workspace / Projects** | `workspace list`, `project list` | Cloud | 3 | 🔲 |
| O | **Issues** | `issue list`, `issue view`, `issue create`, `issue close` | Cloud | 3 | 🔲 |

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
