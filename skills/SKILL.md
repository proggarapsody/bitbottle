---
name: bitbottle
description: >
  Reference for the bitbottle CLI — a gh-style tool for Bitbucket Server/DC
  and Cloud. Load when the user asks about bitbottle commands, auth setup, PRs,
  repos, branches, tags, commits, pipelines, or why a command failed. Load even
  if the user just says "bitbottle", mentions "Bitbucket", or pastes a bitbottle
  error message. Verified against bitbottle 1.72.0. <!-- x-release-please-version -->
---

# bitbottle CLI

A gh-style CLI for Bitbucket Server/DC and Cloud. This file is a router
plus invariant safety rules — load the matching reference for command
detail. **Don't memorize this file; consult `-h` for any flag you're
not sure about.**

## When to load which reference

| Task | Open |
|---|---|
| Auth, hosts.yml, env vars, multi-host setup, auth migrate | `references/auth.md` |
| PR lifecycle (list/view/create/merge/approve/comment/pr activity/review/commits/files/participants/…) | `references/pr.md` |
| Repos, branches, tags, commits, pipelines, webhooks, repo watcher list | `references/repos.md` |
| repo visibility | Get or set repository visibility | `references/repos.md` |
| commit comment list/add/edit/delete/react/unreact | list, add, edit, delete commit comments; react/unreact (Server/DC only) | `references/commit.md` |
| commit files HASH [PROJECT/REPO] | list files changed in a commit — both Cloud and Server/DC | `references/commit.md` |
| commit status PROJECT/REPO HASH | list build/CI statuses for a commit — both Cloud and Server/DC | `references/commit.md` |
| commit status report PROJECT/REPO HASH --key KEY --state STATE | post a build status against a commit hash — both Cloud and Server/DC | `references/commit.md` |
| Raw REST passthrough, pagination, MCP server config | `references/api.md` |
| Issues (list/view/create/close/edit/reopen/assign/comment) — Cloud only | see inline below |
| Deployments (list/view) — Cloud only | `references/deployment.md` |
| Environments (list/create/delete) + variables (list/set/delete) — Cloud only | `references/deployment.md` |
| Top-level variable command (repository/workspace/deployment scope) — Cloud only | `references/variable.md` |
| Named credential profiles (create/use/list/delete) | `references/profile.md` |
| Deploy keys (list/add/delete) — Cloud and Server/DC | `references/deploy-key.md` |
| Branch restriction rules (list/add/delete) — Cloud only | `references/branch-rule.md` |
| SSH keys for current user (list/add/delete) — Cloud only | `references/ssh-key.md` |
| PR default reviewers (list/add/remove) — Cloud and Server/DC | `references/pr.md` |
| PR reviewer groups (list/add/remove) — Server/DC only | `references/pr.md` |
| Diff between refs (`diff REF1..REF2 [--stat]`) — Cloud and Server/DC | `references/diff.md` |
| Workspaces (list) + workspace member list — Cloud only | `references/workspace.md` |
| Workspace webhooks (list/create/delete) — Cloud only | `references/workspace.md` |
| User profile (`user view`) — Cloud and Server/DC | `references/user.md` |

When the user's task spans two areas, load both. Don't load all of
them speculatively.

## Safety rules (always apply)

1. **Never echo tokens.** Pass tokens via stdin (`--with-token`) or
   the `BB_TOKEN` env var. Never put a PAT/App Password on the
   command line — it lands in shell history.
2. **Confirm before destructive ops.** `repo delete`, `repo rename`,
   `branch delete`, `tag delete`, `webhook delete`, `pr decline`,
   `pr merge` are not undoable. (`repo rename` is included because the
   slug change breaks existing clones' `origin` URL on Cloud.) Before
   running, show the user the exact command, the resolved host and
   `PROJECT/REPO`, the PR ID / branch / tag name, then wait for
   explicit confirmation. Reference files reuse this rule — don't
   restate it there.
3. **Don't fabricate flags.** bitbottle has gh-like *shape* but not
   gh-compatible *flags*. If a flag isn't in the reference and the
   user asks for behavior you can't find, run `bitbottle <command> -h`
   first. Phantom flags commonly assumed but **absent**: `--author`,
   `--mine`, `--all`. Phantom **values**: `--state all` (only
   `open`/`closed`/`merged` are accepted); `--reviewer @me` (no
   "self" sentinel — pass the user slug).
4. **Prefer JSON for automation.** Parsing TTY tables is brittle. Every
   `list`/`view` command supports `--json fields --jq 'expr'` plus
   `--limit N`. To discover supported fields for any command, pass a
   bogus value: `bitbottle <cmd> --json X` — the error lists them.
5. **Check the version on behavior mismatches.** If a command behaves
   differently from this file, run `bitbottle --version`. This skill
   was last verified against **1.72.0**. <!-- x-release-please-version -->

## Repo targeting (high-frequency)

Inside a Bitbucket checkout: host/project/repo auto-detected.

Outside one (or to override), pass `-R [HOST/]PROJECT/REPO`:

```bash
bitbottle pr list      -R git.example.com/PROJ/repo
bitbottle pr approve 42 -R git.example.com/PROJ/repo
```

Pin a default for the current checkout (writes `.git/config`):
`bitbottle repo set-default HOST/PROJ/repo`. After this, every
command in that checkout runs without `-R`.

## Cloud vs Server/DC (decision table)

| | Cloud (`bitbucket.org`) | Server/DC (self-hosted) |
|---|---|---|
| Auth context flag | `--email you@…` | `--username your.user` |
| Token type | App Password / API token | PAT (`BBDC-…`) |
| API base path | `2.0/…` | `rest/api/1.0/…` |
| Cloud-only commands | `pipeline *`, `pr request-changes`, `pr comment resolve` | — |
| Server-only commands | — | `branch protect *`, `code-insights *`, `perms *`, `pr task resolve/reopen`, `pr suggestion apply`, `pr comment react/unreact`, `pr comment list --reactions`, `commit comment react/unreact`, `commit comment list --reactions` |

Custom-hostname Cloud Data Center? Force routing in `hosts.yml`:
`backend_type: cloud` (or `server`). See `references/auth.md`.

## Hot-path env vars

| Var | Effect |
|---|---|
| `BB_TOKEN` | Token override for API calls (CI use) |
| `BB_HOST` | Default `--hostname` |
| `BB_REPO` | Default `-R [HOST/]PROJECT/REPO` |
| `BB_PROMPT_DISABLED` | Fail rather than prompt (non-interactive scripts) |
| `BB_CONFIG_DIR` | Override config dir (default `$XDG_CONFIG_HOME/bitbottle`) |

Editor/pager/browser/force-tty/no-color overrides are in
`references/auth.md`.

## Failure-mode hints

When you see one of these messages, you know the fix:

- *"not authenticated; run `bitbottle auth login` first"* → no host
  configured. Run `auth login`. See `references/auth.md`.
- *"multiple hosts configured; specify hostname"* → pass
  `--hostname HOST` or use `-R HOST/PROJ/repo`.
- *"no git remotes found; pass [HOST/]PROJECT/REPO …"* → outside a
  Bitbucket checkout. Pass `-R` or `cd` into the right repo.
- *Cloud auth fails* → most often a missing or wrong `--email`. App
  passwords need the **Atlassian email**, not the username.
- *Server/DC auth fails* → missing `--username`, or `--git-protocol
  ssh` was used with an HTTPS-only PAT.
- *Credential / keychain / TLS / proxy issues* → run
  `bitbottle auth doctor [--hostname HOST]`. It reports the keyring
  backend, whether a token is stored, token format (BBDC- Server PAT /
  ATATT Cloud OAuth / unknown), API base URL reachability, and whether
  the stored token authenticates successfully. Never echoes the token
  value. Exit 0 if all checks pass, 1 otherwise.
- *`code-insights` returns "unsupported on host"* → Code Insights is a
  Bitbucket Server / Data Center feature only. Cloud hosts always return
  this error. Confirm the host is Server/DC with `bitbottle context`.
- *`merge-check` commands return unexpected errors* → The merge-check API
  is partly undocumented; these commands are experimental. Verify the
  report key matches an existing Code Insights report on the same repo.

## Dashboard and navigation quick-reference

```bash
# Show your open PRs (author + reviewer) across a repo
bitbottle status [PROJECT/REPO]
# same view scoped to the pr subcommand
bitbottle pr status [PROJECT/REPO]

# Open a Bitbucket page in the browser
bitbottle browse [PROJECT/REPO]                  # repo home
bitbottle browse [PROJECT/REPO] 42               # pull request #42
bitbottle browse [PROJECT/REPO] abc1234def       # commit (7-40 hex chars)
bitbottle browse [PROJECT/REPO] src/main.go      # file in current branch

# Watch a Cloud pipeline until it completes (exits 0) or fails (exits 1)
bitbottle pipeline watch PROJECT/REPO UUID [--interval 5]

# Trigger a Bitbucket Cloud pipeline on a branch (Cloud only)
bitbottle pipeline trigger [PROJECT/REPO] --branch BRANCH [--variable KEY=VALUE ...]
# --variable is repeatable; omit --branch to use the current git branch

# Pipeline schedules (Cloud only)
bitbottle pipeline schedule list [PROJECT/REPO]
bitbottle pipeline schedule create [PROJECT/REPO] --cron "0 0 * * *" --branch BRANCH [--enabled=false]
bitbottle pipeline schedule delete [PROJECT/REPO] UUID

# Pipeline caches (Cloud only)
bitbottle pipeline cache list [PROJECT/REPO]
bitbottle pipeline cache delete [PROJECT/REPO] UUID
```

MCP tools: `status` (top-level dashboard), `pr_status`, `pr_checks`, `pr_update_branch`, `pipeline_watch`, `trigger_pipeline`, `list_pipeline_schedules`, `create_pipeline_schedule`, `delete_pipeline_schedule`, `list_pipeline_caches`, `delete_pipeline_cache`.

Deployment/environment MCP tools (Cloud only): `list_deployments`, `get_deployment`, `list_environments`, `create_environment`, `delete_environment`. See `references/deployment.md`.

Variable MCP tools (Cloud only, all scopes): `variable_list`, `variable_set`, `variable_delete` — each accepts `scope` (repository/workspace/deployment) and `env_uuid` (required for deployment scope). See `references/variable.md`.

Deploy key MCP tools (both backends): `list_deploy_keys`, `add_deploy_key`, `delete_deploy_key`. See `references/deploy-key.md`.

Branch rule MCP tools (Cloud only): `list_branch_rules`, `add_branch_rule`, `delete_branch_rule`. See `references/branch-rule.md`.

Repo transfer MCP tool (both backends): `transfer_repo` — accepts `repo` (PROJECT/REPO or WORKSPACE/REPO) and `target` (project key or workspace slug). See `references/repo.md`.

SSH key MCP tools (Cloud only): `list_ssh_keys`, `add_ssh_key`, `delete_ssh_key`. See `references/ssh-key.md`.

PR default reviewer MCP tools (both backends): `list_default_reviewers`, `add_default_reviewer`, `remove_default_reviewer`. See `references/pr.md`.

PR reviewer group MCP tools (Server/DC only): `pr_reviewer_group_list`, `pr_reviewer_group_add`, `pr_reviewer_group_remove`. See `references/pr.md`.

Pipeline trigger MCP tool (Cloud only): `trigger_pipeline` — accepts `repo` (WORKSPACE/REPO), `branch`, and optional `variables` (comma-separated `key=value` pairs).

## PR review quick-reference

`pr review` bundles a body + inline comments + an action into one call:

```bash
# Approve with body
bitbottle pr review 42 --approve --body "lgtm"

# Comment-only review with inline comments (PATH:LINE:BODY, repeatable)
bitbottle pr review 42 --comment \
  --inline pkg/foo.go:88:please rename \
  --inline pkg/bar.go:10-15:extract helper          # ranges Cloud only

# Request changes (Cloud only — Server returns host.unsupported)
bitbottle pr review 42 --request-changes --body "see comments"
```

If `--body` or `--inline` is given without an explicit action flag the
review defaults to `--comment`. MCP tool: `submit_pr_review`
(`{action, body, inline_comments[]}`).

## PR task quick-reference _(Server / DC only)_

Tasks are BLOCKER-severity comments with an OPEN/RESOLVED state. Cloud has no
equivalent — resolve/reopen return `host.unsupported` on Cloud.

```bash
# List open tasks on PR 42 (default: --state open)
bitbottle pr task list 42 [--state open|resolved|all]

# Create a task (optionally threaded under an existing comment)
bitbottle pr task create 42 --body "Fix the null check" [--parent COMMENT_ID]

# Mark a task resolved or reopen it
bitbottle pr task resolve 42 TASK_ID
bitbottle pr task reopen  42 TASK_ID
```

MCP tools: `list_pr_tasks` (`{pr_id, state}`), `create_pr_task`
(`{pr_id, body, parent_comment_id}`), `resolve_pr_task` (`{pr_id, task_id}`),
`reopen_pr_task` (`{pr_id, task_id}`).

## PR suggestions _(Server / DC only)_

Apply Bitbucket Server / Data Center inline suggested-change blocks. The server
commits the change directly to the PR source branch — no local file edits needed.

```bash
# Apply a suggestion (commits the change to the PR branch)
bitbottle pr suggestion apply PR_ID COMMENT_ID SUGGESTION_ID

# Preview the suggestion body without applying it
bitbottle pr suggestion apply PR_ID COMMENT_ID SUGGESTION_ID --preview
```

MCP tool: `pr_suggestion_apply` (`{project, slug, pr_id, comment_id, suggestion_id, preview}`).
Cloud returns `host.unsupported`.

## PR comment reactions _(Server / DC only)_

Emoji reactions on individual PR comments. Supported shortcodes: `thumbs_up`,
`thumbs_down`, `heart`, `laugh`, `hooray`, `confused`. Colon-wrapped GitHub
forms (`:thumbsup:`, `:heart:`) are normalised automatically.

```bash
# Add or remove a reaction
bitbottle pr comment react   42 COMMENT_ID --emoji thumbs_up
bitbottle pr comment unreact 42 COMMENT_ID --emoji thumbs_up

# List comments with their reactions (fetched concurrently, up to 4 workers)
bitbottle pr comment list 42 --reactions
```

The `--reactions` flag adds a REACTIONS column rendered as emoji glyphs
(`👍×2 ❤️×1`). With `--json reactions` the field is a typed array of
`{emoji, users[]}`. Cloud returns `host.unsupported` for all three commands.

MCP tools: `list_comment_reactions` (`{project, repo, pr_id, comment_id}`),
`add_comment_reaction` (`{project, repo, pr_id, comment_id, emoji}`),
`remove_comment_reaction` (`{project, repo, pr_id, comment_id, emoji}`).
`list_pr_comments` also accepts `include_reactions: true` to fetch reactions
in a single call.

## Commit comment reactions _(Server / DC only)_

Emoji reactions on individual commit comments.

```bash
# Add or remove a reaction
bitbottle commit comment react   PROJ/REPO HASH COMMENT_ID --emoji thumbs_up
bitbottle commit comment unreact PROJ/REPO HASH COMMENT_ID --emoji thumbs_up

# List comments with their reactions (fetched concurrently, up to 4 workers)
bitbottle commit comment list PROJ/REPO HASH --reactions
```

The `--reactions` flag adds a REACTIONS column rendered as emoji glyphs
(`👍×2 ❤️×1`). With `--json reactions` the field is a typed array of
`{emoji, users:[slug,...]}` objects.

MCP tools: `list_commit_comment_reactions` (`{project, slug, hash, comment_id}`),
`add_commit_comment_reaction` (`{project, slug, hash, comment_id, emoji}`),
`remove_commit_comment_reaction` (`{project, slug, hash, comment_id, emoji}`).
`list_commit_comments` also accepts `include_reactions: true` to fetch reactions
in a single call.

## Code Insights quick-reference _(Server / DC only)_

```bash
# Upsert a report on a commit
bitbottle code-insights report set PROJ/REPO HASH KEY \
  --title "Tool" --result PASS --report-type SECURITY

# Bulk-add annotations (JSON array with path/line/severity/type/message)
bitbottle code-insights annotation add PROJ/REPO HASH KEY \
  --from-json @findings.json

# Single annotation
bitbottle code-insights annotation add PROJ/REPO HASH KEY \
  --path src/main.go --line 42 --severity HIGH --type BUG \
  --message "null ptr"

# Merge check (experimental)
bitbottle code-insights merge-check set PROJ/REPO CHECK_KEY \
  --report-key REPORT_KEY --must-pass --min-severity MEDIUM
```

All subcommands support `--hostname` and `--json / --jq` where applicable.
Merge-check verbs are marked experimental (partly undocumented API).

## Permissions quick-reference _(Server / DC only)_

```bash
# List all grants for a project (ADMIN → WRITE → READ)
bitbottle --hostname HOST perms project list MYPROJ [--json] [--jq EXPR]

# Grant / revoke project permission
bitbottle --hostname HOST perms project grant MYPROJ PERM --user SLUG
bitbottle --hostname HOST perms project grant MYPROJ PERM --group "name"
bitbottle --hostname HOST perms project revoke MYPROJ --user SLUG [--force]

# List / grant / revoke repo permission
bitbottle --hostname HOST perms repo list MYPROJ/REPO [--json] [--jq EXPR]
bitbottle --hostname HOST perms repo grant MYPROJ/REPO PERM --user SLUG
bitbottle --hostname HOST perms repo revoke MYPROJ/REPO --group "name"
```

Valid PERMs: `PROJECT_READ`, `PROJECT_WRITE`, `PROJECT_ADMIN`, `REPO_READ`,
`REPO_WRITE`, `REPO_ADMIN`. Grant warns on TTY if new perm < current;
pass `--force` to skip. MCP tools: `list/grant/revoke_project_permission`,
`list/grant/revoke_repo_permission`.

## Issues (Cloud only)

Issues are gated behind the issue tracker being enabled on the repo. All
commands accept `[PROJECT/REPO]` as an optional first arg; if omitted the
repo is inferred from the current checkout.

```bash
# List / view
bitbottle issue list [PROJECT/REPO] [--state open|new|on-hold|…|all] [--limit N] [--json fields] [--jq expr]
bitbottle issue view [PROJECT/REPO] ID [--json fields] [--jq expr]

# Create / close / reopen
bitbottle issue create [PROJECT/REPO] --title "T" [--body "B"] [--kind bug|enhancement|proposal|task] [--priority trivial|minor|major|critical|blocker]
bitbottle issue close  [PROJECT/REPO] ID
bitbottle issue reopen [PROJECT/REPO] ID

# Edit (all flags optional; supply only what you want to change)
bitbottle issue edit [PROJECT/REPO] ID [--title "T"] [--body "B"] [--kind …] [--priority …] [--assignee USER] [--state …]

# Assign
bitbottle issue assign [PROJECT/REPO] ID USER

# Comments
bitbottle issue comment list   [PROJECT/REPO] ISSUE_ID [--json fields] [--jq expr]
bitbottle issue comment add    [PROJECT/REPO] ISSUE_ID --body "text"
bitbottle issue comment edit   [PROJECT/REPO] ISSUE_ID COMMENT_ID --body "new text"
bitbottle issue comment delete [PROJECT/REPO] ISSUE_ID COMMENT_ID
```

Valid states: `new`, `open`, `resolved`, `on hold`, `invalid`, `duplicate`, `wontfix`, `closed`.
Use `--state on-hold` on the CLI (the hyphen is normalized; the API uses a space).

MCP tools: `list_issues`, `get_issue`, `create_issue`, `close_issue`,
`update_issue`, `reopen_issue`, `assign_issue`, `list_issue_comments`,
`add_issue_comment`, `edit_issue_comment`, `delete_issue_comment`.

## Named Credential Profiles

Profiles are kubectl-context-like named credential sets stored in
`~/.config/bitbottle/profiles.yml`.

```bash
# Create a profile (--hostname and --token are required)
bitbottle profile create work \
  --hostname git.work.com --token BBDC-... \
  --user alice --skip-tls --backend server

# Create a Cloud profile
bitbottle profile create personal \
  --hostname bitbucket.org --token APP_PASSWORD \
  --auth-user you@example.com --backend cloud

# Switch the active host config to a profile
bitbottle profile use work

# List profiles (token is never printed)
bitbottle profile list
bitbottle profile list --json name,hostname,backend_type,skip_tls_verify

# Delete a profile
bitbottle profile delete work --confirm
```

See `references/profile.md` for full flag reference.

## extension
- `extension install USER/REPO` — install from GitHub (repo must be named bitbottle-<name>)
- `extension install --local PATH` — symlink local extension dir
- `extension list` — show installed extensions (name, version, source)
- `extension upgrade NAME [--force]` — upgrade single extension to latest release; prints "local install — skipping" for local installs
- `extension upgrade --all [--force]` — upgrade all non-local extensions
- `extension remove NAME` — delete extension directory entirely
- `extension exec NAME [args...]` — run an installed extension; strips `*KEYRING_PASSPHRASE*`/`*KEYRING_PASSWORD*` from env, injects `BB_TOKEN` (from `$BB_TOKEN`) and `BITBOTTLE_VERSION`; exit code propagated

## Install / version

`npm install -g @proggarapsody/bitbottle` installs the CLI and
auto-registers this skill. For non-npm installs (Homebrew, Go, bare
binary): `bitbottle skill install` (refresh), `skill path` (locate),
`skill remove` (uninstall). Drift check: `python3 skills/scripts/sync_help.py`.
