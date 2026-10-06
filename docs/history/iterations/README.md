# auto-iter reports

Post-stream analysis reports. One file per stream or analysis window.

## Naming convention

`<kind>-<YYYY-MM-DD>-<descriptor>.md`

Where `<kind>` is:
- `stream-` — a `/auto-iter-stream N` run analyzed end-to-end
- `cycles-` — ad-hoc analysis of a cycle range (not a single stream)
- `compare-` — side-by-side era comparison
- `audit-` — focused audit of one dimension (e.g. metric emission, design-judge calibration)

The date is the start of the analyzed window (UTC), not the report
generation date. `<descriptor>` is one of:
- `cycles-NNN-MMM` for stream/cycles reports
- A short slug for thematic audits (`metric-emission`, `dj-false-positives`)

Examples:
- `stream-2026-05-24-cycles-135-144.md`
- `audit-2026-05-26-metric-emission.md`
- `compare-2026-05-30-streams-q2.md`

## When to write a report

- After each `/auto-iter-stream` completes (whether `stream_max_reached`,
  `stream_halted`, or `stream_shutdown`).
- After any ad-hoc range analysis the user asks for.
- After fixing a recurring algorithm bug — capture before/after metrics.

Reports are written via prompt, not script. The minimum sections are:
TL;DR, Throughput, Code quality, Predictability & consistency,
Recommended fixes, Bottom line. Add comparison tables when the prior
era has measurable baselines.

## Why tracked

These are tracked in git (not in `.claude/auto-iter/` which is gitignored
runtime state) because they capture **decisions and observations about
the workflow itself**. Future cycles need to read them to avoid
re-discovering the same recurring bugs.
