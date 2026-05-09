# Tech Stack — bitbottle

## Languages

| Language | Version | Where it's used |
|---|---|---|
| **Go** | 1.25.0 | Primary. All CLI code, API adapters, MCP server. |
| **Python** | 3 (any 3.x) | Tooling scripts. `skills/scripts/sync_help.py` syncs `--help` text into the bundled Claude skill. |
| **JavaScript** | Node (any LTS) | npm install shim. `packages/mcp-npm/install.js` downloads the matching Go binary on `npm postinstall`. |

## Frontend

None. bitbottle is a CLI. The only "UI" is terminal output via
`f.IOStreams` + `fatih/color`, with TTY-awareness for tables.

## Backend (client-side stack)

bitbottle is not a server, so there is no backend framework in the
typical sense. The "backend" it talks to is the Bitbucket REST APIs.
The client-side stack:

| Concern | Library | Notes |
|---|---|---|
| CLI scaffolding | `spf13/cobra` v1.8.0 | Command tree, flag parsing, completion |
| HTTP transport | `net/http` + `api/internal/httpx/` | Custom transport with `ContentTypePolicy`, retry, domain-error classification |
| MCP server | `mark3labs/mcp-go` v0.48.0 | stdio MCP server in `pkg/cmd/mcp/` |
| jq filter | `itchyny/gojq` v0.12.17 | Pure-Go jq, powers `--jq` on every command |
| Credential storage | `zalando/go-keyring` v0.2.8 | OS keychain integration (macOS Keychain, Linux Secret Service, Windows Credential Manager) |
| ANSI color | `fatih/color` v1.19.0 | TTY-aware color output |
| isatty check | `mattn/go-isatty` v0.0.20 | TTY detection |
| Test assertions | `stretchr/testify` v1.11.1 | Test-only |

### Internal architecture (key paths)

```
api/backend/        — shared domain types + Client interface
api/cloud/          — Bitbucket Cloud adapter
api/server/         — Bitbucket Server/DC adapter
api/internal/httpx/ — shared HTTP transport (internal)
internal/           — config, envvars, bbinstance, keyring, etc.
pkg/cmd/            — Cobra commands (one package per noun)
pkg/cmd/mcp/        — MCP stdio server (tools + handlers)
pkg/errfmt/         — user-facing error rendering (titles + hints)
```

## Database

**None / Stateless.** bitbottle holds no domain data of its own; all
state lives in Bitbucket. Local artifacts are just:

- **Auth tokens** in the OS keychain (`zalando/go-keyring`).
- **Hosts config** in `~/.config/bitbottle/hosts.yml`.
- **Per-repo defaults** in `.git/config` keys `bitbottle.host`,
  `bitbottle.project`, `bitbottle.slug` — consulted by `f.BaseRepo()`
  before falling back to remote URL parsing.

## Build & distribution

### CI

GitHub Actions: lint (`golangci-lint`), test (`go test ./... -race`),
build matrix.

### Release pipeline (automated)

1. Merge `feat:`/`fix:` PR to `main`.
2. **release-please** opens (or updates) a release PR — bumps version,
   regenerates `CHANGELOG.md`, updates the `<!-- x-release-please-version
   -->`-tagged lines in `skills/SKILL.md`.
3. Merge the release PR → **GoReleaser** builds cross-platform binaries,
   publishes the GitHub release, pushes Docker images, and triggers npm
   publish.

### Distribution channels

- **GitHub Releases** — pre-built binaries: linux / macOS / windows ×
  amd64 / arm64.
- **Docker images** — `ghcr.io/proggarapsody/bitbottle:<version>` and
  `:latest`.
- **npm** — `@proggarapsody/bitbottle` (the `postinstall` shim downloads
  the matching binary from GitHub Releases).
- **Go toolchain** —
  `go install github.com/proggarapsody/bitbottle/cmd/bitbottle@latest`.

## Tooling files (where the rules live)

| File | What it controls |
|---|---|
| `Makefile` | `make setup`, `make lint`, `make test`, `make build` |
| `.golangci.yml` | All Go lint rules (cited by the Go style guide; do not duplicate) |
| `.goreleaser.yaml` | Cross-platform binary build matrix |
| `release-please-config.json` | Release-please packages, files, and version sources |
| `.github/workflows/` | CI + release automation |
