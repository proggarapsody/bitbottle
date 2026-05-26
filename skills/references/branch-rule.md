# Branch Restriction Rules

Branch restriction rules control who can push to a branch and what conditions
must be met before a merge. **Cloud only** — Bitbucket Server/DC uses
`branch protect` instead.

## Commands

```bash
# List all branch restriction rules for a repository
bitbottle branch-rule list WORKSPACE/REPO

# Add a rule requiring 2 approvals before merging to main
bitbottle branch-rule add WORKSPACE/REPO --kind require_approvals_to_merge --pattern main --value 2

# Add a push restriction (prevent direct pushes to main)
bitbottle branch-rule add WORKSPACE/REPO --kind push --pattern main

# Update a rule by numeric ID (at least one flag required)
bitbottle branch-rule update WORKSPACE/REPO 7 --pattern release/*
bitbottle branch-rule update WORKSPACE/REPO 7 --value 3
bitbottle branch-rule update WORKSPACE/REPO 7 --users alice,bob
bitbottle branch-rule update WORKSPACE/REPO 7 --groups dev-team

# Delete a rule by numeric ID
bitbottle branch-rule delete WORKSPACE/REPO 7
```

`list`, `add`, and `update` support `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |
| `add` | `--kind STRING` | Branch restriction kind (required) |
| `add` | `--pattern STRING` | Branch pattern to restrict (required) |
| `add` | `--value N` | Numeric value for the rule, e.g. required approvals (optional, default 0) |
| `update` | `--pattern STRING` | New branch pattern |
| `update` | `--users u1,u2` | Comma-separated user slugs (replaces existing; empty string clears) |
| `update` | `--groups g1,g2` | Comma-separated group slugs (replaces existing; empty string clears) |
| `update` | `--value N` | Numeric value for the rule |

## Common kinds

| Kind | Value meaning |
|---|---|
| `push` | Prevent direct pushes; no value needed |
| `restrict_merges` | Prevent merges; no value needed |
| `require_approvals_to_merge` | Number of required approvals |
| `require_default_reviewer_approvals_to_merge` | Number of required default reviewer approvals |
| `require_passing_builds_to_merge` | Number of required passing builds |
| `require_tasks_to_be_completed` | No value needed |
| `force` | Prevent force pushes; no value needed |
| `delete` | Prevent branch deletion; no value needed |

## JSON output

```bash
# List as JSON
bitbottle branch-rule list myworkspace/my-service --json

# Extract just IDs and kinds with jq
bitbottle branch-rule list myworkspace/my-service --json --jq '.[].kind'
```

JSON fields: `id` (int), `kind` (string), `pattern` (string), `value` (int, omitted when 0).

## MCP tools

| Tool | Description |
|---|---|
| `list_branch_rules` | List branch restriction rules. Params: `repo` (required), `hostname` |
| `add_branch_rule` | Add a rule. Params: `repo`, `kind` (required), `pattern` (required), `value` (int), `hostname` |
| `update_branch_rule` | Update a rule. Params: `repo`, `id` (required int), `pattern`, `users`, `groups`, `value` (int), `hostname` |
| `delete_branch_rule` | Delete a rule. Params: `repo`, `id` (required int), `hostname` |

## Backend details

**Cloud only** — `GET/POST/PUT/DELETE /repositories/{workspace}/{slug}/branch-restrictions[/{id}]`

Returns paginated `{"values": [...]}`. Each entry: `{"id": 1, "kind": "push", "pattern": "main", "value": 0}`.

Server/DC does not support this API; calling a branch-rule command against a
Server host returns a `host.unsupported` error.
