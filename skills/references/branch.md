# Branch Compare

Compare two branches or commits to see ahead/behind counts and commit lists. Supported on **Cloud and Server/DC**.

## Commands

```bash
# Compare feature branch against main
bitbottle branch compare main..feature WORKSPACE/REPO

# Compare in JSON format
bitbottle branch compare main..feature WORKSPACE/REPO --json

# Limit commits per side
bitbottle branch compare main..feature WORKSPACE/REPO --limit 10
```

`compare` supports `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Flag | Description |
|---|---|
| `--limit N` | Max commits to return per side (default 30) |
| `--hostname HOST` | Override the Bitbucket host |

## JSON output

Fields: `base` (string), `head` (string), `ahead_by` (int), `behind_by` (int), `commits_ahead` (array), `commits_behind` (array).

Each commit in the arrays has the standard Commit shape: `hash`, `message`, `author`, `timestamp`, `web_url`.

## MCP tools

| Tool | Description |
|---|---|
| `compare_refs` | Compare two refs. Params: `repo` (required, `WORKSPACE/REPO`), `base` (required), `head` (required), `limit`, `hostname` |

## Backend details

**Cloud** — Two paginated calls to `GET /2.0/repositories/{ws}/{slug}/commits/{ref}?exclude={other}`: one for ahead commits (commits in HEAD not in BASE) and one for behind commits.

**Server/DC** — `GET /rest/api/1.0/projects/{key}/repos/{slug}/compare/commits?from={ref}&to={other}&limit={n}` called twice.
