# variable — Reference Card

Top-level command group for managing Bitbucket Cloud pipeline variables across three scopes.

## Syntax

```
bitbottle variable list   PROJECT/REPO [--scope repository|workspace|deployment] [--env ENV-UUID] [--json FIELDS] [--jq EXPR] [--hostname HOST]
bitbottle variable view   PROJECT/REPO KEY [--scope repository|workspace|deployment] [--env ENV-UUID] [--json] [--hostname HOST]
bitbottle variable set    PROJECT/REPO KEY [VALUE] [--body VALUE_OR_DASH] [--secured] [--scope ...] [--env ENV-UUID] [--hostname HOST]
bitbottle variable delete PROJECT/REPO KEY [--scope ...] [--env ENV-UUID] [--confirm] [--hostname HOST]
```

## Scopes

| Scope | Description | Backend API |
|-------|-------------|-------------|
| `repository` (default) | Repository-level pipeline variables | `PipelineClient.ListPipelineVariables / GetPipelineVariable / SetPipelineVariable / DeletePipelineVariable` |
| `workspace` | Workspace-level pipeline variables | `WorkspaceVariableClient.ListWorkspaceVariables / SetWorkspaceVariable / DeleteWorkspaceVariable` |
| `deployment` | Deployment environment variables | `DeploymentClient.ListEnvVariables / SetEnvVariable / DeleteEnvVariable` |

All scopes are Bitbucket Cloud only. The command returns `host.unsupported` on Server/DC.

## Notes

- **`view` fetches one variable by key**: resolves via a list-then-filter internally (the Bitbucket API has no single-key GET for these scopes). Returns a `host.not_found` error when the key is absent.
- **Secured variables**: use `--secured` on `set` to mark a variable as secured. The API never returns the value of secured variables; bitbottle renders `<secured>` in their place.
- **Deployment scope requires `--env ENV-UUID`**: the environment UUID must be supplied for all three subcommands when using `--scope deployment`.
- **`set` upserts**: if a variable with the same key already exists it is updated (PUT); otherwise it is created (POST). This applies to all three scopes.
- **`delete` deployment scope**: resolves UUID internally by listing all env variables and matching by key, so callers only need to know the key.
- **Stdin for secrets**: pass `--body=-` to read the variable value from stdin, keeping it out of shell history.

## Examples

```sh
# List repository variables
bitbottle variable list myworkspace/myrepo

# List workspace variables
bitbottle variable list myworkspace/myrepo --scope workspace

# List deployment environment variables
bitbottle variable list myworkspace/myrepo --scope deployment --env aaaabbbb-cccc-dddd-eeee-ffffgggghhhh

# View a single repository variable by key
bitbottle variable view myworkspace/myrepo DEPLOY_ENV

# View a workspace variable
bitbottle variable view myworkspace/myrepo GLOBAL_FLAG --scope workspace

# View a deployment environment variable
bitbottle variable view myworkspace/myrepo DB_URL --scope deployment --env <env-uuid>

# Set a repository variable
bitbottle variable set myworkspace/myrepo DEPLOY_ENV production

# Set a secured variable via stdin
echo -n "$SECRET_VALUE" | bitbottle variable set myworkspace/myrepo API_TOKEN --secured --body=-

# Set a workspace variable
bitbottle variable set myworkspace/myrepo GLOBAL_FLAG true --scope workspace

# Set a deployment environment variable
bitbottle variable set myworkspace/myrepo DB_URL postgres://... --scope deployment --env <env-uuid>

# Delete a repository variable (confirms unless --confirm supplied)
bitbottle variable delete myworkspace/myrepo OLD_VAR --confirm

# Delete a deployment environment variable by key
bitbottle variable delete myworkspace/myrepo DB_URL --scope deployment --env <env-uuid> --confirm
```
