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

[Release Please](https://github.com/googleapis/release-please) opens version PRs
from conventional commits: `fix:` bumps patch, `feat:` bumps minor, and `feat!:`
or `BREAKING CHANGE` bumps major. Merging the version PR creates the release tag.

The shared [candidate workflow](.github/workflows/release-candidate.yml) builds
once at the exact tag version, seals the binaries, packages, npm tarball, Docker
OCI layout, and release notes, then verifies those files on all five native
targets. Both Linux targets also install the packed npm wrapper and run the sealed
Docker image. Publication consumes the same files after all receipts and the
exact-SHA live Cloud evidence pass. Manual release dispatch defaults to
`publish=false`.

Follow [Release process](docs/release-process.md) for the verification command,
publication guards, and recovery when Cloud proof is missing. Candidate CI is
runtime evidence only after its jobs pass at the intended SHA; local validator
tests and older workflow runs do not establish that result.

**Required secrets** (set in repo Settings → Secrets → Actions):

| Secret | Purpose |
|---|---|
| `RELEASE_PLEASE_TOKEN` | PAT with contents + PRs write |
| `DOCKER_PASSWORD` | Docker Hub access token for `proggarapsody` |
| `NPM_TOKEN` | Granular Access Token with read+write on `@proggarapsody/bitbottle` |

---

## Development setup

The Go 1.26.9 toolchain pin includes the macOS arm64 fork/pre-exec crash fix
backported in Go 1.26.5 ([upstream issue #79804](https://github.com/golang/go/issues/79804),
[fix #79806](https://github.com/golang/go/issues/79806)). Keep the
`go 1.25.5` language minimum separate from the selected toolchain. Do not work
around this runtime crash by disabling the race detector or retrying tests.

```bash
# The Go toolchain is selected by go.mod (`toolchain go1.26.9`). The language
# compatibility minimum remains Go 1.25.5 (`go 1.25.5`).
go version

# Fetch dependencies
go mod tidy

# Install git hooks (gofmt pre-commit + golangci-lint pre-push)
make setup

# Build
make build

# Run tests (with race detector)
make test

# Run strict live acceptance checks (requires configured credentials/backend)
make test-acceptance

# Lint
make lint
```

For public CLI/MCP coverage, run `make qa-matrix`. The mandatory release gate
is `make qa-critical-check`, which verifies the reviewed release-critical
profile against a complete test run and its saved evidence. Run
`make qa-matrix-check` to require completion of the entire catalog; it remains
available for exhaustive QA follow-up. Ordinary PRs publish incomplete catalog
results while retaining test gates. See
[`docs/workflows/qa-matrix.md`](docs/workflows/qa-matrix.md) for the gate,
scenario claims, provider applicability, and gap reporting. Local commands
require explicit local success/error proof before their remote
non-applicability can be counted; derived classifications are not provider passes.

Cloud acceptance reads secrets only from the `cloud-acceptance` GitHub
Environment restricted to `main`. Keep bootstrap and actor credentials
separate; cleanup is limited to resources owned by the run manifest and must
be independently verified. A skipped live run is not a pass, and the actor
account cannot serve as an independent reviewer. See
[`docs/workflows/acceptance.md`](docs/workflows/acceptance.md) for the
configuration and workflow details, including the separate manual Cloud
auto-merge diagnostic. It checks authentication on the Cloud `post-build-merge`
route, then exercises a disposable pending-merge fixture with synthetic build
statuses. The report records provider behavior and does not change the release
gate.

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

## Agent skills and project knowledge

Development skills are installed in [`.agents/skills/`](.agents/skills/), pinned
to an upstream commit with provenance in `sources.json`. The root `skills/`
directory remains the Bitbottle consumer skill. See
[the knowledge map](docs/agents/README.md) for current docs and setup status. For planning and shipment, follow [work tracking](docs/workflows/work-tracking.md); keep GitHub Issues authoritative and the local scope index compact.

The custom autonomous iteration loop is retired. Its original procedures,
reports, and early specifications are indexed in [project history](docs/history/README.md).
Completed local sessions and runtime snapshots remain ignored; do not commit
private state. Keep `make test-scripts` and the validation tools covered by CI.
