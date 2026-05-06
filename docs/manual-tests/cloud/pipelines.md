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

### 9. `pipeline steps` lists steps with state and duration

Wait until the run from step 1 has produced at least one step (poll
`pipeline view` until `state` is `IN_PROGRESS` or later, then capture
the first step UUID):

```bash
bitbottle pipeline steps "$BB_TEST_CLOUD_REPO" "$PUUID"
export STEP_UUID=$(bitbottle pipeline steps "$BB_TEST_CLOUD_REPO" "$PUUID" \
  --json uuid --jq '.[0].uuid')
echo "STEP_UUID=$STEP_UUID"
```

TTY table shows `UUID … NAME … STATE … DURATION` columns. The first row
matches the build step from `bitbucket-pipelines.yml`.

### 10. `pipeline steps --json`

```bash
bitbottle pipeline steps "$BB_TEST_CLOUD_REPO" "$PUUID" \
  --json name,state,duration | jq '.[0] | keys | sort'
```

Stdout: `["duration","name","state"]`.

### 11. `pipeline logs` streams plaintext to stdout

```bash
bitbottle pipeline logs "$BB_TEST_CLOUD_REPO" "$PUUID" "$STEP_UUID" | head -20
```

Output is the raw step log (matches what the Bitbucket UI shows under
"View raw"). No table, no JSON, no colour codes when piped.

### 12. `pipeline logs` against a bogus step UUID fails clearly

```bash
bitbottle pipeline logs "$BB_TEST_CLOUD_REPO" "$PUUID" \
  00000000-0000-0000-0000-000000000000
```

Exit code: non-zero. stderr says the log/step was not found.

### 13. `pipeline variable set` creates a variable

```bash
export VKEY="QA_TEST_$(date +%s)"
bitbottle pipeline variable set "$BB_TEST_CLOUD_REPO" "$VKEY" "hello"
```

Stdout: `Set variable $VKEY`. Exit code: `0`.

### 14. `pipeline variable list` shows it (unsecured value visible)

```bash
bitbottle pipeline variable list "$BB_TEST_CLOUD_REPO" \
  | awk -v k="$VKEY" '$1==k'
```

One line, `$VKEY  hello  false` (TTY table). Verify in UI: variable
exists with value `hello`.

### 15. `pipeline variable set --secured` upserts and redacts on read

```bash
export SKEY="QA_SECRET_$(date +%s)"
echo "ssh-key-content" | bitbottle pipeline variable set \
  "$BB_TEST_CLOUD_REPO" "$SKEY" --body=- --secured
```

Stdout: `Set secured variable $SKEY`. Exit code: `0`.

```bash
bitbottle pipeline variable list "$BB_TEST_CLOUD_REPO" \
  | awk -v k="$SKEY" '$1==k'
```

Line shows `$SKEY  <secured>  true`. **Critical:** the actual secret
bytes must NOT appear anywhere in stdout / stderr / log files.

### 16. `pipeline variable list --json` redacts secured values via the same chokepoint

```bash
bitbottle pipeline variable list "$BB_TEST_CLOUD_REPO" \
  --json key,value,secured \
  --jq ".[] | select(.key==\"$SKEY\")"
```

JSON shows `"value":"<secured>"` and `"secured":true`. Plain-text secret
bytes must NOT appear.

### 17. `pipeline variable set` is idempotent (upsert)

```bash
bitbottle pipeline variable set "$BB_TEST_CLOUD_REPO" "$VKEY" "world"
bitbottle pipeline variable list "$BB_TEST_CLOUD_REPO" \
  --json key,value --jq ".[] | select(.key==\"$VKEY\") | .value"
```

Stdout: `"world"`. The variable was updated, not duplicated.

### 18. `pipeline variable delete` requires `--confirm` non-interactively

```bash
bitbottle pipeline variable delete "$BB_TEST_CLOUD_REPO" "$VKEY" </dev/null
```

Exit code: non-zero. stderr says `--confirm required when not running
interactively`.

### 19. `pipeline variable delete --confirm` removes it

```bash
bitbottle pipeline variable delete "$BB_TEST_CLOUD_REPO" "$VKEY" --confirm
bitbottle pipeline variable delete "$BB_TEST_CLOUD_REPO" "$SKEY" --confirm
bitbottle pipeline variable list "$BB_TEST_CLOUD_REPO" \
  --json key --jq '.[].key' | grep -E "^($VKEY|$SKEY)$" || echo "gone"
```

Last line prints `gone`. Both variables are removed.

### 20. Deleting a non-existent variable fails clearly

```bash
bitbottle pipeline variable delete "$BB_TEST_CLOUD_REPO" \
  QA_NEVER_EXISTED --confirm
```

Exit code: non-zero. stderr names the key and says it was not found.

## Cleanup

Pipelines auto-complete; nothing to remove. Variables created above are
already deleted in steps 19. If steps 13–17 left a variable behind on a
failure path, clean up manually:

```bash
bitbottle pipeline variable delete "$BB_TEST_CLOUD_REPO" "$VKEY" --confirm 2>/dev/null || true
bitbottle pipeline variable delete "$BB_TEST_CLOUD_REPO" "$SKEY" --confirm 2>/dev/null || true
```
