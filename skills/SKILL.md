---
name: bitbottle
description: >
  Reference for the bitbottle CLI — a gh-style tool for Bitbucket Server/DC
  and Cloud. Load when the user asks about bitbottle commands, auth setup, PRs,
  repos, branches, tags, commits, pipelines, or why a command failed. Load even
  if the user just says "bitbottle", mentions "Bitbucket", or pastes a bitbottle
  error message. Verified against bitbottle 1.13.1.
---

# bitbottle CLI

A gh-style CLI for Bitbucket Server/DC and Cloud. This file is a router
plus invariant safety rules — load the matching reference for command
detail. **Don't memorize this file; consult `-h` for any flag you're
not sure about.**

## When to load which reference

| Task | Open |
|---|---|
| Auth, hosts.yml, env vars, multi-host setup | `references/auth.md` |
| PR lifecycle (list/view/create/merge/approve/comment/…) | `references/pr.md` |
| Repos, branches, tags, commits, clone | `references/repos.md` |
| Raw REST passthrough, pagination, MCP server config | `references/api.md` |

When the user's task spans two areas, load both. Don't load all of
them speculatively.

## Safety rules (always apply)

1. **Never echo tokens.** Pass tokens via stdin (`--with-token`) or
   the `BB_TOKEN` env var. Never put a PAT/App Password on the
   command line — it lands in shell history.
2. **Confirm before destructive ops.** `repo delete`, `branch delete`,
   `tag delete`, `pr decline`, `pr merge` are not undoable. Before
   running, show the user:
   - the exact command,
   - the target host and `PROJECT/REPO`,
   - the PR ID / branch / tag name,
   then wait for explicit confirmation.
3. **Don't fabricate flags.** bitbottle has gh-like *shape* but not
   gh-compatible *flags*. If a flag isn't in the reference and the
   user asks for behavior you can't find, run `bitbottle <command> -h`
   first. Flags that DO NOT exist in 1.13.1 (commonly assumed):
   `--author`, `--state all`, `--mine`, `--all`, `--reviewer @me`.
4. **Prefer JSON for automation.** Parsing TTY tables is brittle; use
   `--json field1,field2 --jq 'expr'` whenever the output feeds another
   step. Most list/view commands support `--json` and `--jq` — verify
   per-command with `-h`.
5. **Check the version on behavior mismatches.** If a command behaves
   differently from this file, run `bitbottle --version`. This skill
   was last verified against **1.13.1**.

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
| Cloud-only commands | `pipeline *`, `pr request-changes` | — |

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

## Install / version

`npm install -g @proggarapsody/bitbottle` installs the CLI and
auto-registers this skill across detected agent runtimes (Claude Code,
Cursor, Codex, …). To check or refresh:

```bash
bitbottle --version
npx -y skills add proggarapsody/bitbottle --global -y
```
