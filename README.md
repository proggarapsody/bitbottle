# bitbottle 🍶

A command-line interface for **Bitbucket Cloud** and **Bitbucket Server / Data Center** — built with the same philosophy as [GitHub CLI](https://github.com/cli/cli): TTY-aware output, machine-readable pipes, and a clean factory model for easy testing.

---

## ✨ Features

| Area | Status |
|---|---|
| `auth login / logout / status` | ✅ Fully working |
| `repo list / create / delete / clone / view` | ✅ Fully working |
| `pr list / create / merge / approve / view / diff / checkout / edit / decline / unapprove / ready / request-review` | ✅ Fully working |
| `branch list / delete / create / checkout` | ✅ Fully working |
| `tag list / create / delete` | ✅ Fully working |
| `pipeline list / view / run` _(Cloud only)_ | ✅ Fully working |
| `mcp serve` — MCP stdio server for AI assistants | ✅ Fully working |

Works identically against **Bitbucket Cloud** (`bitbucket.org`) and **Bitbucket Server / Data Center** (self-hosted).

---

## 📦 Installation

```bash
go install github.com/proggarapsody/bitbottle/cmd/bitbottle@latest
```

Or build from source:

```bash
git clone https://github.com/proggarapsody/bitbottle
cd bitbottle
make build
```

> Requires Go 1.21+

---

## 🔑 Authentication

### Interactive login

```bash
# Bitbucket Cloud
echo "YOUR_TOKEN" | bitbottle auth login --hostname bitbucket.org --with-token

# Bitbucket Server / Data Center
echo "YOUR_TOKEN" | bitbottle auth login \
  --hostname bitbucket.example.com \
  --with-token \
  --git-protocol https \
  --skip-tls-verify        # for self-signed certs
```

`auth login` validates the token against the API, saves credentials to `~/.config/bitbottle/hosts.yml`, and optionally stores the token in the system keyring.

### Flags

| Flag | Default | Description |
|---|---|---|
| `--hostname` | — | **Required.** Target Bitbucket host |
| `--with-token` | false | Read token from stdin |
| `--git-protocol` | `ssh` | `ssh` or `https` |
| `--skip-tls-verify` | false | Skip TLS cert check (Server/DC self-signed) |

### Check status

```bash
bitbottle auth status
# bitbucket.org: Logged in as alice (Token in keyring: yes)
# bitbucket.example.com: Logged in as bob (Token in keyring: no)
```

### Logout

```bash
bitbottle auth logout --hostname bitbucket.example.com
```

### Manual config

Edit `~/.config/bitbottle/hosts.yml` directly:

```yaml
# Bitbucket Cloud
bitbucket.org:
  oauth_token: <your-access-token>
  git_protocol: ssh

# Bitbucket Server / Data Center
bitbucket.example.com:
  oauth_token: <your-token>
  git_protocol: https
  skip_tls_verify: true

# Force Cloud routing for a non-bitbucket.org host
mycompany.bitbucket.example.com:
  oauth_token: <your-token>
  backend_type: cloud   # "cloud" | "server" | "" (auto)
```

**Token scopes required:**

| Bitbucket Cloud | Bitbucket Server / DC |
|---|---|
| `account:read` for auth commands | `PROJECT_READ` |
| `repository:read` / `repository:write` | `PROJECT_READ` / `PROJECT_WRITE` |
| `pullrequest:read` / `pullrequest:write` | `PROJECT_READ` / `PROJECT_WRITE` |

---

## 🚀 Usage

### 🔐 Auth

```bash
# Login
echo "TOKEN" | bitbottle auth login --hostname bitbucket.org --with-token

# Show all authenticated hosts
bitbottle auth status

# Logout
bitbottle auth logout --hostname bitbucket.org
```

---

### 📁 Repositories

#### List

```bash
# List repositories (auto-detects host)
bitbottle repo list

# Limit results
bitbottle repo list --limit 10

# Target a specific host
bitbottle repo list --hostname bitbucket.example.com

# JSON output (specify comma-separated fields)
bitbottle repo list --json slug,project,scm
```

**TTY output** (aligned table):

```
SLUG              PROJECT     TYPE
my-service        MYPROJ      git
payments-api      MYPROJ      git
infra-tools       PLATFORM    git
```

**Piped / non-TTY output** (tab-separated, no header):

```bash
bitbottle repo list | awk '{print $1}'   # → slugs only
bitbottle repo list | cut -f2            # → projects only
```

#### Create

```bash
bitbottle repo create my-service --project MYPROJ
bitbottle repo create my-service --project MYPROJ --description "Core service" --private=false
```

| Flag | Default | Description |
|---|---|---|
| `--project` | — | **Required.** Project key |
| `--description` | — | Repository description |
| `--private` | `true` | Make repository private |

#### Delete

```bash
# With confirmation prompt (TTY)
bitbottle repo delete MYPROJ/my-service

# Skip prompt
bitbottle repo delete MYPROJ/my-service --confirm
```

#### Clone

```bash
# Clone into ./my-service
bitbottle repo clone MYPROJ/my-service

# Clone into custom directory
bitbottle repo clone MYPROJ/my-service ~/projects/my-service
```

Uses SSH or HTTPS based on `git_protocol` in your config.

#### View

```bash
bitbottle repo view MYPROJ/my-service

# Open in browser
bitbottle repo view MYPROJ/my-service --web
```

---

### 🌿 Branches

#### List

```bash
bitbottle branch list MYPROJ/my-service

# Limit results
bitbottle branch list MYPROJ/my-service --limit 10

# JSON output
bitbottle branch list MYPROJ/my-service --json name,default,hash

# jq filter
bitbottle branch list MYPROJ/my-service --json name --jq .
```

**TTY output:**

```
NAME        DEFAULT   HASH
main        true      a1b2c3d4
feature/x   false     e5f6a7b8
develop     false     c9d0e1f2
```

#### Delete

```bash
bitbottle branch delete MYPROJ/my-service feature/my-branch
```

#### Create

```bash
bitbottle branch create MYPROJ/my-service feature/my-branch --start-at main
bitbottle branch create MYPROJ/my-service hotfix/issue-42 --start-at abc123def456
```

| Flag | Default | Description |
|---|---|---|
| `--start-at` | — | **Required.** Branch name or commit hash to branch from |

#### Checkout

```bash
# Fetch from origin and check out locally
bitbottle branch checkout feature/my-branch
```

Equivalent to `git fetch origin feature/my-branch && git checkout feature/my-branch`.

---

### 🏷️ Tags

#### List

```bash
bitbottle tag list MYPROJ/my-service
bitbottle tag list MYPROJ/my-service --limit 10
bitbottle tag list MYPROJ/my-service --json name,hash
```

**TTY output:**

```
NAME       HASH
v1.2.0     a1b2c3d4
v1.1.0     e5f6a7b8
v1.0.0     c9d0e1f2
```

#### Create

```bash
# Lightweight tag
bitbottle tag create MYPROJ/my-service v1.3.0 --start-at main

# Annotated tag
bitbottle tag create MYPROJ/my-service v1.3.0 --start-at main --message "Release 1.3.0"
```

| Flag | Default | Description |
|---|---|---|
| `--start-at` | — | **Required.** Branch name or commit hash to tag |
| `--message` | — | Tag message (creates annotated tag when non-empty) |

#### Delete

```bash
bitbottle tag delete MYPROJ/my-service v1.3.0
```

---

### 🔀 Pull Requests

#### List

```bash
# List open PRs (auto-detects repo from git remote)
bitbottle pr list

# Explicit PROJECT/REPO
bitbottle pr list MYPROJECT/my-service

# Filter by state
bitbottle pr list --state merged
bitbottle pr list --state closed --limit 5

# Specific host
bitbottle pr list --hostname bitbucket.example.com

# JSON output
bitbottle pr list --json id,title,author,state
```

**TTY output:**

```
TITLE                        AUTHOR     STATE
Fix null pointer in auth     alice      OPEN
Bump lodash to 4.17.21       bob        OPEN
Add retry logic to payments  charlie    OPEN
```

**Piped:**

```bash
# Count open PRs
bitbottle pr list | wc -l

# Get all open PR titles
bitbottle pr list | awk '{print $1}'
```

#### Create

```bash
bitbottle pr create --title "Fix auth bug" --base main

# With description and draft flag
bitbottle pr create \
  --title "Add retry logic" \
  --body "Implements exponential backoff for all HTTP calls." \
  --base main \
  --draft
```

| Flag | Default | Description |
|---|---|---|
| `--title` | — | **Required.** PR title |
| `--base` | — | **Required.** Target branch |
| `--body` | — | PR description |
| `--draft` | false | Mark as draft PR |

Branch is auto-detected from `git rev-parse --abbrev-ref HEAD`.

#### Merge

```bash
# Default merge strategy
bitbottle pr merge 42

# Explicit strategy
bitbottle pr merge 42 --merge      # merge commit
bitbottle pr merge 42 --squash     # squash merge

# Delete source branch after merge
bitbottle pr merge 42 --squash --delete-branch
```

| Flag | Default | Description |
|---|---|---|
| `--merge` | false | Merge commit strategy |
| `--squash` | false | Squash merge strategy |
| `--delete-branch` | false | Delete source branch after merge |

#### Approve

```bash
bitbottle pr approve 42
```

#### View

```bash
bitbottle pr view 42

# Open in browser
bitbottle pr view 42 --web
```

#### Diff

```bash
# Stream diff to terminal
bitbottle pr diff 42

# Pipe to a pager or diff tool
bitbottle pr diff 42 | less
bitbottle pr diff 42 | delta
```

#### Checkout

```bash
# Fetch and checkout the PR's source branch
bitbottle pr checkout 42
```

#### Edit

```bash
# Update title
bitbottle pr edit 42 --title "Fix auth bug (updated)"

# Update description
bitbottle pr edit 42 --body "New description"

# Update both
bitbottle pr edit 42 --title "New title" --body "New body"
```

#### Decline

```bash
bitbottle pr decline 42
```

#### Unapprove

```bash
bitbottle pr unapprove 42
```

#### Ready

```bash
# Promote a draft PR to ready for review
bitbottle pr ready 42
```

#### Request Review

```bash
# Add reviewers (comma-separated usernames/account IDs)
bitbottle pr request-review 42 --reviewer alice --reviewer bob

# Or comma-separated
bitbottle pr request-review 42 --reviewer alice,bob
```

| Flag | Default | Description |
|---|---|---|
| `--reviewer` | — | **Required.** Reviewer username(s); repeatable or comma-separated |

---

### ⚙️ Pipelines _(Bitbucket Cloud only)_

#### List

```bash
bitbottle pipeline list MYWORKSPACE/my-service

# Limit results
bitbottle pipeline list MYWORKSPACE/my-service --limit 10

# JSON output
bitbottle pipeline list MYWORKSPACE/my-service --json buildNumber,state,refName,duration

# jq filter — show only failed
bitbottle pipeline list MYWORKSPACE/my-service --json state --jq 'select(. == "FAILED")'
```

**TTY output:**

```
BUILD   STATE       BRANCH/TAG   DURATION
42      SUCCESSFUL  main         87s
41      FAILED      feature/x    12s
40      SUCCESSFUL  main         91s
```

#### View

```bash
bitbottle pipeline view MYWORKSPACE/my-service {uuid}

# Open in browser
bitbottle pipeline view MYWORKSPACE/my-service {uuid} --web

# JSON output
bitbottle pipeline view MYWORKSPACE/my-service {uuid} --json buildNumber,state,refName,duration,webURL
```

#### Run

```bash
# Trigger a pipeline on a branch (--branch is required)
bitbottle pipeline run MYWORKSPACE/my-service --branch main

# Trigger on a feature branch
bitbottle pipeline run MYWORKSPACE/my-service --branch feature/my-feature
```

| Flag | Default | Description |
|---|---|---|
| `--branch` | — | **Required.** Branch to run the pipeline on |

> **Note:** Pipelines are a Bitbucket Cloud feature. Running any `pipeline` command against a Server / Data Center host returns an error.

---

## 🤖 MCP Server (AI Assistant Integration)

`bitbottle mcp serve` starts a [Model Context Protocol](https://modelcontextprotocol.io) server over stdio. Claude Desktop, Claude Code, and any MCP-compatible client can call Bitbucket operations as native tools — no raw API requests, no output parsing.

### Tools exposed

| Tool | Description |
|---|---|
| `list_hosts` | List all configured Bitbucket hosts |
| `list_repos` | List repositories |
| `get_repo` | Get a single repository |
| `create_repo` | Create a repository |
| `delete_repo` | Delete a repository |
| `list_prs` | List pull requests |
| `get_pr` | Get a single pull request |
| `create_pr` | Create a pull request |
| `merge_pr` | Merge a pull request |
| `approve_pr` | Approve a pull request |
| `get_pr_diff` | Get the unified diff for a pull request |
| `list_branches` | List branches in a repository |
| `create_branch` | Create a new branch |
| `delete_branch` | Delete a branch |
| `list_tags` | List tags in a repository |
| `create_tag` | Create a tag |
| `delete_tag` | Delete a tag |
| `update_pr` | Update PR title and/or description |
| `decline_pr` | Decline a pull request |
| `unapprove_pr` | Remove approval from a pull request |
| `ready_pr` | Mark a draft PR as ready for review |
| `request_review` | Add reviewers to a pull request |
| `list_pipelines` | List pipelines _(Cloud only)_ |
| `get_pipeline` | Get a single pipeline by UUID _(Cloud only)_ |
| `run_pipeline` | Trigger a pipeline on a branch _(Cloud only)_ |
| `get_current_user` | Get the authenticated user |

Every tool accepts an optional `hostname` parameter. When only one host is configured, `hostname` can be omitted.

### Setup

**Claude Desktop** — add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "bitbottle": {
      "command": "bitbottle",
      "args": ["mcp", "serve"]
    }
  }
}
```

**Claude Code** — add to `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "bitbottle": {
      "command": "bitbottle",
      "args": ["mcp", "serve"]
    }
  }
}
```

The MCP server uses the same `~/.config/bitbottle/hosts.yml` config and credentials as the CLI — no separate auth setup needed.

---

## ⚙️ Backend Routing

bitbottle automatically routes API calls to the correct backend:

| Hostname | `backend_type` in config | Routes to |
|---|---|---|
| `bitbucket.org` | _(any / empty)_ | ☁️ Bitbucket Cloud |
| anything else | _(empty)_ | 🏢 Server / Data Center |
| any hostname | `cloud` | ☁️ Bitbucket Cloud (forced) |
| any hostname | `server` | 🏢 Server / DC (forced) |

Same commands, same output format — regardless of backend.

### Cloud vs Server/DC differences (handled internally)

| Concern | Cloud | Server / DC |
|---|---|---|
| REST base | `api.bitbucket.org/2.0` | `HOST/rest/api/1.0` |
| Repo path | `/repositories/{workspace}/{slug}` | `/projects/{key}/repos/{slug}` |
| PR path | `/pullrequests/{id}` (no hyphen) | `/pull-requests/{id}` |
| Approve PR | `POST .../approve` | `PUT .../participants/~` |
| Delete branch | `DELETE .../refs/branches/{branch}` | `DELETE .../branches` (JSON body) |
| Pagination | Cursor (`next` URL) | Keyset (`isLastPage` + `nextPageStart`) |
| Error shape | `{"type":"error","error":{"message":".."}}` | `{"errors":[{"message":".."}]}` |
| Current user | `GET /user` | `GET /users/~` |

---

## 🗂️ Configuration Reference

Config file: `~/.config/bitbottle/hosts.yml`

| Field | Type | Default | Description |
|---|---|---|---|
| `oauth_token` | string | — | Bearer token (preferred) |
| `user` | string | — | Username (populated automatically on login) |
| `git_protocol` | string | `ssh` | `ssh` or `https` |
| `skip_tls_verify` | bool | `false` | Skip TLS cert check (Server/DC self-signed) |
| `backend_type` | string | `""` | `""` (auto), `cloud`, or `server` |

**Auth header precedence:** `Bearer <oauth_token>` → `Basic <user>:<empty>` → none.

---

## 🔌 Architecture

```
bitbottle
├── api/backend/        # Shared domain types + Client interface (12 capabilities)
├── api/cloud/          # Bitbucket Cloud adapter (api.bitbucket.org)
├── api/server/         # Bitbucket Server/DC adapter
├── api/internal/httpx/ # Shared HTTP transport (internal – not importable externally)
├── internal/bbinstance # Host detection, URL builders, version helpers
├── internal/config     # hosts.yml read/write
└── pkg/cmd/            # CLI commands (cobra)
    ├── auth/           # auth login / logout / status
    ├── branch/         # branch list / delete / create / checkout
    ├── mcp/            # mcp serve — MCP stdio server
    ├── pipeline/       # pipeline list / view / run (Cloud only)
    ├── pr/             # pr list / create / merge / approve / view / diff / checkout / edit / decline / unapprove / ready / request-review
    ├── repo/           # repo list / create / delete / clone / view
    └── tag/            # tag list / create / delete
```

The `Backend` factory returns a `backend.Client` — a composite of single-method interfaces. Commands depend only on the methods they use, so they work identically against Cloud and Server with no `if cloud { ... }` branching. Pipeline commands additionally require a `backend.PipelineClient`, which is only implemented by the Cloud adapter. `pr request-changes` uses the Cloud-only optional-interface pattern (type assertion). The MCP server is a thin translation layer on top of the same factory and client.

---

## 🧪 Testing

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Benchmarks (Cloud and Server JSON decode, N=100)
go test -bench=. ./api/cloud/ ./api/server/

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Coverage targets: **≥ 80%** on `api/cloud`, `api/server`, and `pkg/cmd/*`.

---

## 🛠️ Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## 📄 License

MIT
