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

# List workspace-level webhooks
bitbottle workspace hook list WORKSPACE
bitbottle workspace hook list                  # inferred from pinned repo
bitbottle workspace hook list myworkspace --json

# Create a workspace-level webhook
bitbottle workspace hook create WORKSPACE --url URL --events repo:push,pullrequest:created
bitbottle workspace hook create WORKSPACE --url URL --events repo:push --events pullrequest:created --active

# Delete a workspace-level webhook
bitbottle workspace hook delete WORKSPACE UUID
```

## Flags

| Command | Flag | Default | Description |
|---|---|---|---|
| `workspace list` | `--limit INT` | 30 | Max workspaces returned (0 = no cap) |
| `workspace list` | `--hostname HOST` | — | Override the Bitbucket host |
| `workspace member list` | `--limit INT` | 50 | Max members returned (0 = no cap) |
| `workspace member list` | `--hostname HOST` | — | Override the Bitbucket host |
| `workspace hook list` | `--hostname HOST` | — | Override the Bitbucket host |
| `workspace hook create` | `--url URL` | — | Webhook URL (required) |
| `workspace hook create` | `--events E1,E2` | — | Events to subscribe to (required, repeatable) |
| `workspace hook create` | `--active` | true | Whether the webhook is active |
| `workspace hook create` | `--hostname HOST` | — | Override the Bitbucket host |
| `workspace hook delete` | `--hostname HOST` | — | Override the Bitbucket host |
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

### `workspace hook list`

| Field | JSON key | Notes |
|---|---|---|
| UUID | `uuid` | Webhook UUID |
| URL | `url` | Webhook endpoint URL |
| EVENTS | `events` | Comma-separated event list |
| ACTIVE | `active` | Whether the webhook is active |

## Workspace projects

```bash
# Create a project in a workspace
bitbottle workspace project create WORKSPACE --key KEY --name NAME [--description D] [--private]

# View a project
bitbottle workspace project view WORKSPACE KEY
bitbottle workspace project view WORKSPACE KEY --json

# Edit a project (all flags optional — only changed fields are updated)
bitbottle workspace project edit WORKSPACE KEY [--name NAME] [--description D] [--private=BOOL]

# Delete a project
bitbottle workspace project delete WORKSPACE KEY [--confirm]
```

| Flag | Default | Description |
|---|---|---|
| `--key` | — | Project key, e.g. MYPROJ (required for create) |
| `--name` | — | Project name (required for create) |
| `--description` | — | Project description |
| `--private` | false | Make project private |
| `--confirm` | false | Skip deletion confirmation |

Output fields for `view` and `--json`: `key`, `name`, `description`, `is_private`.

## MCP tools

- `list_workspace_members` — `workspace` (required), `hostname`, `limit`
- `workspace_hook_list` — `workspace` (required), `hostname`
- `workspace_hook_create` — `workspace` (required), `url` (required), `events` (required, comma-separated), `active`, `hostname`
- `workspace_hook_delete` — `workspace` (required), `uuid` (required), `hostname`
- `create_workspace_project` — `workspace`, `key`, `name` (all required), `description`, `private`, `hostname`
- `view_workspace_project` — `workspace`, `key` (required), `hostname`
- `edit_workspace_project` — `workspace`, `key` (required), `name`, `description`, `private`, `hostname`
- `delete_workspace_project` — `workspace`, `key` (required), `hostname`
