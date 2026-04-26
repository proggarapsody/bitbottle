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

```
main  ←── PRs only, protected ──── triggers release on tag push
 ↑
dev   ←── daily work, CI runs here
 ↑
feature/*, fix/*, chore/*
```

| Branch | Purpose | Direct push |
|--------|---------|-------------|
| `main` | Production; every commit is a release candidate | Blocked — PRs only |
| `dev` | Integration; CI (test + lint + build) runs on every push | Allowed |
| `feature/*` etc. | Short-lived work branches | Allowed |

**main is protected.** Force-push and direct push are blocked. All changes must arrive via a PR from `dev`. CI (Test, Lint, Build) must pass before a PR can be merged.

---

## Development workflow

```bash
# 1. Start from dev
git checkout dev
git pull

# 2. Create a work branch
git checkout -b feature/my-thing

# 3. Develop, commit
git add ...
git commit -m "feat: add my thing"

# 4. Push and open a PR against dev (for review / sharing)
git push -u origin feature/my-thing
gh pr create --base dev

# 5. After review, merge into dev
# CI runs on dev automatically after merge
```

---

## Release workflow

Releases are fully automated via [Release Please](https://github.com/googleapis/release-please).

```bash
# 1. Open a PR from dev → main
gh pr create --base main --head dev --title "Release vX.Y.Z"

# 2. CI (Test, Lint, Build) must pass — merge when green
```

After the PR merges, Release Please automatically opens a **"Release vX.Y.Z"** PR on `main` with a computed version (based on conventional commit types) and an updated `CHANGELOG.md`. Merge that PR to trigger the release — no manual tagging needed.

**Version bumps follow conventional commits:**

| Commit prefix | Bump |
|---|---|
| `fix:` | patch (0.1.0 → 0.1.1) |
| `feat:` | minor (0.1.0 → 0.2.0) |
| `feat!:` or `BREAKING CHANGE` | major (0.1.0 → 1.0.0) |

Merging the Release Please PR triggers the release workflow, which:
- Builds binaries for Linux, macOS (arm64 + amd64), and Windows
- Creates a GitHub release with a changelog and checksums
- Builds `.deb`, `.rpm`, and `.apk` packages attached to the release
- Pushes multi-arch Docker images to `proggarapsody/bitbottle` on Docker Hub

**Required secrets** (set in repo Settings → Secrets → Actions):

| Secret | Purpose |
|--------|---------|
| `DOCKER_PASSWORD` | Docker Hub password / access token for `proggarapsody` |

**npm MCP wrapper** lives in `packages/mcp-npm/`. Publish manually after each release:
```bash
cd packages/mcp-npm
# Bump version to match the release tag (strip the v prefix)
npm publish --access public
```

**Versioning follows [Semantic Versioning](https://semver.org/):**
- `vMAJOR.MINOR.PATCH` — breaking change / new feature / bug fix
- Use `-rc.N` suffix for release candidates (`v1.0.0-rc.1`)

---

## Development setup

```bash
# Install Go 1.21+
go version

# Fetch dependencies
go mod tidy

# Build
make build

# Run tests (with race detector)
make test

# Lint
make lint
```

---

## Code style

- Format with `gofmt` (enforced by CI).
- Lint with `golangci-lint` — see `.golangci.yml` for enabled linters.
- Error messages: **lowercase**, no trailing punctuation, wrap cause with `%w`.
  ```go
  // good
  return fmt.Errorf("could not load config: %w", err)
  // bad
  return fmt.Errorf("Could not load config: %v.", err)
  ```
- Column headers printed by list commands must be **uppercase** (e.g. `SLUG`, `TITLE`).
- Keep methods short; extract helpers when cyclomatic complexity exceeds 10.
- No comments that restate what the code already says. Only comment the *why* when it is non-obvious.

---

## Adding a new command

1. Create `pkg/cmd/<group>/<action>.go` following the pattern in `repo/list.go`.
2. Register the command in `pkg/cmd/<group>/<group>.go`.
3. Add unit tests in `<action>_test.go` and integration tests in `<action>_integration_test.go`.
4. Use `factory.NewTestFactory` in tests — never touch the real filesystem, keyring, or network.

---

## Testing conventions

- **Unit tests** live alongside the command file (`list_test.go`).
- **Integration tests** use `httptest.NewTLSServer` and `factory.NewTestFactory` (`list_integration_test.go`).
- Always run `go test ./... -race` before opening a PR.
- Test names follow `Test<Package>_<Scenario>_<Outcome>`.
- Use `require` for fatal assertions (stops the test), `assert` for non-fatal ones.
