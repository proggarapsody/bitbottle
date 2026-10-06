# Backlog migration evidence

The raw byte-preserved snapshots are historical evidence, not active instructions: [2026-10-06 BACKLOG](2026-10-06-BACKLOG.md) and [2026-10-06 SHIPPED](2026-10-06-SHIPPED.md). Supporting records: [preservation manifest](preservation.json), [GitHub closed issue index](github-closed-index.json), [issue 654 before migration](issues/654-before.md), and [GitHub migration log](github-migration.json).

The migration covered 3,146 lines, 165 backlog rows, 159 completed scopes, and 6 open scopes before importing 4 scopes into the compact index.

COMMIT-SEARCH shipped 2026-06-02; closed issue [#650](https://github.com/proggarapsody/bitbottle/issues/650) and [PR #651](https://github.com/proggarapsody/bitbottle/pull/651) record it. Source and tests exist in `api/backend/client_commit_search.go`, `pkg/cmd/commit/search.go`, and `pkg/cmd/commit/search_integration_test.go`.

DEPLOY-KEY-PERMISSION shipped 2026-06-02; closed issue [#634](https://github.com/proggarapsody/bitbottle/issues/634) and [PR #635](https://github.com/proggarapsody/bitbottle/pull/635) record it. Source and tests exist in `pkg/cmd/deploykey/add.go` and `api/cloud/deploy_keys.go`.

These corrections follow recorded shipment, closed PR descriptions, and source existence; no fresh functionality tests were run. Issue [#654](https://github.com/proggarapsody/bitbottle/issues/654) retains its original body and remains open for triage. The 150 closed issues remain closed.

The canonical imported tickets are OSSF-BADGE #676, PR-LOCK #677, ADMIN-PUSH-RULES #678, and TESTSCRIPT-BACKFILL-2 #679. Issue #680 is a duplicate corrected by audit; this records the duplicate's existence without assigning a causal theory.
