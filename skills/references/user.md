# User

Display the authenticated user's profile. Works on both Bitbucket Cloud and
Bitbucket Server / Data Center.

## Commands

```bash
# Display the currently authenticated user's profile
bitbottle user view

# JSON output (useful for scripting)
bitbottle user view --json

# Filter with jq
bitbottle user view --json slug,name --jq '.slug'

# Specific host (when multiple hosts are configured)
bitbottle user view --hostname git.example.com
```

## Flags

| Command | Flag | Default | Description |
|---|---|---|---|
| `user view` | `--hostname HOST` | — | Override the Bitbucket host |
| `user view` | `--json [FIELDS]` | — | JSON output (comma-separated field list optional) |
| `user view` | `--jq EXPR` | — | Filter JSON with jq expression |

## Output fields

### `user view`

| Field | JSON key | Notes |
|---|---|---|
| SLUG | `slug` | User's account slug / username |
| NAME | `name` | User's display name |

## MCP tools

- `get_current_user` — `hostname` (optional)
