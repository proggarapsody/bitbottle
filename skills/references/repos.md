# bitbottle repos / branches / tags / commits

All `list` commands below support `--limit N`, `--json fields`, and
`--jq 'expr'` (see SKILL.md safety rule 4 for field discovery).

## Repos

```bash
bitbottle repo list  [PROJ]                     # PROJ optional; defaults to all visible
bitbottle repo view  PROJ/repo [--web]
bitbottle repo create NAME --project PROJ [--description "x"] [--private=false]
bitbottle repo delete PROJ/repo [--confirm]     # destructive
bitbottle repo clone  PROJ/repo [PATH]
bitbottle repo set-default HOST/PROJ/repo       # writes .git/config in cwd
bitbottle repo rename   PROJ/repo NEW-NAME [--confirm]              # both backends; slug derives from name on Cloud — destructive
bitbottle repo fork     WS/repo --into TARGET-WS [--name NAME]       # Cloud only
bitbottle repo transfer PROJ/repo --to TARGET-PROJ                   # both backends; moves repo to another project (Server) or workspace (Cloud)
bitbottle repo file get PROJ/repo PATH --ref REF [--out FILE]        # read file content at a ref
bitbottle repo tree PROJ/repo [PATH] --ref REF [--json fields]       # list directory at a ref
```

`repo rename`, `repo fork`, and `repo transfer` accept `--json fields` and
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

## Branches

```bash
bitbottle branch list   PROJ/repo [--limit N]
bitbottle branch create PROJ/repo BRANCH --start-at main|HASH
bitbottle branch checkout BRANCH                # fetches origin then `git checkout`
bitbottle branch delete PROJ/repo BRANCH        # destructive — confirm first
```

`--start-at` accepts a branch name OR a commit hash. There is no
default; the flag is required.

## Tags

```bash
bitbottle tag list   PROJ/repo [--limit N]
bitbottle tag create PROJ/repo TAG --start-at main|HASH [--message "x"]
bitbottle tag delete PROJ/repo TAG              # destructive
```

`--message` makes the tag annotated. Without it, it's lightweight.

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
