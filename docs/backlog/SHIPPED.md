# bitbottle — Shipped Scopes

> **Append-only record of shipped backlog scopes.** When a scope's `feat:` commit lands on `main`, its row is **moved** from [`BACKLOG.md`](BACKLOG.md) into this file (not flipped in place). See [`docs/workflows/iteration-cycle/quickref.md`](../workflows/iteration-cycle/quickref.md) §"Definition of Done" for the convention and [`docs/workflows/iteration-cycle/README.md`](../workflows/iteration-cycle/README.md) §4 for the iteration-cycle step.

**Layout:** newest scopes first. Each entry carries the ship date, the feat commit, and (if assigned by then) the release-please version. Detail subsections are preserved verbatim from BACKLOG.md so search history stays intact.

**Why a separate file:**
- BACKLOG.md is for **upcoming work** — keeping shipped scopes inline made it grow past 4000 lines and buried the queue under archaeology.
- This file is **not** CHANGELOG.md (release-please owns that, formatted per Conventional Commits). SHIPPED.md tracks **scope-level intent**: what was queued, what shipped, why. CHANGELOG.md tracks **per-PR semver-bumping notes** for end users.
- Append-only means git history is the audit log; no edits-in-place.

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
