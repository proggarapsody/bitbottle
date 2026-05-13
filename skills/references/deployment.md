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

Use `bitbottle variable --scope deployment --env ENV-UUID` for deployment environment variables:

```bash
bitbottle variable list   WORKSPACE/REPO --scope deployment --env ENV-UUID
bitbottle variable set    WORKSPACE/REPO KEY [VALUE] --scope deployment --env ENV-UUID [--secured]
bitbottle variable delete WORKSPACE/REPO KEY --scope deployment --env ENV-UUID --confirm
```

**JSON fields**: `uuid`, `key`, `value`, `secured`

## MCP tools

| Tool | Purpose |
|---|---|
| `list_deployments` | List deployments; `repo` = WORKSPACE/REPO, optional `limit` |
| `get_deployment` | Get single deployment; `repo` + `uuid` |
| `list_environments` | List environments; `repo` |
| `create_environment` | Create environment; `repo`, `name`, `type`, optional `rank` |
| `delete_environment` | Delete environment; `repo`, `env_uuid` |
Use the unified `variable_list`, `variable_set`, `variable_delete` MCP tools with `scope=deployment` and `env_uuid` for deployment environment variables. See `variable.md`.

Secured variable values are **never returned** — the value field is always blank for secured vars.
