# AGENTS.md

See [CONTRIBUTING.md](CONTRIBUTING.md) for full workflow, code style, and testing conventions.

## Key rules for AI agents

- **Branch + commits:** feature/fix/chore branch → PR to `main`. Never push directly to `main`. Use Conventional Commits (`feat:`, `fix:`, `chore:`).
- **Lint:** `make setup` once per clone, then `make lint` before pushing. Hook runs automatically on commit.
- **HTTP:** use `http.NewRequestWithContext` + `client.Do` — never `client.Get/Head/Post` (`noctx` linter).
- **Output:** always via `f.IOStreams`, never `os.Stdout`/`fmt.Println`.
- **Tests:** use `factory.NewTestFactory` — no real filesystem, keyring, or network.
- **New command:** `pkg/cmd/<group>/<action>.go` → register in `<group>.go` → implement in both `api/cloud` and `api/server`.
