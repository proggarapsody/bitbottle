# bitbottle

> Bitbucket CLI for Cloud and Server / Data Center — built with the same philosophy as [GitHub CLI](https://github.com/cli/cli).

[![CI](https://github.com/proggarapsody/bitbottle/actions/workflows/ci.yml/badge.svg)](https://github.com/proggarapsody/bitbottle/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/@proggarapsody/bitbottle)](https://www.npmjs.com/package/@proggarapsody/bitbottle)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Work with Bitbucket from your terminal — pull requests, repos, branches, tags, commits, pipelines, and raw API access. One tool, both backends, same commands.

```
$ bitbottle pr list
#47  feat: seamless audit            main ← feat/audit       OPEN   alice
#46  fix: 409 on concurrent merge    main ← fix/merge-race   MERGED bob

$ bitbottle pr create --title "fix: handle empty diff" --base main
✓ Created PR #48 — https://bitbucket.example.com/projects/PROJ/repos/api/pull-requests/48
```

---

## Installation

**npm** (recommended — works everywhere Node is installed):
```bash
npm install -g @proggarapsody/bitbottle
```

This also registers the bundled AI agent skill (`SKILL.md` + topic
references) with any agent runtimes detected on your machine —
Claude Code, Cursor, Codex, Gemini CLI, and 50+ others — so your
agents know how to drive bitbottle correctly. Skip with
`BITBOTTLE_SKIP_SKILL_INSTALL=1 npm install -g …`.

If you installed via Homebrew, Go install, or the bare binary,
register the skill explicitly:
```bash
bitbottle skill install            # uses npx; needs Node.js >= 18
bitbottle skill path               # show where the skill was installed
bitbottle skill remove             # uninstall
```

**Go install:**
```bash
go install github.com/proggarapsody/bitbottle/cmd/bitbottle@latest
```

**Homebrew / binary / deb / rpm / Docker** — see the [latest release](https://github.com/proggarapsody/bitbottle/releases/latest).

---

## Authentication

```bash
# Bitbucket Cloud
echo "YOUR_APP_PASSWORD" | bitbottle auth login --hostname bitbucket.org --with-token

# Bitbucket Server / Data Center (PAT, self-signed cert)
echo "BBDC-YOUR-PAT" | bitbottle auth login \
  --hostname git.example.com \
  --with-token \
  --git-protocol https \
  --skip-tls-verify

# Verify
bitbottle auth status
```

Credentials are stored in `~/.config/bitbottle/hosts.yml`. Inside a git repo with a Bitbucket remote the host and project/repo are detected automatically. Outside a repo, use `-R HOST/PROJECT/REPO`.

---

## Commands

| Group | Commands |
|---|---|
| `auth` | `login` `logout` `status` `token` `refresh` |
| `pr` | `list` `view` `create` `merge` `approve` `unapprove` `diff` `checkout` `edit` `decline` `ready` `request-review` `comment` |
| `repo` | `list` `view` `create` `delete` `clone` `set-default` `rename` `fork` _(Cloud)_ |
| `branch` | `list` `create` `delete` `checkout` |
| `tag` | `list` `create` `delete` |
| `webhook` | `list` `view` `create` `delete` |
| `commit` | `log` `view` |
| `pipeline` | `list` `view` `run` _(Cloud only)_ |
| `api` | Raw REST passthrough with pagination, `--jq`, variable expansion |
| `alias` | Custom command shortcuts |
| `config` | Editor, pager, git protocol per-host |
| `completion` | `bash` `zsh` `fish` `powershell` |
| `mcp` | MCP stdio server for AI assistants |

All listing commands support `--json fields`, `--jq expr`, `--limit N`, `--hostname HOST`. TTY output is aligned and coloured; piped output is plain tab-separated for scripting.

---

## Key Workflows

### Pull Requests

```bash
# Open PRs in current repo
bitbottle pr list

# Create a PR
bitbottle pr create --title "feat: add retry logic" --base main

# Review
bitbottle pr diff 42 | delta
bitbottle pr checkout 42

# Approve and merge
bitbottle pr approve 42
bitbottle pr merge 42 --squash --delete-branch

# Add reviewers
bitbottle pr request-review 42 --reviewer alice --reviewer bob
```

### Repos & Branches

```bash
bitbottle repo list --limit 20
bitbottle repo create my-service --project MYPROJ
bitbottle repo clone MYPROJ/my-service

bitbottle branch list
bitbottle branch create MYPROJ/my-service feature/x --start-at main
bitbottle branch delete MYPROJ/my-service feature/x
```

### Pipelines _(Cloud only)_

```bash
bitbottle pipeline list  MYWORKSPACE/my-service
bitbottle pipeline run   MYWORKSPACE/my-service --branch main
bitbottle pipeline view  MYWORKSPACE/my-service {uuid} --web

# Drill into a run:
bitbottle pipeline steps MYWORKSPACE/my-service {pipeline-uuid}
bitbottle pipeline logs  MYWORKSPACE/my-service {pipeline-uuid} {step-uuid}

# Repository-level pipeline variables (upsert by KEY):
bitbottle pipeline variable list   MYWORKSPACE/my-service
bitbottle pipeline variable set    MYWORKSPACE/my-service DEPLOY_ENV prod
echo "$TOKEN" | bitbottle pipeline variable set MYWORKSPACE/my-service \
  API_TOKEN --body=- --secured
bitbottle pipeline variable delete MYWORKSPACE/my-service DEPLOY_ENV --confirm
```

Secured variables redact their value on read (TTY column shows
`<secured>`, JSON `value` shows `"<secured>"`). Use `--body=-` to read
the value from stdin so the secret never touches shell history.

### Webhooks

```bash
bitbottle webhook list   MYPROJ/my-service
bitbottle webhook view   MYPROJ/my-service {id}

# Create
bitbottle webhook create MYPROJ/my-service \
  --url https://hooks.example.com/bb \
  --events 'repo:push,pullrequest:created'

# With shared secret from stdin (keeps secret out of shell history)
echo "$WEBHOOK_SECRET" | bitbottle webhook create MYPROJ/my-service \
  --url https://hooks.example.com/bb \
  --events 'repo:push' \
  --secret -

# Delete (destructive — --confirm required when not interactive)
bitbottle webhook delete MYPROJ/my-service {id} --confirm
```

Event keys differ between backends:
- **Cloud** uses `repo:push`, `pullrequest:created`, `pullrequest:approved`, …
- **Server/DC** uses `repo:refs_changed`, `pr:opened`, `pr:merged`, …

`--secret` accepts the raw value, `-` to read from stdin, or `@PATH` to read
from a file. Trailing newlines from stdin / file are trimmed. Webhook
secrets are write-only — neither backend returns them on read.

### Repo extras

```bash
# Rename (both backends — slug derives from new name on Cloud).
# --confirm is required on non-TTY because the slug change breaks existing
# clones' origin URL — run `git remote set-url origin ...` after.
bitbottle repo rename MYPROJ/my-service my-service-v2 --confirm

# Fork into another workspace (Bitbucket Cloud only)
bitbottle repo fork myworkspace/my-service --into otherws
bitbottle repo fork myworkspace/my-service --into otherws --name my-fork
```

`repo fork` returns a typed unsupported-capability error on Bitbucket Server /
Data Center, which has no fork primitive in its REST API. Both `rename` and
`fork` accept `--json fields` and `--jq expr` for structured output.

### Raw API

```bash
# Any endpoint not yet wrapped
bitbottle api 2.0/user
bitbottle api --paginate --jq '.[].full_name' '2.0/repositories/{workspace}'
bitbottle api -X POST -F 'title=hotfix' -F 'source.branch.name=hotfix/x' \
  '2.0/repositories/{workspace}/{repo_slug}/pullrequests'
```

### Outside a git repo

```bash
bitbottle pr list   -R git.example.com/PROJ/api
bitbottle pr merge 42 -R git.example.com/PROJ/api
```

---

## MCP Server — AI Assistant Integration

`bitbottle mcp serve` exposes all CLI operations as [Model Context Protocol](https://modelcontextprotocol.io) tools. Claude Desktop, Claude Code, and any MCP client can call Bitbucket as native tools — no raw API, no output parsing.

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "bitbottle": {
      "command": "bitbottle",
      "args": ["mcp", "serve"]
    }
  }
}
```

**Claude Code** (`.mcp.json` in project root):
```json
{
  "mcpServers": {
    "bitbottle": {
      "command": "bitbottle",
      "args": ["mcp", "serve"]
    }
  }
}
```

**Docker** (no local install required):
```bash
docker run --rm -i \
  -v ~/.config/bitbottle:/root/.config/bitbottle \
  proggarapsody/bitbottle mcp serve
```

The MCP server uses the same `~/.config/bitbottle/hosts.yml` credentials as the CLI — no separate auth setup needed.

---

## Cloud vs Server / Data Center

bitbottle automatically routes to the correct backend based on hostname: `bitbucket.org` → Cloud; anything else → Server/DC. Override with `backend_type: cloud|server` in `hosts.yml`.

The same commands and flags work on both backends. Differences in API shape, pagination style, and endpoint paths are handled internally — no `--cloud` / `--server` flags needed.

---

## Shell Completion

```bash
bitbottle completion --shell bash   >> ~/.bash_profile
bitbottle completion --shell zsh    >> ~/.zshrc
bitbottle completion --shell fish   > ~/.config/fish/completions/bitbottle.fish
bitbottle completion --shell powershell >> $PROFILE
```

---

## Environment Variables

| Variable | Effect |
|---|---|
| `BB_TOKEN` | Override auth token |
| `BB_HOST` | Default hostname |
| `BB_REPO` | Override `[HOST/]PROJECT/REPO` |
| `BB_FORCE_TTY` | Force aligned/coloured output in pipes |
| `NO_COLOR` | Disable colour (standard) |

Full list in [`internal/envvars/envvars.go`](internal/envvars/envvars.go).

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Run `go test ./...` before sending a PR.

## License

[MIT](LICENSE)
