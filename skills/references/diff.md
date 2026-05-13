# Diff Between Refs

Show the unified diff or a file-change summary between two refs (branches,
tags, or commit hashes) in a Bitbucket repository. Works on both Bitbucket
Cloud and Bitbucket Server/Data Center.

## Commands

```bash
# Full unified diff between two refs (single-arg form)
bitbottle diff main..feature

# Full unified diff (two-arg form)
bitbottle diff main feature

# Diff for a specific repository
bitbottle diff main..feature MYPROJECT/my-service

# Show only the summary (files changed, insertions, deletions)
bitbottle diff main..feature --stat

# Diff against a specific host
bitbottle diff main..feature --hostname git.example.com
```

## Flags

| Flag | Description |
|---|---|
| `--stat` | Print a summary of changed files instead of the full diff |
| `--hostname HOST` | Override the Bitbucket host |

## Argument forms

| Form | Example |
|---|---|
| `REF1..REF2` (single arg) | `bitbottle diff main..feature` |
| `REF1 REF2` (two args) | `bitbottle diff main feature` |
| With repo override | `bitbottle diff main..feature PROJ/repo` |
| Two-arg with repo | `bitbottle diff main feature PROJ/repo` |

## Output

Without `--stat`: raw unified diff text (suitable for piping to `patch` or
other tools).

With `--stat`:
```
3 files changed, 15 insertions(+), 4 deletions(-)
  M  api/foo.go   (+10/-2)
  A  api/bar.go   (+5/0)
  D  api/old.go   (0/-2)
```

Status letters: `M` = modified, `A` = added, `D` = deleted, `R` = renamed.

## MCP tools

| Tool | Description |
|---|---|
| `get_diff` | Get the unified diff between two refs. Params: `repo` (required), `from` (required), `to` (required), `hostname` |
| `get_diff_stat` | Get a summary of files changed between two refs. Params: `repo` (required), `from` (required), `to` (required), `hostname` |

### Example MCP usage

```json
// get_diff
{"repo": "myworkspace/my-service", "from": "main", "to": "feature/new-api"}

// get_diff_stat
{"repo": "MYPROJECT/my-service", "from": "v1.0.0", "to": "v1.1.0", "hostname": "git.example.com"}
```

## Backend notes

- **Cloud**: `GET /repositories/{ws}/{slug}/diff/{from}..{to}` (raw text) and
  `GET /repositories/{ws}/{slug}/diffstat/{from}..{to}` (JSON).
- **Server/DC**: `GET /rest/api/1.0/projects/{ns}/repos/{slug}/diff?since={from}&until={to}` (JSON,
  reconstructed as unified diff text).
