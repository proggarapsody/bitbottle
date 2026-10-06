# Project history

Historical records are retained as evidence. Their rules and status claims apply
to their original dates; they do not override current repository guidance.

| Record | Location |
| --- | --- |
| Stream reports and cycle analysis | [iterations/](iterations/) |
| Early feature and interface designs | [designs/specs/](designs/specs/) |
| Original custom workflow and Claude command adapters | [legacy-workflow/](legacy-workflow/) |
| Shipped progress | [SHIPPED](../backlog/SHIPPED.md) |
| Historical planning and functionality map | [BACKLOG snapshot](backlog/2026-10-06-BACKLOG.md) |
| Completed local TDD sessions, EX1 review, initial design notes | `.agents/history/legacy-workflow/` (gitignored) |
| Original local cycle and metrics ledgers | `.claude/auto-iter/`, `auto-iter/cycles.jsonl` (gitignored) |

The 2026-10-06 migration preserves historical contents byte for byte. The tracked
[migration manifest](legacy-workflow/migration-manifest.json) records original
paths, destinations, and SHA-256 hashes. The ignored local manifest also covers
local files, runtime snapshots, and the previously untracked June 18 report.
Runtime originals remain in place for compatibility with retained scripts.
No local settings or credentials were promoted into tracked documentation.

The original backlog snapshot contains open-versus-shipped duplicates; the
active queue was reconciled per migration evidence in [backlog history](backlog/README.md).
- The cycle-analysis folder name says 158–187, while its final report describes
  an actual window of 153–187.
- The June 18 report's explanation of Go indirect dependencies and the old
  local TDD checklist's generalization about Cloud capabilities are historical
  claims, not verified current guidance.
- The old workflow and its Claude adapter disagreed on review/push sequencing;
  its script catalog also incorrectly said shell tests were absent from CI.

Do not normalize raw ledgers, erase completed sessions, or silently reconcile
status claims. Verify against code and linked PRs before changing current plans.
