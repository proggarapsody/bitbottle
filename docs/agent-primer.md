# Agent primer

> Required reading for any agent (subagent or human) implementing a new
> scope in the bitbottle repo. Read this **once** at the start of a
> scope; do not re-read these files in your prompt context unless a
> specific finding requires it. The point is to deduplicate the ~50K
> tokens every implementer would otherwise spend re-discovering the
> same patterns.

## Architecture vocabulary (in order)

1. **`api/backend/client.go`** — the composite `Client` interface and the
   `Feature` type. Per-feature capability interfaces live in
   `api/backend/client_<feature>.go` files (e.g. `client_pr.go`,
   `client_repo.go`, `client_branch.go`, `client_issue.go`,
   `client_search.go`, `client_pipeline.go`, …). Optional interfaces
   (e.g. `IssueClient`, `WorkspaceClient`, `BranchProtector`,
   `PRReopener`) are **not** in `Client`; they're reached via
   `AsXxxClient(c, host)` which returns a typed
   `*DomainError{Kind: ErrUnsupportedOnHost, Code: host.unsupported, …}`
   when the backend doesn't implement the capability. Mirror this shape
   exactly when adding a Cloud-only or Server-only capability. New
   interfaces go in `client_<feature>.go` — do **not** add them to
   `client.go`.

2. **`api/backend/types.go`** — domain type vocabulary. Add new types
   here, keep JSON tags only when the public CLI/MCP contract needs a
   stable wire shape (otherwise leave untagged so the type is reusable
   across surfaces).

3. **`api/backend/errors.go`** — `DomainError`, `Kind` sentinels
   (`ErrNotFound`, `ErrAuth`, `ErrPermission`, `ErrConflict`,
   `ErrUnsupportedOnHost`, `ErrTransport`), `ErrorCode` strings
   (`pr.not_found`, `auth.invalid_token`, `host.unsupported`, …). New
   error paths reuse existing codes; only add a new code when the
   catalogue genuinely lacks one.

4. **`api/internal/paging/`** — `paging.Collect[T]` is the canonical
   helper for any `List*` operation. Never reintroduce the
   per-page-vs-total-cap pattern; pass `limit` through and let
   `Collect` clamp.

5. **`pkg/cmd/factory/factory.go`** — `f.BaseRepo()` resolves
   `bitbottle.host`/`project`/`slug` pinned defaults before falling back
   to git remotes. `f.Backend(host)` returns the configured `Client`
   for a host. Tests should override `BaseRepo` via
   `factorytest.Opts.BaseRepo`, not by enqueueing fake git output.

6. **`pkg/cmdregistry`** — package-level `Register(Builder)` + `All(f)`
   for Cobra subcommand self-registration. New command packages call
   `cmdregistry.Register(NewCmdXxx)` from `init()` so they appear in
   the root without editing `root.go`. The fixed AddCommand list in
   `root.go` is legacy; new commands MUST use this mechanism.

7. **`pkg/cmd/mcp/handlers.go` + `tools.go` + `tools_<domain>.go`** —
   MCP tool registration + handler envelope. The central `tools.go`
   calls `registeredFns()` (from `pkg/cmd/mcp/registry.go`) after the
   fixed list, so new tools self-register from `tools_<domain>.go`
   files via `init()` without editing `tools.go`. Existing handlers
   carry `host.unsupported` typed-error envelopes through to MCP
   clients; new handlers must preserve that shape.

8. **`pkg/errfmt/`** — terminal error rendering with the catalogue.
   Backend adapters classify errors once; the cmd layer stays free of
   per-command boilerplate.

9. **`test/testhelpers/`** — `FakeClient` with `Fn`-suffixed function
   fields. New backend interfaces extend `FakeClient` with one more
   `Fn` field plus a `Fatalf`-on-unset method.

## Exemplar files (copy patterns from these)

| For… | Read |
|---|---|
| Cloud-only command group | `pkg/cmd/issue/list.go` + `pkg/cmd/issue/issue.go` |
| Both-backend command | `pkg/cmd/pr/decline.go` (with its `_test.go`) |
| TTY table + `--json`/`--jq` | `pkg/cmd/repo/view.go` + `pkg/cmd/repo/view_test.go` |
| Cloud paginated list | `api/cloud/issues.go` + `api/cloud/issues_test.go` |
| Server REST adapter | `api/server/prs.go` + `api/server/prs_test.go` |
| Cloud-only optional interface + `AsXxxClient` | `api/backend/client_issue.go` (`AsIssueClient`) and `api/backend/client_workspace.go` (`AsWorkspaceClient`) |
| Server-only optional interface + `AsXxxClient` | `api/backend/client_branch.go` (`AsBranchProtector`) and `api/backend/client_pr.go` (`AsPRReopener`) |
| Server optimistic-concurrency PUT/POST | `MergePR` and `ReadyPR` in
  `api/server/prs.go` |
| MCP envelope (`host.unsupported`) | `searchCode` and `reopenPR` in
  `pkg/cmd/mcp/handlers.go` |

## Invariants (never violate)

- `httpx.Transport` policy: `ContentTypeAlwaysWrite` for Server (avoids
  CSRF 403), `ContentTypeWhenBody` for Cloud (avoids Cloud 400 on
  empty POST). Don't override per-call.
- `httpx.Transport.UseDomainErrors(host)` is the single point where
  HTTP errors become typed `DomainError`s. Adapters wrap operation-
  specific errors (e.g. `stampPRNotFound`) **on top** of that, never
  instead of it.
- `URL` paths interpolating user-supplied values use `url.PathEscape`
  (see `api/cloud/tag.go`, `api/cloud/prs.go`).
- Cloud `pagelen` ≤ 100; the Cloud API rejects larger.
- Server reopen / merge / put endpoints carry `version` for optimistic
  concurrency.

## What the primer does NOT cover

This is the architectural vocabulary, not the spec for your scope.
Read your scope's PRD (linked GitHub issue) for what to build.
