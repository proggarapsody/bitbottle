# Contributing to bitbottle

## Philosophy

bitbottle follows the [GitHub CLI](https://github.com/cli/cli) design philosophy:

- **Factory injection** — every command receives a `*factory.Factory`; no global state.
- **IOStreams** — all output goes through `f.IOStreams`; never write directly to `os.Stdout`.
- **TTY-aware output** — aligned columns and headers in TTY mode; tab-separated, no headers in non-TTY mode (machine-readable).
- **RunE over Run** — Cobra commands use `RunE` so errors propagate correctly.
- **Errors are values** — never panic; always return errors. Wrap with `fmt.Errorf("...: %w", err)`.
- **No global state** — no `init()` side-effects, no package-level mutable vars (sentinel errors are fine).

---

## Branch strategy

All work branches off `main` directly:

```
main  ←── PRs only, protected, triggers release on tag push
 ↑
feature/*, fix/*, docs/*, chore/*   (short-lived; delete after merge)
```

| Branch | Purpose | Direct push |
|---|---|---|
| `main` | Production; every commit is a release candidate | Blocked — PRs only |
| `feature/*` etc. | Short-lived work branches | Allowed |

**main is protected.** Force-push and direct push are blocked. All changes arrive via PR. CI (Test, Lint, Build) must pass before merging.

---

## Development workflow

```bash
# 1. Start from main
git checkout main && git pull

# 2. Create a work branch
git checkout -b feature/my-thing

# 3. Develop and commit (Conventional Commits required)
git add ...
git commit -m "feat: add my thing"

# 4. Push and open a PR against main
git push -u origin feature/my-thing
gh pr create --base main

# 5. Update branch if main has advanced, wait for CI, then merge (squash)
gh pr update-branch <N>
gh pr merge <N> --squash --delete-branch
```

---

## Release workflow

Releases are fully automated via [Release Please](https://github.com/googleapis/release-please).

After a PR with a `feat:` or `fix:` commit merges to `main`, Release Please automatically opens a versioned release PR. Merge that PR to trigger the full release — no manual tagging needed.

**Version bumps follow conventional commits:**

| Commit prefix | Bump |
|---|---|
| `fix:` | patch (1.0.0 → 1.0.1) |
| `feat:` | minor (1.0.0 → 1.1.0) |
| `feat!:` or `BREAKING CHANGE` | major (1.0.0 → 2.0.0) |

Merging the Release Please PR triggers the release workflow, which:
- Builds binaries for Linux, macOS (arm64 + amd64), and Windows
- Creates a GitHub release with changelog and checksums
- Builds `.deb`, `.rpm`, and `.apk` packages
- Pushes multi-arch Docker images to `proggarapsody/bitbottle` on Docker Hub
- Publishes `@proggarapsody/bitbottle` to npm (with README bundled)

**Required secrets** (set in repo Settings → Secrets → Actions):

| Secret | Purpose |
|---|---|
| `RELEASE_PLEASE_TOKEN` | PAT with contents + PRs write |
| `DOCKER_PASSWORD` | Docker Hub access token for `proggarapsody` |
| `NPM_TOKEN` | Granular Access Token with read+write on `@proggarapsody/bitbottle` |

---

## Development setup

```bash
# Install Go 1.21+
go version

# Fetch dependencies
go mod tidy

# Install git hooks (gofmt pre-commit + golangci-lint pre-push)
make setup

# Build
make build

# Run tests (with race detector)
make test

# Lint
make lint
```

---

## Code style

- Format with `gofmt` (enforced on every `git commit` by the pre-commit hook).
- Lint with `golangci-lint` (enforced on every `git push` by the pre-push hook) — see `.golangci.yml` for enabled linters.
- Error messages: **lowercase**, no trailing punctuation, wrap cause with `%w`.
  ```go
  // good
  return fmt.Errorf("could not load config: %w", err)
  // bad
  return fmt.Errorf("Could not load config: %v.", err)
  ```
- Column headers printed by list commands must be **uppercase** (e.g. `SLUG`, `TITLE`).
- Keep methods short; extract helpers when cyclomatic complexity exceeds 10.
- No comments that restate what the code already says. Only comment the *why* when non-obvious.

---

## Adding a new command

1. Create `pkg/cmd/<group>/<action>.go` following the pattern in `repo/list.go`.
2. Register the command in `pkg/cmd/<group>/<group>.go`.
3. Add unit tests in `<action>_test.go` and integration tests in `<action>_integration_test.go`.
4. Use `factory.NewTestFactory` in tests — never touch the real filesystem, keyring, or network.
5. If the command exposes new Bitbucket operations, add the corresponding MCP tool in `pkg/cmd/mcp/`.

---

## Testing conventions

- **Unit tests** live alongside the command file (`list_test.go`).
- **Integration tests** use `httptest.NewTLSServer` and `factory.NewTestFactory` (`list_integration_test.go`).
- Always run `go test ./... -race` before opening a PR.
- Test names follow `Test<Package>_<Scenario>_<Outcome>`.
- Use `require` for fatal assertions, `assert` for non-fatal ones.
- Coverage targets: **≥ 80%** on `api/cloud`, `api/server`, and `pkg/cmd/*`.

---

## Updating OpenAPI specs

Wire types for the Bitbucket Cloud and Server/DC adapters are generated from
OpenAPI specs using `oapi-codegen` in `types`-only mode.

**Specs live at:**
- `api/cloud/gen/openapi.yaml` — Bitbucket Cloud subset
- `api/server/gen/openapi.yaml` — Bitbucket Server/DC subset

**To regenerate types after updating a spec:**

```bash
make gen
```

This installs the pinned `oapi-codegen` version (see `Makefile`) if needed,
then regenerates `api/cloud/gen/types.go` and `api/server/gen/types.go`.

**Layer rule:** only `api/cloud/*.go` may import `api/cloud/gen`, and only
`api/server/*.go` may import `api/server/gen`. The gen packages are
adapter-internal — `api/backend/`, `api/internal/`, and `pkg/cmd/` must never
import them (enforced by `depguard`).

**Adding new wire types:** update the appropriate `openapi.yaml`, run
`make gen`, then use the generated type in the adapter file. Avoid
hand-writing new `wireXxx` structs — the gen packages are the canonical home
for all wire shapes.

---

## Important: do not commit build artifacts

`/dist/` is gitignored — never commit binaries or GoReleaser output. CI will reject tracked files in `dist/` or files larger than 1 MB.
