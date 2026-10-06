# Backend quirks ledger

Real Bitbucket Server/DC and Cloud API behaviors that **no linter, no
unit test against a hand-written fake, and no diff review can infer.**
They are only knowable by hitting the real API — or by reading this file.

Tests built from an incorrect API assumption can agree with the implementation
and still ship a bug. Issue #655 recorded reviewer loss from `pr edit` and a
400 response from `pr request-review`: the full Server PUT rule was known
elsewhere in the code, but absent from the guidance a new write operation read.

## How this ledger is used

- Before designing a write/mutation or new API operation, identify applicable
  `BQ-*` rows and record how the design honors them. Mark suspected new behavior
  `ASSUMED — UNVERIFIED`; settle it with a reality probe before implementation.
- During implementation, honor applicable rows and assert the captured request
  fields in tests, as well as the resulting behavior.
- During [design review](workflows/pre-merge-check.md#6a-architecture-smells),
  violation of an applicable row or stdout-only write tests blocks merging.

## How to add an entry

**Append-only.** Every production bug rooted in real-backend behavior earns
a new `BQ-N` row **in the same fix PR**. Never delete or rewrite a row —
if a quirk is later disproven, add a superseding row that cites it. The
whole value is that the ledger only grows with hard-won, real evidence.

A row needs: backend it applies to · the rule · the symptom when violated
(with the issue/PR that taught us) · a code citation showing the correct
*and* broken pattern · how to honor it.

---

## BQ-1 — Server/DC: a PUT to a PR is a full-object replace

- **Applies to:** Bitbucket Server/DC, `PUT /rest/api/1.0/.../pull-requests/{id}`
  (and PR-shaped PUTs generally).
- **Rule:** the body **replaces the entire resource**. Any field you omit
  is *cleared*, not preserved. You must read-modify-write: GET the current
  object, mutate only the target field, PUT the whole object back.
- **Symptom when violated:** silent data loss. `pr edit --title` removed
  all 21 reviewers from a PR (#655 Bug 1) because the PUT body carried only
  `{version, title, description}`.
- **Evidence:**
  - ✅ correct: `api/server/pr_lifecycle.go:72` (`ReadyPR` GETs `current`, flips one flag, PUTs the full object) — comment at `:69` spells out the rule.
  - ❌ broken: `api/server/pr_lifecycle.go:17` (`UpdatePR` PUTs a partial `{version, title, description}`, dropping `reviewers`).
- **How to honor:** GET-modify-PUT the full object. Never hand-build a
  partial body for a Server PR PUT. If a field must be cleared, send it as
  an explicit empty value — Server needs an explicit empty `reviewers: []`
  to clear (see the `prWithReviewers` note at `api/server/pr_lifecycle.go:123`).

## BQ-2 — Server/DC: PR-mutating writes require the current `version`

- **Applies to:** Server/DC PR-mutating `POST`/`PUT` — update, reopen,
  merge, decline, reviewers, ready/unready.
- **Rule:** optimistic concurrency. Include the PR's current `version`
  (from a fresh GET) in the body. Omit it and the server rejects the write.
- **Symptom when violated:** `HTTP 400 "version must be supplied for this
  request"` (`pr request-review`, #655 Bug 2), or `HTTP 409 "Pull request
  was updated…"` against any non-zero-version PR (reopen/merge).
- **Evidence:**
  - ✅ correct: `api/server/pr_lifecycle.go:46` (`ReopenPR` sends `version`; comment at `:42` explains the 409).
  - ❌ broken: `api/server/pr_lifecycle.go:152` (`RequestReview` builds `RestReviewerPR{Title, Description, Reviewers}` — no `version`).
- **How to honor:** every Server PR-mutating write carries `version` from a
  fresh GET in the same call.

## BQ-3 — Server/DC: write requests must send Content-Type (CSRF filter)

- **Applies to:** Server/DC `POST`/`PUT`/`DELETE`, *including empty-body
  writes*.
- **Rule:** Server's CSRF filter rejects write requests that omit a
  `Content-Type` header (HTTP 403). The Server transport uses
  `ContentTypeAlwaysWrite` for this reason.
- **Symptom when violated:** HTTP 403 on otherwise-correct writes.
- **Evidence:** `api/internal/httpx/httpx.go:83` (`ContentTypeAlwaysWrite`).
- **How to honor:** keep `ContentTypeAlwaysWrite` wired on every Server
  transport — it is the default; do not "simplify" it away.

## BQ-4 — Cloud: empty-body writes must NOT send Content-Type

- **Applies to:** Bitbucket Cloud `POST`/`PUT` with **no** body
  (`ApprovePR`, `DeclinePR`, `RequestChangesPR`).
- **Rule:** the mirror image of BQ-3. Cloud returns HTTP 400 when an
  empty-body write includes a `Content-Type` header. The Cloud transport
  uses `ContentTypeWhenBody`.
- **Symptom when violated:** HTTP 400 on empty-body Cloud writes.
- **Evidence:** `api/internal/httpx/httpx.go:75` (`ContentTypeWhenBody`).
- **How to honor:** keep `ContentTypeWhenBody` wired on every Cloud
  transport.

## BQ-5 — Pagination envelope differs by backend

- **Applies to:** every `List*` operation.
- **Rule:** Cloud paginates with `"next": "<absolute-url>"`; Server/DC
  paginates with `"isLastPage": bool` + `"nextPageStart": N`. The two are
  not interchangeable.
- **Symptom when violated:** truncated lists (only page 1) or infinite
  loops, depending on which envelope you assumed.
- **Evidence:** `api/internal/httpx/httpx.go:90` (`Paginator` interface,
  per-adapter implementations); canonical collector at `api/internal/paging`.
- **How to honor:** route every `List*` through `paging.Collect` with the
  adapter's `Paginator`. Never hand-roll page-walking.

---

## Planned hardening (tracked)

The tooling that turns this ledger from advisory into enforced — so a
violated `BQ-N` fails a test instead of shipping. Off-the-shelf, not
hand-rolled (a bespoke fake would just re-encode our assumptions):

- **#661 VCR-CASSETTES** — `go-vcr` record/replay; cassettes capture real
  Server/Cloud behavior (real `version`-required 400, real full-object
  replace) so BQ-1/BQ-2 regressions fail on replay.
- **#662 OPENAPI-VALIDATE** — `kin-openapi` validates requests/responses
  against the vendored specs; catches schema drift the ledger doesn't cover.
- **#663 ACCEPTANCE-LIVE-WIRE** — gh-style real-backend testscript suite
  (Tier 6); the live run is the reality probe that discovers new BQ rows.
- **#664 E2E-QUEUE-FEEDBACK** / **#665 SMOKE-METRIC** — route real-backend
  failures into the loop queue and track "shipped AND survived a real
  backend" instead of bare "shipped".

Historical initiative detail is preserved in `docs/history/backlog/2026-10-06-BACKLOG.md`; gates already landed in #658.
