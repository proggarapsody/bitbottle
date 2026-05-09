# TASTE — bitbottle UX & agentic experience

Single source of truth for what bitbottle should *feel* like. Read before
designing a new command, flag, error site, MCP tool, or skill entry.

> **Mantra**: bitbottle is `gh` for Bitbucket. **gh philosophy throughout** —
> noun-verb commands, consistent flags, TTY-aware output, thin commands.
> When in doubt, look at how `gh` does it (`reference/gh/` is checked in).

---

## CLI surface

### Command shape
- **Noun-verb**: `bitbottle pr list`, never `list-pr` or `list pr`.
- **Single noun per package**: `pkg/cmd/<noun>/<verb>.go`. Sub-nouns nest:
  `pkg/cmd/pr/comment/list.go`.
- **Args positional, options flagged**: `bitbottle pr view PR_ID`, not
  `--id PR_ID`.

### Standard flags (every applicable command)

| Flag | Behavior |
|---|---|
| `--json` | JSON to stdout. List/view operations only. |
| `--jq EXPR` | Server-side filter on `--json` output. |
| `--hostname HOST` | Override host from config / `bitbottle.host`. |
| `--limit N` | Cap list operations. |
| `--web` | Open the URL in browser. |
| `--confirm` | Required on non-TTY for destructive ops. |
| `--no-color` (global) | Disable color. `NO_COLOR` env honored. |
| `--debug` (global) | Surface raw transport errors. |

### Output
- **TTY**: aligned table with header. State columns colorized (`SUCCESSFUL`
  green, `FAILED` red, `MERGED` magenta) via `iostreams.ColorEnabled()`.
- **Pipe**: tab-separated, no header. Same fields, machine-readable.
- **`--json`**: stable schema across releases. Domain types normalized in
  `api/backend/types.go`.
- **Pager**: long output (`pr diff`, `commit log`, `pipeline logs`) opts in
  via `cmdutil.PagerAnnotation`. Respects `$PAGER`, defaults to `less -FRX`.

### Error UX
- Every error = **title + cause + 0..N hints**. Never raw `fmt.Errorf` for
  user-facing failures.
- Classified at the adapter via `httpx.Transport.UseDomainErrors(host)`;
  codes attached via `errfmt.Wrap`.
- Catalogue in `pkg/errfmt/` — dotted codes (`auth.invalid_token`,
  `pr.merge.conflict`, `host.unsupported`, …). Codes are stable IDs, never
  humanized into the code itself.
- Hints are **deterministic** — keyed off the code, not regex over the
  message.
- TTY: title bold, hints indented. Pipe: single-line `code: cause -- hint1 ; hint2`.
- `--debug` adds wrapped raw error + URL + HTTP status.

### Confirmations
- Destructive ops (`branch delete`, `repo delete`, `webhook delete`,
  `pr decline`, `tag delete`) require interactive confirm OR `--confirm`.
- Non-TTY without `--confirm` → exit with `confirmation required` error.

### Help text
- **Every** cobra command has non-empty `Short` (≤ 60 chars, no trailing period).
- `Long` includes 1–3 example invocations.
- `Examples` field used when there are >2 useful invocation patterns.

---

## Agentic experience (`skills/SKILL.md`)

### Purpose
A single compressed file an LLM agent reads to find the right command in
<2 seconds. Bundled with the npm package; published as part of
`@proggarapsody/bitbottle`.

### Shape
- **Compactness over completeness.** Agents tab-complete with `--help`.
  SKILL.md lists the surface; details live in `--help`.
- **Routes** to `skills/references/*.md` for clusters needing depth (auth, pr,
  repos, api).
- One row per command: `cmd ARGS — short description`. Flags only when
  non-obvious or non-standard.

### Update discipline
- Every new command / flag → SKILL.md row added. Pre-merge §5 doc-sync table
  enforces this as BLOCKER.
- Every new error code → catalogue row in `pkg/errfmt/` + hint phrasing.

### What good looks like
- Agent reads SKILL.md once, identifies the command, runs it with `--json`,
  parses the output. Zero re-discovery.
- Agent gets an error → reads the dotted code → looks up the hint → recovers
  without human help.

---

## MCP server (`pkg/cmd/mcp/`)

### Shape
- One MCP tool per Bitbucket operation. Tools are flat (no nested structures
  in the tool name).
- **Tool naming**: `snake_case_verb_noun` — `list_prs`, `create_branch`,
  `add_pr_comment`. Mirrors gh's MCP style.
- **Inputs**: structured JSON schema mirroring CLI flag names, kebab → snake.
  `pr_id` not `prId` or `PRID`.
- **Outputs**: JSON envelope `{result, error?, error_code?, error_hints?}`.
  `error_code` matches the dotted code from `errfmt`.

### Handler discipline
- **Thin handlers**: validate input → call `backend.Client` → return envelope.
  No business logic in the handler.
- **One handler per tool**, one test per handler.
- Pre-merge §5 doc-sync row: tool ↔ handler ↔ test triplet must be complete.

### Coverage rule
- Every Bitbucket operation that's not interactive-only (i.e., not `browse`)
  gets an MCP tool. CLI-only escape hatches (`api PATH`) don't need MCP
  coverage.

---

## Anti-patterns

- ❌ Verbs as top-level commands (`bitbottle list-prs`).
- ❌ Flag names that don't match existing vocabulary (`--target-name` instead
  of `--name`).
- ❌ JSON output that isn't stable across releases.
- ❌ Errors without codes or hints; bare `fmt.Errorf` for user-facing failures.
- ❌ MCP tools with camelCase names or nested input schemas.
- ❌ Help text that's empty or just paraphrases the cmd name.
- ❌ Color or pager that doesn't respect `NO_COLOR` / `--no-color` / `$PAGER`.
- ❌ Skipping `--confirm` requirement on destructive ops.
