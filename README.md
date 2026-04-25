# bitbottle 🍶

A command-line interface for **Bitbucket Cloud** and **Bitbucket Server / Data Center** — built with the same philosophy as [GitHub CLI](https://github.com/cli/cli): TTY-aware output, machine-readable pipes, and a clean factory model for easy testing.

---

## ✨ Features

| Area | Status |
|---|---|
| `repo list` | ✅ Fully working |
| `pr list` | ✅ Fully working |
| `auth login / status / logout` | 🚧 Scaffolded (coming soon) |
| `repo create / delete / clone / view` | 🚧 Scaffolded (coming soon) |
| `pr create / merge / approve / view / diff / checkout` | 🚧 Scaffolded (coming soon) |

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

Authentication is stored in `~/.config/bitbottle/hosts.yml`. Edit it directly until `auth login` is implemented:

```yaml
# Bitbucket Cloud
bitbucket.org:
  oauth_token: <your-access-token>
  git_protocol: ssh

# Bitbucket Server / Data Center
bitbucket.example.com:
  oauth_token: <your-token>
  git_protocol: ssh

# Skip TLS for self-signed certs (Server/DC only)
bitbucket.internal.corp:
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
| `repository:read` for `repo list` | `PROJECT_READ` |
| `pullrequest:read` for `pr list` | `PROJECT_READ` |

---

## 🚀 Usage

### Repositories

```bash
# List repositories
bitbottle repo list

# Limit results
bitbottle repo list --limit 10

# Target a specific host when multiple are configured
bitbottle repo list --hostname bitbucket.example.com
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

---

### Pull Requests

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
# Get all open PR titles
bitbottle pr list | awk '{print $1}'

# Count open PRs
bitbottle pr list | wc -l
```

---

## ⚙️ Backend Routing

bitbottle automatically routes API calls to the correct backend:

| Hostname | `backend_type` in config | Routes to |
|---|---|---|
| `bitbucket.org` | _(any / empty)_ | ☁️ Bitbucket Cloud |
| anything else | _(empty)_ | 🏢 Server / Data Center |
| any hostname | `cloud` | ☁️ Bitbucket Cloud (forced) |
| any hostname | `server` | 🏢 Server / DC (forced) |

This means `bitbottle` works identically against both backends — same commands, same output format.

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
| `user` | string | — | Username for Basic auth fallback |
| `git_protocol` | string | `ssh` | `ssh` or `https` |
| `skip_tls_verify` | bool | `false` | Skip TLS cert check (Server/DC) |
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
    ├── repo/           # repo list / create / delete / clone / view
    └── pr/             # pr list / create / merge / approve / view / diff / checkout
```

The `Backend` factory function returns a `backend.Client` — a composite of 12 single-method interfaces. Commands depend only on the methods they use, so they work identically against Cloud and Server without any `if cloud { ... }` branching.

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

Coverage targets: **≥ 80%** on `api/cloud` and `api/server`.

---

## 🛠️ Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## 📄 License

MIT
