# Work tracking

GitHub Issues holds current work, specifications, and triage labels. Follow [issue tracker guidance](../agents/issue-tracker.md) and [triage label guidance](../agents/triage-labels.md).

Explicitly invoke triage when categorization or clarification is needed. Imported records marked `needs-triage` are not approved implementation specifications. OSSF-BADGE requires manual external follow-up.

Use relevant current docs and narrow retrieval agents. Avoid loading full archive snapshots into context. Record shipped intent, date, and PR/issue link in [SHIPPED](../backlog/SHIPPED.md), and remove an imported index row in the same implementation change. Detailed notes remain linked in issue history. An implementation PR closes its issue after an authorized merge. This document itself authorizes no external messages or merges.

The legacy table contract is `ID | Scope | Commands | Backends | Tier | Status`; `🔲` means open and `🔲📝` means manual follow-up. When a mutation command returns a running ID, wait on the original session. Recover unknown outcomes through a unique migration marker before retrying, and never rerun a running batch. This last guard derives from an observed duplicate import; its cause is unestablished.
