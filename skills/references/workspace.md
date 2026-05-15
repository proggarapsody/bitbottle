# Workspaces (Cloud only)

Workspaces are the top-level ownership unit on Bitbucket Cloud, containing
repositories and projects. Bitbucket Server/DC has no workspace concept —
commands in this group return a typed `host.unsupported` error against
Server/DC hosts.

## Commands

```bash
# List workspaces the authenticated user belongs to
bitbottle workspace list
bitbottle workspace list --limit 100
bitbottle workspace list --json slug,name --jq '.[].slug'

# List members of a workspace
bitbottle workspace member list WORKSPACE
bitbottle workspace member list                # inferred from pinned repo
bitbottle workspace member list myworkspace --limit 100 --json
```

## Flags

| Command | Flag | Default | Description |
|---|---|---|---|
| `workspace list` | `--limit INT` | 30 | Max workspaces returned (0 = no cap) |
| `workspace list` | `--hostname HOST` | — | Override the Bitbucket host |
| `workspace member list` | `--limit INT` | 50 | Max members returned (0 = no cap) |
| `workspace member list` | `--hostname HOST` | — | Override the Bitbucket host |
| all | `--json [FIELDS]` | — | JSON output (comma-separated field list optional) |
| all | `--jq EXPR` | — | Filter JSON with jq expression |

## Output fields

### `workspace list`

| Field | JSON key | Notes |
|---|---|---|
| SLUG | `slug` | Workspace slug |
| NAME | `name` | Display name |
| URL | `webURL` | Web URL (also `url`, `link`) |
| UUID | `uuid` | JSON-only |

### `workspace member list`

| Field | JSON key | Notes |
|---|---|---|
| SLUG | `slug` | User's account slug |
| NAME | `name` | User's display name |

## MCP tools

- `list_workspace_members` — `workspace` (required), `hostname`, `limit`
