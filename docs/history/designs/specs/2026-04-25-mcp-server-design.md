# Design: bitbottle MCP Server

**Date:** 2026-04-25
**Status:** Approved

---

## Goal

Expose bitbottle's Bitbucket backend as an MCP (Model Context Protocol) server so AI assistants (Claude Desktop, Claude Code) can call Bitbucket operations as native tools — without constructing raw API calls or parsing human-readable CLI output.

Supports both **Bitbucket Cloud** (`bitbucket.org`) and **Bitbucket Server / Data Center** (self-hosted) from a single server.

---

## Decisions

| Question | Decision | Rationale |
|---|---|---|
| Transport | stdio | Standard for local MCP; works with Claude Desktop + Claude Code out of the box; no networking required |
| Binary shape | `bitbottle mcp serve` subcommand | One install, same config, same auth — mirrors `gh mcp-server` precedent |
| Tool scope | All 12 `backend.Client` methods + `list_hosts` | Full surface enables autonomous PR workflows |
| Multi-host | `hostname` optional param on every tool | Omit when single host configured; AI resolves from context when multiple hosts exist |
| Response format | JSON | Structured data AI can reason over; no table parsing |
| SDK | `github.com/mark3labs/mcp-go` | Dominant Go MCP SDK; handles JSON-RPC + stdio boilerplate |

---

## Architecture

```
cmd/bitbottle/main.go
└── pkg/cmd/root.go              (registers "mcp" subcommand)
    └── pkg/cmd/mcp/
        ├── mcp.go               cobra command; starts mcp-go stdio server
        ├── tools.go             registers all 13 tools with descriptions + JSON schemas
        └── handlers.go          one handler func per tool; calls factory.Backend(hostname) → client method
```

### Reuse — nothing changes

| Component | Changes |
|---|---|
| `api/backend.Client` | None |
| `pkg/cmd/factory.Factory` | None |
| `internal/config` | None |
| `internal/bbinstance` | None |
| CLI commands | None |
| `~/.config/bitbottle/hosts.yml` | Same auth config used by MCP server |

The MCP server is a thin translation layer: MCP tool call → `Factory.Backend(hostname)` → `backend.Client` method → JSON response.

---

## Tool Catalogue

### Discovery

| Tool | Description | Inputs | Output |
|---|---|---|---|
| `list_hosts` | List all configured Bitbucket hosts | — | `["bitbucket.org", "git.example.com"]` |

### Repository tools

| Tool | Backend method | Destructive | Inputs | Output |
|---|---|---|---|---|
| `list_repos` | `ListRepos` | — | `hostname?`, `limit?` | `[]Repository` |
| `get_repo` | `GetRepo` | — | `hostname?`, `project`, `slug` | `Repository` |
| `create_repo` | `CreateRepo` | ✅ | `hostname?`, `project`, `name`, `description?`, `private?` | `Repository` |
| `delete_repo` | `DeleteRepo` | ✅ | `hostname?`, `project`, `slug` | `{}` |

### Pull request tools

| Tool | Backend method | Destructive | Inputs | Output |
|---|---|---|---|---|
| `list_prs` | `ListPRs` | — | `hostname?`, `project`, `slug`, `state?`, `limit?` | `[]PullRequest` |
| `get_pr` | `GetPR` | — | `hostname?`, `project`, `slug`, `id` | `PullRequest` |
| `create_pr` | `CreatePR` | ✅ | `hostname?`, `project`, `slug`, `title`, `body?`, `from_branch`, `to_branch`, `draft?` | `PullRequest` |
| `merge_pr` | `MergePR` | ✅ | `hostname?`, `project`, `slug`, `id`, `strategy?` | `PullRequest` |
| `approve_pr` | `ApprovePR` | ✅ | `hostname?`, `project`, `slug`, `id` | `{}` |
| `get_pr_diff` | `GetPRDiff` | — | `hostname?`, `project`, `slug`, `id` | `string` (unified diff) |

### Branch tools

| Tool | Backend method | Destructive | Inputs | Output |
|---|---|---|---|---|
| `delete_branch` | `DeleteBranch` | ✅ | `hostname?`, `project`, `slug`, `branch` | `{}` |

### Auth tools

| Tool | Backend method | Destructive | Inputs | Output |
|---|---|---|---|---|
| `get_current_user` | `GetCurrentUser` | — | `hostname?` | `User` |

---

## Hostname Resolution

Every tool that contacts Bitbucket accepts an optional `hostname` string.

Resolution order:
1. Explicit `hostname` param — use it directly
2. Omitted + single host in config — use that host
3. Omitted + multiple hosts in config — return error: `"multiple hosts configured; specify hostname"`

The `list_hosts` tool exists so the AI can call it at session start to discover available hostnames and resolve ambiguity automatically.

---

## Response Format

All tools return JSON-serialised domain types from `api/backend`:

```go
// Repository
{"slug":"my-service","name":"My Service","namespace":"MYPROJ","scm":"git","web_url":"https://..."}

// PullRequest
{"id":42,"title":"Fix auth","state":"OPEN","author":{"slug":"alice","display_name":"Alice"},...}

// User
{"slug":"alice","display_name":"Alice"}

// Empty success (delete, approve)
{}
```

---

## Error Handling

Backend errors are returned as MCP error responses with the original message from `api/backend/errors.go`. The AI sees the error text and can retry, report, or escalate.

No panics. All errors wrapped with context (tool name + inputs).

---

## Installation & Config

### Install

```bash
go install github.com/proggarapsody/bitbottle/cmd/bitbottle@latest
```

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

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

### Claude Code

Add to `.mcp.json` in your project root:

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

---

## File Changes

| File | Action |
|---|---|
| `go.mod` / `go.sum` | Add `github.com/mark3labs/mcp-go` |
| `pkg/cmd/root.go` | Register `mcp` subcommand |
| `pkg/cmd/mcp/mcp.go` | New — cobra command, starts stdio server |
| `pkg/cmd/mcp/tools.go` | New — 13 tool registrations with JSON schemas |
| `pkg/cmd/mcp/handlers.go` | New — 13 handler funcs |
| `pkg/cmd/mcp/handlers_test.go` | New — unit tests via `FakeClient` |

No other files modified.

---

## Out of Scope

- HTTP / SSE transport (future, if remote access needed)
- Prompt / resource MCP primitives (tools only for now)
- `auth login` as an MCP tool (credential management stays CLI-only)
- Streaming responses for `get_pr_diff` (return full string; pagination later)
