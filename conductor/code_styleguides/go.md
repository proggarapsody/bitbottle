# Go style guide — bitbottle

This guide does **not** redefine rules. It points at the existing
sources of truth and adds project-specific patterns that don't fit in a
linter config.

## Sources of truth (cite, don't duplicate)

| Concern | File | What's there |
|---|---|---|
| Lint rules | [`.golangci.yml`](../../.golangci.yml) | Enabled linters and their settings; `make lint` runs this |
| Project invariants | [`AGENTS.md`](../../AGENTS.md) "Key rules for AI agents" | `IOStreams`, `noctx`, `factory.NewTestFactory`, library-preference rules, dist/ ban, squash-merge gotcha |
| Architecture contract | [`BACKLOG.md`](../../BACKLOG.md) "Architecture Contract (per scope)" | Layer order for any new scope |
| Pre-merge gate | [`docs/workflows/pre-merge-check.md`](../../docs/workflows/pre-merge-check.md) | What's checked before a branch lands |

If a Go-style rule is needed in this guide, it's because it doesn't
belong in a linter config. Anything mechanical lives in `.golangci.yml`.

## Project-specific patterns

### Output

Always go through `f.IOStreams` — never `os.Stdout`, never
`fmt.Println`. The factory wires up `Out`, `ErrOut`, and `In` for
testability and for TTY-aware behavior.

```go
// good
fmt.Fprintln(f.IOStreams.Out, "Pull request created.")

// bad
fmt.Println("Pull request created.")
```

### HTTP

Always `http.NewRequestWithContext` + `client.Do`. Never
`client.Get/Head/Post` — the `noctx` linter blocks them.

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil { return err }
resp, err := client.Do(req)
```

### Tests

Always `factory.NewTestFactory` — never real filesystem, keyring, or
network. The test factory wires fakes for `IOStreams`, config, hosts,
HTTP client, etc.

### Listing operations

`api/internal/paging.Collect[T]` is the canonical helper for any
`List*` operation. Do **not** reintroduce a per-page-vs-total-cap
pattern.

### Errors

Domain errors live in `api/backend/errors.go` (`ErrNotFound`,
`ErrAuth`, `ErrPermission`, `ErrUnsupportedOnHost`, `ErrConflict`,
`ErrTransport`, `DomainError`).

- Adapters wrap raw HTTP errors via
  `httpx.Transport.UseDomainErrors(host)` — keep this wired on any new
  transport.
- User-facing rendering (titles + hints) lives in `pkg/errfmt`. Never
  format error prose inside `cmd/`.
- Stamp call-site-specific codes with
  `backend.StampCode(err, code, resource, id, feature)` after
  `ClassifyHTTPError`.
- Transport-layer errors (TLS, timeout) classify at
  `httpx.Transport.do()` — adapters don't need to wire them.

### Content-Type policy

`httpx.Transport.ContentTypePolicy` is injected:

- `ContentTypeAlwaysWrite` for Server/DC — avoids CSRF 403 on bodyless
  writes.
- `ContentTypeWhenBody` for Cloud — avoids 400 on empty POST.

Don't change these defaults per call; if a new transport is needed,
follow the same pattern.

### Library choices

Prefer well-known, widely adopted libraries. Hand-roll only when the
dependency footprint genuinely outweighs the value. The current
"approved" set lives in `go.mod` — see `tech-stack.md` for the rationale
on each.

### Layered architecture

Every new scope lands in this order:

```
api/backend/client.go    → new capability interface(s)
api/backend/types.go     → new domain type(s)
api/cloud/<domain>.go    → Cloud impl + _test.go
api/server/<domain>.go   → Server/DC impl + _test.go (skip if Cloud-only)
pkg/cmd/<domain>/        → cobra commands
pkg/cmd/mcp/tools.go     → MCP tool registration
pkg/cmd/mcp/handlers.go  → MCP handler method
```

Skipping a layer is allowed only with an explicit reason in the track's
`spec.md` (e.g., "no Server/DC impl: this Bitbucket op is Cloud-only —
returns `ErrUnsupportedOnHost` from the composite client").

### Cloud-only capabilities

Use the optional-interface pattern at the call site:

```go
type PipelineClient interface {
    ListPipelines(...) (...)
}

if p, ok := client.(PipelineClient); ok {
    return p.ListPipelines(...)
}
return backend.ErrUnsupportedOnHost
```

Don't add Cloud-only methods to the composite `Client`.

## Style basics

For everything not covered above, defer to:

- [Effective Go](https://go.dev/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- `gofmt` / `goimports` (run automatically by the pre-commit hook)
- `.golangci.yml` (run by `make lint`)

If those disagree (rare), the linter config wins.
