# Deploy Keys

Deploy keys are SSH public keys attached to a repository that grant read (or
read/write) access without requiring a full user account. Both Bitbucket Cloud
and Server/DC support deploy keys.

## Commands

```bash
# List all deploy keys for a repository
bitbottle deploy-key list PROJECT/REPO
bitbottle deploy-key list WORKSPACE/REPO          # Cloud

# Add a deploy key
bitbottle deploy-key add PROJECT/REPO --key "ssh-rsa AAAA..." --label "CI server"
bitbottle deploy-key add PROJECT/REPO --key "ssh-rsa AAAA..."   # label optional
bitbottle deploy-key add WORKSPACE/REPO --key "ssh-rsa AAAA..." --permission read-write  # Cloud: grant write access

# Delete a deploy key by numeric ID
bitbottle deploy-key delete PROJECT/REPO 42
```

`list` and `add` support `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |
| `add` | `--key STRING` | SSH public key (required) |
| `add` | `--label STRING` | Human-readable label (optional) |
| `add` | `--permission read\|read-write` | Key permission (Cloud only; default: `read`) |

## JSON output

```bash
# List as JSON
bitbottle deploy-key list MYPROJ/my-service --json

# Extract just IDs and labels with jq
bitbottle deploy-key list MYPROJ/my-service --json --jq '.[].label'
```

JSON fields: `id` (int), `label` (string), `key` (string), `readOnly` (bool,
Cloud only — always false on Server/DC).

## MCP tools

| Tool | Description |
|---|---|
| `list_deploy_keys` | List deploy keys. Params: `repo` (required), `hostname` |
| `add_deploy_key` | Add a deploy key. Params: `repo`, `key` (required), `label`, `permission` (`read`\|`read-write`), `hostname` |
| `delete_deploy_key` | Delete a deploy key. Params: `repo`, `id` (required int), `hostname` |

## Backend details

**Cloud** — `GET/POST/DELETE /repositories/{workspace}/{slug}/deploy-keys`

**Server/DC** — `GET/POST/DELETE /rest/api/1.0/projects/{ns}/repos/{slug}/ssh`

The Server wire shape nests the key text inside a `key` object:
`{"id":1,"label":"CI key","key":{"text":"ssh-rsa AAAA...","label":"CI key"}}`.
The CLI maps this transparently — callers see the same `DeployKey` domain type
on both backends.
