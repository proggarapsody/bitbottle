# Bitbucket Cloud Support Design

**Date:** 2026-04-24  
**Scope:** Add `api.bitbucket.org` (Cloud) support alongside existing Bitbucket Server/DC support for all commands.

---

## Goal

`bitbottle` currently targets Bitbucket Server/DC only. This design adds Bitbucket Cloud (`bitbucket.org`) as a first-class backend. The host is auto-detected; the user experience and command syntax are unchanged.

---

## Package Structure

Current `api/` package (Server/DC only) is split into three:

```
api/
  backend/
    types.go    — shared domain types (Repository, PullRequest, User, …)
    client.go   — BackendClient interface hierarchy
  server/
    client.go   — HTTP transport, auth, Server error parsing
    repos.go    — ListRepos, GetRepo, CreateRepo, DeleteRepo
    prs.go      — ListPRs, GetPR, CreatePR, MergePR, ApprovePR, GetPRDiff,
                   DeleteBranch, GetCurrentUser
  cloud/
    client.go   — HTTP transport, auth, Cloud error parsing
    repos.go
    prs.go
```

Nothing outside `api/` imports `api/server/` or `api/cloud/` directly. All consumers import `api/backend/` only. `api/server/` is existing code moved and adapted; `api/cloud/` is new.

---

## SOLID Interface Design

### Capability interfaces (ISP)

```go
// api/backend/client.go

type RepoLister  interface { ListRepos(limit int) ([]Repository, error) }
type RepoReader  interface { GetRepo(ns, slug string) (Repository, error) }
type RepoWriter  interface { CreateRepo(ns string, in CreateRepoInput) (Repository, error) }
type RepoDeleter interface { DeleteRepo(ns, slug string) error }

type PRLister    interface { ListPRs(ns, slug, state string, limit int) ([]PullRequest, error) }
type PRReader    interface { GetPR(ns, slug string, id int) (PullRequest, error) }
type PRCreator   interface { CreatePR(ns, slug string, in CreatePRInput) (PullRequest, error) }
type PRMerger    interface { MergePR(ns, slug string, id int, in MergePRInput) (PullRequest, error) }
type PRApprover  interface { ApprovePR(ns, slug string, id int) error }
type PRDiffer    interface { GetPRDiff(ns, slug string, id int) (string, error) }
type BranchDeleter interface { DeleteBranch(ns, slug, branch string) error }
type UserGetter  interface { GetCurrentUser() (User, error) }
```

### Composite interface (what Factory returns)

Contains only methods that **all** backends can honestly implement (LSP):

```go
type Client interface {
    RepoLister
    RepoReader
    RepoWriter
    RepoDeleter
    PRLister
    PRReader
    PRCreator
    PRMerger
    PRApprover
    PRDiffer
    BranchDeleter
    UserGetter
}
```

### Optional server-only capabilities

Discovered via type assertion — no `ErrNotSupported` hacks:

```go
type ServerCapabilities interface {
    GetApplicationProperties() (AppProperties, error)
}

// usage in auth status:
if sc, ok := client.(backend.ServerCapabilities); ok {
    props, _ = sc.GetApplicationProperties()
}
```

### Command helpers depend on minimal interfaces (ISP)

```go
// repo list internal helper — only depends on what it uses
func runRepoList(client backend.RepoLister, limit int, ios *iostreams.IOStreams) error { … }
```

Tests mock at the right interface level — no need to implement the full `Client`.

---

## Shared Domain Types

Both backends normalize their API responses into these types:

```go
// api/backend/types.go

type Repository struct {
    Slug      string // repo slug
    Name      string // display name
    Namespace string // Server: project key; Cloud: workspace slug
    SCM       string // "git"
    WebURL    string
}

type PullRequest struct {
    ID          int
    Title       string
    Description string
    State       string // OPEN | MERGED | DECLINED
    Draft       bool
    Author      User
    FromBranch  string
    ToBranch    string
    WebURL      string
}

type User struct {
    Slug        string
    DisplayName string
}

type CreateRepoInput struct {
    Name        string
    SCM         string
    Public      bool
    Description string
}

type CreatePRInput struct {
    Title       string
    Description string
    Draft       bool
    FromBranch  string
    ToBranch    string
}

type MergePRInput struct {
    Message  string
    Strategy string
}

type AppProperties struct {
    Version     string
    BuildNumber string
    DisplayName string
}
```

---

## Host Detection

`internal/bbinstance` gains an `IsCloud` function:

```go
// backendType is config.HostConfig.BackendType — passed as a plain string to keep
// bbinstance free of a config package import.
func IsCloud(hostname, backendType string) bool {
    switch backendType {
    case "cloud":  return true
    case "server": return false
    default:       return hostname == "bitbucket.org"
    }
}
```

`bitbucket.org` is Cloud automatically. Any other hostname is Server/DC. `backendType` from `HostConfig` overrides the default for edge cases.

---

## Config Change

`internal/config.HostConfig` gains one field:

```go
type HostConfig struct {
    User          string `yaml:"user"`
    OAuthToken    string `yaml:"oauth_token,omitempty"`
    GitProtocol   string `yaml:"git_protocol"`
    BackendType   string `yaml:"backend_type,omitempty"` // "server"|"cloud"|"" (auto)
    SkipTLSVerify bool   `yaml:"skip_tls_verify,omitempty"`
}
```

---

## Factory Wiring

`Factory.HttpClient` is renamed to `Factory.Backend` and returns `backend.Client`:

```go
// pkg/cmd/factory/factory.go
type Factory struct {
    IOStreams  *iostreams.IOStreams
    Config     func() (*config.Config, error)
    Backend    func(hostname string) (backend.Client, error) // was HttpClient
    GitRunner  func() run.Runner
    Keyring    keyring.Keyring
    Browser    cmdutil.BrowserLauncher
    Editor     cmdutil.EditorLauncher
    BaseURL    func(hostname string) string
    Now        func() time.Time
}

// in New():
Backend: func(hostname string) (backend.Client, error) {
    if err := loadConfig(); err != nil {
        return nil, err
    }
    hostCfg, _ := cfg.Get(hostname)
    hc := buildHTTPClient(hostCfg) // *http.Client with TLS config applied
    token := hostCfg.OAuthToken
    user  := hostCfg.User

    if bbinstance.IsCloud(hostname, hostCfg.BackendType) {
        // Cloud: app-password (Basic) or Bearer token
        return cloud.NewClient(hc, token, user), nil
    }
    // Server/DC: Bearer token or Basic auth
    return server.NewClient(hc, bbinstance.RESTBase(hostname), token, user), nil
},
```

All command files replace `f.HttpClient(host)` with `f.Backend(host)`. This is a mechanical rename — zero logic change in commands.

---

## Cloud API Specifics

`cloud.Client` handles all Cloud-specific concerns internally:

| Concern | Server/DC | Cloud |
|---|---|---|
| Base URL | `https://HOST/rest/api/1.0` | `https://api.bitbucket.org/2.0` |
| Repo paths | `/projects/{key}/repos/{slug}` | `/repositories/{workspace}/{slug}` |
| PR paths | `/projects/{key}/repos/{slug}/pull-requests/{id}` | `/repositories/{ws}/{slug}/pullrequests/{id}` |
| Pagination | `{isLastPage, nextPageStart}` | `{next, pagelen}` (cursor URL) |
| Errors | `{errors:[{message}]}` | `{type:"error",error:{message}}` |
| Auth | Bearer token or Basic auth | App password (Basic) or Bearer |
| Current user | `GET /users/~` | `GET /user` |
| Delete branch | `DELETE /branches` with JSON body | `DELETE /refs/branches/{name}` |

`cloud.Client` has its own `cloudPagedResponse[T]` and `cloudAPIError` types — both unexported and internal to `api/cloud/`.

---

## URL Helpers

`internal/bbinstance` gains Cloud URL helpers alongside existing Server ones:

```go
// existing — untouched
func RESTBase(host string) string
func SSHURL(host, project, slug string) string
func HTTPSURL(host, project, slug string) string
func WebRepoURL(host, project, slug string) string
func WebPRURL(host, project, slug string, id int) string

// new
func CloudRESTBase() string { return "https://api.bitbucket.org/2.0" }
func CloudSSHURL(workspace, slug string) string
func CloudHTTPSURL(workspace, slug string) string
func CloudWebRepoURL(workspace, slug string) string
func CloudWebPRURL(workspace, slug string, id int) string
```

---

## Command Arguments (bbrepo)

`bbrepo.RepoRef{Host, Project, Slug}` is unchanged. For Cloud, `Project` = workspace slug. User syntax is identical:

- Server: `MYPROJECT/my-repo`
- Cloud: `myworkspace/my-repo`

No UX change. Detection and routing happen inside the backend, invisible to commands.

---

## Testing Strategy

- **Unit tests**: each command helper takes a minimal interface — tests inject a one-method fake, no full `Client` mock needed.
- **Backend unit tests**: `server/` and `cloud/` each have their own `*_test.go` files using `httptest.NewTLSServer` (same pattern as existing `api/integration_test.go`).
- **Factory tests**: `factory_integration_test.go` verifies that `IsCloud("bitbucket.org", …)` routes to `cloud.Client` and a custom hostname routes to `server.Client`.
- **No real network calls**: `factory.NewTestFactory` continues to inject a noop HTTP client by default.

---

## Migration Steps (high-level)

1. Create `api/backend/` with types and interfaces.
2. Move `api/` → `api/server/`, adapting response mapping to use `backend.*` types.
3. Implement `api/cloud/` from scratch.
4. Add `IsCloud` to `internal/bbinstance`.
5. Add `BackendType` to `config.HostConfig`.
6. Rename `Factory.HttpClient` → `Factory.Backend`, update factory wiring.
7. Update all command files: `f.HttpClient` → `f.Backend`, type changes.
8. Update `factory.NewTestFactory` and test helpers.
9. Add tests for Cloud backend and detection logic.
