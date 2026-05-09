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
| `pr` | `list` `view` `create` `merge` `approve` `unapprove` `diff` `checkout` `edit` `decline` `reopen` `ready` `request-review` `comment` |
| `repo` | `list` `view` `create` `delete` `clone` `set-default` `rename` `fork` _(Cloud)_ `file get` `tree` |
| `branch` | `list` `create` `delete` `checkout` |
| `tag` | `list` `create` `delete` |
| `webhook` | `list` `view` `create` `delete` |
| `commit` | `log` `view` |
| `pipeline` | `list` `view` `run` _(Cloud only)_ |
| `workspace` | `list` _(Cloud only)_ |
| `project` | `list WORKSPACE` _(Cloud only)_ |
| `issue` | `list` `view` `create` `close` `edit` `reopen` `assign` `comment {list\|add\|edit\|delete}` _(Cloud only)_ |
| `search` | `code QUERY` _(Cloud only)_ |
| `api` | Raw REST passthrough with pagination, `--jq`, variable expansion |
| `alias` | Custom command shortcuts |
| `config` | Editor, pager, git protocol per-host |
| `completion` | `bash` `zsh` `fish` `powershell` |
| `mcp` | MCP stdio server for AI assistants |
| `context` | One-call orientation: host, repo, branch, default branch, ahead/behind, user, backend |

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

# Decline / reopen (reopen is Bitbucket Server / DC only)
bitbottle pr decline 42
bitbottle pr reopen  42

# Read review comments — general + inline (file:line) anchors and replies
bitbottle pr comment list 42                            # all comments; LOCATION column shows path:line for inline
bitbottle pr comment list 42 --inline                   # only inline review comments
bitbottle pr comment list 42 --json id,inline,parentId,resolved,updatedAt

# Write review comments — general, inline, replies
bitbottle pr comment add 42 --body "LGTM!"                              # general comment
bitbottle pr comment add 42 --inline pkg/foo.go:88 --body "rename?"      # inline (new side, single line)
bitbottle pr comment add 42 --inline pkg/foo.go:10-15 --body "..."      # multi-line range (Cloud only)
bitbottle pr comment add 42 --inline pkg/foo.go:7 --side old --body ".." # comment on the old / removed side
bitbottle pr comment add 42 --parent 1234 --body "agreed"                # reply to an existing thread

# Edit / delete / resolve a comment
bitbottle pr comment edit   42 1234 --body "updated"
bitbottle pr comment delete 42 1234
bitbottle pr comment resolve 42 1234                                     # Bitbucket Cloud only
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

#### Branch protection _(Bitbucket Server / DC only)_

Branch restrictions on Bitbucket Server / DC. Cloud has a different model
that isn't wired up yet — invocations against Cloud return a typed
"unsupported on host" error.

```bash
bitbottle branch protect list   MYPROJ/my-service
bitbottle branch protect create MYPROJ/my-service \
  --type fast-forward-only --branch main \
  --user alice --group devs
bitbottle branch protect delete MYPROJ/my-service 42
```

`--type` accepts `read-only`, `no-deletes`, `fast-forward-only`, or
`pull-request-only`. Pass exactly one of `--branch NAME` (literal branch)
or `--pattern GLOB` (e.g. `release/*`).

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

### Reading source at a ref

Read file content and directory listings at any ref (branch, tag, commit
hash) without cloning. Both backends; the bytes round-trip cleanly so
binary files survive `--out`.

```bash
# Read a file at a ref — straight to stdout
bitbottle repo file get MYPROJ/my-service README.md --ref main

# Pin a tag and write to disk (binary-safe)
bitbottle repo file get MYPROJ/my-service logo.png --ref v1.2.0 --out logo.png

# List a directory at a ref. PATH defaults to the repo root.
bitbottle repo tree MYPROJ/my-service --ref main
bitbottle repo tree MYPROJ/my-service cmd --ref main

# Structured output for scripts and agents
bitbottle repo tree MYPROJ/my-service --ref main --json path,type,size
bitbottle repo tree MYPROJ/my-service --ref main --jq '.[]|select(.type=="dir").path'
```

`type` is normalised to `file` or `dir` across both backends. Submodules
surface as `dir` (with the submodule pointer in `hash`) so renderers can
recurse uniformly without a special case. The MCP equivalents are
`get_file_content` and `list_tree`.

### Workspaces & Projects (Cloud only)

```bash
# Workspaces the authenticated user belongs to
bitbottle workspace list
bitbottle workspace list --json slug,name --jq '.[].slug'

# Projects within a workspace
bitbottle project list myworkspace
bitbottle project list myworkspace --limit 100
```

Both commands surface a typed unsupported-capability error against
Bitbucket Server / Data Center hosts (workspaces are a Cloud concept).

### Search _(Cloud only)_

```bash
# Search across the workspace inferred from the current checkout's pinned default
bitbottle search code 'TODO'

# Explicit workspace, custom limit
bitbottle search code 'foo bar' --workspace myws --limit 50

# Path-restricted query, JSON for scripting
bitbottle search code 'path:README' --workspace myws --json path,repository
bitbottle search code 'TODO' --workspace myws --json path --jq '.[].path'
```

Bitbucket Cloud's query language (`path:`, `lang:`, `repo:`, exact-phrase
quoting, etc.) is passed through verbatim — bitbottle does not translate
operators. The matched-segment shape (`pathMatches`, `contentMatches`)
is preserved on the JSON side so renderers can highlight the matched
runs. Bitbucket Server / Data Center has no first-class code-search
REST endpoint; invocations against a Server host return the typed
`host.unsupported` error.

### Code Insights _(Bitbucket Server / DC only)_

Post build / quality / security analysis results as structured annotations
on commits. Requires Bitbucket Server / Data Center.

```bash
# Create or update a report (upsert by key)
bitbottle --hostname git.example.com code-insights report set \
  MYPROJ/my-service abc123def456 my-scanner \
  --title "Security Scan" --result PASS \
  --report-type SECURITY --reporter "semgrep"

# List all reports for a commit
bitbottle --hostname git.example.com code-insights report list \
  MYPROJ/my-service abc123def456

# View a single report as JSON
bitbottle --hostname git.example.com code-insights report view \
  MYPROJ/my-service abc123def456 my-scanner \
  --json key,title,result,details

# Bulk-add annotations from a JSON file
bitbottle --hostname git.example.com code-insights annotation add \
  MYPROJ/my-service abc123def456 my-scanner \
  --from-json @findings.json

# Single annotation
bitbottle --hostname git.example.com code-insights annotation add \
  MYPROJ/my-service abc123def456 my-scanner \
  --path src/main.go --line 42 --severity HIGH --type BUG \
  --message "Potential null dereference"

# List annotations for a report
bitbottle --hostname git.example.com code-insights annotation list \
  MYPROJ/my-service abc123def456 my-scanner --json path,severity,message

# Delete a report and all its annotations
bitbottle --hostname git.example.com code-insights report delete \
  MYPROJ/my-service abc123def456 my-scanner
```

#### Merge checks (experimental)

Configure a Code Insights report as a required gate for PR merges. The
underlying API is partly undocumented — commands are marked experimental.

```bash
# Require the "my-scanner" report to pass before merging
bitbottle --hostname git.example.com code-insights merge-check set \
  MYPROJ/my-service required-scan \
  --report-key my-scanner --must-pass --min-severity MEDIUM

# Inspect the current merge-check configuration
bitbottle --hostname git.example.com code-insights merge-check get \
  MYPROJ/my-service required-scan

# Remove the merge check
bitbottle --hostname git.example.com code-insights merge-check delete \
  MYPROJ/my-service required-scan
```

Invoking any `code-insights` command against a Bitbucket Cloud host returns
the typed `host.unsupported` error.

### Raw API

```bash
# Any endpoint not yet wrapped
bitbottle api 2.0/user
bitbottle api --paginate --jq '.[].full_name' '2.0/repositories/{workspace}'
bitbottle api -X POST -F 'title=hotfix' -F 'source.branch.name=hotfix/x' \
  '2.0/repositories/{workspace}/{repo_slug}/pullrequests'
```

### Context (one-call orientation)

`bitbottle context` collapses three previously independent calls
(`auth status`, `repo view`, `git status`) into a single structured
snapshot of where you are. Especially useful for AI agents that drive
bitbottle through MCP — `get_context` is the standard first call.

```bash
bitbottle context
# Host:           git.example.com
# Backend:        server
# Project:        PROJ
# Slug:           repo
# Branch:         feat/ctx
# Default branch: main
# Ahead/Behind:   2 / 0
# User:           alice (Alice Smith)

bitbottle context --json host,project,slug,branch,default_branch,ahead,behind,user,backend
bitbottle context --json user --jq '.user.slug'
```

Outside a git repo `project`, `slug`, `branch`, and `default_branch`
are empty; `ahead` and `behind` are omitted from the JSON output
entirely (and rendered as `(unknown — run 'git fetch')` in the human
table). They are also omitted whenever `git rev-list` cannot compute
the counts — `0 / 0` would falsely read as "in sync". `host`, `user`,
and `backend` still resolve via the configured (or `--hostname`) host.

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

## Colour & pager

State columns in TTY tables are colourised: `OPEN` / `SUCCESSFUL` /
active webhooks render green; `MERGED` magenta; `DECLINED` /
`SUPERSEDED` / `FAILED` / `ERROR` / `STOPPED` / inactive webhooks red;
`IN_PROGRESS` / `PENDING` yellow. Colour applies to TTY table output
only — JSON (`--json`) and piped output stay raw.

Disable colour with the global `--no-color` flag or by setting
`NO_COLOR=1` in the environment. JSON output is never colourised.

Long-output commands (`pr diff`, `commit log`) pipe through `$PAGER`
when stdout is a TTY (default `less -FRX`). Set `PAGER=cat` to disable.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Run `go test ./...` before sending a PR.

## License

[MIT](LICENSE)
