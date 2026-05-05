# bitbottle api passthrough + MCP server

Load this when the task involves raw REST calls (something
`bitbottle pr/repo/branch/…` doesn't expose) or configuring
bitbottle as an MCP server for an AI agent.

## Raw REST passthrough

```bash
bitbottle api 'PATH'
bitbottle api -X POST -F key=val 'PATH'
bitbottle api --paginate --jq '.[].name' 'PATH'
cat f.json | bitbottle api -X PUT --input - 'PATH'
```

The `PATH` is **relative to the host's REST base** and the prefix
differs by backend:

| Backend | Path prefix | Example |
|---|---|---|
| Cloud (`bitbucket.org`) | `2.0/` | `bitbottle api '2.0/user'` |
| Server/DC | `rest/api/1.0/` | `bitbottle api 'rest/api/1.0/projects'` |

If you pass the wrong prefix for the host, the server returns 404.

### Flag cheat sheet

- `-X METHOD` — HTTP verb. Default GET.
- `-F key=val` — typed field; numbers stay numeric, `true`/`false`
  become bool. JSON output.
- `-f key=val` — raw string field; everything stays a string.
- `--input -` — read body from stdin. Pair with `cat file | …`.
- `-H 'Header: value'` — add a header.
- `--paginate` — follow Cloud `next` and Server `nextPageStart`,
  merging the `values` arrays into one stream.
- `--jq 'expr'` — filter JSON output.

### Path placeholder expansion

Inside a Bitbucket checkout, these placeholders resolve from the
current repo:

| Placeholder | Substitutes |
|---|---|
| `{workspace}` / `{project}` | resolved project key (Server) or workspace slug (Cloud) |
| `{repo_slug}` / `{slug}` | resolved repo slug |

Example:
```bash
bitbottle api '2.0/repositories/{workspace}/{repo_slug}/pullrequests'
```

## MCP server

bitbottle exposes its full surface as an MCP server. Useful when an
AI agent (Claude Desktop, custom workflow, etc.) needs to operate
Bitbucket without shell access.

```bash
bitbottle mcp serve                             # starts MCP server on stdio
```

### Sample agent config

Claude Desktop / Claude Code / Cursor:

```json
{
  "mcpServers": {
    "bitbottle": {
      "command": "bitbottle",
      "args": ["mcp", "serve"]
    }
  }
}
```

The MCP server uses the same `hosts.yml` and env vars as the CLI,
so authentication carries over without extra configuration. Tool
calls take a `hostname` argument; if omitted and exactly one host is
configured, it's used automatically (same rule as CLI commands).
