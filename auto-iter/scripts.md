# auto-iter scripts — canonical interface catalog

> **Source of truth** for every shell script under `auto-iter/scripts/`. The
> `/auto-iter` orchestrator (and its sibling `/auto-iter-stream`) calls these
> scripts for everything mechanical. The orchestrator stays in the LLM only for
> work that needs judgment — TDD dispatch, brainstorm, audit, judge, halt
> phrasing.

---

## Why scripts, not prompts

Mechanical work that read raw `git`/`gh`/`jq` output through the LLM burned ~15–25k
orchestrator tokens per cycle and silently regressed (`metrics.jsonl` writes
stopped between cycles 41–55, see PR #295 audit). Scripts that emit a single
JSON object on stdout fix three problems at once:

1. **Deterministic output.** No malformed JSON from prompt drift.
2. **Testable.** Each script ships with `<name>_test.sh` and runs against a
   sandboxed `git init` repo.
3. **Cheap.** The LLM reads ~200 bytes of JSON instead of multi-KB raw shell
   output.

---

## Convention

Every script in `auto-iter/scripts/` follows this contract:

- **Stdout**: a single JSON object on a single line. Nothing else.
- **Stderr**: progress / log messages (never parsed).
- **Exit 0**: success. Caller reads the JSON.
- **Exit ≠ 0**: halt-class failure. Stdout is `{"halt":"<reason>","details":"..."}`.
- **State dir**: `.claude/auto-iter/` (gitignored) — created on demand.
- **Source helpers**: `. "$(dirname "$0")/_common.sh"` for `emit_json`,
  `halt`, `repo_root`, `auto_iter_dir`, `now_iso`.
- **Tests**: paired `<name>_test.sh` that runs in `mktemp -d` with `git init`.

---

## Catalog

Status legend: ✅ implemented · 🔲 planned · 🟡 partial.

### Implemented

| Script | Status | Inputs | Output (success) |
|---|---|---|---|
| [`_common.sh`](scripts/_common.sh) | ✅ | n/a (sourced) | helpers: `emit_json`, `halt`, `repo_root`, `auto_iter_dir`, `now_iso` |
| [`metric.sh`](scripts/metric.sh) | ✅ | `--cycle=N --step=NAME [--key=val ...]` | `{"cycle":N,"step":"...","ts":"...",...}` — also appended to `metrics.jsonl` |
| [`log-cycle.sh`](scripts/log-cycle.sh) | ✅ | `--cycle=N --outcome=... [...]` or `--stream=started\|completed [...]` | `{...}` — also appended to `cycles.jsonl` |
| [`lock.sh`](scripts/lock.sh) | ✅ | `acquire \| release \| status` | `{"acquired":true,"age_min":0}` or `{"halt":"recent_lock","age_min":N}` or `{"held":bool,"age_min":N}` or `{"released":true}` |
| [`preflight.sh`](scripts/preflight.sh) | ✅ | (none) | `{"clean":bool,"branch":"...","on_main":bool,"ahead":N,"behind":N,"open_prs":[{...}],"findings":[...]}`. Exit 1 if any finding is halt-class. |

### Planned

| Script | Status | Inputs | Output (success) | Replaces in orchestrator |
|---|---|---|---|---|
| `pick-mode.sh` | 🔲 | (none — reads `cycles.jsonl` + `BACKLOG.md`) | `{"mode":"iteration\|architecture\|brainstorm\|stop","cycle":N,"reason":"..."}` | §1 mode-pick block; cycle-counter parsing |
| `pick-scope.sh` | 🔲 | (none — reads `BACKLOG.md`) | `{"slug":"...","backend":"Cloud\|Server\|Both","pointer":N,"details_anchor":"#..."}` | §1 scope-pick BACKLOG-table parsing |
| `bundle-check.sh` | 🔲 | `<slug1> <slug2>` | `{"bundle":bool,"reason":"...","predicted_files":[...]}` | §1 bundle disjointness rules |
| `overlap-check.sh` | 🔲 | `<scope-slug>` | `{"overlapping_pr":N\|null,"matched_keywords":[...]}` | §0B open-PR overlap scan |
| `worktree.sh` | 🔲 | `create <slug>` \| `remove <path>` | `{"path":"...","branch":"..."}` or `{"removed":true}` | §2 worktree create/remove |
| `await-ci.sh` | 🔲 | `<pr-number>` | `{"all_passed":bool,"failed":[...],"elapsed_min":N}` | §6 CI wait loop |
| `await-release-pr.sh` | 🔲 | (none) | `{"release_pr":N,"version":"...","timed_out":bool}` | §7 release-please PR poll |
| `await-publish.sh` | 🔲 | `<version>` | `{"github":bool,"npm":bool,"version":"..."}` | §7 publish wait |
| `pre-merge-mechanical.sh` | 🔲 | (none — runs on current branch) | `{"findings":[...],"blocker":bool}` — bundles `smell-scan.sh`, `tdd-check.sh`, dist/-check, conventional-commit lint, release-please boundary grep, gitleaks | §5 mechanical pre-merge gate |

---

## Testing

Each script ships with `<name>_test.sh`. Tests run in a sandboxed `mktemp -d`
with `git init -b main`, so they never touch real state.

```bash
# Run every auto-iter script test:
for t in auto-iter/scripts/*_test.sh; do bash "$t" || exit 1; done

# Or via make (see Makefile):
make test-scripts
```

CI doesn't yet wire the script tests — the `Smell scan` job is the only shell
gate. Adding a `Script tests` job is a follow-up; the existing Go `Test` job
is unaffected.

---

## Anti-patterns

- ❌ Multiple JSON objects on stdout (callers must `jq -c .` a single line).
- ❌ Mixing log messages and JSON on stdout (use stderr).
- ❌ Reading state from anywhere other than the repo root and `.claude/auto-iter/`.
- ❌ Hard-coded `~/.claude/...` paths — tests must be able to sandbox.
- ❌ Calling out to gh/git without a `command -v` guard or a graceful no-op fallback (tests run without GitHub auth).
- ❌ Shipping a script without `<name>_test.sh`. Mechanical = testable.

---

## How the orchestrator dispatches

The local `.claude/commands/auto-iter.md` (gitignored) replaces inline shell
blocks with one-liners:

```bash
# §0a — lock
out=$(bash auto-iter/scripts/lock.sh acquire) || exit_skip "$out"

# §0 — preflight
out=$(bash auto-iter/scripts/preflight.sh) || halt_to_phone "$out"

# §4 — metric writes (one per step, near the end of each phase)
bash auto-iter/scripts/metric.sh --cycle=$CYCLE --step=step2_tdd \
  --duration_ms=$DUR --subagent_tokens=$TOK >/dev/null

# §5 — cycle log + release lock
bash auto-iter/scripts/log-cycle.sh --cycle=$CYCLE --mode=iteration \
  --scope=$SCOPE --outcome=shipped --pr=$PR --release=$REL >/dev/null
bash auto-iter/scripts/lock.sh release >/dev/null
```

The LLM receives only the JSON. The shell side stays compact and testable.
