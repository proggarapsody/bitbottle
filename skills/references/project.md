# Project command reference

The `project` command group covers both Bitbucket Cloud and Server/DC,
but each verb applies to only one platform.

## Cloud-only

```
project list WORKSPACE [--limit N] [--hostname HOST] [--json] [--jq EXPR]
```

Lists projects inside a Bitbucket Cloud workspace. Each project is a logical
grouping of repositories.

## Server/DC-only

The following verbs operate on Bitbucket Server / Data Center projects (the
top-level namespace that groups repositories). All return `host.unsupported`
on Bitbucket Cloud.

```
project server-list [--filter PREFIX] [--limit N] [--hostname HOST] [--json] [--jq EXPR]
project view KEY [--hostname HOST] [--json] [--jq EXPR]
project create KEY --name NAME [--description TEXT] [--public] [--hostname HOST]
project edit KEY [--name NAME] [--description TEXT] [--public=BOOL] [--hostname HOST]
project delete KEY [--confirm] [--hostname HOST]
```

## Examples

```bash
# Cloud: list projects in a workspace
bitbottle project list myworkspace --hostname bitbucket.org

# Server: list all projects
bitbottle project server-list --hostname git.example.com

# Server: filter projects by name prefix
bitbottle project server-list --filter Dev --hostname git.example.com

# Server: view a project
bitbottle project view PRJ --hostname git.example.com

# Server: create a project
bitbottle project create PRJ --name "My Project" --hostname git.example.com

# Server: edit a project's name
bitbottle project edit PRJ --name "New Name" --hostname git.example.com

# Server: delete a project (non-interactive)
bitbottle project delete PRJ --confirm --hostname git.example.com
```

## MCP tools (Server/DC only)

| Tool | Description |
|---|---|
| `list_server_projects` | List projects (optional `filter`, `limit`) |
| `get_server_project` | Get a single project (`key` required) |
| `create_server_project` | Create a project (`key`, `name` required; optional `description`, `public`) |
| `update_server_project` | Update a project (`key` required; optional `name`, `description`, `public`) |
| `delete_server_project` | Delete a project (`key` required) |

All Server/DC MCP tools accept an optional `hostname` parameter and return
`host.unsupported` on Cloud.

## Notes

- Project `key` is a short uppercase identifier (e.g. `PRJ`, `DEV`). It is
  used as the project namespace in repo paths (`/projects/PRJ/repos/...`).
- `type` is always `NORMAL` for user-created projects; `PERSONAL` projects
  are auto-created by Bitbucket for each user's home namespace and cannot
  be managed via these commands.
- `project edit` requires at least one of `--name`, `--description`, or
  `--public` to be provided; omitting all flags is an error.
- `project delete` requires `--confirm` in non-interactive (non-TTY) mode.
