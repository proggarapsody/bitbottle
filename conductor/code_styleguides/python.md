# Python style guide — bitbottle

Python is a tooling-only language in this project. The total surface
fits in `skills/scripts/`. This guide is intentionally short — the goal
is to keep the small scripts consistent and dependency-light, not to
become a Python codebase.

## When Python is appropriate

Tooling that lives outside the shipped CLI: code generators, doc
syncers, repo-housekeeping scripts. Examples on disk:

- `skills/scripts/sync_help.py` — re-syncs `--help` text from the Go
  binary into `skills/SKILL.md` and the per-noun reference files.
- `skills/scripts/test_sync_help.py` — tests for the syncer.

If a script is under ~50 lines and would otherwise be a fragile shell
pipeline, Python wins. If it grows, reconsider whether it should be a
Go subcommand or a Makefile target.

## Versions and dependencies

- **Python 3.x**, any modern minor version. The scripts must run on
  whatever ships with macOS / Ubuntu LTS — no `3.13`-only syntax.
- **Standard library only.** No `requirements.txt`, no `pyproject.toml`,
  no virtualenv. If you reach for a third-party package, that's a sign
  the script should be Go instead.
- **No build step.** Scripts run directly via `python3 path/to/script.py`.

## Style

Follow [PEP 8](https://peps.python.org/pep-0008/) with these
project-specific notes:

- **Imports**: standard library only; group as `import X` then `from X
  import Y`, alphabetised within each group.
- **Type hints**: encouraged on functions that take or return non-trivial
  data (paths, dicts, parsed structures). Skip them on tiny helpers.
- **Docstrings**: one-line summary on every public function. Module
  docstring at the top of every script explaining what it does and how
  it's invoked.
- **Errors**: print to `sys.stderr` and exit with a non-zero code. Don't
  swallow exceptions silently.
- **Subprocess calls**: prefer `subprocess.run(..., check=True,
  text=True, capture_output=True)` so failures are loud and output is
  decoded.
- **Paths**: use `pathlib.Path`, not `os.path`.

## Tests

Where they exist (`*_test.py`), keep them in the same directory as the
script and runnable with plain `python3 -m unittest`. No pytest, no
fixtures, no plugins.

```python
# skills/scripts/test_sync_help.py
import unittest
from pathlib import Path

class TestSyncHelp(unittest.TestCase):
    def test_routes_table_round_trips(self):
        ...
```

## Formatting

Use `black` if you have it installed locally; don't add it to CI.
Manual `python3 -m py_compile path/to/script.py` is the bare-minimum
syntax check.
