# Pipeline Schedules

Pipeline schedules let you run a Bitbucket Cloud pipeline on a recurring
cron schedule. This feature is **Cloud only**.

## Commands

```bash
# List all pipeline schedules for a repository
bitbottle pipeline schedule list WORKSPACE/REPO

# Create a schedule (runs daily at midnight UTC on main)
bitbottle pipeline schedule create WORKSPACE/REPO \
  --cron "0 0 * * *" --branch main

# Create a disabled schedule (can be enabled later)
bitbottle pipeline schedule create WORKSPACE/REPO \
  --cron "0 12 * * 1" --branch develop --enabled=false

# Delete a schedule by UUID
bitbottle pipeline schedule delete WORKSPACE/REPO {uuid}
```

`list` and `create` support `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |
| `create` | `--cron EXPR` | Cron expression, e.g. `"0 0 * * *"` (required) |
| `create` | `--branch BRANCH` | Branch to run the pipeline on (required) |
| `create` | `--enabled` | Whether the schedule is active; default `true` |

## JSON output

```bash
# List schedules as JSON
bitbottle pipeline schedule list MYWORKSPACE/my-service --json

# Extract UUIDs with jq
bitbottle pipeline schedule list MYWORKSPACE/my-service --json \
  --jq '.[].uuid'
```

JSON fields: `uuid` (string), `enabled` (bool), `cronExpression` (string),
`branch` (string).

## MCP tools

| Tool | Description |
|---|---|
| `list_pipeline_schedules` | List schedules. Params: `repo` (required), `hostname` |
| `create_pipeline_schedule` | Create a schedule. Params: `repo`, `cron`, `branch` (all required), `enabled` (bool, default true), `hostname` |
| `delete_pipeline_schedule` | Delete a schedule. Params: `repo`, `uuid` (all required), `hostname` |

## Backend details

**Cloud** — `GET/POST/DELETE /repositories/{workspace}/{slug}/pipelines_config/schedules`

List response: `{"values": [{"uuid": "{abc}", "enabled": true, "cron_expression": "0 0 * * *", "target": {"branch": "main"}}, ...]}`

Create body:
```json
{
  "enabled": true,
  "cron_expression": "0 0 * * *",
  "target": {"ref_type": "branch", "ref_name": "main", "type": "pipeline_ref_target"}
}
```

Delete returns HTTP 204. The UUID must be wrapped in curly braces in the
URL path (e.g. `{abc-123}`); the CLI handles this automatically.

**Server/DC** — not supported. Commands return a `host.unsupported` error.
