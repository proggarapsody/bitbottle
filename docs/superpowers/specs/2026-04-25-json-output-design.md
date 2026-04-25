# JSON Output (`--json` + `--jq`) Design

**Date:** 2026-04-25
**Status:** Approved

---

## Goal

Add `--json <fields>` and `--jq <expr>` flags to all output-producing commands
(`repo list/view/create`, `pr list/view/create`), following the same UX as
GitHub CLI. A single deep `internal/format` package owns all rendering so
commands declare their fields once and get every output mode for free.

---

## Package: `internal/format`

### API

```go
package format

import "io"

// Field describes one output column and JSON key for type T.
type Field[T any] struct {
    Name    string       // --json key; lowercased, camelCase (e.g. "fromBranch")
    Header  string       // table column header; defaults to uppercase Name
    Extract func(T) any  // extracts the value from an item
}

// Printer renders a slice of T in the correct output mode.
type Printer[T any] struct { /* unexported */ }

// New constructs a Printer.
//   w          — destination writer (f.IOStreams.Out)
//   isTTY      — whether the output fd is a terminal
//   jsonFields — comma-separated field names ("" = table/tsv mode)
//   jqExpr     — jq filter expression ("" = no filter)
func New[T any](w io.Writer, isTTY bool, jsonFields, jqExpr string) *Printer[T]

// AddField registers a field. Call once per field during command setup.
func (p *Printer[T]) AddField(f Field[T])

// AddItem enqueues one result item.
func (p *Printer[T]) AddItem(item T)

// Render writes all items in the appropriate format.
// Returns an error for unknown field names, --jq without --json, or jq eval errors.
func (p *Printer[T]) Render() error
```

### Output modes

| `jsonFields` | `jqExpr` | TTY | Output |
|---|---|---|---|
| `""` | `""` | yes | aligned table, header row |
| `""` | `""` | no | tab-separated, no header |
| `"id,title"` | `""` | either | JSON array, requested fields only |
| `"id,title"` | `".[] \| .id"` | either | jq result written to stdout |
| `""` | anything | either | **error**: `--jq requires --json` |
| `"bad"` | — | either | **error**: unknown field "bad"; valid fields: … |

### JSON shape

Each item becomes a JSON object with only the requested fields:

```json
[
  {"id": 42, "title": "Fix auth", "state": "OPEN"},
  {"id": 43, "title": "Bump deps", "state": "OPEN"}
]
```

Single-item commands (`repo view`, `pr view`) emit a JSON object, not an array.

### `--jq` implementation

Uses `github.com/itchyny/gojq` (pure Go, no cgo). The printer:
1. Builds the JSON value (array or object) in memory.
2. Compiles the jq expression.
3. Runs the iterator, writing each output value as a JSON line.

---

## Dependency

Add to `go.mod`:

```
github.com/itchyny/gojq vX.Y.Z
```

One new dependency. Pure Go, no cgo, well-maintained, used by many Go CLIs.

---

## Commands affected

### `repo list`

Fields: `slug`, `name`, `namespace`, `scm`, `webURL`

```
bitbottle repo list --json slug,name
bitbottle repo list --json slug,webURL --jq '.[] | .webURL'
```

### `repo view`

Fields: same as `repo list`. Single-item → JSON object output.

### `repo create`

Fields: same as `repo list`. Returns the created repo as JSON object.

### `pr list`

Fields: `id`, `title`, `state`, `draft`, `author`, `fromBranch`, `toBranch`, `webURL`

```
bitbottle pr list --json id,title,state
bitbottle pr list --json id,state --jq '.[] | select(.state == "OPEN") | .id'
```

### `pr view`

Fields: same as `pr list` + `description`. Single-item → JSON object output.

### `pr create`

Fields: same as `pr list`. Returns the created PR as JSON object.

### Commands explicitly out of scope

| Command | Reason |
|---|---|
| `auth status` | Per-host text output, not structured data |
| `repo delete` | Side-effect only, no output |
| `repo clone` | Side-effect only |
| `pr merge` | Side-effect only |
| `pr approve` | Side-effect only |
| `pr diff` | Raw unified diff text |
| `pr checkout` | Side-effect only |
| MCP handlers | Already return JSON directly |

---

## Flag wiring

Each affected command gains two new flags:

```go
cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
```

`--jq` without `--json` is validated inside `Printer.Render()`, not at flag-parse time, so the error message can name the valid fields.

---

## Migration: tableprinter → format.Printer

Before (repo list):
```go
tp := tableprinter.New(f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), 80)
tp.AddHeader("SLUG", "PROJECT", "TYPE")
for _, r := range repos {
    tp.AddField(r.Slug)
    tp.AddField(r.Namespace)
    tp.AddField(r.SCM)
    tp.EndRow()
}
return tp.Render()
```

After:
```go
p := format.New[backend.Repository](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
p.AddField(format.Field[backend.Repository]{Name: "slug",      Header: "SLUG",    Extract: func(r backend.Repository) any { return r.Slug }})
p.AddField(format.Field[backend.Repository]{Name: "name",      Header: "NAME",    Extract: func(r backend.Repository) any { return r.Name }})
p.AddField(format.Field[backend.Repository]{Name: "namespace", Header: "PROJECT", Extract: func(r backend.Repository) any { return r.Namespace }})
p.AddField(format.Field[backend.Repository]{Name: "scm",       Header: "TYPE",    Extract: func(r backend.Repository) any { return r.SCM }})
p.AddField(format.Field[backend.Repository]{Name: "webURL",    Header: "URL",     Extract: func(r backend.Repository) any { return r.WebURL }})
for _, r := range repos {
    p.AddItem(r)
}
return p.Render()
```

`tableprinter` stays in the codebase for any future commands that do not need `--json`.

---

## Testing strategy

### `internal/format` unit tests (TDD, written first)

- TTY table output matches expected aligned string
- non-TTY TSV output, no header
- `--json id,title` → correct JSON array
- `--json` single item → JSON object (not array)
- Unknown field name → descriptive error
- `--jq` without `--json` → error
- Valid `--jq` expression → correct filtered output
- Invalid `--jq` expression → compile error surfaced

### Command-level tests (update existing)

Existing `repo list`, `pr list`, etc. tests stay green (no `--json` flag → same table/TSV behavior). Add new test cases:
- `--json slug,name` → JSON output matches fixture
- `--json id --jq '.[] | .id'` → raw id list

---

## Implementation order

1. Add `gojq` to `go.mod` / `go.sum`
2. `internal/format` package + full unit test suite (TDD)
3. `repo list` migration + tests
4. `repo view`, `repo create`
5. `pr list` migration + tests
6. `pr view`, `pr create`
