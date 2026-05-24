# Group management reference

Bitbucket Server / Data Center only. Cloud returns `host.unsupported`.

## Commands

```
group list [--filter PREFIX] [--hostname HOST] [--limit N] [--json] [--jq EXPR]
group create NAME [--hostname HOST]
group delete NAME [--confirm] [--hostname HOST]
group member list NAME [--hostname HOST] [--limit N] [--json] [--jq EXPR]
group member add NAME USER [--hostname HOST]
group member remove NAME USER [--confirm] [--hostname HOST]
```

## Examples

```bash
# List all groups
bitbottle group list --hostname git.example.com

# Filter by prefix
bitbottle group list --filter dev --hostname git.example.com

# Create a group
bitbottle group create qa-team --hostname git.example.com

# Delete a group (non-interactive)
bitbottle group delete oldgroup --confirm --hostname git.example.com

# List members of a group
bitbottle group member list developers --hostname git.example.com

# Add a user to a group
bitbottle group member add developers alice --hostname git.example.com

# Remove a user from a group (non-interactive)
bitbottle group member remove developers alice --confirm --hostname git.example.com
```

## MCP tools

| Tool | Description |
|---|---|
| `list_groups` | List admin groups (optional `filter`, `limit`) |
| `create_group` | Create an admin group (`name` required) |
| `delete_group` | Delete an admin group (`name` required) |
| `list_group_members` | List members of a group (`group` required, optional `limit`) |
| `add_group_member` | Add user to a group (`group`, `user` required) |
| `remove_group_member` | Remove user from a group (`group`, `user` required) |

All tools accept an optional `hostname` parameter. All return `host.unsupported` on Cloud.

## Notes

- Groups are Bitbucket Server/DC internal security groups (different from Bitbucket Cloud workspace groups).
- `perms project grant --group NAME` and `perms repo grant --group NAME` reference these group names.
- `pr reviewer-group` conditions reference these group names via the reviewer slug.
