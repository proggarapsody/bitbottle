# Pre-merge check

Run the applicable checks before a branch lands on `main`. Report blockers with
file/line or CI evidence; warnings do not block. In review mode, report findings
without changing the implementation. This gate applies to human and agent work;
it does not start an iteration loop or authorize a merge.

## 0. Scope first

Inspect the branch, fetch `origin/main`, and inspect changed files and commit
subjects with `git diff --name-only origin/main...HEAD` and
`git log --format='%s' origin/main..HEAD`. For uncommitted review, include the
working-tree diff. Scale review and tests to the actual change.

## 1. Branch and tree hygiene

Use `feature/`, `fix/`, `docs/`, or `chore/`; never push directly to `main`.
Before merge, the tree must be clean and the PR must target `main`.
A draft PR is a warning. Being behind `origin/main` alone is not a blocker;
conflicts or required failing checks are blockers.

## 2. Conventional commits and PR title

Use Conventional Commits. If the change contains `feat:` or `fix:`, use a matching
release-triggering squash PR title. GitHub uses that title as the squash subject.
Docs/chore/refactor/test changes may use their corresponding non-release prefix.
Confirm an intended major bump with the author before a breaking release.

## 3. Build artifacts and repository cleanliness

No tracked `dist/`, root binary, files over 1 MB, `.DS_Store`, logs, or
`coverage.out`. Do not commit ignored local history, runtime state, personal
configuration, credentials, or private backend details.

## 4. Lint and tests

For a pushed PR, inspect `gh pr view --json statusCheckRollup` and require every
required CI check to succeed. For failed or pending checks, report the exact
name and log URL; read the failure before fixing it. Do not repeat green checks
without a new change or unresolved concern.

Without a PR, run `make lint` and `make test` for code changes. `make test` enables
the race detector. Run `make test-scripts` when shell tooling or its paths change.
For documentation-only changes, verify links, path references, and preservation
of any moved records; code design review is skipped. Report exact failures.

## 5. Documentation and progress sync

- New commands self-register, implement each supported adapter, and add an MCP
  tool when they map to a Bitbucket operation. Preserve typed unsupported-host
  behavior for host-specific capabilities. Update the consumer `skills/SKILL.md`.
- Changed flags update curated consumer references; exhaustive details stay in
  command help. Verify the consumer skill router after changing its references.
- UX/output/new subcommands update README. A manual smoke guide for a new
  user-visible flow is recommended.
- Branch, commit, release, setup, or workflow changes update CONTRIBUTING and
  AGENTS together. Record new invariants and patterns in the appropriate source.
- Auth/config/token changes update the consumer auth reference.
- Backend client, types, or errors changes keep the agent primer accurate.
- When completing an imported backlog scope, remove its index row and add a
  dated SHIPPED entry in the same implementation change, retaining issue/PR
  links. Detailed specs remain in linked issue/history records. Preserve earlier progress.

## 6. Design-judge

New commands, interfaces, packages, transports, MCP tools, and error sites must
comply with [TASTE](../TASTE.md) and [ARCHITECTURE](../ARCHITECTURE.md). Cite a
compliant exemplar or a violation with its file/line and applicable principle.
Skip this review for documentation, CI, dependencies, or progress-record-only
diffs. Architecture violations block merging unless an allowed exception is
explicitly justified in the PR description.

## 6a. Architecture smells

Run `scripts/smell-scan.sh` as appropriate; retain the CI smell gate. Review for:

- A repeated three-way capability switch in at least three files (nine or more
  conversion hits): introduce a shared resolver instead of another clone.
- Structurally identical per-command test-helper packages: share the helper.
- New commands overlapping an open backlog surface: unify the form or deprecate
  the older surface in the same PR.
- A third or fourth conversion-function pair for a backend type: prefer typed
  enum marshaling methods.
- Added Go comment density above roughly 5%, except new exported-package docs:
  review whether comments explain intent or narrate code.
- Write tests that assert only stdout: also assert captured request fields.
  Honor applicable rows in [the backend quirks ledger](../backend-quirks.md).

These findings block merging unless explicitly justified in the PR description.

## 7. Release Please boundaries

Do not hand-edit CHANGELOG, the release manifest, or Release Please version
markers. The exception is an actual `release-please--*` release branch. Follow
[the release process](../release-process.md).

## 8. Secret leak scan

Inspect the diff for tokens, app passwords, credentials, private corporate
hostnames, internal usernames, and personal emails that belong in local config.
Use the repository secret-scanning CI check; investigate hits before merging.
The maintainer's private Server host and username must never be tracked.

## 9. Final report

Give a READY or NOT READY verdict with blockers, warnings, relevant evidence,
changed-file scope, commit count, PR title, release impact, and checks performed.
Do not claim merge readiness while required checks remain unverified.

The [original checklist](../history/legacy-workflow/pre-merge-check.md) is retained
for historical reference. Current instructions are in this file.
