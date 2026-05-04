# AGENTS.md

See [CONTRIBUTING.md](CONTRIBUTING.md) for full workflow, code style, and testing conventions.

## Reference implementations

`reference/gh/` contains a shallow clone of [github.com/cli/cli](https://github.com/cli/cli). When in doubt about CLI design patterns (flag naming, config structs, auth flows), check how `gh` does it there first.

## Key rules for AI agents

- **Branch + commits:** `feature/*` / `fix/*` / `docs/*` branch → PR to `main`. Never push directly to `main`. Use Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`).
- **Lint:** `make setup` once per clone, then `make lint` before pushing. Hook runs automatically on commit.
- **HTTP:** use `http.NewRequestWithContext` + `client.Do` — never `client.Get/Head/Post` (`noctx` linter).
- **Output:** always via `f.IOStreams`, never `os.Stdout`/`fmt.Println`.
- **Tests:** use `factory.NewTestFactory` — no real filesystem, keyring, or network.
- **New command:** `pkg/cmd/<group>/<action>.go` → register in `<group>.go` → implement in both `api/cloud` and `api/server` → add MCP tool in `pkg/cmd/mcp/` if it maps to a Bitbucket operation.
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
docs/               — manual test guides, design notes
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

## ContentTypePolicy

Injected into `httpx.Transport`:
- `ContentTypeAlwaysWrite` for Server/DC (avoids CSRF 403 on bodyless writes)
- `ContentTypeWhenBody` for Cloud (avoids 400 on empty POST)
