# Tracks Registry

Tracks are the units of work in Conductor — one per scope, bug, or
refactor. Each track gets its own directory at
`conductor/tracks/<track-id>/` containing `spec.md` and `plan.md`.

| Status | Track ID | Title | Created | Updated |
| ------ | -------- | ----- | ------- | ------- |

<!-- Tracks registered by /conductor:new-track -->

## How to use

- **Create**: `/conductor:new-track`
- **Implement**: `/conductor:implement` (within an existing track)
- **Status overview**: `/conductor:status`
- **Manage** (archive / restore / rename / delete): `/conductor:manage`
- **Revert** (undo at track / phase / task granularity): `/conductor:revert`

## Source of scope

New tracks should pick from [`BACKLOG.md`](../BACKLOG.md) — the
authoritative list of open scopes — unless they're bug fixes or
refactors driven by something else. Don't invent scopes that don't
appear in the backlog.
