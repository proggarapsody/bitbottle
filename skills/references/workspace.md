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

# Search workspaces by slug/name prefix with optional role filter
bitbottle workspace search
bitbottle workspace search --query myws
bitbottle workspace search --role owner
bitbottle workspace search --query myws --role collaborator --limit 10
bitbottle workspace search --json

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
| `workspace search` | `--query Q` | — | Slug/name prefix to match |
| `workspace search` | `--role ROLE` | — | Filter by role: owner, collaborator, or member |
| `workspace search` | `--limit INT` | 30 | Max workspaces returned (0 = no cap) |
| `workspace search` | `--hostname HOST` | — | Override the Bitbucket host |
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

- `search_workspaces` — `query`, `role`, `hostname`, `limit`
- `list_workspace_members` — `workspace` (required), `hostname`, `limit`
- `workspace_hook_list` — `workspace` (required), `hostname`
- `workspace_hook_create` — `workspace` (required), `url` (required), `events` (required, comma-separated), `active`, `hostname`
- `workspace_hook_delete` — `workspace` (required), `uuid` (required), `hostname`
- `create_workspace_project` — `workspace`, `key`, `name` (all required), `description`, `private`, `hostname`
- `view_workspace_project` — `workspace`, `key` (required), `hostname`
- `edit_workspace_project` — `workspace`, `key` (required), `name`, `description`, `private`, `hostname`
- `delete_workspace_project` — `workspace`, `key` (required), `hostname`
- `list_workspace_perms` — `workspace` (required), `limit`
- `list_workspace_repo_perms` — `workspace` (required), `limit`
- `grant_workspace_perm` — `workspace`, `user`, `permission` (all required)
- `revoke_workspace_perm` — `workspace`, `user` (both required)
- `list_workspace_project_perms` — `workspace`, `project_key` (both required), `hostname`
- `grant_workspace_project_perm` — `workspace`, `project_key`, `permission` (all required), `user_slug` or `group_slug`, `hostname`
- `revoke_workspace_project_perm` — `workspace`, `project_key`, `subject_slug` (all required), `is_group`, `hostname`
- `list_workspace_pipeline_vars` — `workspace` (required), `hostname`
- `get_workspace_pipeline_var` — `workspace`, `key` (both required), `hostname`
- `set_workspace_pipeline_var` — `workspace`, `key`, `value` (all required), `secured`, `hostname`
- `delete_workspace_pipeline_var` — `workspace`, `key` (both required), `hostname`

## Workspace Pipeline Variables (Cloud only)

Workspace-level pipeline variables are shared across all repositories in the
workspace and injected into every pipeline step. They are distinct from
repo-level pipeline variables (`variable --scope repository`) and deployment
env vars. Returns `host.unsupported` on Bitbucket Server/DC.

```bash
# List all workspace pipeline variables
bitbottle workspace pipeline-variable list WORKSPACE
bitbottle workspace pipeline-variable list                    # inferred from pinned repo
bitbottle workspace pipeline-variable list myworkspace --json

# Get a workspace pipeline variable by key
bitbottle workspace pipeline-variable get WORKSPACE KEY
bitbottle workspace pipeline-variable get myworkspace MY_VAR --json

# Create or update a workspace pipeline variable (upsert by key)
bitbottle workspace pipeline-variable set WORKSPACE KEY VALUE
bitbottle workspace pipeline-variable set myworkspace MY_VAR myvalue
bitbottle workspace pipeline-variable set myworkspace SECRET_VAR s3cr3t --secured

# Delete a workspace pipeline variable by key
bitbottle workspace pipeline-variable delete WORKSPACE KEY [--confirm]
bitbottle workspace pipeline-variable delete myworkspace MY_VAR --confirm
```

| Flag | Default | Description |
|---|---|---|
| `--secured` | false | Mark as secured (value redacted on read) |
| `--confirm` | false | Skip deletion confirmation prompt |
| `--hostname HOST` | — | Override the Bitbucket host |
| `--json [FIELDS]` | — | JSON output |
| `--jq EXPR` | — | Filter JSON with jq expression |

Output fields for `list` and `get`: `key`, `value`, `secured`, `uuid` (JSON-only).

## Workspace Permissions (Cloud only)

Manage workspace-level user membership and per-repository effective permissions.
Returns `host.unsupported` on Bitbucket Server/DC.

### workspace perms list

List member-level permissions for a workspace.

```bash
bitbottle workspace perms list myworkspace
bitbottle workspace perms list myworkspace --json
bitbottle workspace perms list myworkspace --limit 100
```

Output columns: USER, PERMISSION

### workspace perms repo list

List effective per-repository permissions for all users in a workspace.

```bash
bitbottle workspace perms repo list myworkspace
bitbottle workspace perms repo list myworkspace --json
```

Output columns: REPO, USER, PERMISSION

### workspace perms grant

Grant a user a workspace-level permission. Valid permissions: `member`, `collaborator`, `owner`.

```bash
bitbottle workspace perms grant myworkspace --user alice --permission member
bitbottle workspace perms grant myworkspace --user bob --permission owner
```

### workspace perms revoke

Revoke a user's workspace-level permission.

```bash
bitbottle workspace perms revoke myworkspace --user alice --confirm
# Omit --confirm on a TTY to get an interactive confirmation prompt
```

## Workspace Project Permissions (Cloud only)

Manage per-project user and group permissions within a Cloud workspace. These
are distinct from workspace-level membership (`workspace perms`) and from
Server/DC project permissions (`perms project`). Returns `host.unsupported`
on Bitbucket Server/DC.

### workspace project perms list

List user and group permissions for a workspace project.

```bash
bitbottle workspace project perms list myworkspace PROJ
bitbottle workspace project perms list myworkspace PROJ --json
```

Output columns: SUBJECT, TYPE, PERMISSION

### workspace project perms grant

Grant a user or group a permission on a workspace project.
Valid permissions: `read`, `write`, `admin`, `create-repo`.

```bash
bitbottle workspace project perms grant myworkspace PROJ --user alice --permission write
bitbottle workspace project perms grant myworkspace PROJ --group devs --permission read
```

### workspace project perms revoke

Revoke a user or group permission on a workspace project.

```bash
bitbottle workspace project perms revoke myworkspace PROJ --user alice --confirm
bitbottle workspace project perms revoke myworkspace PROJ --group devs --confirm
# Omit --confirm on a TTY to get an interactive confirmation prompt
```

## Workspace Project Default Reviewers (Cloud only)

Manage default reviewers at the project level. Default reviewers cascade to
every repository in the project. These are distinct from per-repo default
reviewers (`repo default-reviewer *`) and workspace project permissions.
Returns `host.unsupported` on Bitbucket Server/DC.

### workspace project default-reviewer list

List default reviewers for a workspace project.

```bash
bitbottle workspace project default-reviewer list myworkspace PROJ
bitbottle workspace project default-reviewer list myworkspace PROJ --json
```

Output columns: ACCOUNT ID, DISPLAY NAME, NICKNAME

### workspace project default-reviewer add

Add a user as a default reviewer on a workspace project.

```bash
bitbottle workspace project default-reviewer add myworkspace PROJ --user abc123
```

### workspace project default-reviewer remove

Remove a user from the default reviewers of a workspace project.

```bash
bitbottle workspace project default-reviewer remove myworkspace PROJ --user abc123 --confirm
# Omit --confirm on a TTY to get an interactive confirmation prompt
```
