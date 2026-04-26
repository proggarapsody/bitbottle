# Scope F Design: Commits

**Date:** 2026-04-26
**Scope:** F (Commits)
**Execution:** Sequential, after M and P land on main

---

## Philosophy

New domain following the established layer pattern. Both backends. Context-aware
branch detection (gh-style). No precomputed derived fields in domain types —
formatters handle display concerns.

---

## Domain Layer

### `api/backend/client.go`

Two new interfaces added to composite `Client`:

```go
type CommitLister interface {
    ListCommits(ns, slug, branch string, limit int) ([]Commit, error)
}
type CommitReader interface {
    GetCommit(ns, slug, hash string) (Commit, error)
}
```

### `api/backend/types.go`

```go
type Commit struct {
    Hash      string
    Message   string    // subject line only (first line of commit message)
    Author    User
    Timestamp time.Time
    WebURL    string
}
```

No `ShortHash` — formatters compute `hash[:7]` for display. Full hash available
for `--json` output and `commit view`.

### `api/backend/fake_client.go`

```go
ListCommitsFn func(ns, slug, branch string, limit int) ([]Commit, error)
GetCommitFn   func(ns, slug, hash string) (Commit, error)
```

Both default to `t.Fatalf("unexpected call to ListCommits / GetCommit")`, same
loud-failure pattern as all other fake methods.

---

## Commands

### Package layout

```
pkg/cmd/commit/
    commit.go     # parent cobra group, wired into root
    log.go        # commit log
    log_test.go
    view.go       # commit view
    view_test.go
```

### `commit log PROJECT/REPO`

List commits on a branch. TTY: aligned table. Pipe: tab-separated, no header.

**Flags:**

| Flag | Short | Default | Description |
|---|---|---|---|
| `--branch` | `-b` | _(see resolution)_ | Branch to list commits from |
| `--limit` | — | 30 | Maximum number of results |
| `--json` | — | — | Comma-separated fields to output as JSON |
| `--jq` | — | — | jq filter applied to JSON output |
| `--hostname` | — | — | Target Bitbucket host |

**Branch resolution order:**
1. `--branch` flag (explicit)
2. `git rev-parse --abbrev-ref HEAD` — current local branch if inside a git repo
3. `"main"` — hardcoded fallback (avoids a round-trip; user can always override with `--branch`)

**TTY output:**

```
HASH     MESSAGE                           AUTHOR   DATE
abc1234  Fix null pointer in auth          alice    2 days ago
def5678  Bump lodash to 4.17.21            bob      3 days ago
c9d0e1f  Add retry logic to payments       charlie  5 days ago
```

Hash column shows first 7 characters. Message truncated to ~60 chars on narrow
terminals. Date shown as relative ("2 days ago") on TTY, RFC3339 in `--json`.

**Pipe output** (tab-separated, no header):

```
abc1234def456abc1234def456abc1234def456ab\tFix null pointer in auth\talice\t2026-04-24T10:00:00Z
```

### `commit view PROJECT/REPO HASH`

Show full details for a single commit.

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--web` | false | Open commit URL in browser |
| `--json` | — | Comma-separated fields |
| `--jq` | — | jq filter |
| `--hostname` | — | Target host |

**TTY output:**

```
commit abc1234def456abc1234def456abc1234def456ab

Fix null pointer in auth middleware

Author:  alice
Date:    2026-04-24 10:00:00 +0000 UTC
Web:     https://bitbucket.org/myws/my-service/commits/abc1234def456
```

---

## Backend Adapters

### Cloud (`api/cloud/commit.go`)

**List:**
```
GET /repositories/{workspace}/{slug}/commits?branch={branch}&pagelen={limit}
```

Wire type:
```go
type wireCommit struct {
    Hash    string `json:"hash"`
    Message string `json:"message"`
    Author  struct {
        Raw  string `json:"raw"`
        User struct {
            AccountID   string `json:"account_id"`
            DisplayName string `json:"display_name"`
        } `json:"user"`
    } `json:"author"`
    Date  time.Time `json:"date"`
    Links struct {
        HTML struct{ Href string `json:"href"` } `json:"html"`
    } `json:"links"`
}
```

`toDomain()`: `Hash` → `Hash`, first line of `Message` → `Message`,
`Author.User.DisplayName` (or `Author.Raw` if no user) → `Author.Login`,
`Date` → `Timestamp`, `Links.HTML.Href` → `WebURL`.

Pagination: Cloud cursor (`next` URL), same helper as existing list endpoints.

**Get:**
```
GET /repositories/{workspace}/{slug}/commit/{hash}
```

Same wire type, no pagination.

### Server (`api/server/commit.go`)

**List:**
```
GET /rest/api/1.0/projects/{key}/repos/{slug}/commits?until={branch}&limit={limit}
```

Wire type:
```go
type wireCommit struct {
    ID        string `json:"id"`
    Message   string `json:"message"`
    Author    struct {
        Name          string `json:"name"`
        EmailAddress  string `json:"emailAddress"`
    } `json:"author"`
    AuthorTimestamp int64 `json:"authorTimestamp"` // Unix ms
}
```

`toDomain()`: `ID` → `Hash`, first line of `Message` → `Message`,
`Author.Name` → `Author.Login`, `AuthorTimestamp/1000` → `Timestamp`.
`WebURL` constructed: `https://{host}/projects/{key}/repos/{slug}/commits/{hash}`.

Pagination: Server keyset (`isLastPage` + `nextPageStart`), same helper as
existing list endpoints.

**Get:**
```
GET /rest/api/1.0/projects/{key}/repos/{slug}/commits/{hash}
```

Same wire type, no pagination.

---

## MCP Tools

Registered in `pkg/cmd/mcp/tools.go`, handlers in `pkg/cmd/mcp/handlers.go`.

| Tool | Description |
|---|---|
| `list_commits` | List commits for a repository |
| `get_commit` | Get a single commit by hash |

`list_commits` parameters: `hostname` (opt), `project` (req), `slug` (req),
`branch` (opt), `limit` (opt).

`get_commit` parameters: `hostname` (opt), `project` (req), `slug` (req),
`hash` (req).

---

## Testing

### `api/cloud/commit_test.go`

- `TestCloudClient_ListCommits_IssuesCorrectPath` — verifies URL, pagelen param,
  branch param; returns one commit; asserts `toDomain` mapping.
- `TestCloudClient_GetCommit_IssuesCorrectPath` — verifies URL path; asserts
  mapping.

### `api/server/commit_test.go`

- `TestServerClient_ListCommits_IssuesCorrectPath` — verifies URL, `until`
  param, `limit` param; asserts mapping including WebURL construction.
- `TestServerClient_GetCommit_IssuesCorrectPath` — verifies URL path.

### `pkg/cmd/commit/log_test.go`

- Happy path: `FakeClient` returns 3 commits; verify TTY table output includes
  truncated hash, subject, author.
- Branch flag: `--branch feat/x` passed through to `ListCommitsFn`.
- Branch default: no flag, no git repo → `"main"` passed to `ListCommitsFn`.
- `--json` output: raw hashes, full timestamps.

### `pkg/cmd/commit/view_test.go`

- Happy path: `FakeClient` returns commit; verify TTY detail output.
- `--web`: `WebURL` opened via browser opener stub.

### `pkg/cmd/mcp/handlers_test.go`

- `TestListCommits_CallsClientAndReturnsJSON`
- `TestGetCommit_CallsClientAndReturnsJSON`

---

## Definition of Done

- [ ] `api/backend/client.go` — `CommitLister`, `CommitReader` interfaces; both
      added to composite `Client`
- [ ] `api/backend/types.go` — `Commit` type
- [ ] `api/backend/fake_client.go` — `ListCommitsFn`, `GetCommitFn` with loud
      defaults
- [ ] `api/cloud/commit.go` + `_test.go`
- [ ] `api/server/commit.go` + `_test.go`
- [ ] `pkg/cmd/commit/` — `commit.go`, `log.go`, `view.go` + tests
- [ ] `pkg/cmd/mcp/tools.go` — `list_commits`, `get_commit` registered
- [ ] `pkg/cmd/mcp/handlers.go` — `listCommits`, `getCommit` handlers + tests
- [ ] README — new Commits section
- [ ] BACKLOG.md — `commit log`, `commit view` rows → ✅; F row → ✅
- [ ] `go test ./... -race` green
