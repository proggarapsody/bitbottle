# bitbottle — Shipped Scopes

> **Append-only record of shipped backlog scopes.** When a scope's `feat:` commit lands on `main`, its row is **moved** from [`BACKLOG.md`](BACKLOG.md) into this file (not flipped in place). See [`docs/workflows/iteration-cycle/quickref.md`](../workflows/iteration-cycle/quickref.md) §"Definition of Done" for the convention and [`docs/workflows/iteration-cycle/README.md`](../workflows/iteration-cycle/README.md) §4 for the iteration-cycle step.

**Layout:** newest scopes first. Each entry carries the ship date, the feat commit, and (if assigned by then) the release-please version. Detail subsections are preserved verbatim from BACKLOG.md so search history stays intact.

**Why a separate file:**
- BACKLOG.md is for **upcoming work** — keeping shipped scopes inline made it grow past 4000 lines and buried the queue under archaeology.
- This file is **not** CHANGELOG.md (release-please owns that, formatted per Conventional Commits). SHIPPED.md tracks **scope-level intent**: what was queued, what shipped, why. CHANGELOG.md tracks **per-PR semver-bumping notes** for end users.
- Append-only means git history is the audit log; no edits-in-place.

---

## 2026-06-02 — REF-UX — branch/tag create --start-at positional (Option A)

`branch create` and `tag create` now accept START_AT as a 3rd positional argument
(`[PROJECT/REPO] NAME [START_AT]`). The `--start-at` flag remains for backward
compatibility. Closes #621.

---

## 2026-05-31 — CLOUD-WIRE

### CLOUD-WIRE — Cloud API path + response-struct drift (BB-08, BB-09, BB-10)

- **Fix commit:** `fix(cloud): CLOUD-WIRE — /permissions-config/ path, /commit/ singular, pipeline trigger response struct` (2026-05-31)
- **Backends:** Cloud only
- **Estimate (planned):** 1.5 days

**Status:** ✅ — sourced from the 2026-05-27 CLI-comparison audit.

**Why P1:** Three Cloud-only wire-level defects. (1) `workspace project perms list` calls `/permissions/users` but the correct path is `/permissions-config/users` — 404. (2) `commit comment {list,add,edit,delete}` all hardcode `/commits/{hash}/comments` (plural) but Bitbucket Cloud uses `/commit/{hash}/comments` (singular) — 404 on all four operations. (3) `pipeline trigger`'s response struct types `CloudTriggerResponseLinks.links.self` as `[]CloudTriggerResponseSelfLink` (slice) but the API returns a single object — JSON unmarshal failure, no useful output (and exit 0, see BB-12).

**Shape:**

1. **Endpoint fixes** — `api/cloud/workspace_project_perms.go`: change `/permissions/users` → `/permissions-config/users`. `api/cloud/commit_comment.go` lines 32, 52, 65, 74: change `/commits/{hash}/comments` → `/commit/{hash}/comments`. Both are ~4 lines + regression tests against captured fixtures.
2. **Pipeline trigger struct** — fix the generator (or hand-correct the type) so `links.self` is `*CloudTriggerResponseSelfLink` (single, nullable). Add a unit test loading a captured pipeline-trigger response fixture and asserting unmarshal success.
3. **Add live-wire (tier 6) tests** for each of the three fixed endpoints to prevent silent re-drift if Atlassian moves the paths again.

**Definition of Done:**

- [x] `api/cloud/workspace_project_perms.go` — path corrected to `/permissions-config/`.
- [x] `api/cloud/commit_comment.go` — all four operations use `/commit/` (singular).
- [x] `api/cloud/gen/openapi.yaml` + `types.go` regenerated — `links.self` typed as `*CloudTriggerResponseSelfLink` + unmarshal regression test green.
- [x] `test/script/script_test.go` — bb-fake Cloud stubs updated to `/permissions-config/` paths.
- [ ] `test/testdata/fixtures/pipeline_trigger_response.json` — captured live response (deferred: requires live Cloud sandbox).
- [ ] `.txtar` scripts: `commit_comment_lifecycle.txtar`, `pipeline_trigger.txtar` (deferred: no bb-fake stubs for these endpoints yet).
- [ ] Nightly `BITBOTTLE_E2E=1` runs against real Cloud sandbox (deferred: infra not yet wired).

---

## 2026-05-31 — PR-GUARDS

### PR-GUARDS — PR state-machine + `--state` enum validation (BB-07, BB-11, BB-20, BB-21)

- **Fix commit:** `fix(pr): PR-GUARDS — state-machine guard + --state enum validation` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 1.5 days

**Status:** ✅ — sourced from the 2026-05-27 CLI-comparison audit.

**Why P1:** Two related defect classes. (1) `pr approve` succeeds with exit 0 + "Approved pull request #N" on a DECLINED PR — because the Bitbucket Cloud API returns 200 for participant approval regardless of PR state. `pr decline` and `pr merge` correctly reject DECLINED/MERGED PRs (because the API returns 400 there). So bitbottle is "right by accident" three times, and wrong on the one path the API is permissive on. (2) `pr list --state INVALID_STATE` (and `--state ""`) silently returns all 11 PRs across MERGED/OPEN/DECLINED — no client-side validation.

**Shape:**

1. **`ValidateMutablePRState(pr) error`** in `api/backend/pr_state.go` — returns typed `*DomainError{Kind: ErrConflict}` if state ∈ {DECLINED, MERGED, SUPERSEDED}. Called by `pr approve`, `pr unapprove`, `pr request-changes`, `pr edit`, before the mutation request.
2. **`--state` enum validation** via `pkg/cmd/internal/enumflag/` `pflag.Value` helper. Rejects `""`, `INVALID_STATE`, and any off-enum value at parse time.

**Definition of Done (as shipped):**

- [x] `api/backend/pr_state.go` — `ValidateMutablePRState` helper + tests.
- [x] `pkg/cmd/pr/{approve,unapprove,request-changes,edit}` — call the guard before the API request.
- [x] `pkg/cmd/internal/enumflag/enumflag.go` — generic enum flag helper + tests.
- [x] `pkg/cmd/pr/list.go` — `--state` uses enum flag, rejects invalid values.
- [x] `.txtar` scripts: `pr_approve_on_declined.txtar`, `pr_list_state_invalid.txtar`.
- [x] `pr approve N` on DECLINED PR exits 1 with typed error.

---

## 2026-05-31 — MCP-TAXONOMY

### MCP-TAXONOMY — Unify tool catalog (MCP-04, MCP-05, MCP-16)

- **Fix commit:** `fix(mcp): unify MCP tool catalog (MCP-04, MCP-05, MCP-16)` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 1.5 days

Collapsed three structural inconsistencies in the 254-tool MCP catalog that
forced AI clients to special-case bitbottle:

- **MCP-04 — canonical `{project, slug}` repo-arg shape.** Migrated the five
  BACKLOG-named alternates — `compare_refs`, `list_pr_commits`,
  `list_pr_files` (were `{repo}` = `WORKSPACE/REPO`), and
  `get_repo_pr_settings` / `set_repo_pr_settings` (were `{project, repo}`) —
  to `{project, slug}`. The old shape still works for one release; the tool
  result prepends a `DEPRECATION` text block. Shared resolver helpers live in
  `pkg/cmd/mcp/repo_arg.go` (`repoFromProjectSlugOrRepo`,
  `repoFromProjectSlugOrProjectRepo`, `withDeprecation`). Many other tools
  intentionally keep `{repo}` — the regression test is an allowlist of the
  five migrated tools, not a global ban.
- **MCP-05 — reject unknown hostnames.** `handlers.resolveBackend` now calls
  `requireConfiguredHost`, which returns a typed
  `*DomainError{Kind: ErrUnknownHost, Code: host.unknown}` listing the
  configured hosts when a non-empty `hostname` isn't in `hosts.yml` — before
  any HTTP or Server-vs-Cloud URL inference. Added `ErrUnknownHost` kind +
  `CodeHostUnknown` code (wired through `AllCodes` + the errfmt catalogue).
- **MCP-16 — `_meta.backends` + pre-HTTP gating.** A small registration
  wrapper (`addGatedTool` in `pkg/cmd/mcp/tools_backends.go`) stamps
  `_meta.backends` (`["server"]` / `["cloud"]` / both) onto a tool's
  `tools/list` entry and wraps its handler with a pre-HTTP gate: a
  single-backend tool invoked against the wrong flavour returns
  `host.unsupported` (with the allowed list) before the backend is dialed.
  Scoped to backend-specific tools — the five migrated tools were rewired
  through the wrapper; the broader catalog defaults to both backends (no
  gating), avoiding 250+ hand-edits. The optional Phase-2 host-filtered
  `tools/list` (`BITBOTTLE_MCP_HOST_FILTER=1`) was skipped as out-of-scope
  for the diff budget.

## 2026-05-31 — MCP-INPUT-VALIDATION

### MCP-INPUT-VALIDATION — Tighten client-side validators across MCP tools (MCP-06 through MCP-14)

- **Fix commit:** `fix(mcp): tighten client-side arg validators (MCP-06..MCP-14)` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 2.5 days

Shipped a shared `pkg/cmd/mcp/argval` package of typed extractors —
`Int` (with `Required`/`Min`/`Max`, treats explicit 0 as present, rejects
wrong-type and out-of-range), `Hash(minLen=7)`, `RefName`, `EnumOneOf`
(panics on an empty member so `""` can never re-enter an enum),
`MutuallyRequired`, and `OneOfRequired` — each returning a structured
`{code, field, got, message}` envelope (`arg.missing` / `arg.invalid_type`
/ `arg.out_of_range` / `arg.invalid_value`). All numeric-id handlers across
the MCP surface were migrated to a shared `requireIntArg` helper, and the
six affected tools (`merge_pr`, `add_pr_comment`, `add_commit_comment`,
`create_branch`, `update_pr`, `compare_refs`) now reject malformed input
client-side instead of forwarding it to a generic upstream 404. The
`merge_pr` `strategy` schema gained an explicit `Enum(merge, squash,
rebase)`. A 14-case acceptance matrix (`handlers_input_validation_test.go`)
replays each of MCP-06 … MCP-14 against the structured envelope, alongside
the `argval` package's own unit tests.

Note vs. the original DoD: the scope was delivered as a `fix(mcp):` commit
(behavioural correctness fixes, not a new feature). The structured
envelope and 14-case Go matrix cover the acceptance criteria in lieu of a
`.txtar` corpus; migration touched the affected + mechanically-safe
numeric-id handlers rather than all 254 (the PRD explicitly de-scoped a
full sweep to avoid ballooning the diff).

### MCP-INPUT-VALIDATION — original backlog detail

**Status:** ✅ — sourced from the 2026-05-27 MCP sweep (Phase 3).

**Why P1:** MCP arg validation is inconsistent — some fields validate beautifully (`inline_side` enum, 1-segment repo on `compare_refs`, missing-required-string), but adjacent fields on the same tools forward garbage to the HTTP layer and return a generic upstream 404. AI agents reading the error can't tell whether their input was malformed or the resource doesn't exist. Concretely, the negative-input matrix surfaced **nine** distinct gaps:

- **MCP-06** Wrong-type `id` → reported as "missing required parameter: id" instead of "id must be integer".
- **MCP-07** `id: 0` → falsely "missing" (Go zero-value issue in MCP unmarshal). Hits every numeric-id tool.
- **MCP-08** Negative `id` → passes through, generic 404.
- **MCP-09** `merge_pr.strategy` enum lists `""` as a valid value; error messages show "must be one of , merge, squash, …" (note the bare comma).
- **MCP-10** `add_pr_comment` inline-anchor asymmetry: `inline_path` w/o `inline_line` is caught client-side, but `inline_line` w/o `inline_path` is not.
- **MCP-11** `add_commit_comment.hash` not validated for length/hex; "a" or "NOT_HEX_!@#" reaches Cloud and returns generic 404.
- **MCP-12** `create_branch.name` accepts `"/"`, leading/trailing slashes, and other refs that Git refuses by spec.
- **MCP-13** `update_pr` with neither `title` nor `body` hits the API instead of returning a clean "nothing to update".
- **MCP-14** `compare_refs.repo` rejects 1-segment input cleanly but silently accepts 3-segment (`bitbucket.org/proj/repo`).

These are individually small but collectively they prevent any safe automated retry policy on the MCP surface.

**Shape:**

1. **Shared `argval` helper** in `pkg/cmd/mcp/argval/` — typed extractors: `Int(name, required, min=)`, `Hash(name, minLen=7)`, `RefName(name)`, `EnumOneOf(name, allowed)`, `MutuallyRequired(field, dependency)`, `OneOfRequired(fields)`. Each extractor returns a structured error mapped to MCP's tool-error envelope so the client gets `{code: "arg.invalid_type", field: "id", got: "string"}`-style payloads.
2. **Migrate every handler** in `pkg/cmd/mcp/handlers_*.go` to call the new helpers in their first 3–5 lines.
3. **Fix the strategy enum** for `merge_pr` — drop `""` from the canonical list, treat empty as "use default" via a separate branch.
4. **Sweep all existing tool registrations** for similar zero-value pitfalls — anywhere a number can legally be 0 (limits, pagination), confirm `_, ok := args[name]` is used not `args[name].(int) > 0`.
5. **Tabulate per-handler arg specs** in a generated table — one row per `{tool, field, type, required, validators}`. Becomes the spec source for both runtime check and JSONSchema export.

**Definition of Done (as shipped):**

- [x] `pkg/cmd/mcp/argval/argval.go` lands with typed extractors + tests.
- [x] Affected + mechanically-safe numeric-id handlers migrated to the shared helpers (full 254-handler sweep de-scoped per PRD).
- [x] Test matrix replaying MCP-06 … MCP-14 — every one returns a structured error with the right `code` and `field`.
- [x] `merge_pr` strategy enum no longer lists `""`.
- [x] Go acceptance matrix covers each of MCP-06 through MCP-14 with the corrected error envelope (`.txtar` not required given Go coverage).
- [ ] CHANGELOG line per bug class — deferred to release-please (owns CHANGELOG.md).

---

## 2026-05-31 — FMT-CONTRACT

### FMT-CONTRACT — `--jq` / `--template` / hint consistency (BB-17, BB-18, BB-19, MCP-15)

- **Fix commit:** `fix(fmt): unify --jq guard, template trailing newline, neutral not-found hint, user-view Cloud IDs` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 1.25 days

Output-format middleware silently misbehaved, breaking scripts. Four fixes, one PR:

1. **`--jq` consistency (BB-17)** — root's `PersistentPreRunE` carried the `--jq requires --json` guard plus the json/yaml/template mutual-exclusion, but cobra runs only the single *deepest* `PersistentPreRunE` in a command chain. `factory.EnableRepoOverride` (the `-R` flag wiring) installs its own hook on every command group, which shadowed the root guard — so `pr view --jq .x` / `repo view --jq .x` silently ignored `--jq` and returned the full text view, while `status`/`pr list` correctly errored. Extracted the contract into `cmdutil.ValidateOutputFlags`, called from **both** root's hook and the repo-override hook. Now `--jq` without `--json` errors identically on `pr view`, `repo view`, `pr list`, `branch list`; `--jq` with `--json` applies on all of them.
2. **`--template` trailing newline (BB-18)** — `format.WriteTemplate` now newline-terminates rendered output (text/template appends none), matching `--json`/`--yaml` encoders.
3. **Hint accuracy (BB-19)** — the `pr.not_found` errfmt catalogue hint no longer claims the PR "may have been deleted" (misleading when `pr view 99999` never existed); reworded to the neutral "No pull request #N exists in this repo."
4. **MCP-15** — `backend.User` gained `AccountID`/`UUID`/`CreatedOn`/`Links.HTML.Href` (Cloud-stable identifiers, populated from `GET /user`). `user view --json` exposes them (JSONOnly so the TTY table stays slug+name) and the MCP `get_current_user` tool marshals them, giving AI clients durable machine-readable handles.

A `.txtar` corpus (`fmt_contract_jq_requires_json`, `fmt_contract_jq_applies`, `fmt_contract_template_newline`, `fmt_contract_user_view_fields`) proves all four behaviours end-to-end.

---

## 2026-05-31 — SCRIPT-TRUST

### SCRIPT-TRUST — Exit codes + `-R` flag + ref parser (BB-12, BB-13, BB-14)

- **Fix commits:** `fix(cli): unify repo-ref parser; accept HOST/PROJECT/REPO everywhere` + `fix(cli): wire -R/--repo on every command; add exit-code contract test` (2026-05-31)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 3 days

Scripts could not trust bitbottle's contract surface. Three fixes, one PR:

1. **Exit-code contract** — added `pkg/cmd/contract_test.go`, which walks every leaf cobra command, runs it with intentionally-bad input, and asserts that any command writing to stderr exits non-zero (guards against the `print-then-return-nil` anti-pattern). Leaves that shell out to external processes (`skill`, `extension`, `mcp`, `completion`, `alias`) are skipped — their exit-code behaviour is covered by their own package tests.
2. **`-R`/`--repo` unification** — `-R` was advertised in every command's INHERITED FLAGS but silently ignored on most non-PR commands because their `cobra.ExactArgs(N)` forced a positional repo ref. Loosened the leading repo positional to optional (`ExactArgs(N)` → `MaximumNArgs(1)` / `RangeArgs(N-1, N)`) on ~30 commands across branch, tag, commit, pipeline, codeinsights, webhook, deployment, environment and variable. Each splits args via the new `repoarg.SplitLeadingRepo` and falls back to `factory.ResolveTarget`, so `-R` / `BB_REPO` / pinned defaults now resolve the repo everywhere.
3. **Ref-parser unification** — added `pkg/cmd/internal/repoarg/ref.go` with the single `ParseRef` helper accepting `PROJECT/REPO` and `HOST/PROJECT/REPO`. `factory.ResolveTarget` routes through it, and `repo view` migrated off the 2-part-only `bbrepo.Parse`. Regression: `repo view bitbucket.org/ws/repo` (3-part) is now accepted everywhere.

A `.txtar` corpus (`script_trust_repo_flag.txtar`, `script_trust_error_exit.txtar`) proves `-R` resolves the positional, 3-part refs are accepted, and error paths exit non-zero.

---

## 2026-05-29 — CLOUD-DISCOVERY

### CLOUD-DISCOVERY — Migrate off deprecated `/2.0/workspaces` (MCP-01, MCP-02, MCP-03)

- **Feat commit:** `feat(workspace): migrate workspace list/search to /user/permissions/workspaces` (2026-05-29)
- **Backends:** Cloud
- **Estimate (planned):** 1 day

Atlassian deprecated `GET /2.0/workspaces` under CHANGE-2770; the endpoint returns HTTP 410 Gone. Both `workspace list` and `workspace search` were broken, leaving fresh users with no in-tool way to find their workspace slug. Migration target: `GET /2.0/user/permissions/workspaces` (workspaces now wrapped under `value[].workspace`). Added `cloudWorkspacePermission` DTO, updated both list/search unmarshal paths, fixed all test fixtures. Also added `ErrEndpointDeprecated` / `CodeEndpointDeprecated` to the errors catalogue so future 410s get a clean typed message with an upgrade link.

---

## 2026-05-29 — HOST-INFO

### HOST-INFO — Host Capabilities & Version Info

- **Feat commit:** `feat(host): add host info command (Cloud + Server)` (2026-05-29)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 2 days

Today an agent starting work against a new Bitbucket host has no single endpoint to discover (a) which backend type it is, (b) what version (matters for Server feature gates like `pr unready` 8.0+ requirement), (c) which optional `AsXxxClient` capabilities the host implements. The pieces exist internally — `ServerCapabilities.GetApplicationProperties()`, `api/server/version.go`'s parsed `Version`, and `AllFeatureSpecs` — but no command surfaces them. `context` (✅) answers "where am I?"; `host info` answers "what can I do here?".

**Interface (shipped):**
```go
// api/backend/client_host_info.go
type HostInfoClient interface {
    GetHostInfo() (HostInfo, error)
}

type HostInfo struct {
    BackendType       string   `json:"backend_type"`
    BaseURL           string   `json:"base_url"`
    Version           string   `json:"version,omitempty"`
    BuildNumber       string   `json:"build_number,omitempty"`
    DisplayName       string   `json:"display_name"`
    SupportedFeatures []string `json:"supported_features"`
}
```

**Commands:** `host info [--hostname H] [--json]`

**MCP tools:** `get_host_info`

---

## 2026-05-25 — Bootstrap migration

Three scopes pulled from BACKLOG.md "Up Next" that had shipped before the file split. Each has a corresponding `feat:` commit; ✅ markers in the Full Functionality Map cross-reference these.

### PR-SETTINGS — Per-Repo PR Settings

- **Feat commit:** `0be675b feat(repo): add repo pr-settings get/set commands (Cloud + Server/DC)` (2026-05-25)
- **Backends:** Both (Cloud + Server/DC)
- **Estimate (planned):** 2 days

Bitbucket exposes per-repo pull-request configuration that's distinct from branch protections and from the branching model: required number of approvers, "all approvers must approve", "all tasks must be resolved", required green builds, and which merge strategies are allowed. Before this scope shipped, these could only be set in the web UI (Server) or by hand-rolling a `PUT /repositories/{ws}/{slug}` payload (Cloud). The scope shipped the read/write CLI surface that completes the trio: BRANCH-RULE (Cloud ref restrictions ✅), BRANCH-PROTECT (Server ref restrictions ✅), PR-SETTINGS (PR merge gate, both backends).

**Interface (shipped):**
```go
// api/backend/client_repo_pr_settings.go
type RepoPRSettingsClient interface {
    GetRepoPRSettings(ns, slug string) (RepoPRSettings, error)
    UpdateRepoPRSettings(ns, slug string, in RepoPRSettingsInput) (RepoPRSettings, error)
}

type RepoPRSettings struct {
    RequiredApprovers           int
    RequiredAllApprovers        bool
    RequiredAllTasksComplete    bool
    RequiredSuccessfulBuilds    int
    MergeStrategy               string   // default
    AllowedStrategies           []string // subset of {ff, ff-only, rebase, squash, merge-commit}
}

// Pointer fields → only set the ones the user passed.
type RepoPRSettingsInput struct {
    RequiredApprovers           *int
    RequiredAllApprovers        *bool
    RequiredAllTasksComplete    *bool
    RequiredSuccessfulBuilds    *int
    MergeStrategy               *string
    AllowedStrategies           *[]string
}
```

**Commands shipped:** `bitbottle repo pr-settings get [PROJECT/REPO] [--json]`, `bitbottle repo pr-settings set [PROJECT/REPO] [--required-approvers N] [--required-all-approvers[=false]] [--required-all-tasks-complete[=false]] [--required-successful-builds N] [--merge-strategy STRAT] [--allowed-strategies s1,s2,...]`

**Backends:** Server (`GET /rest/api/1.0/projects/{ns}/repos/{slug}/settings/pull-requests`, `POST` same path with merge config body). Cloud (`GET /repositories/{ws}/{slug}` for read, `PUT /repositories/{ws}/{slug}` for the subset Cloud supports — `fork_policy` + the project-level merge defaults; fields not honoured by Cloud return `ErrUnsupportedOnHost` per-field rather than failing the whole call).

**MCP tools shipped:** `get_repo_pr_settings`, `set_repo_pr_settings`

**Definition of Done (verified):**
- [x] `api/backend/client_repo_pr_settings.go` — interface + `RepoPRSettings` + `RepoPRSettingsInput`
- [x] `api/cloud/repo_pr_settings.go` + `_test.go` (subset support + typed errors on Cloud-unsupported fields)
- [x] `api/server/repo_pr_settings.go` + `_test.go`
- [x] `pkg/cmd/repo/pr-settings/` (get, set) + tests
- [x] MCP pair (`tools.go` + `handlers.go` + `handlers_test.go`)
- [x] `test/script/testdata/repo_pr_settings_*.txtar` for both backends
- [x] `skills/SKILL.md` + `skills/references/repos.md`

### PIPE-RERUN — Pipeline rerun

- **Feat commit:** `cd9cc7c feat(pipeline): add pipeline rerun command (Cloud)` (2026-05-25)
- **Backend:** Cloud only
- **Estimate (planned):** 1 day

Re-trigger a pipeline at the same commit. Cloud-only because Server/DC has no equivalent rerun primitive. Command shipped: `pipeline rerun UUID`. Functionality map row at `docs/backlog/BACKLOG.md` cross-references this scope.

### CHERRY-PICK — Cherry-pick commit onto a branch

- **Feat commit:** `979db8a feat(commit): add commit cherry-pick command (Server/DC)` (2026-05-25)
- **Release tag:** `v1.108.0`
- **Backend:** Server/DC only (branch-utils plugin)
- **Estimate (planned):** 1 day

Cherry-pick an arbitrary commit onto a named branch. Server/DC only — Cloud has no equivalent endpoint. Command shipped: `commit cherry-pick HASH BRANCH`. Functionality map row at `docs/backlog/BACKLOG.md` cross-references this scope.

---

## How to add an entry (for /auto-iter and humans)

When a scope's `feat:` commit lands on `main`:

1. **Same commit** that ships the code: cut its Up-Next table row and its `### SCOPE — Title` detail section from `docs/backlog/BACKLOG.md`.
2. Prepend a new entry to SHIPPED.md under a dated subsection (`## YYYY-MM-DD — <SCOPE>`). Include:
   - `Feat commit:` short SHA + subject + date
   - `Release tag:` if assigned by release-please at write time (else add later when known)
   - `Backend:` Cloud / Server-DC / Both
   - `Estimate (planned):` what was budgeted vs reality (optional callout)
   - The scope's detail section verbatim, with all checkboxes in the Definition of Done flipped to `[x]`.
3. Leave the Functionality Map `✅` markers in `BACKLOG.md` in place — they're cross-references, not status duplicates.
4. **Never** rewrite history in SHIPPED.md. If a shipped scope later breaks, file a fresh BACKLOG scope to fix it; do not edit the SHIPPED entry.
