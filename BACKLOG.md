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

## Backlog

| ID | Scope | Commands | Backends | Tier | Status |
|---|---|---|---|---|---|
| L | Branch Create | `branch create` | Both | 1 | 🔲 |
| E | Tags | `tag list`, `tag create`, `tag delete` | Both | 1 | 🔲 |
| G | PR Lifecycle | `pr decline`, `pr unapprove`, `pr request-changes` | Both / Cloud | 1 | 🔲 |
| M | Shell Completion | `completion bash\|zsh\|fish\|powershell` | N/A | DX | 🔲 |
| F | Commits | `commit log`, `commit view` | Both | 1 | 🔲 |
| H | Pipeline Depth | `pipeline steps`, `pipeline logs`, `pipeline variable *` | Cloud | 1 | 🔲 |
| I | Webhooks | `webhook list`, `webhook create`, `webhook delete`, `webhook view` | Both | 2 | 🔲 |
| J | PR Comments | `pr comment list`, `pr comment add` | Both | 2 | 🔲 |
| K | Commit Statuses | `commit status` | Both | 2 | 🔲 |
| N | Workspace / Projects | `workspace list`, `project list` | Cloud | 3 | 🔲 |
| O | Issues | `issue list`, `issue view`, `issue create`, `issue close` | Cloud | 3 | 🔲 |

---

## Scope Details

### L — Branch Create

**New interface** (`api/backend/client.go`):
```go
type BranchCreator interface {
    CreateBranch(ns, slug string, in CreateBranchInput) (Branch, error)
}
```
Add `BranchCreator` to the composite `Client` interface.

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

**MCP tools**: `create_branch`

---

### E — Tags

**New interfaces**:
```go
type TagLister  interface { ListTags(ns, slug string, limit int) ([]Tag, error) }
type TagCreator interface { CreateTag(ns, slug string, in CreateTagInput) (Tag, error) }
type TagDeleter interface { DeleteTag(ns, slug, name string) error }
```
All three go in the composite `Client` interface (both backends support tags).

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
    Message string // empty = lightweight tag; non-empty = annotated
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

No new types needed. Three new interfaces:

```go
type PRDecliner         interface { DeclinePR(ns, slug string, id int) error }
type PRUnapprover       interface { UnapprovePR(ns, slug string, id int) error }
type PRChangesRequester interface { RequestChangesPR(ns, slug string, id int) error }
```

`DeclinePR` and `UnapprovePR` → composite `Client` (both backends).
`PRChangesRequester` → Cloud-only optional interface (type assertion, not in `Client`).

**Commands**:

| Command | Args | Flags |
|---|---|---|
| `pr decline PR_ID` | 1 | `--hostname` |
| `pr unapprove PR_ID` | 1 | `--hostname` |
| `pr request-changes PR_ID` | 1 | `--hostname` _(Cloud only — error if Server)_ |

**MCP tools**: `decline_pr`, `unapprove_pr`

---

### M — Shell Completion

No backend changes. Single file `pkg/cmd/completion/completion.go`.
Cobra provides `GenBashCompletion`, `GenZshCompletion`, `GenFishCompletion`, `GenPowerShellCompletion`.

```go
rootCmd.AddCommand(&cobra.Command{
    Use:       "completion [bash|zsh|fish|powershell]",
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    Args:      cobra.ExactValidArgs(1),
    RunE:      func(...) { /* dispatch to cobra gen method */ },
})
```

**No MCP tool needed.**

---

### F — Commits

**New interfaces**:
```go
type CommitLister interface { ListCommits(ns, slug, branch string, limit int) ([]Commit, error) }
type CommitReader interface { GetCommit(ns, slug, hash string) (Commit, error) }
```
Both go in the composite `Client` interface.

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

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `pipeline steps PROJECT/REPO UUID` | 2 | — | `--json`, `--jq`, `--hostname` |
| `pipeline logs PROJECT/REPO PIPELINE-UUID STEP-UUID` | 3 | — | `--hostname` |
| `pipeline variable list PROJECT/REPO` | 1 | — | `--json`, `--jq`, `--hostname` |
| `pipeline variable set PROJECT/REPO KEY VALUE` | 3 | — | `--secured`, `--hostname` |
| `pipeline variable delete PROJECT/REPO KEY` | 2 | — | `--hostname` |

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
All go in the composite `Client` interface.

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
Both go in the composite `Client` interface.

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
Goes in composite `Client`.

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

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `commit status PROJECT/REPO HASH` | 2 | — | `--json`, `--jq`, `--hostname` |

Note: lives in the `commit` package alongside `commit log` / `commit view` (Scope F). Implement after F.

**MCP tools**: `list_commit_statuses`

---

### N — Workspace / Projects _(Cloud only)_

**New interfaces** (Cloud-only optional, not in composite `Client`):
```go
type WorkspaceClient interface {
    ListWorkspaces(limit int) ([]Workspace, error)
    ListProjects(workspace string, limit int) ([]Project, error)
}
```

**New types**:
```go
type Workspace struct {
    Slug string
    Name string
}

type Project struct {
    Key  string
    Name string
}
```

**Commands**:

| Command | Args | Optional flags |
|---|---|---|
| `workspace list` | 0 | `--limit`, `--json`, `--jq`, `--hostname` |
| `project list WORKSPACE` | 1 | `--limit`, `--json`, `--jq`, `--hostname` |

**MCP tools**: `list_workspaces`, `list_projects`

---

### O — Issues _(Cloud only)_

**New interfaces** (Cloud-only optional):
```go
type IssueClient interface {
    ListIssues(ns, slug, status string, limit int) ([]Issue, error)
    GetIssue(ns, slug string, id int) (Issue, error)
    CreateIssue(ns, slug string, in CreateIssueInput) (Issue, error)
    UpdateIssue(ns, slug string, id int, in UpdateIssueInput) (Issue, error)
}
```

**New types**:
```go
type Issue struct {
    ID       int
    Title    string
    Status   string // new | open | resolved | on hold | invalid | duplicate | wontfix | closed
    Priority string
    Author   User
    WebURL   string
}

type CreateIssueInput struct {
    Title    string
    Content  string
    Priority string // trivial | minor | major | critical | blocker
}

type UpdateIssueInput struct {
    Status   string
    Priority string
}
```

**Commands**:

| Command | Args | Required flags | Optional flags |
|---|---|---|---|
| `issue list PROJECT/REPO` | 1 | — | `--status`, `--limit`, `--json`, `--jq`, `--hostname` |
| `issue view PROJECT/REPO ID` | 2 | — | `--web`, `--json`, `--hostname` |
| `issue create PROJECT/REPO` | 1 | `--title` | `--body`, `--priority`, `--hostname` |
| `issue close PROJECT/REPO ID` | 2 | — | `--hostname` |

**MCP tools**: `list_issues`, `get_issue`, `create_issue`, `close_issue`

---

## Implementation Order

Ship complete verticals — one scope at a time, not horizontal slices.

| Order | Scope | Rationale |
|---|---|---|
| 1 | **L** Branch Create | Extends existing package; trivial backend change |
| 2 | **E** Tags | Pure git-ref domain; good template for new domains |
| 3 | **G** PR Lifecycle | Extends existing pr package; no new types |
| 4 | **M** Completion | Zero backend work; high DX value |
| 5 | **F** Commits | New domain; both backends |
| 6 | **H** Pipeline Depth | Cloud-only; extends existing pipeline |
| 7 | **I** Webhooks | New domain; both backends |
| 8 | **J** PR Comments | New domain; both backends |
| 9 | **K** Commit Statuses | Extends commit domain from F |
| 10 | **N** Workspace/Projects | Cloud-only; lower priority |
| 11 | **O** Issues | Cloud-only; many teams use Jira |
