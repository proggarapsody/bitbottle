# Project knowledge map

Use the repository docs as durable memory. Read the relevant source for a task;
do not load the full historical archive into every session.

This repository configures the agent harness through project instructions, skills, tools, and checks. The agent runtime provides the execution loop; GitHub Issues supplies current work, and repository docs supply relevant knowledge.

| Need | Source |
| --- | --- |
| Contribution, setup, testing conventions | [CONTRIBUTING](../../CONTRIBUTING.md) |
| CLI, MCP, consumer skill behavior | [TASTE](../TASTE.md) |
| Architecture and boundaries | [ARCHITECTURE](../ARCHITECTURE.md) |
| Concise implementation context | [Agent primer](../agent-primer.md) |
| Real backend behavior and evidence | [Backend quirks](../backend-quirks.md) |
| Recorded decisions | [ADRs](../adr/) |
| Active work and scope index | [GitHub Issues](https://github.com/proggarapsody/bitbottle/issues?q=is%3Aissue+is%3Aopen), [Backlog](../backlog/BACKLOG.md) |
| Shipped progress and PRD links | [Shipped ledger](../backlog/SHIPPED.md) |
| Merge requirements | [Pre-merge check](../workflows/pre-merge-check.md) |
| Planning and shipping actions | [Work tracking](../workflows/work-tracking.md) |
| Acceptance tests and manual checks | [Acceptance](../workflows/acceptance.md), [manual tests](../manual-tests/README.md) |
| Historical designs, reports, workflow | [History](../history/README.md) |
| Development skills, provenance, updates | [Project skills](../../.agents/skills/README.md) |

The custom autonomous loop is retired from the default workflow. Choose a skill
for the task at hand; preserve the repository's design, test, and release rules.
`auto-iter/scripts/` remains as compatibility tooling because CI tests it and
some utilities still read the original local ledger paths. Installation does
not start a loop or schedule future work.

Skill setup is complete: [GitHub Issues](issue-tracker.md), [default triage labels](triage-labels.md), and [single-context domain docs](domain.md). Existing GitHub PRD links and backlog records remain intact.

The domain convention is a root `GLOSSARY.md` for terms
and the existing `docs/adr/` for decisions. Create a glossary when there are
terms to record; do not copy the architecture document into it.

For new knowledge, update the appropriate source above, citing evidence and
separating observations from inference. Temporary task notes belong in ignored
local state; retain useful conclusions in durable docs. History is evidence of
past work, not a current instruction source or proof of current behavior.
