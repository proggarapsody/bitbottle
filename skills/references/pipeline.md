# Pipeline Config

Enable or disable Bitbucket Cloud Pipelines at the repository level. This feature is **Cloud only**.

## Commands

```bash
# Get pipeline config for a repository
bitbottle pipeline config get WORKSPACE/REPO

# Enable pipelines for a repository
bitbottle pipeline config enable WORKSPACE/REPO

# Disable pipelines for a repository
bitbottle pipeline config disable WORKSPACE/REPO
```

`get` supports `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--hostname HOST` | Override the Bitbucket host |

## MCP tools

| Tool | Description |
|---|---|
| `get_pipeline_config` | Get pipeline config. Params: `project`, `slug` (required), `hostname` |
| `enable_pipelines` | Enable pipelines. Params: `project`, `slug` (required), `hostname` |
| `disable_pipelines` | Disable pipelines. Params: `project`, `slug` (required), `hostname` |

## Backend details

**Cloud** — `GET/PUT /2.0/repositories/{workspace}/{slug}/pipelines_config`

Response: `{"enabled": true, "type": "repository_pipeline_settings"}`

**Server/DC** — not supported. Returns a `host.unsupported` error.

---

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

# Pipeline Rerun

Re-trigger a finished pipeline at the same commit. This feature is **Cloud only**.

## Commands

```bash
# Rerun a pipeline by UUID (WORKSPACE/REPO optional when BaseRepo is configured)
bitbottle pipeline rerun UUID WORKSPACE/REPO

# When BaseRepo is configured, WORKSPACE/REPO is optional
bitbottle pipeline rerun UUID
```

## Flags

| Flag | Description |
|---|---|
| `--hostname HOST` | Override the Bitbucket host |

## MCP tools

| Tool | Description |
|---|---|
| `rerun_pipeline` | Re-run a pipeline at the same commit. Params: `repo` (required), `pipeline_uuid` (required), `hostname` |

## Backend details

**Cloud** — Fetches the source pipeline (`GET /2.0/repositories/{ws}/{slug}/pipelines/{uuid}`) to read the target ref and commit hash, then POSTs a new pipeline run (`POST /2.0/repositories/{ws}/{slug}/pipelines/`) with the same ref and commit. When the source pipeline has no commit hash (custom pipelines), falls back to a ref-only trigger. UUIDs are automatically wrapped in curly braces.

**Server/DC** — not supported. Returns a `host.unsupported` error.

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

---

## Pipeline Artifacts (Cloud only)

List and download per-step build artifacts declared via `artifacts:` in `bitbucket-pipelines.yml`.

```bash
# List artifacts for a step
bitbottle pipeline artifact list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID
bitbottle pipeline artifact list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID --json
bitbottle pipeline artifact list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID --limit 20

# Download an artifact (writes to current directory by default)
bitbottle pipeline artifact download PIPELINE_UUID WORKSPACE/REPO \
  --step STEP_UUID --name build.tar.gz

# Download to a specific path
bitbottle pipeline artifact download PIPELINE_UUID WORKSPACE/REPO \
  --step STEP_UUID --name build.tar.gz --out /tmp/build.tar.gz

# Write to stdout
bitbottle pipeline artifact download PIPELINE_UUID WORKSPACE/REPO \
  --step STEP_UUID --name build.tar.gz --out -
```

## Flags

| Command | Flag | Description |
|---|---|---|
| `artifact list` | `--step STEP_UUID` | Step UUID (required) |
| `artifact list` | `--limit N` | Max results (default 50) |
| `artifact list` | `--json` / `--jq` | JSON/jq output |
| `artifact download` | `--step STEP_UUID` | Step UUID (required) |
| `artifact download` | `--name FILE` | Artifact filename (required) |
| `artifact download` | `--out PATH` | Output path; `-` for stdout |

## MCP tools

| Tool | Description |
|---|---|
| `list_pipeline_artifacts` | List artifacts for a step. Params: `project`, `slug`, `pipeline_uuid`, `step_uuid` (all required), `limit`, `hostname` |
| `download_pipeline_artifact` | Download artifact as base64. Params: `project`, `slug`, `pipeline_uuid`, `step_uuid`, `name` (all required), `hostname`. Returns `{"name": "...", "content_base64": "..."}`. Artifacts > 5 MB return an error — use the CLI `--out PATH` instead. |

## Backend details

**Cloud** — List: `GET /repositories/{ws}/{slug}/pipelines/{pipeline_uuid}/steps/{step_uuid}/artifacts` (paginated). Download: `GET /repositories/{ws}/{slug}/pipelines/{pipeline_uuid}/steps/{step_uuid}/artifacts/{name}` (binary stream). UUIDs are automatically wrapped in curly braces.

**Server/DC** — not supported. Returns a `host.unsupported` error.

---

# Pipeline Test Reports

View JUnit test results for a Bitbucket Cloud pipeline step. This feature is **Cloud only**.

## Commands

```bash
# View test report summary for a pipeline step
bitbottle pipeline test-report view PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID

# List individual test cases
bitbottle pipeline test-case list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID

# Filter test cases by status
bitbottle pipeline test-case list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID --status FAILED

# Limit results
bitbottle pipeline test-case list PIPELINE_UUID WORKSPACE/REPO --step STEP_UUID --limit 20
```

`test-report view` and `test-case list` support `--json`, `--jq`, `--yaml`, and `--template`.

## Flags

| Command | Flag | Description |
|---|---|---|
| all | `--step STEP_UUID` | Step UUID (required) |
| all | `--hostname HOST` | Override the Bitbucket host |
| `test-case list` | `--status STATUS` | Filter by `PASSED`, `FAILED`, or `SKIPPED` |
| `test-case list` | `--limit N` | Max results (default 50) |

## JSON output

`test-report view` fields: `total` (int), `passed` (int), `failed` (int), `skipped` (int), `duration_ms` (int).

`test-case list` fields: `name` (string), `class_name` (string), `status` (string), `duration_ms` (int), `failure_message` (string, omitted when empty).

## MCP tools

| Tool | Description |
|---|---|
| `get_pipeline_test_report` | Get test report summary. Params: `project`, `slug`, `pipeline_uuid`, `step_uuid` (all required), `hostname` |
| `list_pipeline_test_cases` | List test cases. Params: `project`, `slug`, `pipeline_uuid`, `step_uuid` (all required), `status`, `limit`, `hostname` |

## Backend details

**Cloud** — Summary: `GET /2.0/repositories/{ws}/{slug}/pipelines/{uuid}/steps/{step_uuid}/test_reports`. Cases: `GET .../test_reports/test_cases` (paginated). UUIDs are automatically wrapped in curly braces. Duration is returned in seconds (`duration_in_seconds`) and converted to milliseconds.

**Server/DC** — not supported. Returns a `host.unsupported` error.
