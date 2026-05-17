# Scenario: Cloud pipeline lifecycle (trigger → list → watch → view → schedule)

**Backend:** Cloud only. Bitbucket Pipelines is a Cloud-exclusive feature;
Server/DC returns `host.unsupported` for these commands.

End-to-end: `pipeline trigger` → `pipeline list` → `pipeline watch` →
`pipeline view` → `pipeline steps` → `pipeline logs` → `pipeline schedule
{create,list,delete}`. Exercises real CI state and long-polling.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST` (Cloud).
- `BB_TEST_CLOUD_REPO` has **Pipelines enabled** and a valid
  `bitbucket-pipelines.yml` on `main` (a one-step `echo` pipeline is fine).
- A non-trivial pipeline definition that runs for ≥ 30s (so `watch` has
  time to observe a state transition).

### Minimal `bitbucket-pipelines.yml` for the scratch repo

```yaml
pipelines:
  default:
    - step:
        name: QA smoke
        script:
          - echo "manual test"
          - sleep 30
```

Commit and push it once before running this scenario.

## Setup

```bash
rm -rf /tmp/bb-pipe
bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-pipe
cd /tmp/bb-pipe
git checkout main
```

## Steps

### 1. `pipeline trigger` on `main`

```bash
bitbottle pipeline trigger "$BB_TEST_CLOUD_REPO" --branch main
```

Exit code: `0`. stdout/stderr prints the new pipeline UUID and a web URL.
Capture the UUID:

```bash
export PIPE_UUID=$(bitbottle pipeline list "$BB_TEST_CLOUD_REPO" \
  --json uuid --limit 1 | jq -r '.[0].uuid')
echo "PIPE_UUID=$PIPE_UUID"
```

`PIPE_UUID` is a UUID-shaped string (with curly braces on Cloud:
`{xxxxxxxx-...}`).

### 2. `pipeline trigger` with `--variable` (typed inline overrides)

```bash
bitbottle pipeline trigger "$BB_TEST_CLOUD_REPO" \
  --branch main \
  --variable QA_RUN=1 \
  --variable QA_LABEL=manual-smoke
```

Exit code: `0`. The new pipeline UUID is printed. Verify in the UI that
the two variables appear on the pipeline run.

### 3. `pipeline list` includes the new run

```bash
bitbottle pipeline list "$BB_TEST_CLOUD_REPO" --limit 5
```

Stdout is a table with columns `BUILD#`, `STATE`, `BRANCH`, `TRIGGER`,
`CREATED`. The most-recent row matches `main` and `manual`/`api` trigger
type. JSON variant:

```bash
bitbottle pipeline list "$BB_TEST_CLOUD_REPO" --json uuid,state,branch --limit 3
```

Parses as JSON; `.[0].state.name` is one of `PENDING`/`IN_PROGRESS`/
`SUCCESSFUL`/`FAILED`.

### 4. `pipeline watch` blocks until terminal

```bash
bitbottle pipeline watch "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" --interval 3
```

Output streams state transitions every 3 seconds. Exits `0` on
`SUCCESSFUL`, non-zero on `FAILED`/`STOPPED`. The final line names the
terminal state. Verify Ctrl-C interrupts cleanly with no stack trace.

### 5. `pipeline view` shows the terminal pipeline

```bash
bitbottle pipeline view "$BB_TEST_CLOUD_REPO" "$PIPE_UUID"
```

Stdout shows build #, state, branch, trigger, creator, and a step
summary. Exit code: `0`.

### 6. `pipeline view --web` opens the browser

```bash
bitbottle pipeline view "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" --web
```

A browser tab opens to the pipeline page on `$BB_TEST_CLOUD_HOST`. The URL
matches `https://$BB_TEST_CLOUD_HOST/<workspace>/<repo>/pipelines/results/<num>`.
Exit code: `0`.

### 7. `pipeline steps` enumerates steps

```bash
bitbottle pipeline steps "$BB_TEST_CLOUD_REPO" "$PIPE_UUID"
export STEP_UUID=$(bitbottle pipeline steps "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" \
  --json uuid | jq -r '.[0].uuid')
echo "STEP_UUID=$STEP_UUID"
```

Stdout lists one row per step (`UUID`, `NAME`, `STATE`, `DURATION`).
`STEP_UUID` is non-empty.

### 8. `pipeline logs` streams build output

```bash
bitbottle pipeline logs "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" "$STEP_UUID" | head -50
```

Stdout contains `manual test` (the echo from the `bitbucket-pipelines.yml`).
Exit code: `0`.

### 9. Bogus UUID gives a clear error

```bash
bitbottle pipeline view "$BB_TEST_CLOUD_REPO" "{00000000-0000-0000-0000-000000000000}"
```

Exit code: non-zero. stderr mentions "not found" with the typed
`ErrNotFound` wording — not a raw 404 dump.

### 10. `pipeline schedule create`

```bash
export SCHED_UUID=$(bitbottle pipeline schedule create "$BB_TEST_CLOUD_REPO" \
  --branch main \
  --cron "0 0 * * *" \
  --enabled=false \
  --json uuid 2>/dev/null | jq -r '.uuid' || \
  bitbottle pipeline schedule create "$BB_TEST_CLOUD_REPO" \
    --branch main --cron "0 0 * * *" --enabled=false 2>&1 \
  | grep -oE '[0-9a-f-]{36}' | head -1)
echo "SCHED_UUID=$SCHED_UUID"
```

Exit code: `0`. A schedule UUID is printed.

### 11. `pipeline schedule list` includes the new schedule

```bash
bitbottle pipeline schedule list "$BB_TEST_CLOUD_REPO" | grep "$SCHED_UUID"
```

`grep` exits `0`.

### 12. `pipeline schedule delete` removes it

```bash
bitbottle pipeline schedule delete "$BB_TEST_CLOUD_REPO" "$SCHED_UUID"
bitbottle pipeline schedule list "$BB_TEST_CLOUD_REPO" | grep "$SCHED_UUID" || echo "deleted"
```

Final line prints `deleted`. Exit code: `0`.

### 12.5. `pipeline cache list` (PIPELINE-CACHE)

```bash
bitbottle pipeline cache list "$BB_TEST_CLOUD_REPO"
```

Stdout is a table of cache entries with columns `UUID`, `NAME`, `FILE_SIZE`,
`CREATED`. If pipelines haven't populated any caches yet (a fresh repo
with no docker/npm/etc. caches), the list is empty — exit `0`. With
`--json`:

```bash
bitbottle pipeline cache list "$BB_TEST_CLOUD_REPO" --json uuid,name | jq 'length'
```

Output is a non-negative integer.

### 12.6. `pipeline cache delete` (skip if no caches)

```bash
CACHE_UUID=$(bitbottle pipeline cache list "$BB_TEST_CLOUD_REPO" \
  --json uuid | jq -r '.[0].uuid // empty')
if [ -n "$CACHE_UUID" ]; then
  bitbottle pipeline cache delete "$BB_TEST_CLOUD_REPO" "$CACHE_UUID"
  bitbottle pipeline cache list   "$BB_TEST_CLOUD_REPO" --json uuid \
    | jq 'map(.uuid) | index("'"$CACHE_UUID"'")'
fi
```

Final `jq` prints `null`. Skip cleanly if there were no caches to begin
with.

### 13. Pipeline commands on Server/DC return `host.unsupported`

```bash
bitbottle pipeline list "$BB_TEST_SERVER_REPO" --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr mentions "unsupported on host" (mapped from
`ErrUnsupportedOnHost`) — not an opaque 404. Skip this step if no Server
session is configured.

## Cleanup

```bash
# Optional — the test pipelines accumulate in the run history. They are
# harmless but you can stop in-flight runs from the UI if needed.
cd - >/dev/null
rm -rf /tmp/bb-pipe
```
