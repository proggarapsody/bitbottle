# Conductor — bitbottle

Navigation hub for project context. These artifacts capture *what* bitbottle
is and *how* it gets built, so any contributor (human or agent) lands with
the same baseline.

## Quick links

- [Product Definition](./product.md) — what bitbottle is, who it's for, why
- [Product Guidelines](./product-guidelines.md) — voice, tone, design principles
- [Tech Stack](./tech-stack.md) — languages, libraries, distribution
- [Workflow](./workflow.md) — TDD, commits, review, verification
- [Tracks Registry](./tracks.md) — open/closed tracks (auto-populated by `/conductor:new-track`)
- [Code Style Guides](./code_styleguides/) — per-language conventions

## How this relates to the rest of the repo

Conductor sits *alongside* the existing project docs, not on top of them.
The existing files remain the source of truth for execution; Conductor
provides a standard scaffold for tracking individual work items.

| Concern | Source of truth |
|---|---|
| Backlog of scopes | [`BACKLOG.md`](../BACKLOG.md) |
| Agent contract / project rules | [`AGENTS.md`](../AGENTS.md) |
| Workflow checklists | [`docs/workflows/`](../docs/workflows/) |
| Manual test scenarios | [`docs/manual-tests/`](../docs/manual-tests/) |
| Per-track spec + plan | [`conductor/tracks/<name>/`](./) |

## Active tracks

<!-- Auto-populated by /conductor:new-track -->

_None yet — run `/conductor:new-track` to create your first track._

## Getting started

1. Read [`product.md`](./product.md) to understand what bitbottle is.
2. Read [`workflow.md`](./workflow.md) before starting any change.
3. Run `/conductor:new-track` when picking up a new scope from `BACKLOG.md`.
