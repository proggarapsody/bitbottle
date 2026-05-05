---
name: bitbottle
description: >
  Reference for the bitbottle CLI — a gh-style tool for Bitbucket Server/DC
  and Cloud. Load when the user asks about bitbottle commands, auth setup, PRs,
  repos, branches, tags, commits, pipelines, or why a command failed. Load even
  if the user just says "bitbottle", mentions "Bitbucket", or pastes a bitbottle
  error message. Verified against bitbottle 1.13.0.
---

# bitbottle CLI (1.13.0)

Install: `npm install -g @proggarapsody/bitbottle`

> When in doubt, run `bitbottle <command> --help`. Flag names below were
> verified against 1.13.0; older releases may differ.

## Agent safety rules

- **Never echo tokens to logs.** Use `--with-token` (reads stdin) or
  the `BB_TOKEN` env var; do not put a PAT/App Password on the command
  line where it lands in shell history.
- **Confirm before destructive ops.** `repo delete`, `branch delete`,
  `tag delete`, `pr decline`, `pr merge` are not undoable. Show the
  command and ask before running.
- **Don't fabricate flags.** If the user asks for behavior not in
  this file, run `--help` first; do not invent flags like `--author`,
  `--state all`, `--all`, `--mine` — they don't exist as of 1.13.0.

## Auth

Cloud (Bitbucket.org) needs an **email**; Server/DC needs a **username**.
The token comes from stdin via `--with-token`.

```bash
# Bitbucket Cloud (App Password or API token)
echo "$APP_PASSWORD" | bitbottle auth login \
  --hostname bitbucket.org \
  --email you@example.com \
  --with-token

# Bitbucket Server / Data Center (PAT, often "BBDC-…")
echo "$BBDC_PAT" | bitbottle auth login \
  --hostname git.example.com \
  --username your.user \
  --with-token \
  --git-protocol https \
  --skip-tls-verify             # only for self-signed certs

# Lifecycle
bitbottle auth status
bitbottle auth token   [--hostname HOST]
bitbottle auth refresh [--hostname HOST]
bitbottle auth logout  --hostname HOST
```

Credentials are stored in `~/.config/bitbottle/hosts.yml`.
`--skip-tls-verify` is set once at login and remembered per host.

## Repo targeting

Inside a git repo with a Bitbucket origin, host/project/repo is
auto-detected. Outside one, pass `-R` (or set `BB_REPO`):

```bash
bitbottle pr list      -R git.example.com/PROJ/repo
bitbottle pr approve 42 -R git.example.com/PROJ/repo
```

Pin a default for the current checkout (writes to `.git/config`):

```bash
bitbottle repo set-default git.example.com/PROJ/repo
```

## Common flags

Most list/view commands accept these — but verify with `--help`,
since not every command implements every one:

`--json field1,field2` `--jq 'expr'` `--limit N` `--hostname HOST`

## pr

```bash
bitbottle pr list   [PROJ/repo] [--state open|closed|merged]   # default: open
bitbottle pr view   42 [--web]
bitbottle pr create --title "x" --base main [--body "x"] [--draft] [--head BRANCH]
bitbottle pr merge  42 [--merge|--squash] [--delete-branch]
bitbottle pr approve   42
bitbottle pr unapprove 42
bitbottle pr diff      42                       # unified diff; pipes to pager on TTY
bitbottle pr checkout  42
bitbottle pr edit      42 [--title "x"] [--body "x"]
bitbottle pr decline   42
bitbottle pr ready     42                       # draft → ready
bitbottle pr request-review  42 --reviewer alice [--reviewer bob]
bitbottle pr request-changes 42                 # Cloud only
bitbottle pr comment list 42
bitbottle pr comment add  42 --body "x"
```

`--state` accepts only `open|closed|merged`. There is no `all` and
no `--author` flag in 1.13.0.

## repo / branch / tag

```bash
bitbottle repo list  [PROJ]
bitbottle repo view  PROJ/repo [--web]
bitbottle repo create NAME --project PROJ [--description "x"] [--private=false]
bitbottle repo delete PROJ/repo [--confirm]
bitbottle repo clone  PROJ/repo [PATH]
bitbottle repo set-default HOST/PROJ/repo

bitbottle branch list   PROJ/repo [--limit N]
bitbottle branch create PROJ/repo BRANCH --start-at main|HASH
bitbottle branch checkout BRANCH                # fetches origin then checks out
bitbottle branch delete PROJ/repo BRANCH

bitbottle tag list   PROJ/repo [--limit N]
bitbottle tag create PROJ/repo TAG --start-at main|HASH [--message "x"]
bitbottle tag delete PROJ/repo TAG
```

## commit

```bash
# Branch resolution: --branch flag → current local branch → "main"
bitbottle commit log    PROJ/repo [--branch main] [--limit 5]
bitbottle commit view   PROJ/repo HASH [--web]
bitbottle commit status PROJ/repo HASH          # build/CI statuses for the commit
```

## pipeline (Cloud only)

```bash
bitbottle pipeline list WORKSPACE/repo [--limit N]
bitbottle pipeline view WORKSPACE/repo UUID [--web]
bitbottle pipeline run  WORKSPACE/repo --branch BRANCH
```

## api (raw REST)

```bash
bitbottle api 'PATH'
bitbottle api -X POST -F key=val 'PATH'
bitbottle api --paginate --jq '.[].name' 'PATH'
cat f.json | bitbottle api -X PUT --input - 'PATH'
```

`--paginate` follows Cloud `next` and Server `nextPageStart`, merging
the `values` arrays into a single stream.

## Other commands

```bash
bitbottle alias set NAME 'COMMAND'              # also `alias list`, `alias delete`
bitbottle config set git_protocol https --host HOST
bitbottle completion --shell bash|zsh|fish|powershell
```

## MCP server

bitbottle exposes its full surface as an MCP server for AI agents:

```bash
bitbottle mcp serve                             # starts MCP server on stdio
```

Example Claude Code / Cursor MCP config:

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

## Environment variables

| Var | Effect |
|---|---|
| `BB_TOKEN` | Override the stored token for API calls (useful in CI) |
| `BB_HOST` | Default hostname; equivalent to `--hostname` on every command |
| `BB_REPO` | `[HOST/]PROJECT/REPO` override; equivalent to `-R` |
| `BB_EDITOR` / `BB_PAGER` / `BB_BROWSER` | Per-tool overrides; precede `$EDITOR` / `$PAGER` / config |
| `BB_FORCE_TTY` | Force aligned/colored output even when piped (mirrors `GH_FORCE_TTY`) |
| `BB_PROMPT_DISABLED` | Fail rather than prompt; required for non-interactive scripts |
| `BB_CONFIG_DIR` | Override config dir (default `$XDG_CONFIG_HOME/bitbottle`) |
| `NO_COLOR` | Standard convention — disables colored output |

## hosts.yml reference

Stored at `~/.config/bitbottle/hosts.yml`. The `backend_type` key
forces routing when the hostname is ambiguous (e.g. a self-hosted
Bitbucket Cloud-DC behind `git.internal`):

```yaml
git.internal.example.com:
  oauth_token: BBDC-…
  user: alice
  git_protocol: https
  skip_tls_verify: true
  backend_type: server          # or "cloud" — overrides hostname-based dispatch
```

## Cloud vs Server/DC quick reference

- **Auth tokens**: Cloud uses App Password / API token; Server/DC uses
  PAT (`BBDC-…`).
- **Cloud-only commands**: `pipeline *`, `pr request-changes`.
- **`--skip-tls-verify`**: only meaningful for Server/DC with
  self-signed certs.
- **XSRF**: Server/DC rejects writes without a Content-Type header;
  bitbottle handles this internally — no user action needed.

## Failure-mode hints (for agents)

- *"not authenticated; run `bitbottle auth login` first"* — no host in
  config. Run auth login.
- *"multiple hosts configured; specify hostname"* — pass
  `--hostname HOST` or use `-R HOST/PROJ/repo`.
- *"no git remotes found; pass [HOST/]PROJECT/REPO …"* — running
  outside a Bitbucket checkout. Pass `-R` or `cd` into a repo.
- *Auth errors on a Cloud host*: most often a missing or wrong
  `--email`. App passwords need the Atlassian email, not the username.
- *Auth errors on Server/DC*: missing `--username`, or `--git-protocol
  ssh` was used with an HTTPS-only PAT.
