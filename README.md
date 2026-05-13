# bitbottle

> Bitbucket CLI for Cloud and Server / Data Center — built with the same philosophy as [GitHub CLI](https://github.com/cli/cli).

[![CI](https://github.com/proggarapsody/bitbottle/actions/workflows/ci.yml/badge.svg)](https://github.com/proggarapsody/bitbottle/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/@proggarapsody/bitbottle)](https://www.npmjs.com/package/@proggarapsody/bitbottle)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/proggarapsody/bitbottle/badge)](https://securityscorecards.dev/viewer/?uri=github.com/proggarapsody/bitbottle)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/XXXX/badge)](https://www.bestpractices.dev/projects/XXXX)

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

**Nix:**
```sh
nix run github:proggarapsody/bitbottle -- --version
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

Tokens are intentionally stripped from `hosts.yml` on every save. If you have an existing `oauth_token` in `hosts.yml` (from an older version), run `bitbottle auth migrate` to move it to the OS keyring.

---

## Commands

| Group | Commands |
|---|---|
| `auth` | `login` `logout` `status` `token` `refresh` |
| `pr` | `list` `view` `create` `merge` `approve` `unapprove` `diff` `checkout` `edit` `decline` `reopen` `ready` `request-review` `comment` `default-reviewer {list\|add\|remove}` |
| `repo` | `list` `view` `create` `delete` `clone` `set-default` `rename` `fork` _(Cloud)_ `file get` `tree` |
| `branch` | `list` `create` `delete` `checkout` |
| `tag` | `list` `create` `delete` |
| `webhook` | `list` `view` `create` `delete` |
| `deploy-key` | `list` `add` `delete` |
| `branch-rule` | `list` `add` `delete` _(Cloud only)_ |
| `ssh-key` | `list` `add` `delete` _(Cloud only)_ |
| `commit` | `log` `view` `status` `comment {list\|add\|edit\|delete}` |
| `pipeline` | `list` `view` `run` _(Cloud only)_ |
| `deployment` | `list` `view` _(Cloud only)_ |
| `environment` | `list` `create` `delete` _(Cloud only)_ |
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

# Submit a compound review (body + inline comments + action) in one call
bitbottle pr review 42 --approve --body "lgtm overall"
bitbottle pr review 42 --request-changes --body "see comments" \
    --inline pkg/foo.go:88:please rename                                  # request-changes is Cloud only
bitbottle pr review 42 --comment --inline pkg/foo.go:10-15:extract helper # comment-only review

# PR activity event stream (approvals, comments, updates, merges)
bitbottle pr activity 42
bitbottle pr activity 42 --limit 20
bitbottle pr activity 42 --json type,actor,createdAt,detail
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
bitbottle variable list   MYWORKSPACE/my-service
bitbottle variable set    MYWORKSPACE/my-service DEPLOY_ENV prod
echo "$TOKEN" | bitbottle variable set MYWORKSPACE/my-service \
  API_TOKEN --body=- --secured
bitbottle variable delete MYWORKSPACE/my-service DEPLOY_ENV --confirm
```

Secured variables redact their value on read (TTY column shows
`<secured>`, JSON `value` shows `"<secured>"`). Use `--body=-` to read
the value from stdin so the secret never touches shell history.

### Deployments & Environments _(Cloud only)_

```bash
# List recent deployments
bitbottle deployment list  MYWORKSPACE/my-service
bitbottle deployment list  MYWORKSPACE/my-service --limit 25

# View a single deployment
bitbottle deployment view  MYWORKSPACE/my-service {deployment-uuid}

# List environments
bitbottle environment list MYWORKSPACE/my-service

# Create an environment (type must be Test, Staging, or Production)
bitbottle environment create MYWORKSPACE/my-service --name "QA" --type Test
bitbottle environment create MYWORKSPACE/my-service --name "Prod" --type Production --rank 10

# Delete an environment (destructive — --confirm required on non-TTY)
bitbottle environment delete MYWORKSPACE/my-service {env-uuid} --confirm

# Environment variables — use variable --scope deployment (see deprecation note below)
bitbottle variable list   MYWORKSPACE/my-service --scope deployment --env {env-uuid}
bitbottle variable set    MYWORKSPACE/my-service DEPLOY_KEY prod-val --scope deployment --env {env-uuid}
bitbottle variable set    MYWORKSPACE/my-service API_TOKEN secret --scope deployment --env {env-uuid} --secured
bitbottle variable delete MYWORKSPACE/my-service DEPLOY_KEY --scope deployment --env {env-uuid} --confirm
```

All commands support `--json fields` and `--jq expr`. Secured variable values
are never returned by the API once set.

These commands return a typed unsupported-capability error on Bitbucket Server /
Data Center — deployments and environments are a Cloud-only feature.

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

### Deploy keys

```bash
# List deploy keys
bitbottle deploy-key list MYPROJ/my-service

# Add a deploy key
bitbottle deploy-key add MYPROJ/my-service --key "ssh-rsa AAAA..." --label "CI server"

# Delete a deploy key by ID
bitbottle deploy-key delete MYPROJ/my-service 42
```

All three subcommands work on both Bitbucket Cloud and Server/DC. `list` and
`add` support `--json` / `--jq` / `--yaml` for structured output.

### Branch restriction rules (Cloud only)

```bash
# List branch restriction rules
bitbottle branch-rule list myworkspace/my-service

# Add a rule requiring 2 approvals before merging to main
bitbottle branch-rule add myworkspace/my-service --kind require_approvals_to_merge --pattern main --value 2

# Add a push restriction (no direct pushes to main)
bitbottle branch-rule add myworkspace/my-service --kind push --pattern main

# Delete a rule by ID
bitbottle branch-rule delete myworkspace/my-service 7
```

Cloud only. `list` and `add` support `--json` / `--jq` / `--yaml` for structured output.

### SSH keys (Cloud only)

```bash
# List SSH keys for the current user
bitbottle ssh-key list

# Add an SSH key
bitbottle ssh-key add --key "ssh-rsa AAAA..." --label "Laptop"

# Delete an SSH key by ID
bitbottle ssh-key delete 42
```

`list` and `add` support `--json` / `--jq` / `--yaml` for structured output.

### PR default reviewers

```bash
# List default reviewers
bitbottle pr default-reviewer list MYPROJ/my-service

# Add a default reviewer (Server: user slug; Cloud: account ID or nickname)
bitbottle pr default-reviewer add MYPROJ/my-service jsmith

# Remove a default reviewer
bitbottle pr default-reviewer remove MYPROJ/my-service jsmith
```

Both Cloud and Server/DC are supported. `list` supports `--json` / `--jq` / `--yaml`.

### Repo extras

```bash
# Rename (both backends — slug derives from new name on Cloud).
# --confirm is required on non-TTY because the slug change breaks existing
# clones' origin URL — run `git remote set-url origin ...` after.
bitbottle repo rename MYPROJ/my-service my-service-v2 --confirm

# Fork into another workspace (Bitbucket Cloud only)
bitbottle repo fork myworkspace/my-service --into otherws
bitbottle repo fork myworkspace/my-service --into otherws --name my-fork

# Transfer to another project (Server) or workspace (Cloud)
bitbottle repo transfer MYPROJ/my-service --to NEWPROJ
bitbottle repo transfer myworkspace/my-service --to otherws
```

`repo fork` returns a typed unsupported-capability error on Bitbucket Server /
Data Center, which has no fork primitive in its REST API. Both `rename` and
`fork` accept `--json fields` and `--jq expr` for structured output.
`repo transfer` works on both backends and also accepts `--json`/`--jq`.

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

### Permissions _(Bitbucket Server / DC only)_

Manage user and group permissions for projects and repositories.
Requires `PROJECT_ADMIN` on the target project.

```bash
# List all grants for a project (users + groups, sorted ADMIN → WRITE → READ)
bitbottle --hostname git.example.com perms project list MYPROJ

# Grant a user PROJECT_WRITE on a project
bitbottle --hostname git.example.com perms project grant MYPROJ PROJECT_WRITE --user alice

# Grant a group PROJECT_READ on a project
bitbottle --hostname git.example.com perms project grant MYPROJ PROJECT_READ --group "qa team"

# Revoke a user's permission on a project
bitbottle --hostname git.example.com perms project revoke MYPROJ --user bob

# List all grants for a repository
bitbottle --hostname git.example.com perms repo list MYPROJ/my-service

# Grant a user REPO_WRITE on a repository
bitbottle --hostname git.example.com perms repo grant MYPROJ/my-service REPO_WRITE --user carol

# Revoke a group's permission on a repository
bitbottle --hostname git.example.com perms repo revoke MYPROJ/my-service --group "qa team"

# Structured output
bitbottle --hostname git.example.com perms project list MYPROJ --json permission,subject
bitbottle --hostname git.example.com perms project list MYPROJ --jq '.[] | select(.permission == "PROJECT_ADMIN")'
```

If a grant call would **downgrade** an existing permission (e.g. ADMIN → READ),
`grant` prompts for confirmation on a TTY. Pass `--force` to skip the prompt.

### Commit Comments

Add and manage review-style comments on individual commits. Both Bitbucket
Cloud and Server / Data Center are supported.

```bash
# List all comments on a commit
bitbottle commit comment list MYPROJ/my-service abc1234

# Add a comment
bitbottle commit comment add MYPROJ/my-service abc1234 --body "LGTM"

# Edit an existing comment (COMMENT_ID from list --json id)
bitbottle commit comment edit MYPROJ/my-service abc1234 1234 --body "Updated text"

# Delete a comment
bitbottle commit comment delete MYPROJ/my-service abc1234 1234

# Structured output
bitbottle commit comment list MYPROJ/my-service abc1234 --json id,author,body
bitbottle commit comment list MYPROJ/my-service abc1234 --jq '.[].body'

# Emoji reactions (Bitbucket Server / DC only)
bitbottle commit comment list MYPROJ/my-service abc1234 --reactions   # adds REACTIONS column
bitbottle commit comment react   MYPROJ/my-service abc1234 1234 --emoji thumbs_up
bitbottle commit comment unreact MYPROJ/my-service abc1234 1234 --emoji thumbs_up
```

On Bitbucket Server / Data Center, `edit` and `delete` use optimistic
concurrency (fetches the current `version` before writing). A 409 Conflict
means the comment changed between calls — retry.

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

## Extensions

bitbottle supports third-party extensions. Extension repositories must be named `bitbottle-<name>`.

```bash
bitbottle extension install USER/bitbottle-<name>  # install from GitHub
bitbottle extension install --local PATH           # install from local dir
bitbottle extension list                           # list installed
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Run `go test ./...` before sending a PR.

## License

[MIT](LICENSE)
