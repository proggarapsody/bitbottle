# Contributing to bitbottle

## Philosophy

bitbottle follows the [GitHub CLI](https://github.com/cli/cli) design philosophy:

- **Factory injection** — every command receives a `*factory.Factory`; no global state.
- **IOStreams** — all output goes through `f.IOStreams`; never write directly to `os.Stdout`.
- **TTY-aware output** — aligned columns and headers in TTY mode; tab-separated, no headers in non-TTY mode (machine-readable).
- **RunE over Run** — Cobra commands use `RunE` so errors propagate correctly.
- **Errors are values** — never panic; always return errors. Wrap with `fmt.Errorf("...: %w", err)`.
- **No global state** — no `init()` side-effects, no package-level mutable vars (sentinel errors are fine).

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

## Adding a new command

1. Create `pkg/cmd/<group>/<action>.go` following the pattern in `repo/list.go`.
2. Register the command in `pkg/cmd/<group>/<group>.go`.
3. Add unit tests in `<action>_test.go` and integration tests in `<action>_integration_test.go`.
4. Use `factory.NewTestFactory` in tests — never touch the real filesystem, keyring, or network.

## Testing conventions

- **Unit tests** live alongside the command file (`list_test.go`).
- **Integration tests** use `httptest.NewTLSServer` and `factory.NewTestFactory` (`list_integration_test.go`).
- Always run `go test ./... -race` before opening a PR.
- Test names follow `Test<Package>_<Scenario>_<Outcome>`.
- Use `require` for fatal assertions (stops the test), `assert` for non-fatal ones.

## Pull request process

1. Fork the repository and create a feature branch.
2. Make small, focused commits with clear messages.
3. Run `make test lint` and verify everything passes.
4. Open a PR; the description should explain *why*, not just *what* changed.
