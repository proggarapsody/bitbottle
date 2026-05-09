# Product — bitbottle

## What it is

A CLI tool that brings `gh`-style ergonomics — noun-verb commands,
consistent flags, TTY-aware output — to both Bitbucket Cloud and
Server/DC.

## Problem

Bitbucket users have no first-class CLI. `gh` exists for GitHub, `glab`
for GitLab, but Atlassian ships only the web UI and a handful of
unmaintained third-party tools. Teams running on Bitbucket are stuck
context-switching to a browser for routine operations, or hand-rolling
REST calls when they need automation.

## Target users

In priority order:

1. **AI coding agents** (Claude Code, Cursor, Codex, etc.) — primary
   audience. Need structured `--json` output and an MCP server to operate
   on Bitbucket programmatically. Every command supports `--json` /
   `--jq` / `--hostname` for this reason.
2. **Developers using Bitbucket day-to-day** — want PRs, repos,
   pipelines, branches, tags, commits from the terminal instead of the
   web UI.

## Key goals

1. **Feature parity with `gh`** — every applicable `gh` command has a
   `bitbottle` equivalent against Bitbucket. Parity gaps live in
   `BACKLOG.md`.
2. **One UX, two backends** — the same commands and flags work against
   Bitbucket Cloud and Server/DC. Differences appear only where the
   underlying API genuinely differs (e.g., pipelines are Cloud-only;
   branch protections look different on Server).
3. **First-class agent surface** — every command supports `--json` /
   `--jq`, and the same operations are exposed via an MCP server
   (`pkg/cmd/mcp/`) so agents and the CLI share one surface.
4. **Predictable errors** — typed domain errors with dotted codes and
   actionable hints. `pkg/errfmt` renders them; adapters classify HTTP
   responses through `httpx.Transport.UseDomainErrors(host)`.

## Non-goals

- Replacing `git` — bitbottle complements it (like `gh`); `branch
  checkout` is a thin `git fetch + checkout` wrapper, not a
  reimplementation.
- Hosting Bitbucket-side state — bitbottle is stateless. All data lives
  in Bitbucket; local artifacts are just config + cached creds.
- Bitbucket Server/DC features that have no API surface — if Atlassian
  doesn't expose it via REST, bitbottle can't.

## Reference

- [`reference/gh/`](../reference/gh/) — shallow clone of the `gh` CLI
  source. The design reference for any unclear flag, command, or auth
  flow.
