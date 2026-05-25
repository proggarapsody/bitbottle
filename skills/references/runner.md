# Pipeline Runners (Cloud only)

Self-hosted runners allow Bitbucket Cloud Pipelines to execute steps on your
own infrastructure. Runners are scoped to a workspace. Commands in this group
return a typed `host.unsupported` error against Server/DC hosts.

## Commands

```bash
# List workspace runners
bitbottle runner list WORKSPACE
bitbottle runner list                # inferred from pinned repo
bitbottle runner list myworkspace --json

# Create a runner
bitbottle runner create WORKSPACE --name NAME [--platform PLATFORM] [--label LABEL ...]
bitbottle runner create             --name my-runner                 # inferred workspace
bitbottle runner create myworkspace --name my-runner --platform linux_arm64 --label self.hosted --label linux

# Delete a runner
bitbottle runner delete WORKSPACE RUNNER_UUID
bitbottle runner delete RUNNER_UUID              # inferred workspace
```

## Flags

| Command | Flag | Default | Description |
|---|---|---|---|
| `runner create` | `--name NAME` | — | Runner name (required) |
| `runner create` | `--platform PLATFORM` | `linux_amd64` | Runner platform (see below) |
| `runner create` | `--label LABEL` | — | Runner label (repeatable; comma-separated values also accepted) |
| `runner create` | `--hostname HOST` | — | Override the Bitbucket host |
| `runner list` | `--hostname HOST` | — | Override the Bitbucket host |
| `runner delete` | `--hostname HOST` | — | Override the Bitbucket host |
| all | `--json` | — | JSON output |
| all | `--jq EXPR` | — | Filter JSON with jq expression |

## Platforms

| Value | OS | Architecture |
|---|---|---|
| `linux_amd64` | Linux | x86-64 |
| `linux_arm64` | Linux | ARM 64-bit |
| `windows_amd64` | Windows | x86-64 |
| `macos_arm64` | macOS | ARM 64-bit (Apple Silicon) |

## Output fields

| Field | JSON key | Notes |
|---|---|---|
| UUID | `uuid` | Runner UUID |
| NAME | `name` | Runner display name |
| STATE | `state` | `ONLINE`, `OFFLINE`, `DISABLED` |
| PLATFORM | `platform` | `operating` + `arch` sub-fields |
| LABELS | `labels` | Array of label strings |

## MCP tools

- `list_runners` — `workspace` (required), `hostname`
- `create_runner` — `workspace` (required), `name` (required), `platform` (default `linux_amd64`), `labels` (array), `hostname`
- `delete_runner` — `workspace` (required), `uuid` (required), `hostname`
