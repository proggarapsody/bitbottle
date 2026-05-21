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

---

# Pipeline Stop

Stop a running pipeline. This feature is **Cloud only**.

## Commands

```bash
# Stop a pipeline by UUID (TTY — no --confirm needed)
bitbottle pipeline stop UUID WORKSPACE/REPO

# Stop a pipeline in a non-TTY environment (CI, scripts)
bitbottle pipeline stop UUID WORKSPACE/REPO --confirm

# When BaseRepo is configured, WORKSPACE/REPO is optional
bitbottle pipeline stop UUID --confirm
```

## Flags

| Flag | Description |
|---|---|
| `--confirm` | Required when stdout is not a TTY (e.g. CI scripts) |
| `--hostname HOST` | Override the Bitbucket host |

## MCP tools

| Tool | Description |
|---|---|
| `stop_pipeline` | Stop a running pipeline. Params: `repo` (required), `uuid` (required), `hostname` |

## Backend details

**Cloud** — `POST /2.0/repositories/{workspace}/{slug}/pipelines/{uuid}/stopPipeline` (empty body, returns HTTP 204).
The UUID is automatically wrapped in curly braces (e.g. `{abc-123}`) as required by the Cloud API.

**Server/DC** — not supported. Returns a `host.unsupported` error.
