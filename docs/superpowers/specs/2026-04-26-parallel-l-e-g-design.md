# Parallel Implementation: Scopes L, E, G

**Date:** 2026-04-26  
**Scopes:** L (Branch Create + Checkout), E (Tags), G (PR Lifecycle)  
**Strategy:** Pre-stage shared backend additions in one commit, then run three agents in parallel on separate worktrees.

---

## Pre-staging Commit (shared setup — done before any parallel work)

Both `api/backend/client.go` and `api/backend/types.go` need additive changes from all three scopes. A single commit on `main` lands these before worktrees branch off, eliminating file conflicts.

### `api/backend/types.go` additions

```go
// Scope L
type CreateBranchInput struct {
    Name    string
    StartAt string // branch name or commit hash
}

// Scope E
type Tag struct {
    Name   string
    Hash   string
    WebURL string
}

type CreateTagInput struct {
    Name    string
    StartAt string // branch name or commit hash
    Message string // empty = lightweight tag; non-empty = annotated tag
}

// Scope G
type UpdatePRInput struct {
    Title       string // empty = no change
    Description string // empty = no change
}
```

### `api/backend/client.go` additions

New interfaces — all added to composite `Client` except `PRChangesRequester`:

```go
// Scope L
type BranchCreator interface {
    CreateBranch(ns, slug string, in CreateBranchInput) (Branch, error)
}

// Scope E
type TagLister interface {
    ListTags(ns, slug string, limit int) ([]Tag, error)
}
type TagCreator interface {
    CreateTag(ns, slug string, in CreateTagInput) (Tag, error)
}
type TagDeleter interface {
    DeleteTag(ns, slug, name string) error
}

// Scope G
type PREditor interface {
    UpdatePR(ns, slug string, id int, in UpdatePRInput) (PullRequest, error)
}
type PRDecliner interface {
    DeclinePR(ns, slug string, id int) error
}
type PRUnapprover interface {
    UnapprovePR(ns, slug string, id int) error
}
type PRReadier interface {
    ReadyPR(ns, slug string, id int) error // draft → open
}
type PRReviewRequester interface {
    RequestReview(ns, slug string, id int, users []string) error
}

// Cloud-only optional interface (NOT in composite Client — access via type assertion)
type PRChangesRequester interface {
    RequestChangesPR(ns, slug string, id int) error
}
```

Composite `Client` embeds: `BranchCreator`, `TagLister`, `TagCreator`, `TagDeleter`, `PREditor`, `PRDecliner`, `PRUnapprover`, `PRReadier`, `PRReviewRequester`.

---

## Scope L — Branch Create + Checkout

### File Ownership

```
api/cloud/branch.go          (add CreateBranch — file already exists)
api/server/branch.go         (add CreateBranch — file already exists)
pkg/cmd/branch/create.go     (new)
pkg/cmd/branch/checkout.go   (new)
pkg/cmd/mcp/handlers.go      (add createBranch handler)
pkg/cmd/mcp/tools.go         (register create_branch tool)
```

### Backend API

| Backend | Endpoint | Body |
|---|---|---|
| Cloud | `POST /2.0/repositories/{ws}/{slug}/refs/branches` | `{"name":"..","target":{"hash":"<start_at>"}}` |
| Server | `POST /rest/api/1.0/projects/{key}/repos/{slug}/branches` | `{"name":"..","startPoint":"<start_at>"}` |

Cloud requires `target.hash` to be a commit hash; if `--start-at` is a branch name, resolve to hash first via `GET .../refs/branches/{name}`. Server accepts both branch names and hashes in `startPoint`.

### Commands

**`branch create PROJECT/REPO NAME --start-at REF`**

- `--start-at` is required
- On success: `Created branch NAME` to stdout
- Flags: `--start-at` (required), `--hostname`

**`branch checkout NAME`**

- No backend call — thin git wrapper (same pattern as `pr checkout`)
- Detects repo from `.git/config` remote; `--hostname` overrides host detection
- Logic:
  1. `git fetch origin NAME`
  2. If branch exists locally: `git checkout NAME`
  3. Else: `git checkout -b NAME --track origin/NAME`
- Flags: `--hostname`

### MCP Tool

**`create_branch`** — params: `project` (required), `slug` (required), `name` (required), `start_at` (required), `hostname` (optional).

### Tests

- `api/cloud/branch_test.go` — unit test for `CreateBranch` JSON decode (Cloud response shape)
- `api/server/branch_test.go` — unit test for `CreateBranch` JSON decode (Server response shape)
- `pkg/cmd/branch/create_test.go` — command unit test with stub backend
- `pkg/cmd/branch/checkout_test.go` — command test with mock git runner (same pattern as `pr/checkout_test.go`)
- `pkg/cmd/mcp/handlers_test.go` — handler test for `createBranch`

---

## Scope E — Tags

### File Ownership

```
api/cloud/tag.go             (new)
api/server/tag.go            (new)
pkg/cmd/tag/tag.go           (new — cobra group)
pkg/cmd/tag/list.go          (new)
pkg/cmd/tag/create.go        (new)
pkg/cmd/tag/delete.go        (new)
pkg/cmd/mcp/handlers.go      (add listTags, createTag, deleteTag handlers)
pkg/cmd/mcp/tools.go         (register list_tags, create_tag, delete_tag tools)
pkg/cmd/root/root.go         (register tag command group)
```

### Backend API

**List:**

| Backend | Endpoint |
|---|---|
| Cloud | `GET /2.0/repositories/{ws}/{slug}/refs/tags` (cursor pagination) |
| Server | `GET /rest/api/1.0/projects/{key}/repos/{slug}/tags` (keyset pagination) |

**Create:**

| Backend | Endpoint | Body |
|---|---|---|
| Cloud | `POST /2.0/repositories/{ws}/{slug}/refs/tags` | `{"name":"..","target":{"hash":"<start_at>"},"message":".."}` |
| Server | `POST /rest/api/1.0/projects/{key}/repos/{slug}/tags` | `{"name":"..","startPoint":"<start_at>","message":".."}` |

**Delete:**

| Backend | Endpoint |
|---|---|
| Cloud | `DELETE /2.0/repositories/{ws}/{slug}/refs/tags/{name}` |
| Server | `DELETE /rest/api/1.0/projects/{key}/repos/{slug}/tags/{name}` |

### Commands

**`tag list PROJECT/REPO`**

- TTY output: aligned table, columns `NAME  HASH  MESSAGE` (MESSAGE empty for lightweight tags, first line for annotated)
- Piped: tab-separated, no header
- Flags: `--limit` (default 30), `--json`, `--jq`, `--web` (open tags page in browser), `--hostname`

**`tag create PROJECT/REPO NAME --start-at REF`**

- `--start-at` is required
- `--message` optional; non-empty creates an annotated tag
- On success: print tag web URL to stdout
- Flags: `--start-at` (required), `--message`, `--hostname`

**`tag delete PROJECT/REPO NAME`**

- TTY: confirmation prompt before deleting; `--confirm` skips prompt
- Silent success (exit 0, no output)
- Flags: `--confirm`, `--hostname`

### MCP Tools

- **`list_tags`** — params: `project`, `slug`, `limit` (optional, default 30), `hostname` (optional)
- **`create_tag`** — params: `project`, `slug`, `name`, `start_at`, `message` (optional), `hostname` (optional)
- **`delete_tag`** — params: `project`, `slug`, `name`, `hostname` (optional)

### Tests

- `api/cloud/tag_test.go` — unit tests for list/create/delete JSON decode
- `api/server/tag_test.go` — unit tests for list/create/delete JSON decode
- `pkg/cmd/tag/list_test.go`, `create_test.go`, `delete_test.go` — command unit tests
- `pkg/cmd/mcp/handlers_test.go` — handler tests for all three tag tools

---

## Scope G — PR Lifecycle

### File Ownership

```
api/cloud/pr_lifecycle.go       (new)
api/server/pr_lifecycle.go      (new)
pkg/cmd/pr/edit.go              (new)
pkg/cmd/pr/decline.go           (new)
pkg/cmd/pr/unapprove.go         (new)
pkg/cmd/pr/ready.go             (new)
pkg/cmd/pr/request_review.go    (new)
pkg/cmd/pr/request_changes.go   (new)
pkg/cmd/pr/pr.go                (register 6 new subcommands)
pkg/cmd/mcp/handlers.go         (add 5 new handlers)
pkg/cmd/mcp/tools.go            (register 5 new tools)
```

### Backend API

| Op | Cloud | Server |
|---|---|---|
| edit | `PUT /2.0/repositories/{ws}/{slug}/pullrequests/{id}` body `{"title":"..","description":".."}` | `PUT /rest/api/1.0/projects/{key}/repos/{slug}/pull-requests/{id}` same body |
| decline | `POST .../pullrequests/{id}/decline` | `POST .../pull-requests/{id}/decline` |
| unapprove | `DELETE .../pullrequests/{id}/approve` | `DELETE .../pull-requests/{id}/participants/~` |
| ready | `PUT .../pullrequests/{id}` body `{"draft":false}` | same |
| request-review | `POST .../pullrequests/{id}/participants` body `{"user":{"account_id":".."},"role":"REVIEWER"}` (one request per reviewer) | `GET` PR first, merge new usernames into `reviewers` array, `PUT` full PR body back (Server has no single-participant-add endpoint) |
| request-changes | `POST .../pullrequests/{id}/request-changes` | ❌ returns error: `"request-changes is not supported on Bitbucket Server/DC"` |

### Commands

All commands resolve `PROJECT/REPO` from `.git/config` when not provided as argument (via `f.ResolveRef`).

**`pr edit PR_ID`**
- Requires at least one of `--title` or `--body`; errors if neither provided
- On success: `Updated pull request #42` + PR URL
- Flags: `--title`, `--body`, `--hostname`

**`pr decline PR_ID`**
- No confirmation prompt (not a data-deletion operation)
- On success: `Declined pull request #42`
- Flags: `--hostname`

**`pr unapprove PR_ID`**
- On success: `Removed approval from pull request #42`
- Flags: `--hostname`

**`pr ready PR_ID`**
- On success: `Marked pull request #42 as ready for review` + PR URL
- Flags: `--hostname`

**`pr request-review PR_ID --reviewer USERS`**
- `--reviewer` is required; accepts comma-separated usernames (e.g. `--reviewer alice,bob`)
- On success: `Requested review on pull request #42`
- Flags: `--reviewer` (required), `--hostname`

**`pr request-changes PR_ID`**
- Cloud only; on Server/DC returns error: `"request-changes is not supported on Bitbucket Server/DC"`
- On success: `Requested changes on pull request #42`
- Flags: `--hostname`

### MCP Tools

- **`update_pr`** — params: `project`, `slug`, `id`, `title` (optional), `body` (optional), `hostname` (optional)
- **`decline_pr`** — params: `project`, `slug`, `id`, `hostname` (optional)
- **`unapprove_pr`** — params: `project`, `slug`, `id`, `hostname` (optional)
- **`ready_pr`** — params: `project`, `slug`, `id`, `hostname` (optional)
- **`request_review`** — params: `project`, `slug`, `id`, `reviewers` (comma-separated string), `hostname` (optional)

`request_changes` is omitted from MCP tools — Cloud-only escape hatch with narrow use; accessible via `api` passthrough if needed.

### Tests

- `api/cloud/pr_lifecycle_test.go` — unit tests for all 6 operations
- `api/server/pr_lifecycle_test.go` — unit tests for 5 operations (request-changes excluded)
- `pkg/cmd/pr/edit_test.go`, `decline_test.go`, `unapprove_test.go`, `ready_test.go`, `request_review_test.go`, `request_changes_test.go`
- `pkg/cmd/mcp/handlers_test.go` — handler tests for all 5 MCP tools

---

## Parallel Execution Plan

```
main ──[pre-stage commit]──┬── worktree/scope-L ── (Agent L) ── PR → merge
                            ├── worktree/scope-E ── (Agent E) ── PR → merge
                            └── worktree/scope-G ── (Agent G) ── PR → merge
```

1. Pre-stage commit lands on `main` (types + interfaces for all three scopes)
2. Three git worktrees branch off `main`: `feat/scope-l`, `feat/scope-e`, `feat/scope-g`
3. Each agent works exclusively in its owned files — zero overlap
4. Each agent runs `go test ./... -race` before marking done
5. Merge order: any order; no inter-scope dependencies

### Definition of Done (each scope)

- [ ] Backend implementations compile and pass unit tests
- [ ] Commands registered in cobra group
- [ ] MCP tools registered and handler tests pass
- [ ] `go test ./... -race` green
- [ ] `go build ./...` green (all three backends satisfy updated `Client` interface)
