# ARCHITECTURE — bitbottle

Single source of truth for *how* bitbottle is built. Read before designing a
new package, interface, transport, or refactor.

> **Mantra**: composition over inheritance, interfaces over implementations,
> deep modules over shallow ones. When unsure, read the closest exemplar
> file (each section below names one).

---

## Layered structure

```
api/backend/        — pure domain types + Client interface (NO I/O)
api/cloud/          — Bitbucket Cloud adapter (impl details hidden)
api/cloud/gen/      — spec-derived wire types for Cloud (oapi-codegen, types-only)
api/server/         — Bitbucket Server/DC adapter (impl details hidden)
api/server/gen/     — spec-derived wire types for Server/DC (oapi-codegen, types-only)
api/internal/httpx/ — shared HTTP transport
api/internal/paging/— canonical paginator
internal/           — config, envvars, keyring, bbinstance (cross-cutting non-domain)
pkg/cmd/            — Cobra CLI (thin: flag-parse → backend → format)
pkg/cmd/mcp/        — MCP server (thin: schema → handler → backend)
pkg/errfmt/         — error code catalogue + Render
pkg/iostreams/      — stdin/stdout/pager/color abstraction
test/testhelpers/   — fakes, fixtures
```

**Layer rule**: a package never imports from a layer above it. `pkg/cmd/**`
may import `api/backend`; **may not** import `api/cloud`, `api/server`, or their
`gen/` sub-packages. The `gen/` packages are adapter-internal — only `api/cloud/*.go`
may import `api/cloud/gen`, and only `api/server/*.go` may import `api/server/gen`.
The only path through the codebase is `cmd → backend → adapter`.

**Enforcement**: `depguard` in `.golangci.yml` + Rule 5 in `scripts/smell-scan.sh`. Both fail CI on violations. Two principled exceptions are encoded in both checks:
- `pkg/cmd/factory/` — the **composition root**; it must know concrete adapters to wire them together (canonical hex/clean-architecture exemption).
- `pkg/cmd/**/*_integration_test.go` — integration tests legitimately construct real adapter clients against `httptest` servers to verify wire compatibility.

---

## Design decisions

### 1. Composite `Client` + optional interfaces
- `backend.Client` composite embeds **only** what BOTH backends implement
  (RepoLister, PRLister, BranchLister, …).
- Cloud-only (issues, deployments, search, workspaces) and Server-only
  (branch protect, code-insights) capabilities are **optional interfaces**:
  `IssueClient`, `DeploymentClient`, `CodeInsightsClient`, etc.
- Access via type assertion: `if issues, ok := backend.AsIssueClient(c); ok { … }`.
  The `AsX` helpers attach `ErrUnsupportedOnHost` automatically when the
  assertion fails.
- **Why**: each backend implements only what it can; composite stays honest
  about cross-backend symmetry; capability gaps surface as typed errors, not
  panics.
- **Exemplar**: `api/backend/client.go` + `AsIssueClient` / `AsPipelineClient`
  / `AsBranchProtector`.

### 2. Interface segregation per capability
- One capability = one interface. `TagLister`, `TagCreator`, `TagDeleter` —
  not a `TagClient` with three methods.
- **Why**: clients can mock individual capabilities; future split between
  read/write hosts becomes trivial; ISP literally.
- **Exemplar**: `api/backend/client_tag.go`.

### 3. `paging.Collect[T]` for every `List*`
- Single canonical paginator at `api/internal/paging/`. Hides Cloud's `next`
  URL pattern AND Server's `nextPageStart` cursor pattern behind one generic
  call.
- **Never** reintroduce the per-page-vs-total-cap pattern. **Never** call
  paginator state from caller code.
- **Exemplar**: `api/cloud/repos.go` `ListRepos`.

### 4. `httpx.Transport` with policies
- All HTTP I/O goes through `api/internal/httpx`. Raw `net/http.Client` outside
  `httpx` is a depguard violation.
- **`ContentTypePolicy`** injected per host: `ContentTypeAlwaysWrite` on
  Server (avoids CSRF 403), `ContentTypeWhenBody` on Cloud (avoids 400 on
  empty POST).
- **`UseDomainErrors(host)`** wraps generic HTTP errors into
  `Err{NotFound,Auth,Permission,UnsupportedOnHost,Conflict,Transport}` and
  attaches dotted error codes from `errfmt`.
- **Why**: cross-cutting concerns (content-type, error classification, retry,
  paging) live in one place; adapters stay focused on endpoint shapes.

### 5. Typed errors with dotted codes
- Vocabulary in `api/backend/errors.go`; catalogue in `pkg/errfmt/`
  (`auth.invalid_token`, `pr.merge.conflict`, `host.unsupported`, …).
- Adapters classify once via `httpx.Transport.UseDomainErrors`; cmd layer
  renders via `errfmt.Render` (TTY-aware, color, hint catalogue).
- **Hints are deterministic** — table-keyed by code, never heuristic regex
  over the server message.
- **Exemplar**: `pkg/errfmt/codes.go`.

### 6. `cmd` layer is thin
- Every cobra command: `parse flags → call backend interface → format output`.
  **Zero business logic.**
- Every applicable command supports `--json` / `--jq` / `--hostname`
  (`cmdutil.AddJSONFlags`).
- TTY-aware output via `iostreams`; color via `iostreams.ColorEnabled()`.
- **Exemplar**: `pkg/cmd/repo/list.go`.

---

## SOLID

| Principle | Application |
|---|---|
| **S**ingle Responsibility | One interface = one capability. One package = one concern. |
| **O**pen/Closed | Composite `Client` extends via new optional interfaces; never modifies existing methods. |
| **L**iskov | Cloud and Server impls of the same interface are interchangeable; differences surface via `ErrUnsupportedOnHost`, not by changing semantics. |
| **I**nterface Segregation | See decision §2 above. |
| **D**ependency Inversion | `pkg/cmd` depends on `api/backend.Client` (abstraction), not on `api/cloud` / `api/server` (concrete). |

---

## Deep modules (Ousterhout)

A "deep module" hides a lot of complexity behind a small interface. Internals
can change freely; callers don't notice.

| Module | Surface | Hidden complexity |
|---|---|---|
| `httpx.Transport` | Constructor + `Do(req)` | retry, content-type policy, domain-error classification, paging, custom headers, TLS config |
| `paging.Collect[T]` | One generic function | Cloud's `next` URL vs Server's `nextPageStart`, page-size caps, cursor exhaustion |
| `errfmt.Render(io, err)` | One function | TTY detection, color, dotted-code lookup, hint catalogue, `--debug` mode |
| `backend.AsIssueClient(c)` | Type assertion + bool | "is this backend even capable of issues?" + `ErrUnsupportedOnHost` wrapping |
| `iostreams` | `In/Out/ErrOut` + `StartPager/StopPager` + `ColorEnabled()` | terminal detection, `$PAGER`, `NO_COLOR`, `--no-color` global flag |

**Anti-pattern (shallow module)**: a package whose interface size is comparable
to its implementation size. Refactor — extract or collapse.

---

## Clean code

- **Functions short, single-purpose.** If a function has >2 nested levels,
  extract.
- **Names communicate intent.** `CreateBranchInput`, not `BranchData`.
  `ListPRComments`, not `GetPRStuff`.
- **No dead code.** Off-the-shelf linters catch syntactic dead code
  (`unused`, `staticcheck`, `gocritic` via golangci-lint). They do **not**
  catch semantic no-ops — e.g. `if x == "Y" { x = "Y" }`,
  `if cond { return v } else { return v }`,
  `slice = append(slice, item); slice = slice[:len(slice)]`. Those are
  on the human/judge reviewer (see Anti-patterns).
- **Test names are sentences.** `TestListPRs_FiltersByState_ReturnsMatching`,
  not `TestListPRs2`.
- **One PR = one logical change.** Conventional Commits
  (`feat(scope): ...`, `fix(scope): ...`).
- **TDD**: red → green → refactor. One commit per phase if non-trivial.
- **Comments explain *why*, not *what*.** The code already says what.

---

## `gh` references

`reference/gh/` is checked into the repo. Read it before designing:

- **Command layout**: `gh pr list` → `bitbottle pr list`. Don't invent your
  own noun-verb shape.
- **Flag patterns**: `gh pr view --json fields` is the model for
  `--json` / `--jq`.
- **Exit codes**: 0 success, 1 generic, 2 usage, 4 auth, 8 unsupported.
  Mirror `gh` exactly.
- **MCP tool naming**: `gh`'s `snake_case_verb_noun` is the canonical;
  reproduce.
- **TTY-aware output**: `gh repo list` formatting (table on TTY, TSV on pipe,
  no header on pipe) is the model.

---

## Anti-patterns

- ❌ Adding a new transport that doesn't go through `httpx`.
- ❌ Adding a backend method to the composite `Client` when it's only
  implemented on one backend.
- ❌ Business logic in `pkg/cmd/`.
- ❌ Calling pagination from cmd code instead of via `paging.Collect`.
- ❌ Bare `fmt.Errorf` for user-facing errors (no code, no hint).
- ❌ A "client" interface with >5 methods covering >1 concern (split per ISP).
- ❌ A package whose API surface is the same size as its implementation
  (shallow module — collapse or extract).
- ❌ `pkg/cmd/**` importing `api/cloud` or `api/server` directly.
- ❌ **Tautological / dead-branch logic.** Conditional whose body assigns
  the same value the condition already pins (`if x == "Y" { x = "Y" }`),
  or whose branches return identical values, or whose work is undone on
  the next line. Standard Go linters don't flag this — review for it
  explicitly. Real example: cyc 55 / PR #292 shipped
  `if state == "CHANGES_REQUESTED" { state = "CHANGES_REQUESTED" }` in
  `api/cloud/pr_participants.go` (removed in PR #295).

---

## Test tiers

| Tier | What it owns | Location | Catches |
|---|---|---|---|
| 1. **Unit** | one function / one struct method | `api/cloud/**/*_test.go`, `api/server/**/*_test.go`, `internal/**/*_test.go` | logic bugs inside a leaf module |
| 2. **Adapter / integration** | one cobra command against an `httptest` fake | `pkg/cmd/**/*_integration_test.go` | wire compatibility for one route + handler |
| 3. **Script (testscript)** | whole binary, real argv/env/exit codes | `test/script/testdata/*.txtar` | flag wiring, errfmt rendering, `bitbottle.host` defaulting, capability gaps |
| 4. **Contract** | cross-package invariants that can't live in a leaf | `test/contract/*_test.go` | hint↔flag drift, future cross-cutting regressions |

### Tier 3 — testscript harness

- **Entry point**: `test/script/main_test.go` registers `bitbottle: app.Run` via `testscript.RunMain`.
- **Env isolation**: each script gets a fresh `HOME`, scrubbed `BB_*`/`GIT_*`/`XDG_*` env, and its own config dir at `$WORK/.config/bitbottle`.
- **Fake servers**: `bb-fake server` / `bb-fake cloud` in a script start an `httptest.TLSServer` with fixture stubs and write `hosts.yml` so `bitbottle` authenticates without touching the OS keyring.
- **Cloud override**: `BB_CLOUD_BASE_URL` redirects Cloud API calls to the fake server (see `internal/bbinstance/bbinstance.go`).
- **Run**: `make test-script` or `go test ./test/script/... -race`.
- **New feature gate**: every new user-visible command must add at least one `.txtar` script before merge.
