# bitbottle deployments & environments _(Cloud only)_

All `list` commands support `--limit N`, `--json fields`, and
`--jq 'expr'` (see SKILL.md safety rule 4 for field discovery).

Deployment and environment commands require Bitbucket Cloud. Invocations
against a Server / DC host return a typed "unsupported on host" error.

## Deployments

```bash
bitbottle deployment list  WORKSPACE/REPO [--limit 10]
bitbottle deployment view  WORKSPACE/REPO UUID
```

`deployment list` shows the most recent deployments (default 10).
TTY output includes UUID, State, Environment, Release name, and short
commit hash. Pipe through `--json` for automation.

`deployment view` shows a single deployment by UUID. Prints UUID, State,
Environment name, Release name, and a 7-character commit hash prefix.

**JSON fields**: `uuid`, `state`, `environment`, `release`

**States**: `COMPLETED`, `FAILED`, `IN_PROGRESS`, `UNDEPLOYED`, `STOPPED`

## Environments

```bash
bitbottle environment list   WORKSPACE/REPO
bitbottle environment create WORKSPACE/REPO --name NAME --type TYPE [--rank N]
bitbottle environment delete WORKSPACE/REPO ENV-UUID [--confirm]
```

`environment list` shows all environments for the repo (UUID, Name, Type,
Rank).

`environment create` creates a new deployment environment. `--type` must
be one of `Test`, `Staging`, or `Production`. `--rank` sets numeric
ordering (optional).

`environment delete` is destructive — pass `--confirm` on non-TTY or
the command errors out. Confirm in interactive sessions when prompted.

**JSON fields**: `uuid`, `name`, `type`, `rank`

## Environment Variables

```bash
bitbottle environment variable list   WORKSPACE/REPO ENV-UUID
bitbottle environment variable set    WORKSPACE/REPO ENV-UUID KEY VALUE [--secured]
bitbottle environment variable delete WORKSPACE/REPO ENV-UUID VAR-UUID
```

`environment variable list` shows variables for an environment. Secured
variable values are shown as `<secured>`.

`environment variable set` creates or updates a variable by key. Pass
`--secured` to mark the value as secured (it will be redacted on
subsequent reads).

`environment variable delete` removes a variable by UUID (not by key).
Obtain the UUID first with `environment variable list --json uuid,key`.

**JSON fields**: `uuid`, `key`, `value`, `secured`

## MCP tools

| Tool | Purpose |
|---|---|
| `list_deployments` | List deployments; `repo` = WORKSPACE/REPO, optional `limit` |
| `get_deployment` | Get single deployment; `repo` + `uuid` |
| `list_environments` | List environments; `repo` |
| `create_environment` | Create environment; `repo`, `name`, `type`, optional `rank` |
| `delete_environment` | Delete environment; `repo`, `env_uuid` |
| `list_env_variables` | List variables (secured values blanked); `repo`, `env_uuid` |
| `set_env_variable` | Upsert variable; `repo`, `env_uuid`, `key`, `value`, optional `secured` |
| `delete_env_variable` | Delete variable by UUID; `repo`, `env_uuid`, `key` (the variable UUID) |

Secured variable values are **never returned** by `list_env_variables` or
`set_env_variable` — the value field is always blank for secured vars.
