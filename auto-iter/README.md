# auto-iter

Self-contained library for bitbottle's autonomous iteration loop. Everything that's auto-iter-specific lives here; everything broader (workflow, pre-merge gate, taste/architecture rules, BACKLOG) is referenced out.

## What's in here

| File | Role |
|---|---|
| [`quickref.md`](quickref.md) | Declarative contract — model tier per phase, halt routing, cycle/metrics log schemas, brainstorm rules, anti-patterns, cadence. The orchestrator reads this at the top of every run. |
| [`scripts.md`](scripts.md) | Interface catalog for every shell script under `scripts/`. JSON contracts, exit-code semantics, implementation status (✅ / 🔲). |
| [`scripts/`](scripts/) | Mechanical building blocks the orchestrator calls instead of running inline shell. Each script ships with a paired `_test.sh` that sandboxes via `mktemp -d` + `git init`. |
| [`metrics.csv`](metrics.csv) | Hand-rolled per-cycle ledger (early-cycle history). The `.jsonl` files in `.claude/auto-iter/` are the authoritative real-time log; this CSV is a curated subset. |

## What's NOT here (and why)

| Lives at | Why outside auto-iter/ |
|---|---|
| [`docs/workflows/iteration-cycle.md`](../docs/workflows/iteration-cycle.md) | Canonical workflow — used by `/iteration-cycle` (humans) AND `/auto-iter` (autonomous). Not auto-iter-specific. |
| [`docs/workflows/pre-merge-check.md`](../docs/workflows/pre-merge-check.md) | Repo-wide gate. Referenced by `AGENTS.md` and manually invoked on PRs. Auto-iter calls it as one step among many. |
| [`docs/TASTE.md`](../docs/TASTE.md), [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) | Repo-wide design rules. The design-judge subagent reads them; they apply to every PR, not just auto-iter. |
| [`BACKLOG.md`](../BACKLOG.md) | Project backlog. |
| `.claude/auto-iter/*.jsonl` + `.lock` | Runtime state. Gitignored. Read/written by scripts here, but not part of the library itself. |
| `.claude/commands/auto-iter.md` | The orchestrator entry point. Gitignored, local-only. Defers to `quickref.md` + `scripts.md` here for everything mechanical. |

## Run the script tests

```bash
make test-scripts
```

## Convention

Every script in `scripts/` emits a **single JSON object on stdout**. Stderr is for progress. Exit 0 = success; exit ≠ 0 = halt-class with `{"halt":"<reason>","details":"..."}`. Full contract in [`scripts.md`](scripts.md).
