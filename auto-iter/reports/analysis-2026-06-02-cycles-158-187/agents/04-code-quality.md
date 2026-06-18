# 04 — Code Quality Analysis: Cycles 158–187

> Analyst: code-quality agent · Dataset: `dataset.json` + git log v1.119.0..HEAD + `stream-2026-06-01-cycles-178-187.md`

---

## TL;DR

Six DJ BLOCKERs escaped to main in cycles 178–187, all confirmed real (0 false positives). The defect classes are: missing test file (×2), API-default mismatch (×1), duplicated abstraction (×1), wrong format helper (×1), zero-value flag bug (×1). Core repo conventions — `paging.Collect`, typed errors, ContentTypePolicy, conventional commits — are consistently followed in the new feature code. The one structural gap that predates this window and persists through it is systemic under-coverage of tier-2 integration tests: 42 command packages exist, 8 have `*_integration_test.go` files (19%). Cycles 158–177 added ~20 new command packages and none got tier-2 tests. DJ only started flagging this in cycles 185 and 187 because those packages (`repo/sync`, `commit`) already had integration test neighbours in the same directory tree.

---

## Escaped-defect catalogue

All 6 BLOCKERs were caught by DJ *after* auto-merge (post-merge race, not pre-merge miss).

| Cycle | Scope | Defect | Class | Root cause | Fix commit |
|---|---|---|---|---|---|
| 180 | REPO-HOOK-SCRIPTS | `handlers_repohook_test.go` not created | Missing test — MCP triplet incomplete | TDD subagent wrote `handlers_repohook.go` + `tools_repohook.go` but omitted the `*_test.go` counterpart required by the MCP handler triplet convention | `0a3b2e9` (+142 lines) |
| 183 | DEPLOY-KEY-PERMISSION | `add_deploy_key` MCP tool defaulted `permission=""` while CLI `--permission` defaulted to `"read"` | API-default mismatch | `GetString("permission", "")` in the handler used empty string; the CLI flag used `"read"` as the default. The two surfaces became inconsistent | `a1c6116` (+46 lines) |
| 184 | REPO-PIPELINE-VAR-VIEW | `variableView` handler inlined a 3-case scope switch identical to the logic in `shared.ResolveVariableOps` / `ops.GetVariableByKey`; also `skills/references/variable.md` not updated with the new `view` subcommand | Duplicated logic + missing doc | TDD subagent implemented scope dispatch directly in the MCP handler instead of delegating to the existing shared ops layer. The reference doc omission was a separate checklist miss | `f7b2ba6` (−70/+41 LOC net; 4 files) |
| 185 | REPO-SYNC | Custom `isJSONRequested(cmd)` function used instead of standard `format.ConfigFromCmd(cmd)` + `format.RegisterOutputFlags(cmd)`; `sync_integration_test.go` absent | Wrong abstraction + missing test | The subagent wrote a 14-line private helper that re-implements flag lookup by name. The format package already provides this as a one-liner, also wires `--json/--yaml/--jq/--template` as a bundle | `e01a4c3` (+138 lines; replaces the helper, adds 4-test integration suite) |
| 186 | ADMIN-RATE-LIMIT | `setRun` guarded `RequestsPerHour != 0` and `ThrottleWaitMS != 0` to decide whether to apply a partial update; setting either flag to `0` was silently ignored | Zero-value flag bug | The common `flag.Changed()` pattern (already used for `--enabled` via `EnabledSet`) was not extended to the two `int` flags; their zero value was the guard instead of `Flags().Changed()` | `b3ca126` (+50 lines; adds `RequestsPerHourSet`/`ThrottleWaitMSSet`) |
| 187 | COMMIT-SEARCH | `search_integration_test.go` absent | Missing test | Same root cause as cycle 185: the TDD subagent produced unit tests (`search_test.go`), API tests, MCP handler tests, and a txtar script but did not produce a tier-2 `pkg/cmd/commit/search_integration_test.go` | `7bfe6b6` (+159 lines) |

---

## Test-discipline findings

### Integration test omissions (confirmed: cycles 185, 187)

DJ flagged two consecutive cycles for missing `*_integration_test.go`. Both fixes required adding the test as a standalone commit (follow-up PR). The ARCHITECTURE.md test-tier table at `docs/ARCHITECTURE.md` defines tier-2 as "one cobra command against an httptest fake … wire compatibility" and the "New feature gate" note reads: "every new user-visible command must add at least one `.txtar` script before merge" — this names txtar (tier 3) as the hard gate, not tier-2. DJ appears to apply a stricter-than-written rule: require tier-2 when the command's package already has integration tests. This is sound engineering judgment (consistency within a package) but it is implicit.

### Systemic under-coverage of tier-2 (cycles 158–177)

Scanning all 20+ feature commits across cycles 158–177 (REPO-DOWNLOAD, MILESTONES, ISSUE-VERSIONS, WORKSPACE-PROJECT-CRUD, MIRROR, WORKSPACE-PERMS, WORKSPACE-PIPELINE-VAR, REPO-CLONE, PR-PARTICIPANT, ISSUE-ACTIVITY, BRANCH-RULE, WORKSPACE-PROJECT-PERMS, WORKSPACE-SEARCH, PR-MERGE-PREVIEW, SNIPPET-COMMENTS, PIPELINE-OIDC, HOST-INFO, REPO-CLONE, CLOUD-DISCOVERY, SCRIPT-TRUST, FMT-CONTRACT, MCP-INPUT-VALIDATION, MCP-TAXONOMY, PR-GUARDS, CLOUD-WIRE):

- **Integration tests added**: 0 in cycles 158–177 feature commits
- **Txtar (tier 3) added**: 1–4 per cycle, always present
- **Packages with tier-2 today**: `pkg/cmd/{pr,pr/approve,repo/create,repo/delete,repo/list,repo/view,repo/sync,commit}` = 8 of 42 (19%)
- **DJ did not flag these** because none of those new packages already had integration tests to be consistent with

The gap is real but DJ's lack of flags here is internally consistent: it flags regression within a package but does not retroactively impose tier-2 on brand-new packages.

### MCP handler triplet adherence

Pre-analysis-window: the 8d1f3a4 refactor split a monolith `handlers.go` into ~18 per-feature files; none of these legacy split files have `_test.go` counterparts (handlers_branch, handlers_pr, handlers_issue, handlers_webhook, handlers_repo, handlers_tag, handlers_commit, etc. — 17 handler files without tests).

Post-refactor new handlers (cycles 158+): all new handler files created during this window have tests **except** `handlers_repohook.go` (cycle 180, caught and fixed). The pattern is strongly established for new features; the legacy debt is inherited.

---

## Convention-adherence spot-checks

### 1. `paging.Collect[T]` for list operations (PASS)

Every new list operation added in cycles 158–187 uses the canonical helper:

```
api/server/repo_hook.go    — paging.Collect (REPO-HOOK-SCRIPTS, cycle 180)
api/cloud/commit_search.go — paging.Collect with opts.Limit cap (COMMIT-SEARCH, cycle 187)
api/cloud/pipeline_variables.go — paging.Collect (REPO-PIPELINE-VAR-VIEW, cycle 184)
api/cloud/repo_sync.go     — sync is a mutation op; no list involved (REPO-SYNC, cycle 185)
api/server/deploy_keys.go  — paging.Collect (DEPLOY-KEY-PERMISSION, cycle 183)
```

No hand-rolled pagination loop was found in any new adapter in this window.

### 2. Typed errors (`api/backend/errors.go`) (MOSTLY PASS, one escaped miss)

New API adapters consistently rely on transport-level error wrapping (`UseDomainErrors`) rather than constructing raw `fmt.Errorf`. The one case that escaped: `pkg/cmd/variable/shared/ops.go` (pre-cycle-184) used bare `fmt.Errorf` for three validation paths (`unknown scope`, `missing env_uuid`, `variable not found`). These were corrected by fix commit `f7b2ba6` which wrapped them in `&backend.DomainError{Kind: ErrInvalidRequest/ErrNotFound}`. The fix was correct and added the `Resource`/`ID`/`Message` fields properly.

No `fmt.Errorf` without `%w` wrapping was found in the new `api/cloud/*.go` or `api/server/*.go` files for cycles 178–187.

### 3. `format.ConfigFromCmd` vs custom JSON detection (BLOCKER caught cycle 185)

`pkg/cmd/repo/sync/sync.go` (commit `d839aeb`) shipped with a 14-line private function:

```go
// isJSONRequested returns true when the --json flag is present on the command
// or any of its ancestors (persistent flag).
func isJSONRequested(cmd *cobra.Command) bool {
    f := cmd.Flags().Lookup("json")
    if f == nil {
        f = cmd.InheritedFlags().Lookup("json")
    }
    if f == nil {
        return false
    }
    return f.Changed
}
```

This re-implements flag-name lookup and misses `--yaml`, `--jq`, `--template` (the full output flag bundle). The standard pattern — `format.RegisterOutputFlags(cmd)` in the builder, `format.ConfigFromCmd(cmd)` in RunE — is used in 30+ commands across the codebase. Fix `e01a4c3` replaced the helper and wired all flags properly.

This is a clear convention breach: a new private function substituting a standard helper rather than an intentional extension.

### 4. `flag.Changed()` for zero-value flags (BLOCKER caught cycle 186)

`pkg/cmd/admin/ratelimit/set/set.go` (commit `7a92e26`) used `!= 0` guards for `RequestsPerHour` and `ThrottleWaitMS`:

```go
if opts.RequestsPerHour != 0 {
    in.RequestsPerHour = opts.RequestsPerHour
}
```

The existing pattern in the same `Options` struct for `--enabled` used `EnabledSet bool` populated from `cmd.Flags().Changed("enabled")`. The inconsistency was present within the same file. Fix `b3ca126` extended the `*Set` pattern to both int flags, making `--requests-per-hour=0` and `--throttle-wait-ms=0` functional as intended.

Root cause: the TDD subagent correctly applied `flag.Changed()` for the bool flag (which has an obvious true/false zero-value problem) but did not notice the same issue applies to int flags with a documented `0` as a valid user intent.

### 5. MCP/CLI default parity (BLOCKER caught cycle 183)

`pkg/cmd/mcp/handlers_deploykey.go` (commit `2bebaf1`) used `GetString("permission", "")` while the CLI flag in `pkg/cmd/deploykey/add.go` used `StringVar(..., "read", ...)`. Result: MCP callers who omit `permission` get `""` (passed through to API) while CLI users get `"read"`. The fix `a1c6116` changed the default to `"read"` and added explicit validation rejecting unknown values.

This is a pattern-level failure: when a new MCP tool wraps an existing CLI flag, the tool's default must match the CLI's flag default. The subagent wired the mapping logic but did not copy the default value.

### 6. ContentTypePolicy and alt-transport wiring (PASS)

`newAltTransport` in `api/server/client.go` correctly applies `httpx.ContentTypeAlwaysWrite` to all Server/DC alternative transports including the new `cherryPickHTTP` (cycle 154), `sshHTTP`, and `mirrorHTTP`. No new transport was wired with the wrong policy in this window.

---

## Refactor / cleanliness observations

### Handler duplication (cycle 184)

The `variableView` handler in `handlers_variable.go` before fix `f7b2ba6` contained ~70 lines implementing a 3-case scope switch (`repository`, `workspace`, `deployment`) with a linear scan for the matching key in each case. The `shared/ops.go` package already exposed `ResolveVariableOps()` which returns a unified `VariableOps` interface, and a `GetVariableByKey(key)` operation was added as part of the same cycle. The handler could have called `ops.GetVariableByKey(key)` directly in ~10 lines. The inline duplication was a classic "implement in the handler first" shortcut rather than delegating to the shared layer.

### Zero-delta test for integration test (cycle 187)

The standalone fix commit `7bfe6b6` added `search_integration_test.go` at 159 lines — no other changes. This is a clean, minimal fix. The pattern of "add the test as a standalone PR" is functional but wastes a release bump. The correct outcome would have been: test present in the initial feat commit `336d4a4`.

---

## Net quality verdict

**The at-merge quality is marginal but DJ is carrying an increasing load.**

Quantitative summary for the 178–187 stream (10 feature cycles):
- 5 of 10 cycles had post-merge DJ BLOCKERs (50% BLOCKER rate)
- 6 individual defects in 5 cycles
- 0 false positives from DJ (100% precision)
- All defects were medium-severity (not data-loss or security issues)

The code that *does* merge is structurally sound: paging, typed errors, transport policy, conventional commits all hold. What escapes are implementation-discipline misses: forgetting a file, using a private helper instead of the standard one, not fully applying an existing pattern. These are the kind of mistakes that a focused code review catches immediately — and that is exactly what DJ is doing.

**The concern is trend, not absolute level.** The BLOCKER rate went from 0% (cycles 153–167, all pre-stream) to 20% (168–177) to 50% (178–187). If the rate continues to climb, DJ becomes a mandatory remediation loop rather than a quality gate. Fix 1 from the stream report (arm auto-merge only after DJ returns SHIP) eliminates the structural race. Fix 2 (explicit integration test requirement in TDD prompt) addresses the most recurring pattern.

The legacy gap — 34 command packages with no tier-2 integration tests — is not a new regression from this window; it predates the analysis range. It is worth scheduling a dedicated "test-coverage uplift" iteration but should not be attributed to quality decline in cycles 158–187.

---

## Recommendations

1. **Gate auto-merge on DJ verdict** (unrelated to code quality per se, but structurally prevents all 6 of these defects from reaching main). File paths: `docs/workflows/iteration-cycle/README.md` §3.6, `.claude/commands/auto-iter.md`.

2. **Extend TDD subagent prompt with explicit integration test checklist item**:
   > "If adding a command in a package that already contains `*_integration_test.go` files, create `<command>_integration_test.go` using the httptest pattern. See `pkg/cmd/repo/sync/sync_integration_test.go`."
   
   File: `.claude/commands/auto-iter.md`.

3. **Add MCP/CLI default parity rule to the MCP handler checklist**:
   > "For every optional MCP parameter that wraps a CLI flag, verify `GetString("param", "CLI_DEFAULT")` uses the *exact* default from the CLI flag definition."

4. **Add `flag.Changed()` requirement for any int/string flag that accepts 0/empty as a valid user intent**. Can be a comment in the template for new `set` commands.

5. **Consider a lint rule (or smell-scan entry)** to detect inline scope switches that shadow logic already in `shared/ops.go`. The variableView duplication (cycle 184) is a prototype for a pattern that could recur as more composite operations are added.

---

## Commits inspected

| SHA | Cycle | Purpose |
|---|---|---|
| `6a8319d` | 180 | REPO-HOOK-SCRIPTS feat |
| `0a3b2e9` | 180 | Missing handler test fix |
| `2bebaf1` | 183 | DEPLOY-KEY-PERMISSION feat |
| `a1c6116` | 183 | Permission default fix |
| `4fb66e8` | 184 | REPO-PIPELINE-VAR-VIEW feat |
| `f7b2ba6` | 184 | Handler refactor + typed errors fix |
| `d839aeb` | 185 | REPO-SYNC feat |
| `e01a4c3` | 185 | Format pattern + integration test fix |
| `7a92e26` | 186 | ADMIN-RATE-LIMIT feat |
| `b3ca126` | 186 | Zero-value flag.Changed() fix |
| `336d4a4` | 187 | COMMIT-SEARCH feat |
| `7bfe6b6` | 187 | Missing integration test fix |
| `5493bac` | 156 | PIPE-CONFIG+SSH-KEY-SERVER (earlier era spot-check) |
| `bc0fc21` | 158 | REPO-DOWNLOAD+MILESTONES (tier-2 gap baseline) |
| `44f3f3b` | 164 | REPO-CLONE (tier-2 gap) |
| `bba7c42` | 181 | CLOUD-CODE-INSIGHTS (clean cycle, no BLOCKER) |
| `8d1f3a4` | pre-window | MCP handler refactor (legacy untested files origin) |
