# Product Guidelines — bitbottle

## Voice and tone

| Surface | Tone |
|---|---|
| **CLI output** | Concise and direct, `gh`-style. Short imperatives. Minimal prose. Examples: `Pull request created.`, `Authentication required.`, `Webhook deleted.` Tab-separated rows on pipes; aligned tables on TTY. |
| **README / contributor docs** | Approachable but professional. Full sentences when teaching; terminal-style blocks when showing examples. |
| **Error messages** | Direct + actionable. The `pkg/errfmt` "title + hint" format: a human-readable title (what failed) plus a hint (what to do about it), indexed by a dotted error code. |

## Design principles

These are non-negotiable. They show up in every PR review and in the
pre-merge gate.

1. **`gh` philosophy throughout.**
   - Noun-verb commands (`bitbottle pr list`, `bitbottle tag create`).
   - Consistent flags on every applicable command: `--limit`, `--json`,
     `--jq`, `--web`, `--hostname`.
   - TTY-aware output: aligned tables on TTY; tab-separated, no header,
     on pipes.
   - Thin commands — `cmd/` parses flags → calls a backend interface →
     formats output. Zero business logic in `cmd/`.

2. **Interface segregation.** Each capability is its own interface in
   `api/backend/client.go`. The composite `Client` embeds only what
   *both* backends implement. Cloud-only operations use the
   optional-interface pattern (type assertion at the call site, e.g.
   `PipelineClient`).

3. **Layered architecture, per-scope contract.** Every new scope follows
   the same layer order. No exceptions:

   ```
   api/backend/client.go    → new capability interface(s)
   api/backend/types.go     → new domain type(s)
   api/cloud/<domain>.go    → Cloud impl + _test.go
   api/server/<domain>.go   → Server/DC impl + _test.go
   pkg/cmd/<domain>/        → cobra commands
   pkg/cmd/mcp/tools.go     → MCP tool registration
   pkg/cmd/mcp/handlers.go  → MCP handler method
   ```

4. **Prefer well-known libraries.** When a problem has a clear
   community-standard solution (color: `fatih/color`, MCP: `mark3labs/mcp-go`,
   keyring: `zalando/go-keyring`), use it. Hand-roll only when the
   dependency footprint outweighs the value.

5. **One MCP per Bitbucket op.** Every command that maps to a Bitbucket
   REST call gets a matching MCP tool. Agents and humans share one
   surface; we never want a command available to humans but not to
   agents (or vice versa).

## When in doubt

- For CLI design: check `reference/gh/`.
- For project rules: check [`AGENTS.md`](../AGENTS.md).
- For workflow: check [`docs/workflows/`](../docs/workflows/).
