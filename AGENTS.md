# AGENTS.md — AI Agent Guidelines

This file tells AI agents (Claude, Copilot, Cursor, etc.) how to work in this
repository. Follow these rules exactly. They override any generic defaults.

---

## Branch strategy — follow CONTRIBUTING.md

```
main  ←── PRs only, protected ───── triggers Release Please on merge
 ↑
feature/*, fix/*, chore/*  ←── short-lived work branches
```

**Never push directly to `main`.** All changes must arrive via a pull request.

### Workflow for every change

```bash
# 1. Start from an up-to-date main
git checkout main
git pull

# 2. Create a work branch using the correct prefix
git checkout -b feature/<slug>   # new functionality
git checkout -b fix/<slug>       # bug fixes
git checkout -b chore/<slug>     # tooling, deps, CI

# 3. Develop and commit with Conventional Commits
git add <files>
git commit -m "feat: <description>"   # triggers minor bump
git commit -m "fix: <description>"    # triggers patch bump
git commit -m "chore: <description>"  # no version bump

# 4. Push and open a PR against main
git push -u origin <branch>
gh pr create --base main
```

CI (Test, Lint, Build) must pass before merging. Do not merge if lint or tests
are red.

---

## Commit message format

Use [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix | Effect | Example |
|--------|--------|---------|
| `feat:` | minor bump | `feat(auth): add SSO support` |
| `fix:` | patch bump | `fix(pr): handle empty PR list` |
| `chore:` | no bump | `chore: update golangci-lint` |
| `feat!:` or `BREAKING CHANGE` | major bump | `feat!: remove legacy XML output` |

Scope is optional but preferred: `feat(auth):`, `fix(repo):`, etc.

---

## Code style — from CONTRIBUTING.md

- **Factory injection** — every command receives `*factory.Factory`; no global state.
- **IOStreams** — all output goes through `f.IOStreams`; never `fmt.Println` or `os.Stdout` directly.
- **TTY-aware** — aligned columns in TTY mode; tab-separated, no headers in non-TTY.
- **RunE over Run** — always use `RunE` so errors propagate correctly.
- **Errors** — lowercase, no trailing punctuation, wrap with `%w`:
  ```go
  return fmt.Errorf("could not load config: %w", err)
  ```
- **No comments** that restate the code. Only comment the *why* when non-obvious.
- **No global mutable state** — no `init()` side-effects; sentinel errors are fine.

---

## Linting — run before every commit

```bash
make setup   # one-time: activates .githooks/pre-commit (runs golangci-lint)
make lint    # run manually at any time
```

The pre-commit hook blocks the commit if `golangci-lint` finds issues. Fix all
lint errors before pushing — CI uses the same linter config (`.golangci.yml`).

Common rules enforced:

| Linter | Rule |
|--------|------|
| `noctx` | Use `http.NewRequestWithContext` + `client.Do`, not `client.Get/Head/Post` |
| `errcheck` | Check all errors; use `_ =` only for `resp.Body.Close()` |
| `bodyclose` | Always close response bodies |
| `unused` | Delete dead code, do not comment it out |

---

## Testing — from CONTRIBUTING.md

- Unit tests: `<action>_test.go` alongside the command file.
- Integration tests: `<action>_integration_test.go` using `httptest.NewTLSServer`.
- Never touch the real filesystem, keyring, or network in tests — use `factory.NewTestFactory`.
- Always run `make test` (includes `-race`) before opening a PR.
- Test names: `Test<Package>_<Scenario>_<Outcome>`.
- Use `require` for fatal assertions, `assert` for non-fatal.

---

## Adding a new command

1. Create `pkg/cmd/<group>/<action>.go` following the pattern in `repo/list.go`.
2. Register it in `pkg/cmd/<group>/<group>.go`.
3. Add unit and integration tests.
4. Wire any new backend methods through `api/backend/client.go` and `api/backend/types.go`.
5. Both Cloud (`api/cloud`) and Server (`api/server`) backends must implement new interfaces.

---

## What NOT to do

- Do not push directly to `main`.
- Do not use `--no-verify` to skip the pre-commit hook.
- Do not call `client.Get`, `client.Head`, or `client.Post` — use `client.Do` with a context.
- Do not write to `os.Stdout`/`os.Stderr` directly — use `f.IOStreams`.
- Do not add `fmt.Println` debug statements.
- Do not leave commented-out code.
- Do not open a PR without passing CI.
