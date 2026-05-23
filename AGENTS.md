# AGENTS.md

See [CONTRIBUTING.md](CONTRIBUTING.md) for full workflow, code style, and testing conventions.

## Reference implementations

`reference/gh/` contains a shallow clone of [github.com/cli/cli](https://github.com/cli/cli). When in doubt about CLI design patterns (flag naming, config structs, auth flows), check how `gh` does it there first.

## Design principles

Read before designing any new command, interface, package, transport, MCP tool, or error site. Pre-merge-check §6 ("Design-judge") gates merges against both.

- [`docs/TASTE.md`](docs/TASTE.md) — UX (gh philosophy, standard flags, TTY-aware output, error format), agentic skill experience, MCP tool shape.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — SOLID, layered structure, composite + optional interfaces, deep modules, design decisions, gh references.

## Automation

The autonomous iteration loop is documented in `docs/workflows/iteration-cycle/`.
Agent-specific wrappers (e.g. `.claude/commands/auto-iter.md`) inherit one-way from
these canonical docs — they add delivery details (model names, tool invocation syntax)
without duplicating procedural content.

## Workflow

End-to-end procedures live in [`docs/workflows/`](docs/workflows/) and are tool-neutral — humans, Codex, Cursor, Aider, and Claude all follow them.

- [`docs/agent-primer.md`](docs/agent-primer.md) — required architectural reading for any subagent implementing a new scope.
- [`docs/workflows/iteration-cycle/`](docs/workflows/iteration-cycle/) — the iteration loop: [`README.md`](docs/workflows/iteration-cycle/README.md) (canonical procedure), [`quickref.md`](docs/workflows/iteration-cycle/quickref.md) (halt routing, outcome enum, cadence), [`scripts.md`](docs/workflows/iteration-cycle/scripts.md) (script catalog), [`autonomous.md`](docs/workflows/iteration-cycle/autonomous.md) (autonomous-mode deltas), [`parallel-mode.md`](docs/workflows/iteration-cycle/parallel-mode.md) (multi-scope iteration).
- [`docs/workflows/pre-merge-check.md`](docs/workflows/pre-merge-check.md) — the merge gate (sections 0–9). Must pass before any branch lands on `main`.

Agent-specific wrappers (e.g., `.claude/commands/`) are tracked in git and inherit one-way from these docs — see `.claude/commands/` for Claude-specific delivery details.

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
api/cloud/gen/      — spec-derived wire types for Cloud (oapi-codegen); only api/cloud/*.go may import this
api/server/         — Bitbucket Server/DC adapter
api/server/gen/     — spec-derived wire types for Server/DC (oapi-codegen); only api/server/*.go may import this
api/internal/httpx/ — shared HTTP transport (internal)
internal/           — config, envvars, bbinstance, keyring, etc.
pkg/cmd/            — Cobra commands (one package per noun)
pkg/cmd/mcp/        — MCP stdio server (tools + handlers)
skills/SKILL.md     — Claude skill file for bitbottle (all commands in one file)
docs/manual-tests/  — manual test guides
docs/workflows/     — contributor + agent workflow checklists (pre-merge-check, iteration-cycle/)
packages/mcp-npm/   — npm wrapper (downloads Go binary on postinstall, bundles README)
```

## Wire types

**All adapter wire types live in `api/cloud/gen/types.go` and `api/server/gen/types.go`.**
Do not hand-write new `wireXxx` structs in `api/cloud/*.go` or `api/server/*.go` —
add the schema to the corresponding `openapi.yaml`, run `make gen`, and use the
generated type. The gen packages are adapter-internal: `api/backend/`, `api/internal/`,
and `pkg/cmd/` must never import them (enforced by `depguard`).

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
