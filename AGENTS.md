# AGENTS.md

See [CONTRIBUTING.md](CONTRIBUTING.md) for full workflow, code style, and testing conventions.

## Reference implementations

`reference/gh/` contains a shallow clone of [github.com/cli/cli](https://github.com/cli/cli). When in doubt about CLI design patterns (flag naming, config structs, auth flows), check how `gh` does it there first.

## Workflow

End-to-end procedures live in [`docs/workflows/`](docs/workflows/) and are tool-neutral — humans, Codex, Cursor, Aider, and Claude all follow them.

- [`docs/workflows/iteration-cycle.md`](docs/workflows/iteration-cycle.md) — pick a scope → spec → TDD → docs → pre-merge-check → PR → release → close PRD → manual tests. The full loop.
- [`docs/workflows/pre-merge-check.md`](docs/workflows/pre-merge-check.md) — the merge gate (sections 0–8). Must pass before any branch lands on `main`.

Tool-specific auto-trigger wrappers (e.g. `.claude/skills/`) defer to these docs and stay local-only — they're not committed.

## Key rules for AI agents

- **Branch + commits:** `feature/*` / `fix/*` / `docs/*` branch → PR to `main`. Never push directly to `main`. Use Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`).
- **Lint:** `make setup` once per clone, then `make lint` before pushing. Hook runs automatically on commit.
- **HTTP:** use `http.NewRequestWithContext` + `client.Do` — never `client.Get/Head/Post` (`noctx` linter).
- **Output:** always via `f.IOStreams`, never `os.Stdout`/`fmt.Println`.
- **Tests:** use `factory.NewTestFactory` — no real filesystem, keyring, or network.
- **New command:** `pkg/cmd/<group>/<action>.go` → register in `<group>.go` → implement in both `api/cloud` and `api/server` → add MCP tool in `pkg/cmd/mcp/` if it maps to a Bitbucket operation.
- **Self-registration:** new commands self-register via `pkg/cmdregistry`. New MCP tools self-register from per-domain files in `pkg/cmd/mcp/`. Capability interfaces live in `api/backend/client_<feature>.go`.
- **Libraries:** prefer well-known, widely-adopted libraries over hand-rolled solutions. Pick the most popular/maintained option (e.g. `fatih/color` for ANSI color). Hand-roll only when dependency footprint outweighs value.
- **No build artifacts:** `/dist/` is gitignored. Never `git add dist/` or commit binaries. CI rejects tracked files in `dist/` or files > 1 MB.
- **Squash-merge gotcha:** GitHub's squash uses the PR title as the commit subject. Title PRs with `feat:`/`fix:` so release-please picks them up — otherwise a follow-up `feat:` commit is needed to trigger the release.

## Repository layout (key paths)

```
api/backend/        — shared domain types + Client interface
api/cloud/          — Bitbucket Cloud adapter
api/server/         — Bitbucket Server/DC adapter
api/internal/httpx/ — shared HTTP transport (internal)
internal/           — config, envvars, bbinstance, keyring, etc.
pkg/cmd/            — Cobra commands (one package per noun)
pkg/cmd/mcp/        — MCP stdio server (tools + handlers)
skills/SKILL.md     — Claude skill file for bitbottle (all commands in one file)
docs/manual-tests/  — manual test guides
docs/workflows/     — contributor + agent workflow checklists (pre-merge-check, iteration-cycle)
packages/mcp-npm/   — npm wrapper (downloads Go binary on postinstall, bundles README)
```

## Release pipeline (automated)

1. Merge `feat:`/`fix:` PR to `main`
2. Release Please opens a release PR automatically
3. Merge release PR → GoReleaser builds binaries, publishes GitHub release, pushes Docker images, publishes to npm (with `README.md` bundled)

## paging.Collect[T]

`api/internal/paging.Collect[T]` is the canonical helper for any `List*` operation. Never reintroduce a per-page-vs-total-cap pattern.

## Typed errors

`api/backend/errors.go` — `ErrNotFound`, `ErrAuth`, `ErrPermission`, `ErrUnsupportedOnHost`, `ErrConflict`, `ErrTransport`, `DomainError`. Adapters wrap via `httpx.Transport.UseDomainErrors(host)` — keep that wired on any new transport.

User-facing error rendering (titles + hints) lives in `pkg/errfmt`. Attach a `backend.ErrorCode` at the classification layer (and append it to `backend.AllCodes`); add the matching catalogue entry in `pkg/errfmt/errfmt.go`. Never format prose in `cmd/`.

Adapter call sites stamp call-site-specific codes via `backend.StampCode(err, code, resource, id, feature)` after the generic `ClassifyHTTPError` runs — used for `repo.not_found`, `pr.not_found`, `pr.merge.{conflict,behind}`, `pr.create.duplicate_branch`, `pr.reviewer.unknown`, `branch.protected`. Inspect `de.HTTPStatus()` to decide whether to stamp. Transport-layer errors (TLS, timeout) classify at `httpx.Transport.do()` via `backend.ClassifyTransportError` — no adapter wiring required. The MCP error envelope (`pkg/cmd/mcp/handlers.go`) surfaces the dotted code + hints from `errfmt.HintsFor(de)`.

## ContentTypePolicy

Injected into `httpx.Transport`:
- `ContentTypeAlwaysWrite` for Server/DC (avoids CSRF 403 on bodyless writes)
- `ContentTypeWhenBody` for Cloud (avoids 400 on empty POST)
