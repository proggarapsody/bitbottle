# bitbottle repos / branches / tags / commits

All `list` commands below support `--limit N`, `--json` (full object as
JSON), and `--jq 'EXPR'` to filter fields. `--json` is boolean — don't
pass `--json field1,field2`.

## Repos

```bash
bitbottle repo list  [PROJ]                     # PROJ optional; defaults to all visible
bitbottle repo view  PROJ/repo [--web]
bitbottle repo create NAME --project PROJ [--description "x"] [--private=false]
bitbottle repo delete PROJ/repo [--confirm]     # destructive
bitbottle repo clone  [HOST/]PROJ/repo [DIR] [--ssh|--https]    # API-resolved URL; writes bitbottle.* git config; MCP: clone_repo
bitbottle repo set-default HOST/PROJ/repo       # writes .git/config in cwd
bitbottle repo rename   PROJ/repo NEW-NAME [--confirm]              # both backends; slug derives from name on Cloud — destructive
bitbottle repo fork     WS/repo --into TARGET-WS [--name NAME]       # Cloud only
bitbottle repo transfer PROJ/repo --to TARGET-PROJ                   # both backends; moves repo to another project (Server) or workspace (Cloud)
bitbottle repo file get PROJ/repo PATH --ref REF [--out FILE]        # read file content at a ref
bitbottle repo tree PROJ/repo [PATH] --ref REF [--json]       # list directory at a ref
bitbottle repo watcher list PROJ/repo                                 # list users watching a repo
bitbottle repo visibility PROJ/repo                                   # get visibility: "public" or "private"
bitbottle repo visibility PROJ/repo public                            # set repo public
bitbottle repo visibility PROJ/repo private                           # set repo private
bitbottle repo edit PROJ/repo --description "new desc"                # update description (both backends)
bitbottle repo edit PROJ/repo --website https://example.com           # update website (Cloud only)
bitbottle repo edit PROJ/repo --language Go                           # update language (Cloud only)
bitbottle repo edit PROJ/repo --fork-policy allow_forks               # set fork policy (Cloud only)
bitbottle repo edit PROJ/repo --enable-issues                         # enable issue tracker (Cloud only)
bitbottle repo edit PROJ/repo --disable-wiki                          # disable wiki (Cloud only)
```

`repo rename`, `repo fork`, and `repo transfer` accept `--json` and
`--jq expr` for structured output, like every other mutation.
`repo rename --confirm` is required on non-TTY: the slug change breaks
existing clones' `origin` URL, so users must run
`git remote set-url origin ...` after.

`repo create --private=false` makes the repo public. Default is
private. `--private=true` is the explicit form.

`repo fork` is Cloud-only — Bitbucket Server / Data Center has no fork
primitive in its REST API and the command returns a typed
unsupported-capability error on Server hosts.

`repo transfer` works on both backends. On Server it moves the repo to
a different project (identified by project key). On Cloud it moves the repo
to a different workspace (identified by workspace slug). MCP tool:
`transfer_repo(repo, target)`.

`repo file get` writes raw bytes to stdout (use `--out FILE` for binary).
`repo tree` normalises `type` to `file` or `dir` across both backends —
submodules surface as `dir` with the submodule pointer in `hash` so
agents can recurse uniformly. PATH defaults to the repo root. Both
commands require `--ref` (branch / tag / commit hash) and accept
`--hostname`. MCP equivalents: `get_file_content` and `list_tree`.

`repo watcher list` lists all users watching a repository. Works on
both Cloud and Server/DC. Columns: DISPLAY_NAME, USERNAME. Supports
`--json`, `--jq expr`, `--hostname`. MCP tool: `list_repo_watchers(repo)`.

## Downloads

Manage repository download artifacts. **Cloud only** — Server/DC returns a
typed `host.unsupported` error.

```bash
bitbottle repo download list   [WORKSPACE/REPO] [--limit N] [--json]
bitbottle repo download upload [WORKSPACE/REPO] FILE [--name NAME]
bitbottle repo download get    [WORKSPACE/REPO] NAME [--out PATH]
bitbottle repo download delete [WORKSPACE/REPO] NAME [--confirm]
```

`list` shows name, size (human-readable), download count, and creation date.
`upload` sends the file as a multipart form POST; `--name` overrides the filename.
`get` streams the artifact to a local file; `--out` sets the destination path.
`delete` requires `--confirm` outside a TTY.
MCP tools: `list_repo_downloads`, `upload_repo_download` (base64 content), `delete_repo_download`.

## Labels

```bash
bitbottle repo label list   [PROJECT/REPO] [--json]
bitbottle repo label create [PROJECT/REPO] --name N [--color C]
bitbottle repo label update [PROJECT/REPO] ID [--name N] [--color C]
bitbottle repo label delete [PROJECT/REPO] ID
```

Works on both Cloud and Server/DC. MCP tools: `list_repo_labels`,
`create_repo_label`, `update_repo_label`, `delete_repo_label`.

`repo visibility PROJ/repo` prints `public` or `private`. With a second
argument (`public` or `private`) it sets the visibility. Works on both
Cloud and Server/DC. MCP tool: `repo_visibility(repo[, visibility])`.

`repo edit PROJ/repo` updates mutable metadata fields. On Bitbucket Server /
Data Center only `--description` is forwarded; Cloud-only flags (`--website`,
`--language`, `--fork-policy`, `--enable-issues`, `--disable-issues`,
`--enable-wiki`, `--disable-wiki`) are accepted but silently ignored on Server
hosts. At least one flag is required. `--enable-issues` and `--disable-issues`
are mutually exclusive; same for `--enable-wiki` and `--disable-wiki`. MCP
tool: `edit_repo(repo, [description, website, language, fork_policy, has_issues, has_wiki])`.

## PR gate settings (Server/DC only)

```bash
bitbottle repo pr-settings get PROJ/repo              # show PR gate settings
bitbottle repo pr-settings get PROJ/repo --json       # as JSON

bitbottle repo pr-settings set PROJ/repo \
  --required-approvers 2 \
  --required-all-approvers \
  --required-successful-builds 1 \
  --merge-strategy no-ff \
  --allowed-strategies squash,no-ff
```

`repo pr-settings get` returns the current PR merge gate configuration:
required approvers, all-approvers flag, all-tasks-complete flag, required
successful builds, default merge strategy, and allowed merge strategies.

`repo pr-settings set` accepts any subset of the above flags; omitted flags
leave the current value unchanged (fetch-then-merge semantics). At least one
flag is required.

**Server/DC only.** Bitbucket Cloud does not expose this API; both commands
return a typed `host.unsupported` error on Cloud hosts.

MCP tools: `get_repo_pr_settings(project, repo)`, `set_repo_pr_settings(project, repo, [required_approvers, required_all_approvers, required_all_tasks_complete, required_successful_builds, merge_strategy, allowed_strategies])`.

## Branches

```bash
bitbottle branch list   PROJ/repo [--limit N]
bitbottle branch create PROJ/repo BRANCH START_AT           # positional (recommended)
bitbottle branch create PROJ/repo BRANCH --start-at main|HASH  # flag form
bitbottle branch checkout BRANCH                # fetches origin then `git checkout`
bitbottle branch delete PROJ/repo BRANCH        # destructive — confirm first
```

`--start-at` accepts a branch name OR a commit hash. Pass it as the
third positional (`[PROJECT/REPO] NAME START_AT`) or via `--start-at`;
both forms are equivalent. start-at is required — there is no default.

## Tags

```bash
bitbottle tag list   PROJ/repo [--limit N]
bitbottle tag create PROJ/repo TAG START_AT [--message "x"]      # positional (recommended)
bitbottle tag create PROJ/repo TAG --start-at main|HASH [--message "x"]  # flag form
bitbottle tag delete PROJ/repo TAG              # destructive
```

`--message` makes the tag annotated. Without it, it's lightweight.
`--start-at` is required — there is no default.

## Commits

```bash
# Branch resolution: --branch flag → current local branch → "main"
bitbottle commit log    PROJ/repo [--branch main] [--limit 5]
bitbottle commit view   PROJ/repo HASH [--web]
bitbottle commit status PROJ/repo HASH          # build/CI statuses for the commit
```

`commit log` runs against the resolved branch. Outside a checkout
without `--branch`, it falls back to `main` — pass `--branch` if the
default branch is named differently (`master`, `develop`, …).

## Pipelines (Cloud only)

```bash
bitbottle pipeline list  WORKSPACE/repo [--limit N]
bitbottle pipeline view  WORKSPACE/repo UUID [--web]
bitbottle pipeline run   WORKSPACE/repo --branch BRANCH

# Drill into a run:
bitbottle pipeline steps WORKSPACE/repo PIPELINE-UUID
bitbottle pipeline logs  WORKSPACE/repo PIPELINE-UUID STEP-UUID

```

Pipeline / step UUIDs are returned by `--json uuid`.
Self-hosted Server/DC has no pipelines product — these commands will
error out with "unsupported on host" against any non-Cloud host.

**Logs:** plain text, streams to stdout. Pipe to `less`, `grep`, redirect
to a file. No `--json`.

**Repository-level pipeline variables** — use `bitbottle variable --scope repository`:

```bash
bitbottle variable list   WORKSPACE/repo
bitbottle variable set    WORKSPACE/repo KEY [VALUE] [--body=-] [--secured]
bitbottle variable delete WORKSPACE/repo KEY --confirm
```

Secured values are redacted on read. The TTY column shows `<secured>` and
`--json` shows `"value":"<secured>"` — *the same chokepoint*, so secrets
cannot leak by switching output modes. Always use `--body=-` to feed a
secret value via stdin; never put it on the command line.

**`variable set` is upsert by KEY.** Existing → PUT, missing →
POST. No separate `update` command. Same for `delete`: takes the
user-friendly KEY, looks up the UUID internally.

## Webhooks (both backends)

```bash
bitbottle webhook list   PROJ/repo [--limit N]
bitbottle webhook view   PROJ/repo ID
bitbottle webhook create PROJ/repo --url URL --events EV1,EV2[,…] [--secret SECRET|-|@PATH] [--active=true|false]
bitbottle webhook delete PROJ/repo ID [--confirm]            # destructive
```

`--events` is a comma-separated list. Trim/dedupe is applied. Whitespace-only
input is rejected with a clear error.

**Event keys are backend-specific.** Cloud uses dotted-style — `repo:push`,
`pullrequest:created`, `pullrequest:approved`. Server/DC uses
`repo:refs_changed`, `pr:opened`, `pr:merged`. Passing a Cloud key to a Server
host (or vice versa) will surface as a backend validation error.

**`--secret`:** the raw string by default. `-` reads from stdin (recommended
for keeping secrets out of shell history); `@PATH` reads from a file. A
trailing newline from stdin/file is stripped. Secrets are write-only — neither
backend ever returns them on read.

**ID shape differs:** Cloud returns UUIDs (with curly-brace internal form
that `bitbottle` strips); Server returns numeric IDs. Both surface as
strings in the `id` field of `--json` output.

## Other

```bash
bitbottle alias set NAME 'COMMAND'
bitbottle alias list
bitbottle alias delete NAME

bitbottle config set git_protocol https --host HOST
bitbottle config get KEY [--host HOST]
bitbottle config list

bitbottle completion --shell bash|zsh|fish|powershell
```

## Destructive ops

`repo delete`, `branch delete`, `tag delete`, `webhook delete` follow the
canonical destructive-op rule in SKILL.md (safety rule 2).
