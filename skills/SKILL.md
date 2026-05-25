---
name: bitbottle
description: >
  Reference for the bitbottle CLI — a gh-style tool for Bitbucket Server/DC
  and Cloud. Load when the user asks about bitbottle commands, auth setup, PRs,
  repos, branches, tags, commits, pipelines, or why a command failed. Load even
  if the user just says "bitbottle", mentions "Bitbucket", or pastes a bitbottle
  error message.
---

# bitbottle CLI

A gh-style CLI for Bitbucket Server/DC and Cloud. This file is a router
plus invariant safety rules — load the matching reference for command
detail. **Don't memorize this file; run `bitbottle <cmd> -h` for the
authoritative shape and flag list.** Run `bitbottle --version` if behavior
disagrees with this doc; the binary wins.

## When to load which reference

| Task | Open |
|---|---|
| Auth, hosts.yml, env vars, multi-host, `auth migrate`, `auth doctor` | `references/auth.md` |
| PR lifecycle (list/view/create/merge/approve/comment/activity/review/commits/files/participants/ready/unready/task/suggestion/comment-react/default-reviewer/reviewer-group/…) | `references/pr.md` |
| Repos, branches, tags, file/tree, visibility, edit, transfer, watcher, repo label list/create/update/delete, repo pr-settings get, repo pr-settings set, repo download list/upload/get/delete, repo clone (MCP: `clone_repo`) | `references/repos.md` |
| Commits (view/files/status/comment/comment-react) | `references/commit.md` |
| Pipelines, pipeline config get/enable/disable, schedules, caches, watch, trigger, test-report view, test-case list, SSH key pair, known hosts (Cloud only) | `references/pipeline.md` |
| Pipeline runners — list / create / delete (Cloud only) | `references/runner.md` |
| Code Insights reports/annotations/merge-check (Server/DC only) | `references/code-insights.md` |
| Issues, comments, attachments (list/delete), vote/unvote, watch/unwatch, activity/change log (Cloud only) | `references/issues.md` |
| Issue milestones list/view (Cloud only) | `references/milestone.md` |
| Issue versions list/view/create/delete (Cloud only) | `references/version.md` |
| snippet list/view/create/delete (snippet snippet group, Cloud only) | `references/snippet.md` |
| Deployments + environments + variables (Cloud only) | `references/deployment.md`, `references/variable.md` |
| Deploy keys (both), branch-rules (Cloud only), ssh-keys (Cloud + Server/DC) | `references/deploy-key.md`, `references/branch-rule.md`, `references/ssh-key.md` |
| Diff between refs, `diff REF1..REF2` | `references/diff.md` |
| Branch compare (`branch compare BASE..HEAD`) — ahead/behind counts + commit lists (Cloud + Server/DC) | `references/branch.md` |
| Workspaces + webhooks + workspace project CRUD + workspace project perms + workspace perms + workspace pipeline variables (Cloud only), `user view` | `references/workspace.md`, `references/user.md` |
| Mirror servers (list/view/repo list) (Server/DC only) | `references/mirror.md` |
| Named credential profiles (`profile create/use/list/delete`) | `references/profile.md` |
| Raw REST passthrough, pagination, MCP server config | `references/api.md` |
| Extensions (`extension install/list/upgrade/remove/exec`, extension scaffold) | `references/extension.md` |
| Group management — group list, group create, group delete, group member list, group member add, group member remove (Server/DC only) | `references/groups.md` |
| Project CRUD — project server-list, project create, project view, project edit, project delete (Server/DC only); project list WORKSPACE (Cloud only) | `references/project.md` |
| auth pat list/create/revoke (Server/DC only) | `references/auth.md` |

Load both when a task spans areas. Don't load speculatively.

## Safety rules (always apply)

1. **Never echo tokens.** Pass tokens via stdin (`--with-token`) or the
   `BB_TOKEN` env var. Never put a PAT/App Password on the command line —
   it lands in shell history.
2. **Confirm before destructive ops.** `repo delete`, `repo rename`,
   `branch delete`, `tag delete`, `webhook delete`, `pr decline`,
   `pr merge` are not undoable (`repo rename` breaks existing clones'
   `origin` URL on Cloud). Show the exact command + resolved
   host/PROJECT/REPO + PR/branch/tag name, then wait for explicit confirm.
3. **Don't fabricate flags.** bitbottle has gh-like *shape* but not
   gh-compatible *flags*. If a flag isn't in `bitbottle <cmd> -h`,
   don't pass it. Common phantoms: `--author`, `--mine`, `--reviewer @me`
   (no "self" sentinel — pass the user slug).
4. **`--json` is a BOOLEAN flag, not a field selector.** `--json` emits
   the full object as JSON; filter with `--jq EXPR`. Passing
   `--json title,body` is parsed as `--json` plus a positional argument
   (`title,body`) and fails with `accepts N arg(s), received N+1`. To
   discover field names: run with `--json` once, inspect the object.

## Argument shape (high-frequency footgun)

bitbottle is inconsistent about repo targeting — match what `-h` says:

| Command group | Repo arg | Notes |
|---|---|---|
| `pr *` (view/list/create/merge/…) | `-R [HOST/]PROJECT/REPO` flag | `-R` is the only path; no positional |
| `repo *`, `branch *`, `tag *`, `webhook *` | positional `PROJECT/REPO` | `-R` is in the inherited flag list but **silently ignored** — pass `PROJECT/REPO` as a positional and `--hostname HOST` separately |
| `commit view PROJECT/REPO HASH` | positional, **PROJ/REPO first** | |
| `commit files HASH [PROJECT/REPO]` | positional, **HASH first** | inverse of `commit view` |
| `commit cherry-pick HASH BRANCH [PROJECT/REPO]` | positional, **HASH then BRANCH** | Server/DC only (`branch-utils` plugin); `--message/-m` overrides commit msg |
| `issue *`, `code-insights *` | `[PROJECT/REPO]` positional or `-R` | optional; defaults to checkout |

Outside any Bitbucket checkout, the relevant positional/`-R` is mandatory.

## Repo targeting & TLS

```bash
# Inside a Bitbucket checkout: host/project/repo auto-detected.
# Outside one (or to override):
bitbottle pr list      -R git.example.com/PROJ/repo
bitbottle repo view    PROJ/repo --hostname git.example.com
bitbottle repo edit    PROJ/repo --description "new description"

# Pin a default for the current checkout (writes .git/config):
bitbottle repo set-default HOST/PROJ/repo
```

Self-signed CA on Server/DC? `-k` / `--skip-tls-verify` is an inherited
flag on **every** command. For permanence, set `skip_tls_verify: true`
on the host entry in `hosts.yml` (see `references/auth.md`).

## Cloud vs Server/DC

| | Cloud (`bitbucket.org`) | Server/DC (self-hosted) |
|---|---|---|
| Auth context flag | `--email you@…` | `--username your.user` |
| Token type | App Password / API token | PAT (`BBDC-…`) |
| API base path | `2.0/…` | `rest/api/1.0/…` |
| Cloud-only | `pipeline *` (list/view/run/stop/trigger/watch/logs/steps/schedule/cache/variable/config/test-report/test-case), `runner list`, `runner create`, `runner delete`, `issue *` (including `issue attachment list/delete`, `issue vote/unvote`, `issue watch/unwatch`, `issue activity`), `snippet list [--workspace W]`, `snippet view`, `snippet create`, `snippet delete`, `pr request-changes`, `pr comment resolve`, `branch-rule *`, `branch-model *`, `workspace *` (including `workspace search`, `workspace perms *`, `workspace project perms *`), `search`, `project list WORKSPACE` | — |
| Server-only | — | `admin user list/activate/deactivate/rename`, `admin license`, `admin cluster`, `admin secrets rotate`, `admin logging get/set`, `code-insights *`, `pr task *`, `pr suggestion apply`, `pr/commit comment react/unreact`, `pr reviewer-group *`, `repo pr-settings get`, `repo pr-settings set`, `group list`, `group create`, `group delete`, `group member list`, `group member add`, `group member remove`, `mirror list`, `mirror view`, `mirror repo list`, `project server-list`, `project create`, `project view`, `project edit`, `project delete`, `auth pat list`, `auth pat create`, `auth pat revoke`, `commit cherry-pick HASH BRANCH` |

For custom-hostname Bitbucket Data Center, force routing with
`backend_type: cloud|server` in `hosts.yml`.

## Hot-path env vars

| Var | Effect |
|---|---|
| `BB_TOKEN` | Token override (CI use) |
| `BB_HOST` | Default `--hostname` |
| `BB_REPO` | Default `-R [HOST/]PROJECT/REPO` |
| `BB_PROMPT_DISABLED` | Fail rather than prompt (non-interactive) |
| `BB_CONFIG_DIR` | Override config dir (default `$XDG_CONFIG_HOME/bitbottle`) |

## Failure-mode hints

| Message | Fix |
|---|---|
| `not authenticated; run bitbottle auth login` | No host configured — run `auth login`. |
| `multiple hosts configured; specify hostname` | Pass `--hostname HOST` or `-R HOST/PROJ/repo`. |
| `no git remotes found; pass [HOST/]PROJECT/REPO` | Outside a checkout — pass positional or `-R`. |
| `accepts N arg(s), received N+1` after `--json …` | You wrote `--json field1,field2`; bitbottle's `--json` is boolean. Drop the field list and use `--jq`. |
| `warning: token found in hosts.yml — run bitbottle auth migrate` | Printed on every command until tokens are migrated to OS keyring. Run `bitbottle auth migrate` once. |
| `code-insights … unsupported on host` | Server/DC-only feature. Confirm host with `bitbottle context`. |
| Cloud auth fails | Usually a missing/wrong `--email` — App Passwords need the **Atlassian email**, not the username. |
| Server/DC auth fails | Missing `--username`, or `--git-protocol ssh` was used with an HTTPS-only PAT. |
| Cred/keychain/TLS/proxy weirdness | `bitbottle auth doctor [--hostname HOST]` reports keyring backend, token presence/format, base-URL reachability, auth success. Never echoes the token. |

## Pipeline lifecycle (Cloud only)

```bash
# Trigger, watch, stop, view
bitbottle pipeline trigger WORKSPACE/REPO --branch main
bitbottle pipeline watch WORKSPACE/REPO UUID
bitbottle pipeline stop UUID WORKSPACE/REPO --confirm
bitbottle pipeline rerun UUID WORKSPACE/REPO
bitbottle pipeline view UUID WORKSPACE/REPO

# Pipeline config (Cloud only)
bitbottle pipeline config get WORKSPACE/REPO
bitbottle pipeline config enable WORKSPACE/REPO
bitbottle pipeline config disable WORKSPACE/REPO

# List and download step artifacts
bitbottle pipeline artifact list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID
bitbottle pipeline artifact download PIPELINE_UUID WORKSPACE/REPO \
  --step STEP_UUID --name build.tar.gz [--out PATH|-] [--clobber]

# Test report and test cases (Cloud only)
bitbottle pipeline test-report view PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID
bitbottle pipeline test-case list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID [--status FAILED|PASSED|SKIPPED] [--limit N]

# Pipeline SSH configuration (Cloud only)
bitbottle pipeline ssh key-pair view [WORKSPACE/REPO]
bitbottle pipeline ssh key-pair regenerate [WORKSPACE/REPO] [--bits 2048|4096] [--confirm]
bitbottle pipeline ssh known-hosts list [WORKSPACE/REPO]
bitbottle pipeline ssh known-hosts view UUID [WORKSPACE/REPO]
bitbottle pipeline ssh known-hosts add HOSTNAME [WORKSPACE/REPO] [--key KEY] [--key-type rsa|ecdsa|ed25519]
bitbottle pipeline ssh known-hosts delete UUID [WORKSPACE/REPO] [--confirm]
```

```bash
# Branch compare (Cloud + Server/DC)
bitbottle branch compare main..feature WORKSPACE/REPO
bitbottle branch compare main..feature WORKSPACE/REPO --json
```

## Dashboard

```bash
# Open PRs (author + reviewer) across a repo, or globally with --hostname
bitbottle status [PROJECT/REPO]
bitbottle pr status [PROJECT/REPO]

# Open a Bitbucket page in the browser
bitbottle browse [PROJECT/REPO]              # repo home
bitbottle browse [PROJECT/REPO] 42           # PR #42
bitbottle browse [PROJECT/REPO] abc1234      # commit
bitbottle browse [PROJECT/REPO] src/main.go  # file
```

## Install / version

`npm install -g @proggarapsody/bitbottle` installs the CLI and
auto-registers this skill. For non-npm installs (Homebrew, Go, bare
binary): `bitbottle skill install` (refresh), `skill path` (locate),
`skill remove` (uninstall).
