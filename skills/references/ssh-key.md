# SSH Keys

User SSH keys grant SSH access to Bitbucket for the authenticated user.
Supported on both **Bitbucket Cloud** and **Bitbucket Server/DC**.

## Commands

```bash
# List SSH keys for the current user
bitbottle ssh-key list

# Add an SSH key
bitbottle ssh-key add --key "ssh-rsa AAAA..." --label "Laptop"
bitbottle ssh-key add --key "ssh-rsa AAAA..."   # label optional

# Delete an SSH key by numeric ID
bitbottle ssh-key delete 42
```

`list` and `add` support `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |
| `add` | `--key STRING` | SSH public key (required) |
| `add` | `--label STRING` | Human-readable label (optional) |

## JSON output

```bash
# List as JSON
bitbottle ssh-key list --json

# Extract just IDs and labels with jq
bitbottle ssh-key list --json --jq '.[].label'
```

JSON fields: `id` (int), `label` (string), `key` (string).

## MCP tools

| Tool | Description |
|---|---|
| `list_ssh_keys` | List SSH keys for the current user. Params: `hostname` |
| `add_ssh_key` | Add an SSH key. Params: `key` (required), `label`, `hostname` |
| `delete_ssh_key` | Delete an SSH key. Params: `id` (required int), `hostname` |

## Backend details

**Cloud** — calls `GET /user` to resolve the current user's nickname, then:
- `GET /users/{username}/ssh-keys` — paginated list
- `POST /users/{username}/ssh-keys` — add key
- `DELETE /users/{username}/ssh-keys/{id}` — delete key

**Server/DC** — calls the `/rest/ssh/1.0` REST root:
- `GET /keys` — paginated list (uses `isLastPage`/`nextPageStart`)
- `POST /keys` — add key; body `{"text": "<pub key>", "label": "<label>"}`
- `DELETE /keys/{id}` — delete by integer ID

Server wire format uses `"text"` for the key field; the CLI maps it to the `key` JSON field in output.
