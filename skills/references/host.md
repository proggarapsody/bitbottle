# host — Host Capabilities & Version Info

## Commands

```
bitbottle host info [--hostname HOST] [--json]
```

Prints metadata about the connected Bitbucket backend: type (cloud/server/datacenter), base URL, server version (Server/DC only), and the list of supported feature capabilities.

## Output (default table)

**Cloud:**
```
Backend:  cloud
Host:     bitbucket.org
Features: branch_model, branch_rules, ...
```

**Server/DC:**
```
Backend:  server
Host:     git.example.com
Version:  8.19.0
Features: admin, branch-protect, ...
```

## Flags

| Flag | Description |
|---|---|
| `--hostname HOST` | Override auto-detected host |
| `--json` | Emit JSON with all fields |
| `--jq EXPR` | Filter JSON with a jq expression |

## JSON fields

```json
{
  "backend_type": "cloud | server | datacenter",
  "base_url": "https://...",
  "version": "8.19.0",
  "build_number": "80190000",
  "display_name": "Bitbucket",
  "supported_features": ["issues", "pipelines", ...]
}
```

`version` and `build_number` are omitted for Cloud (rolling release).

## Use cases

- Discover whether a host is Cloud or Server/DC before issuing capability-gated calls.
- Check the server version to gate features requiring a minimum version (e.g. `pr unready` requires Server 8.0+).
- Enumerate available features for a host before calling `AsXxxClient`-gated commands.

## MCP tool

`get_host_info` — accepts `hostname` (optional).
