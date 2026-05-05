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
```

`repo create --private=false` makes the repo public. Default is
private. `--private=true` is the explicit form.

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
bitbottle pipeline list WORKSPACE/repo [--limit N]
bitbottle pipeline view WORKSPACE/repo UUID [--web]
bitbottle pipeline run  WORKSPACE/repo --branch BRANCH
```

Pipeline UUIDs are returned by `pipeline list --json uuid`.
Self-hosted Server/DC has no pipelines product — these commands
will error out with "unsupported on host" against any non-Cloud host.

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

`repo delete`, `branch delete`, `tag delete` follow the canonical
destructive-op rule in SKILL.md (safety rule 2).
