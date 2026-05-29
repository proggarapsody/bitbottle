# bitbottle — Shipped Scopes

> **Append-only record of shipped backlog scopes.** When a scope's `feat:` commit lands on `main`, its row is **moved** from [`BACKLOG.md`](BACKLOG.md) into this file (not flipped in place). See [`docs/workflows/iteration-cycle/quickref.md`](../workflows/iteration-cycle/quickref.md) §"Definition of Done" for the convention and [`docs/workflows/iteration-cycle/README.md`](../workflows/iteration-cycle/README.md) §4 for the iteration-cycle step.

**Layout:** newest scopes first. Each entry carries the ship date, the feat commit, and (if assigned by then) the release-please version. Detail subsections are preserved verbatim from BACKLOG.md so search history stays intact.

**Why a separate file:**
- BACKLOG.md is for **upcoming work** — keeping shipped scopes inline made it grow past 4000 lines and buried the queue under archaeology.
- This file is **not** CHANGELOG.md (release-please owns that, formatted per Conventional Commits). SHIPPED.md tracks **scope-level intent**: what was queued, what shipped, why. CHANGELOG.md tracks **per-PR semver-bumping notes** for end users.
- Append-only means git history is the audit log; no edits-in-place.

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
