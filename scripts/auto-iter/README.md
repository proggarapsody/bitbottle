# scripts/auto-iter/

Mechanical building blocks for the `/auto-iter` orchestrator.

**Canonical interface catalog**: [`docs/auto-iter/scripts.md`](../../docs/auto-iter/scripts.md). That's where every script's input contract, JSON output shape, and replacement target lives. This README is just an entry point — don't duplicate the catalog here.

## Convention (one paragraph)

Each script emits a **single JSON object** on stdout. Stderr is for progress, never parsed. Exit 0 = success, exit ≠ 0 = halt-class failure with `{"halt":"<reason>","details":"..."}`. State lives under `.claude/auto-iter/` (gitignored). All scripts source `_common.sh` for `emit_json`, `halt`, `repo_root`, `auto_iter_dir`, `now_iso`.

## Run the tests

```bash
make test-scripts          # or:
for t in scripts/auto-iter/*_test.sh; do bash "$t" || exit 1; done
```

Tests sandbox in `mktemp -d` + `git init`, so they never touch real `.claude/auto-iter/` state.

## Adding a new script

1. Read [`docs/auto-iter/scripts.md`](../../docs/auto-iter/scripts.md) → Planned table — pick a row or add one.
2. Create `<name>.sh` + `<name>_test.sh` following the existing pattern (start from `metric.sh` for the simplest case).
3. `chmod +x` both.
4. Move the row from Planned → Implemented in the catalog.
5. Replace the corresponding inline block in `.claude/commands/auto-iter.md` with a script invocation.
