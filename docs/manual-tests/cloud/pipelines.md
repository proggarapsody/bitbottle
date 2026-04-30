# Scenario: Cloud Pipelines

**Backend:** Cloud only.

Trigger a pipeline, list pipelines, view one.

## Prerequisites

- `$BB_TEST_CLOUD_REPO` has Pipelines enabled and a `bitbucket-pipelines.yml`
  on `main` defining at least one default step.
- Token has `pipeline:write` scope.

## Steps

### 1. `pipeline run --branch main`

```bash
bitbottle pipeline run "$BB_TEST_CLOUD_REPO" --branch main
```

Exit code: `0`. stdout/stderr prints the new build number and a web URL.
Capture the UUID:

```bash
sleep 5
export PUUID=$(bitbottle pipeline list "$BB_TEST_CLOUD_REPO" --limit 1 \
  --json uuid --jq '.[].uuid')
echo "PUUID=$PUUID"
```

### 2. `pipeline list` shows the new run

```bash
bitbottle pipeline list "$BB_TEST_CLOUD_REPO" --limit 5
```

TTY table header `BUILD … STATE … BRANCH/TAG … DURATION`. The first row
is the run we just triggered (state likely `PENDING` or `IN_PROGRESS`).

### 3. `pipeline list --json`

```bash
bitbottle pipeline list "$BB_TEST_CLOUD_REPO" --limit 5 \
  --json buildNumber,state,refName,duration | jq '.[0] | keys | sort'
```

Stdout: `["buildNumber","duration","refName","state"]`.

### 4. `pipeline list --jq` filtering for state

```bash
bitbottle pipeline list "$BB_TEST_CLOUD_REPO" --limit 20 \
  --json buildNumber,state --jq '.[] | select(.state=="FAILED") | .buildNumber'
```

Either prints zero or more build numbers. Exit code: `0` regardless of
whether the filter matched.

### 5. `pipeline view` by UUID

```bash
bitbottle pipeline view "$BB_TEST_CLOUD_REPO" "$PUUID"
```

Stdout includes build number, state, ref, duration, and a web URL line.

### 6. `pipeline view --web`

```bash
bitbottle pipeline view "$BB_TEST_CLOUD_REPO" "$PUUID" --web
```

Browser opens at the pipeline run page. Exit code: `0`.

### 7. `pipeline run` without `--branch` is rejected

```bash
bitbottle pipeline run "$BB_TEST_CLOUD_REPO"
```

Exit code: non-zero. stderr names `--branch` as required.

### 8. `pipeline view` of a bogus UUID fails clearly

```bash
bitbottle pipeline view "$BB_TEST_CLOUD_REPO" 00000000-0000-0000-0000-000000000000
```

Exit code: non-zero. stderr says the pipeline was not found.

## Cleanup

Pipelines auto-complete; nothing to remove.
