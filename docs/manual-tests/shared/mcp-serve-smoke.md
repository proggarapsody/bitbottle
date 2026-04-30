# Scenario: `mcp serve` handshake smoke test

**Backend:** N/A (the MCP server itself; tools may hit a backend, but this
scenario only verifies startup + protocol handshake).

## Prerequisites

- `jq` installed.
- Authenticated to at least one host (so MCP tools can be discovered).

## Steps

### 1. Server starts and responds to `initialize`

```bash
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"manual-test","version":"0"}}}' \
  | bitbottle mcp serve \
  | head -1 \
  | jq .
```

Stdout is a single JSON-RPC response object with:

- `jsonrpc: "2.0"`
- `id: 1`
- `result.protocolVersion` set
- `result.serverInfo.name` containing `bitbottle`
- `result.capabilities` present (object)

Exit code: `0`.

### 2. Tool list is non-empty

```bash
{
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"manual-test","version":"0"}}}'
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  printf '%s\n' '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
} | bitbottle mcp serve | jq -c 'select(.id==2) | .result.tools | length'
```

Stdout: a positive integer (number of registered tools). Exit code: `0`.

### 3. Malformed JSON-RPC is rejected without crashing

```bash
printf 'not json\n' | bitbottle mcp serve
```

Server emits an error response (or logs to stderr) and does not panic /
segfault. Exit code may be `0` (clean shutdown on EOF) or non-zero — record
which.

## Cleanup

None.
