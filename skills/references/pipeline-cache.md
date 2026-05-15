# Pipeline Caches

Pipeline caches store build dependencies between Bitbucket Cloud pipeline runs
to speed up subsequent builds. This feature is **Cloud only**.

## Commands

```bash
# List all pipeline caches for a repository
bitbottle pipeline cache list WORKSPACE/REPO

# Delete a cache by UUID
bitbottle pipeline cache delete WORKSPACE/REPO {uuid}
```

`list` supports `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |

## JSON output

```bash
# List caches as JSON
bitbottle pipeline cache list MYWORKSPACE/my-service --json

# Extract UUIDs with jq
bitbottle pipeline cache list MYWORKSPACE/my-service --json \
  --jq '.[].uuid'
```

JSON fields: `uuid` (string), `name` (string), `path` (string),
`fileSizeBytes` (int64), `createdOn` (string).

## MCP tools

| Tool | Description |
|---|---|
| `list_pipeline_caches` | List caches. Params: `repo` (required), `hostname` |
| `delete_pipeline_cache` | Delete a cache. Params: `repo`, `uuid` (all required), `hostname` |

## Backend details

**Cloud** — `GET /repositories/{workspace}/{slug}/pipelines_config/caches/`
           `DELETE /repositories/{workspace}/{slug}/pipelines_config/caches/{uuid}`

List response: `{"values": [{"uuid": "{abc-123}", "name": "node_modules", "path": "/app/node_modules", "file_size_bytes": 12345678, "created_on": "2024-01-01T00:00:00.000Z"}, ...]}`

Delete returns HTTP 204. The UUID must be wrapped in curly braces in the
URL path (e.g. `{abc-123}`); the CLI handles this automatically.

**Server/DC** — not supported. Commands return a `host.unsupported` error.
