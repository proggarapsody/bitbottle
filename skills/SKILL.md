---
name: bitbottle
description: >
  Reference for the bitbottle CLI — a gh-style tool for Bitbucket Server/DC
  and Cloud. Load when the user asks about bitbottle commands, auth setup, PRs,
  repos, branches, tags, commits, pipelines, or why a command failed. Load even
  if the user just says "bitbottle" or pastes a bitbottle error.
---

# bitbottle CLI

Install: `npm install -g @proggarapsody/bitbottle`

## Auth

```bash
echo "TOKEN" | bitbottle auth login --hostname git.example.com --with-token \
  --git-protocol https --skip-tls-verify   # --skip-tls-verify for self-signed certs

bitbottle auth status
bitbottle auth token [--hostname HOST]
bitbottle auth logout --hostname HOST
bitbottle auth refresh [--hostname HOST]
```
Credentials stored in `~/.config/bitbottle/hosts.yml`.
`--skip-tls-verify` is set once at login; remembered per host automatically.

## Repo targeting

Inside a git repo with a Bitbucket origin → host/project/repo auto-detected.
Outside a git repo, always use `-R`:
```bash
bitbottle pr list -R git.example.com/PROJ/repo
bitbottle pr approve 42 -R git.example.com/PROJ/repo
```
Pin a default: `bitbottle repo set-default git.example.com/PROJ/repo`

## Common flags (all listing commands)

`--json field1,field2` `--jq 'expr'` `--limit N` `--hostname HOST`

## pr

```bash
bitbottle pr list [PROJ/repo] [--state open|merged|closed|all] [--author @me]
bitbottle pr view 42 [--web]
bitbottle pr create --title "x" --base main [--body "x"] [--draft] [--head BRANCH]
bitbottle pr merge 42 [--merge|--squash] [--delete-branch]
bitbottle pr approve 42
bitbottle pr unapprove 42
bitbottle pr diff 42            # unified diff, pipe to delta/less
bitbottle pr checkout 42
bitbottle pr edit 42 [--title "x"] [--body "x"]
bitbottle pr decline 42
bitbottle pr ready 42           # draft → ready
bitbottle pr request-review 42 --reviewer alice [--reviewer bob]
bitbottle pr request-changes 42 # Cloud only
bitbottle pr comment list 42
bitbottle pr comment add 42 --body "x"
```

## repo / branch / tag

```bash
bitbottle repo list|view|create|delete|clone PROJ/repo
bitbottle repo create NAME --project PROJ [--description "x"] [--private=false]
bitbottle repo delete PROJ/repo [--confirm]
bitbottle repo clone PROJ/repo [PATH]
bitbottle repo set-default HOST/PROJ/repo

bitbottle branch list|create|delete|checkout PROJ/repo [BRANCH]
bitbottle branch create PROJ/repo BRANCH --start-at main|HASH

bitbottle tag list|create|delete PROJ/repo [TAG]
bitbottle tag create PROJ/repo TAG --start-at main|HASH [--message "x"]
```

## commit

```bash
# branch resolution: --branch flag → current local branch → main
bitbottle commit log PROJ/repo [--branch main] [--limit 5]
bitbottle commit view PROJ/repo HASH [--web]
bitbottle commit status PROJ/repo HASH   # build/CI status
```

## pipeline (Cloud only)

```bash
bitbottle pipeline list WORKSPACE/repo [--limit N]
bitbottle pipeline view WORKSPACE/repo UUID [--web]
bitbottle pipeline run WORKSPACE/repo --branch BRANCH
```

## api (raw REST)

```bash
bitbottle api 'PATH'
bitbottle api -X POST -F key=val 'PATH'
bitbottle api --paginate --jq '.[].name' 'PATH'
cat f.json | bitbottle api -X PUT --input - 'PATH'
```
`--paginate` follows Cloud `next` / Server `nextPageStart` and merges `values`.

## Other

```bash
bitbottle alias set NAME 'COMMAND'
bitbottle config set git_protocol https --host HOST
bitbottle completion --shell bash|zsh|fish|powershell
bitbottle --mcp   # run as MCP server
```

## Cloud vs Server/DC

- Auth: Cloud uses App Password/OAuth; Server uses PAT (`BBDC-…`)
- `pr request-changes`, `pipeline *`: Cloud only
- XSRF: Server rejects write requests without Content-Type (handled internally)
- `--skip-tls-verify`: Server only (self-signed certs)
