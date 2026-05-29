# Bitbucket CLI Comparison — Verified Report

> **Date:** 2026-05-27
> **Tested against:** `bitbucket.org/proggarapsody_main/cli-comparison-test`
> **Auth:** Atlassian API token (Basic, email:[TOKEN]); token redacted in all commands
> **Tools under test:**
> - **bitbottle** v1.121.0 (Go, npm-distributed; Bitbucket Server/DC + Cloud)
> - **bb** v0.18.0 (gildas/bitbucket-cli; Cloud-only)
> - **bkt** v0.26.6 (avivsinai/bitbucket-cli; Cloud + DC)

---

## 0. Methodology — read this first

This report has **two evidence tiers**:

- **Tier A (verified 2026-05-27, fresh):** bug repros run in the current session against the live test repo. Every claim has an exact command and observed output.
- **Tier B (carried forward from prior agent sweep 2026-05-26):** broader capability claims (75 use cases) made by background agents in a prior session. The raw `.meta` files were on `/tmp/` and have since been wiped by macOS. **Corroborated** by side-effects still visible on the test repo (11 PRs whose titles match what agents claimed they created), so the prior testing did happen — but individual claims cannot be replayed without the raw outputs.

The verdict is built **only on Tier A**. Tier B is summarized in the appendix as supporting context.

What's **deliberately excluded** from this report (was in earlier drafts, was synthetic):
- Round-number "5× parallel speedup" benchmarks — never witnessed with `time`
- Per-tool 75-UC matrices — built by agents with a thesis to support
- "Cleanup completed" claims — the test repo still has 11 PRs and several branches from the sweep

---

## 1. Verified bitbottle bugs (Tier A)

All 6 bugs reproduced in the current session. Each is fix-actionable and belongs in `BACKLOG.md`.

### BB-07 — `pr approve` succeeds on DECLINED PR

**Command:**
```
bitbottle pr approve 6 -R bitbucket.org/proggarapsody_main/cli-comparison-test
```
**Observed:** `Approved pull request #6` — exit 0
**State at time of test:** PR #6 was DECLINED
**Why it's a bug:** Bitbucket's REST API returns 200 for participant approval on a declined PR (API permissiveness). bitbottle echoes "Approved" without checking PR state. Scripts will see success when the action was semantically meaningless.
**Fix scope:** Check PR `state` in `Approve()` and return `ErrInvalidRequest` if `state ∈ {DECLINED, MERGED}`.

### BB-08 — `workspace project perms list` always 404

**Command:**
```
bitbottle workspace project perms list proggarapsody_main cli-comparison-test --hostname bitbucket.org
```
**Observed:** `Resource not found.` — exit 1
**Root cause:** bitbottle calls `/permissions/users`; the correct Bitbucket Cloud endpoint is `/permissions-config/users`.
**Fix scope:** Rename the path in `api/cloud/workspace_project_perms.go`.

### BB-09 — All `commit comment` subcommands 404 on Cloud

**Command:**
```
bitbottle commit comment list proggarapsody_main/cli-comparison-test b893d112fba64be0387ab65ab4a10979dfab9b57 --hostname bitbucket.org
```
**Observed:** `Resource not found.` — **exit 0** (silent failure)
**Source confirmation:** `api/cloud/commit_comment.go:32,52,65,74` all hardcode `/commits/{hash}/comments` (plural). Bitbucket Cloud's endpoint is `/commit/{hash}/comments` (singular). All four operations — list, add, edit, delete — share the bug.
**Fix scope:** Change `commits` → `commit` at the four `fmt.Sprintf` call sites. ~5 LOC + a regression test.

### BB-10 — `pipeline trigger` silently fails with JSON unmarshal error

**Command:**
```
bitbottle pipeline trigger proggarapsody_main/cli-comparison-test --branch main --hostname bitbucket.org
```
**Observed:**
```
json: cannot unmarshal object into Go struct field CloudTriggerResponseLinks.links.self of type []gen.CloudTriggerResponseSelfLink
```
Exit: **0** (silent failure)
**Severity note:** Earlier draft called this a "crash" — incorrect. It's worse than a crash. The CLI prints the error to stderr-ish and exits 0, so any script wrapping it sees success. The pipeline does get triggered server-side (verifiable via API), but no useful response data reaches the user.
**Root cause:** Generated `CloudTriggerResponseLinks.links.self` is typed `[]CloudTriggerResponseSelfLink` (slice) but Bitbucket returns a single object. Schema-generated struct vs. real API drift.
**Fix scope:** Fix the generated type (or hand-correct it) so `self` is a struct, not a slice. Add a unit test against a captured fixture.
**Workaround:** Use `bitbottle pipeline run` (different code path, works).

### BB-11 — `pr list --state INVALID_STATE` silently returns all PRs

**Command:**
```
bitbottle pr list -R bitbucket.org/proggarapsody_main/cli-comparison-test --state INVALID_STATE
```
**Observed:** Returns 11 PRs spanning MERGED, OPEN, and DECLINED. Exit 0. No warning.
**Root cause:** No client-side validation of `--state`; invalid value is forwarded to API as a query param, API ignores unknown values and returns the unfiltered list.
**Fix scope:** Validate `--state` against `{OPEN, MERGED, DECLINED, SUPERSEDED}` (case-insensitive) in `pr/list.go`. Reject with `ErrInvalidRequest` on miss.

### BB-12 (new) — Inconsistent exit codes on errors

Discovered during this verification round, not in the prior sweep.

| Failure | Output | Exit |
|---|---|---|
| BB-08 (workspace perms 404) | "Resource not found." | **1** |
| BB-09 (commit comment 404) | "Resource not found." | **0** |
| BB-10 (pipeline trigger unmarshal error) | "json: cannot unmarshal..." | **0** |

Same error string, different exit codes; runtime errors swallowed at exit 0 in two places. **Scripts cannot trust `$?` from bitbottle.** This is the most dangerous defect of all six because it makes every other bug invisible to automation.

**Fix scope:** Audit all command runners in `cmd/` for the `fmt.Println(err); return nil` anti-pattern. Should be `return err`. Likely a one-afternoon sweep of ~10–20 sites.

---

## 2. Tier-A bitbottle observations (positive)

These were also verified in this session:

- **Auth works clean.** `bitbottle auth status` correctly shows both bitbucket.org (Cloud) and the user's DC host. No spurious prompts. Token in macOS Keychain.
- **`pr list` and `pr view`** return correct data with sensible default columns. Multi-host disambiguation via `-R bitbucket.org/...` works (when omitted with multiple hosts, errors with a clear "specify HOST/PROJECT/REPO or use --hostname").
- **`pr commits N`** works and produces an exact commit hash + author + timestamp.
- **`pr approve` and `branch list`** return real data and execute their action server-side.

bitbottle is **not broken** as a tool — it's broken at specific endpoints and broken in its exit-code discipline. The shape of the CLI is sound.

---

## 3. Tier-B summary (prior sweep, not re-verified)

Corroborated by side effects (PR titles, branch names match agent claims), but individual claims cannot be replayed. Treat as directional, not authoritative.

### bitbottle uniquely strong areas (prior claims):
- Bitbucket Server/DC + Cloud parity (most CLIs are one or the other)
- MCP server with 254 tools — only tool here that exposes itself to AI agents
- Typed error catalogue (`api/backend/errors.go`) with structured DomainErrors

### Known bitbottle gaps (prior claims, not re-verified this round):
- No `commit list` or `commit view` standalone commands (workaround: `pr commits N`)
- No `--author` filter on `pr list` (BB-06)
- BB-01: `commit files` returns empty list on Cloud
- BB-02: `-R` flag inconsistent across some subcommands
- BB-08 already covered above

### Competitor positioning (prior claims, not re-verified):

**bb (gildas):**
- Strengths: fastest cold-start (~6 ms), CSV/TSV/columns/sort/dry-run output features, SSH/GPG key management, OAuth grant
- Gaps: no `branch create`/`branch delete` (silently no-op, exits 0), no webhook commands, no commit list/view, `tag create` sends malformed JSON payload, `pullrequest update` constructs bare relative URL
- Verdict positioning: fastest but Cloud-only and incomplete on writes

**bkt (avivsinai):**
- Strengths: named contexts (host+workspace+repo as one selector), `auth doctor` diagnostics, DC/Cloud parity labels in help, OSSF Scorecard + SBOM (supply-chain hygiene)
- Gaps: BUG-1 — unknown subcommands silently print parent help and exit 0 (e.g. `pr request-changes`, `pipeline stop`, `commit view`); BUG-2 YAML uses Go struct field names; BUG-3 `--template` broken for nested fields
- Verdict positioning: cleanest UX, but BUG-1 is a correctness hazard for scripts

---

## 4. Verdict

### bitbottle's competitive position

**Strongest of the three when you need:**
- Bitbucket Server/DC support (bb is Cloud-only; bkt does both but bitbottle's DC paths are more battle-tested per repo history)
- AI/MCP integration (only tool with an MCP server)
- A typed error catalogue for programmatic consumption

**Behind when you need:**
- Reliable exit codes for shell scripts → **BB-12 is the critical blocker**
- Commit-level workflows on Cloud (BB-09 blocks all commit comments; no `commit list`)
- Pipeline triggers that return usable response data (BB-10)

### Recommended fix order (impact × effort)

1. **BB-12 (exit-code audit)** — highest impact, ~1 day. Makes every other bug catch-able in CI.
2. **BB-09 (commit comment endpoint)** — ~5 LOC + test. Unblocks all four commit-comment operations.
3. **BB-08 (workspace perms endpoint)** — ~3 LOC + test.
4. **BB-11 (state validation)** — ~10 LOC + test.
5. **BB-10 (pipeline trigger struct)** — needs a generator fix or hand-correction; ~half-day.
6. **BB-07 (approve on DECLINED)** — needs a pre-check; ~20 LOC.

All six are local, scoped fixes. None require architectural changes. Total estimated work: **2–3 days** to close every Tier-A bug in this report.

### What the comparison says about bitbottle's market positioning

bitbottle's differentiation is **DC + MCP**, not raw feature count. bb beats it on output formatting and cold-start; bkt beats it on context ergonomics. Where bitbottle uniquely wins — Bitbucket Server users and AI-driven workflows — those audiences don't care about a 6 ms cold-start delta but **do** care about exit-code reliability and complete commit-comment support. **Fixing BB-12 + BB-09 alone closes the most visible quality gap with bkt** while preserving bitbottle's structural advantages.

---

## 5. Appendix — test artifact inventory

These persist in `bitbucket.org/proggarapsody_main/cli-comparison-test` and were **not** cleaned up by the prior sweep despite a "completed" task:

- PRs #3, #4 — DECLINED (titled `cli-test-bkt:` / `cli-test-bitbottle: EDITED PR title`)
- PRs #5, #6 — DECLINED (`cli-test-bb:` ...)
- PRs #7, #8, #9, #10, #11 — MERGED (test merge/delete-branch flows)
- PR #1 (`feat: add greet function`), PR #2 (`fix: correct typo`) — OPEN, predate the sweep

Test branches that may still exist: `cli-test-bitbottle-write`, `cli-test-bitbottle-write-2`, `cli-test-bb-write`, `cli-test-bkt-write`, `cli-test-bb-pr-branch`, `cli-test-bitbottle-pr-branch`, `cli-test-bkt-pr-branch`, `feature/add-greeting`, `fix/typo-in-readme`, `docs/add-contributing`.

Cleanup is a separate task — not done by this report.

---

## 6. Phase 2 — Depth pass (added 2026-05-27)

The first version of this report had only 6 verified bugs. That count was suspicious — too thin for a young CLI's first big external audit. A depth pass on negative inputs, state-machine doubles, output-format consistency, and reference-parsing patterns turned up **9 more bugs** plus several observations. All bugs in this section have been reproduced live in the current session and have logs in `~/cli-comparison-phase2/`.

### BB-13 — `-R` flag silently ignored on most non-PR commands ⭐ HIGH IMPACT

The `--repo` / `-R` flag is listed under `INHERITED FLAGS` in `--help` for every subcommand, implying it works everywhere. In reality it **only works on `pr` subcommands**. On these commands it's silently parsed and ignored, producing confusing "accepts N arg(s), received 0" errors:

| Command | `-R` accepted? | Observed |
|---|---|---|
| `pr list` / `pr view` / `pr create` / `pr approve` / ... | Yes | ✅ Works |
| `branch list -R bitbucket.org/ws/repo` | **No** | "accepts 1 arg(s), received 0" |
| `branch create -R ... NAME` | **No** | Same |
| `branch delete -R ... BRANCH` | **No** | Same |
| `repo view -R ...` | **No** | Same |
| `tag list -R ...` | **No** | Same |
| `tag delete -R ...` | **No** | Same |
| `pipeline list -R ...` | **No** | Same |
| `pipeline run -R ...` | **No** | Same |

**User-facing impact:** users adopt `-R` after one `pr` command, then it silently fails on every other command. The error message ("received 0") doesn't hint that `-R` was the problem — it suggests a missing positional. Workaround is undocumented: use positional `PROJECT/REPO` + `--hostname` instead, or 3-part positional `HOST/PROJECT/REPO` on commands that accept it.

**Fix scope:** in each command's `RunE`, before validating positional count, resolve `-R` flag to populate the expected positional. Or — better — remove `-R` from `PersistentFlags()` and only add it where actually wired. The first approach preserves help-text expectations; the second is honest.

**Verification logs:** `branch-list-R-flag.log`, `tag-list.log` (no -R, just `-R` flag), `pipeline-list.log`, `repo-view-R-flag.log`.

### BB-14 — `repo view` rejects 3-part `HOST/PROJECT/REPO` format

```
bitbottle repo view bitbucket.org/proggarapsody_main/cli-comparison-test
→ invalid repo ref "bitbucket.org/proggarapsody_main/cli-comparison-test": expected PROJECT/slug
EXIT: 1
```

But `branch list bitbucket.org/proggarapsody_main/cli-comparison-test` accepts the same 3-part format and works. So the parser used by `repo view` is stricter (2-part only) than the parser used by `branch list` (2 or 3 parts). This is the same root cause as BB-02 from the prior sweep but with a fresh repro on a different command surface.

**Fix scope:** unify the ref-parsing helper. There should be one function for `HOST/PROJECT/REPO | PROJECT/REPO` and every command should use it.

### BB-15 — `branch create PROJECT/REPO NAME` requires `--start-at` despite signature

Help text:
```
USAGE
  bitbottle branch create PROJECT/REPO NAME [flags]
```

Actual behaviour:
```
$ bitbottle branch create proggarapsody_main/cli-comparison-test mybranch --hostname bitbucket.org
required flag(s) "start-at" not set
EXIT: 1
```

The signature implies `NAME` is enough, but `--start-at` is mandatory. Either the signature should be `PROJECT/REPO NAME START_AT` (matching `git branch NAME START_POINT` ergonomics), or `--start-at` should default to `HEAD` / current branch HEAD.

**Fix scope:** EITHER promote `--start-at` to a 3rd positional, OR default it to the repo's HEAD when omitted. The current "required flag that looks like a positional" pattern is the worst of both worlds.

### BB-16 — `tag create` has the same `--start-at` issue as `branch create`

```
$ bitbottle tag create existing-tag HEAD -R ...
required flag(s) "start-at" not set
```

Same diagnosis and same fix as BB-15. Two commands sharing the bug suggests a shared helper / cobra cargo-cult pattern that should be factored.

### BB-17 — `--jq` flag silently ignored when `--json` not set (inconsistent across commands) ⭐ HIGH IMPACT

`--jq` is documented to "Filter JSON output with a jq expression" in every command's INHERITED FLAGS. Behavior in practice:

| Command | `--jq` without `--json` | Behavior |
|---|---|---|
| `pr list --jq '.[].title'` | Errors clearly | "--jq requires --json" exit 1 ✅ |
| `branch list --jq '.[].name'` | Errors clearly | "--jq requires --json" exit 1 ✅ |
| `pr view --jq '.title'` | **Silently ignored** | Returns full text view ❌ |
| `repo view --jq '.name'` | **Silently ignored** | Returns full text view ❌ |
| Same commands with `--json --jq ...` | All work | ✅ |

So the rule is "`--jq` requires `--json`" but only two of the four commands enforce it. The other two silently produce wrong output that looks correct (a full table when the user expected a single field). This is a script-trap: a CI step like `state=$(bitbottle pr view 1 --jq .state)` will set `state` to the entire PR view text instead of "OPEN".

**Fix scope:** centralize the `--jq requires --json` check in the formatter middleware, before any command's `RunE` body. Every command's output path goes through one place.

### BB-18 — `--template` output has no trailing newline (minor)

```
$ bitbottle pr view 1 --template '{{.title}}'
feat: add greet function with unit tests$ # shell prompt next to output, no newline
```

Causes ugly terminal output and breaks `output=$(cmd); echo "[$output]"` patterns.

**Fix scope:** append `"\n"` to the template execution result in the formatter.

### BB-19 — "Not found" hint suggests the resource was deleted, even when it never existed (minor)

```
$ bitbottle pr view 99999 -R ...
Pull request #99999 not found on api.bitbucket.org.
Cause: HTTP 404: Not Found
Hint:  It may have been deleted. Run `bitbottle pr list` to see open PRs.
```

The repo only has 11 PRs. #99999 was never created. The hint "may have been deleted" is misleading. Real-world impact: low (this is "polite verbiage"), but bitbottle's pitch is "Cause + Hint" — the hint should be accurate.

**Fix scope:** check `repo.maxPRId` (cheap query) before asserting "may have been deleted". Or generalize the hint: "Pull request #N does not exist in this repo."

### BB-20 — State-machine enforcement inconsistent across PR actions ⭐ HIGH IMPACT (combines with BB-07)

Tested today on PR #6 (DECLINED) and #7 (MERGED):

| Action | Server-side state | bitbottle behavior |
|---|---|---|
| `pr decline 6` on already-DECLINED | HTTP 400 "already closed" | ✅ Errors with proper hint |
| `pr merge 6` on DECLINED | HTTP 400 "already closed" | ✅ Errors with proper hint |
| `pr merge 7` on already-MERGED | HTTP 400 "already closed" | ✅ Errors with proper hint |
| `pr approve 6` on DECLINED | HTTP 200 (API permissive) | ❌ Reports "Approved" — **BB-07** |

So bitbottle correctly catches *most* invalid-state operations because the API returns 400 — but `pr approve` is the lone exception because the API is permissive for participant approvals. The bug is **the asymmetry between approve and decline/merge**: a user reading the source would expect the same check on all three.

**Fix scope:** combined with BB-07, add a `validateMutablePRState()` precheck that any PR-mutation command can call. Reject `DECLINED|MERGED|SUPERSEDED` for approve / request-changes / decline / merge / edit.

### BB-21 — `pr list --state ""` silently returns all PRs (same root cause as BB-11)

Confirmed: empty string state value is forwarded to API, which ignores it. Same fix as BB-11 (client-side validation against the known state enum).

### New positive observations (Tier A, verified today)

To be honest about what's good, not just bad:

1. **`pr decline` / `pr merge` correctly enforce state.** Most-of state machine is right (see BB-20).
2. **Unicode body in `pr comment add`** works perfectly. Comment #801530234 created with `🎉 émoji тест 中文`.
3. **Long comment body (5000 chars)** works. Comment #801530235 created.
4. **Empty `--body ""` is caught client-side** with a clear "--body is required" error. ✅
5. **JSON and YAML field names match** (both use camelCase like `fromBranch`, `toBranch`, `webURL`). Better than bkt's BUG-2 where YAML used Go struct names.
6. **`--jq` works correctly** when paired with `--json`. Output is valid JSON-derived.
7. **`pr view abc`** validates client-side: "invalid PR ID 'abc': must be a positive integer". ✅
8. **`--limit 0` / `--limit -1`** validated client-side: "--limit must be at least 1, got 0". ✅
9. **`--state open` (lowercase)** works — case-insensitive matching. ✅
10. **404 errors have helpful Cause+Hint** on most missing-repo / missing-PR paths.

---

## 7. Updated verdict (after Phase 2)

**Total verified bitbottle bugs: 14** (BB-07 through BB-21, excluding gaps in numbering where Phase 1 numbers were reserved).

| ID | Severity | Title | Phase |
|---|---|---|---|
| BB-07 | High | `pr approve` succeeds on DECLINED PR | 1 |
| BB-08 | High | `workspace project perms list` wrong endpoint | 1 |
| BB-09 | High | `commit comment` wrong endpoint (all 4 ops broken) | 1 |
| BB-10 | High | `pipeline trigger` silent JSON unmarshal failure | 1 |
| BB-11 | Med | `pr list --state INVALID` silently returns all | 1 |
| BB-12 | **Critical** | Inconsistent exit codes across error paths | 1 |
| BB-13 | **Critical** | `-R` flag silently ignored on most non-PR commands | 2 |
| BB-14 | Med | `repo view` rejects 3-part HOST/PROJECT/REPO format | 2 |
| BB-15 | Med | `branch create` `--start-at` required despite signature | 2 |
| BB-16 | Med | `tag create` same `--start-at` issue | 2 |
| BB-17 | High | `--jq` silently ignored on `pr view`/`repo view` without `--json` | 2 |
| BB-18 | Low | `--template` missing trailing newline | 2 |
| BB-19 | Low | "Not found" hint misleading when never existed | 2 |
| BB-20 | High | State-machine enforcement inconsistent (sister of BB-07) | 2 |
| BB-21 | Low | `--state ""` silently returns all (sister of BB-11) | 2 |

### The two critical ones drive everything else

**BB-12 (exit codes) + BB-13 (silent -R)** between them mean: **a shell script using bitbottle cannot trust either the exit code OR the flag behavior.** That's the headline finding of the entire audit. Every other bug is downstream of "scripts can't tell bitbottle is doing the wrong thing." Until these two are fixed, the other 12 bugs are individually invisible to automation.

### Revised fix order

1. **BB-13** (`-R` flag) — biggest user-facing UX bug. Either remove from PersistentFlags or wire it into every command's positional resolution. ~1 day to do properly across all commands.
2. **BB-12** (exit codes) — audit every `cmd/*/*.go` for `fmt.Println(err); return nil`. Should be `return err`. ~1 day.
3. **BB-17** (`--jq` consistency) — central formatter check. ~half day.
4. **BB-09** (commit comment endpoint) — 4 lines + a regression test. ~half hour.
5. **BB-20 + BB-07** (PR state-machine) — one `validateMutablePRState()` helper, wire into 4 commands. ~half day.
6. **BB-08** (workspace perms endpoint) — 3 lines + test.
7. **BB-15 + BB-16** (`--start-at` UX) — one design decision (positional vs flag), apply to both. ~half day.
8. **BB-10** (pipeline trigger struct) — generator fix or manual struct correction. ~half day.
9. **BB-11 + BB-21** (state validation) — one validation helper. ~hour.
10. **BB-14** (3-part ref on repo view) — share the parser used by `branch list`. ~hour.
11. **BB-18** (template newline) — 1 line.
12. **BB-19** (hint accuracy) — pick less misleading wording. ~10 min.

**Total estimate: 4–5 days of focused fix work** to close every Tier-A bug in this report.

### What's still UNTESTED (honest gaps)

The depth pass focused on **negative-input + state-machine + output-format** on the most-used bitbottle commands. The following high-yield surfaces remain untested:

- **MCP server (254 tools)** — zero tested. Likely 10–30 more bugs here. This is bitbottle's biggest differentiator and biggest blind spot.
- **Bitbucket Server / DC** — only Cloud tested. Cause+Hint format pitch implies DC has equivalent paths; not verified.
- **Concurrency / large data** — no tests with >100 PRs, no two-writers race, no pagination boundary tests.
- **Auth edge cases beyond happy path** — token revocation mid-session, expired token, wrong scope.
- **Shell completion** — 4 shells (bash/zsh/fish/powershell) advertised; none verified to actually run.
- **Help-text examples** — `--help` output for each command contains usage examples; none have been executed as-is to verify they work.
- **Output-format full matrix** — only spot-checked json/yaml/template/jq on `pr view` and `pr list`; not done for all commands.

Conservative estimate for **bugs likely to still exist in untested surfaces: 20–40**. The 14 found so far represent breadth + targeted depth on most-used commands. A full audit would likely double this number, with the MCP server alone being the single highest-yield surface.

---

## 8. Methodology audit (what to fix next time)

The sweep had structural problems independent of the per-tool findings:

1. **No contract written before execution.** Tests were "run command, eyeball output." Should have been a pre-declared `{cmd, expected_exit, expected_stdout_regex, expected_api_state}` for each UC.
2. **Trusted CLI's own success messages.** BB-07, BB-09, BB-10, BB-11 all share "CLI said OK, was lying" as the failure mode. Without side-effect verification (re-GET the resource after every mutation), these bugs would only appear by accident.
3. **No raw-API ground truth.** Endpoint-path bugs (BB-08, BB-09) only surface when you compare the CLI's claimed action against the actual REST surface. A contract test layer should hit the API directly first to establish ground truth.
4. **Asymmetric per-tool sweeps.** Each tool was swept end-to-end before moving to the next. When bb's `--close-source-branch` accidentally deleted a seed branch mid-sweep, downstream tool runs ran against different seed state.
5. **No exit-code contract.** BB-12 was missed because every test ran "exit code is whatever, look at the output." A contract harness would have caught it on the first failing command.
6. **Cleanup never happened.** Marked "completed" without verification — the 11 PRs still in the repo prove it.

For the next CLI comparison: build a YAML test matrix with `setup → cmd → assert_exit → assert_stdout → assert_api → teardown` per row, runner executes per-row with seed reset, output matrix drives the report. BB-12 would have been the first finding from such a harness, not the last.

---

## 9. Phase 3 — MCP server sweep (2026-05-27, same session)

**Scope.** Drive `bitbottle mcp serve` over stdio JSON-RPC. 20-tool sample selected to mirror the CLI bug classes (PR state machine, formatters, references, host detection, error envelopes). Total of **38 negative-input cases** plus 2 catalog probes plus 4 live calls. **No working sacrificial repo was reachable** at the start (workspace `proggarapsody_main` not guessed from username slug `proggarapsody`, and the workspace-discovery surface is itself broken — see MCP-01/02/03), so live state-machine tests were deferred by user request. All findings below come from catalog inspection + negative-input matrix + source-code verification.

**Test artefacts.** `/Users/aleksey/cli-comparison-phase3-mcp/` — `driver.py` (stdio JSON-RPC client), `cases-*.json` (test matrices), `out-*.jsonl` (full request/response log), `tools-list.jsonl` (254-tool catalog), `tool-schemas.txt`, `buckets.txt`.

### Verified MCP findings

| ID | Sev | Title | Evidence |
|---|---|---|---|
| **MCP-01** | **P0** | `workspace list` returns HTTP 410 — endpoint `/2.0/workspaces` deprecated by Atlassian (CHANGE-2770). | `bitbottle workspace list` reproduces; root cause `api/cloud/workspaces.go:26`. |
| **MCP-02** | **P0** | `workspace search` returns HTTP 410 — same endpoint, same deprecation. | Reproduces; root cause `api/cloud/workspaces.go:60`. |
| **MCP-03** | **P0** | **No discovery path.** With MCP-01+02 broken, a fresh user has no way through bitbottle to find their workspace slug — making every other Cloud command unusable without out-of-band knowledge. | Demonstrated live: I (with valid auth) couldn't find the workspace name for my own account from inside bitbottle. |
| **MCP-04** | P1 | **Repo-arg schema is inconsistent across MCP tools.** Three competing shapes for the same concept: `{project, slug}` (15+ tools), `{repo}` (compare_refs, list_pr_commits, list_pr_files), `{project, repo}` (get/set_repo_pr_settings). | `tools-list.jsonl` per-tool inputSchema. |
| **MCP-05** | P1 | **Unknown hostname silently defaults to Server URL path.** `get_repo` with `hostname=not-a-real-host.example` builds `https://.../rest/api/1.0/projects/x/repos/y` (Server route) rather than rejecting unknown host. Misleading error on typos. | `out-negative.jsonl` case `07_get_repo_wrong_hostname`. |
| **MCP-06** | P1 | **Wrong-type input on `id` reported as "missing required parameter"** instead of "id must be integer". `id: "abc"` → "missing required parameter: id". | Case `10_get_pr_string_id`. |
| **MCP-07** | P1 | **`id: 0` falsely reported as "missing"** — Go zero-value confusion in MCP arg unmarshalling. Affects every numeric-id tool. | Cases `12_get_pr_zero_id`, `21_decline_pr_zero_id`. |
| **MCP-08** | P2 | **Negative `id` (e.g., `-1`) passes client-side validation** and hits HTTP, returning generic 404. | Case `11_get_pr_negative_id`. |
| **MCP-09** | P2 | **`merge_pr` strategy enum lists empty string `""` as a valid value**, surfaced in error messages: `must be one of , merge, squash, …` (note the bare comma). | Case `22_merge_pr_bad_strategy`. |
| **MCP-10** | P1 | **`add_pr_comment` inline-anchor validation is asymmetric.** `inline_path` without `inline_line` → caught client-side. `inline_line` without `inline_path` → NOT caught, hits API. | Cases `31`, `32`. |
| **MCP-11** | P2 | **`add_commit_comment` doesn't validate `hash` client-side** (length, hex). "a" or "NOT_HEX_!@#" forwarded to API; generic 404 returned. | Cases `41`, `42`. |
| **MCP-12** | P2 | **`create_branch` accepts malformed branch names** ("/", trailing slashes) and forwards to API instead of rejecting at the boundary. | Case `53_create_branch_slash_only`. |
| **MCP-13** | P2 | **`update_pr` with neither `title` nor `body` hits API** instead of returning a clean "nothing to update". | Case `70_update_pr_no_fields`. |
| **MCP-14** | P2 | **`compare_refs.repo` validator is asymmetric** — rejects 1-segment ("only-one-segment" → clean error) but silently accepts 3-segment ("bitbucket.org/proj/repo" → API call). | Cases `80`, `81`. |
| **MCP-15** | P2 | **`user view --json` projects DTO to `{name, slug}` only**, dropping account_id/uuid/links/created_on that `/2.0/user` returns. AI clients reading by-name lose stable identifiers. | Live CLI: `bitbottle user view --json` → `[{"name":...,"slug":...}]`. |
| **MCP-16** | P2 | **Server-only tools are advertised on Cloud (and likely vice-versa) with the host-gating only in description prose, not in structured metadata.** `set_repo_pr_settings` is registered for both backends; on Cloud it always returns `host.unsupported`. An AI client picking tools by name + first sentence has no machine-readable signal. | Cases `95–97`; source `pkg/cmd/mcp/tools_repo_pr_settings.go:30`. |

### What's good (Tier A positive observations)

1. **Empty-string and missing required strings are caught uniformly** with `missing required parameter: X` (cases 01–04, 30, 40, 51, 61, 91, 92). Schema enforcement is consistent for the simple case.
2. **`compare_refs` repo-format check for 1-segment input is correct** ("repo must be in WORKSPACE/REPO format, got 'only-one-segment'").
3. **Enum validators for `inline_side` and `merge_pr.strategy` exist and fire correctly** — the bugs are about *what's in the enum* (MCP-09) and *which fields are validated* (MCP-10), not whether enum validation works at all.
4. **Multi-host detection works** — case 08 cleanly rejects with "multiple hosts configured; specify hostname".
5. **Server-side error envelopes are typed and structured** — `{"code":"pr.not_found","host":"api.bitbucket.org","resource":"pull-request"}`. Better than most CLIs we tested.

### Phase 3 verdict

**16 new findings**, of which **3 are P0** (CHANGE-2770 deprecation cluster — onboarding-blocking), **5 are P1** (input contract, host detection, schema consistency), and **8 are P2** (validation polish).

**Most important takeaway:** the MCP surface inherits **all** the CLI bug classes plus three new ones unique to MCP:
- *Schema inconsistency across tools* (MCP-04) — agents have to learn three different argument conventions.
- *Lost host-gating signal* (MCP-16) — tools that always fail on the current host should be filtered from `tools/list`, not just annotated in prose.
- *Onboarding cliff* (MCP-01/02/03) — equally affects CLI and MCP, but feels worse over MCP because there's no `--help` shoreline to fall back to.

### Combined report card

**Total verified bugs across the audit: 30** (14 CLI + 16 MCP). All categorized. All evidence-cited.

