# Backend quirks ledger

Real Bitbucket Server/DC and Cloud API behaviors that **no linter, no
unit test against a hand-written fake, and no diff review can infer.**
They are only knowable by hitting the real API — or by reading this file.

This ledger exists because the iteration loop's verification chain (TDD
fakes, design-judge, CI) all validate the implementer's *assumptions*
against each other, never against the real backend. A wrong assumption
about how Server behaves is reproduced identically in the code, the fake,
and the test — so all three agree and ship the bug. See issue #655: `pr edit`
wiped reviewers and `pr request-review` 400'd, and the "full Server PUT"
rule was *already* known and correctly applied three functions away in the
same file — just not written down anywhere a new write op would consult it.

## How this ledger is used in the loop

- **PRD time** (`docs/workflows/iteration-cycle/README.md` §2) — every
  write/mutation or new-API scope lists, in its `## Assumptions & Evidence`
  section, which `BQ-*` rows apply and how the design honors them. A
  suspected new quirk with no row here is `ASSUMED — UNVERIFIED` and must
  be settled by a reality probe before TDD.
- **TDD time** (§3) — the subagent honors every applicable row and asserts
  on the **captured request** (which fields were sent), not just stdout.
- **Design-judge** (`docs/workflows/pre-merge-check.md` §6a) — a write-op
  PR that violates an applicable row, or whose test only asserts stdout, is
  a BLOCKER.

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
