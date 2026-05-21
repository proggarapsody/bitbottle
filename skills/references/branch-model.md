# Branching Model

The branching model controls the development/production branch configuration and
branch-type naming prefixes used by Bitbucket's in-UI "Create branch" wizard.
**Cloud only** — Bitbucket Server/DC returns `host.unsupported`.

Distinct from `branch-rule` (which controls enforcement restrictions).

## Commands

```bash
# Show the effective branching model for a repository
bitbottle branch-model get WORKSPACE/REPO

# Show as JSON
bitbottle branch-model get WORKSPACE/REPO --json

# Update the development branch name
bitbottle branch-model set WORKSPACE/REPO --dev-branch develop

# Enable and name the production branch
bitbottle branch-model set WORKSPACE/REPO --prod-branch main --prod-enabled

# Update branch type prefixes (comma-separated kind=prefix pairs)
bitbottle branch-model set WORKSPACE/REPO --branch-type-prefix feature=feat/,hotfix=hf/

# Combine flags
bitbottle branch-model set WORKSPACE/REPO --dev-branch develop --branch-type-prefix feature=feat/
```

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |
| all | `--json` | Output as JSON |
| `set` | `--dev-branch NAME` | Development branch name |
| `set` | `--prod-branch NAME` | Production branch name |
| `set` | `--prod-enabled` | Enable the production branch |
| `set` | `--branch-type-prefix kind=prefix` | Repeatable or comma-separated kind=prefix pairs |

All `--branch-type-prefix` overrides are merged into existing kinds; kinds not
mentioned are preserved unchanged.

## Common branch kinds

| Kind | Default prefix |
|---|---|
| `feature` | `feature/` |
| `bugfix` | `bugfix/` |
| `hotfix` | `hotfix/` |
| `release` | `release/` |

## MCP tools

| Tool | Description |
|---|---|
| `get_branch_model` | Get the effective branching model. Params: `repo` (required), `hostname` |
| `set_branch_model` | Update branching model settings. Params: `repo` (required), `dev_branch`, `prod_branch`, `prod_enabled`, `branch_type_prefixes` (object), `hostname` |

`branch_type_prefixes` is a JSON object mapping kind to prefix:
```json
{"feature": "feat/", "hotfix": "hf/"}
```

## Backend details

**Cloud only** — three endpoints:

| Operation | Path |
|---|---|
| `GetBranchModel` | `GET /repositories/{workspace}/{slug}/branching-model` |
| `GetBranchModelSettings` | `GET /repositories/{workspace}/{slug}/branching-model/settings` |
| `UpdateBranchModelSettings` | `PUT /repositories/{workspace}/{slug}/branching-model/settings` |

`get` reads the effective model (resolved branch names).
`set` reads settings first to merge `branch_type_prefixes`, then PUTs the result.
