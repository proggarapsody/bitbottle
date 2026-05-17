# Scenario: Cloud deployment + environment + deployment-scoped variables

**Backend:** Cloud only. Bitbucket Deployments is a Cloud-exclusive feature;
Server/DC returns `host.unsupported`.

End-to-end of the **DEP** scope plus the deployment-scope path of the
**VAR** scope: `environment create` → `variable set --scope deployment` →
`variable list --scope deployment` → trigger a deployment pipeline →
`deployment list` → `deployment view` → cleanup.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST` (Cloud).
- `BB_TEST_CLOUD_REPO` has Pipelines enabled.
- `bitbucket-pipelines.yml` on `main` contains a deployment step. Example:

  ```yaml
  pipelines:
    default:
      - step:
          name: build
          script:
            - echo "build artifact"
      - step:
          name: deploy
          deployment: qa-manual
          script:
            - echo "deploying to $BITBUCKET_DEPLOYMENT_ENVIRONMENT"
            - echo "secret=$QA_SECRET"
  ```

  The deployment slug must match the environment created in step 1
  (`qa-manual`). Commit and push it before running this scenario.

- `jq` available locally.

## Setup

```bash
rm -rf /tmp/bb-deploy
bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-deploy
cd /tmp/bb-deploy
```

## Steps

### 1. `environment create`

```bash
bitbottle environment create "$BB_TEST_CLOUD_REPO" \
  --name qa-manual --type Test
```

Exit code: `0`. stdout includes the new environment UUID (Cloud format,
`{xxxxxxxx-...}`). Capture it:

```bash
export ENV_UUID=$(bitbottle environment list "$BB_TEST_CLOUD_REPO" --json uuid,name \
  | jq -r '.[] | select(.name=="qa-manual") | .uuid')
echo "ENV_UUID=$ENV_UUID"
```

### 2. `environment list` includes it

```bash
bitbottle environment list "$BB_TEST_CLOUD_REPO" | grep qa-manual
```

`grep` exits `0`. Tabular columns: `NAME`, `TYPE`, `UUID`, `CATEGORY` (or
similar — record the exact header order).

### 3. `variable set --scope deployment --env <UUID>` (plain)

```bash
bitbottle variable set "$BB_TEST_CLOUD_REPO" QA_LABEL \
  --body "manual-smoke" \
  --scope deployment --env "$ENV_UUID"
```

Exit code: `0`. stderr confirms creation.

### 4. `variable set --scope deployment --secured` (secret)

```bash
bitbottle variable set "$BB_TEST_CLOUD_REPO" QA_SECRET \
  --body "super-secret-value" \
  --scope deployment --env "$ENV_UUID" --secured
```

Exit code: `0`. stderr confirms creation.

### 5. `variable list --scope deployment --env <UUID>`

```bash
bitbottle variable list "$BB_TEST_CLOUD_REPO" \
  --scope deployment --env "$ENV_UUID"
```

Stdout lists both variables. The `QA_SECRET` value column shows `<secured>`
or `*****` — never the raw value. The `QA_LABEL` value shows
`manual-smoke`.

### 5.5. `variable set --scope repository` (VAR — basic CI variable)

```bash
bitbottle variable set "$BB_TEST_CLOUD_REPO" QA_REPO_VAR \
  --body "repo-scope" \
  --scope repository
bitbottle variable list "$BB_TEST_CLOUD_REPO" --scope repository | grep QA_REPO_VAR
```

Exit `0`. `grep` exits `0`. Repo-scope variables apply to every pipeline
run regardless of environment.

### 5.6. `variable set --scope workspace` (VAR — workspace-wide)

```bash
bitbottle variable set "$BB_TEST_CLOUD_REPO" QA_WS_VAR \
  --body "ws-scope" \
  --scope workspace
bitbottle variable list "$BB_TEST_CLOUD_REPO" --scope workspace | grep QA_WS_VAR
```

Exit `0`. `grep` exits `0`. Note: requires workspace-admin permission;
record permission errors verbatim if they occur (typed `ErrPermission`).

### 5.7. `variable set --body -` reads value from stdin

```bash
echo "from-stdin" | bitbottle variable set "$BB_TEST_CLOUD_REPO" QA_STDIN \
  --body - --scope repository
bitbottle variable list "$BB_TEST_CLOUD_REPO" --scope repository \
  --json key,value | jq '.[] | select(.key=="QA_STDIN").value'
```

Final `jq` prints `"from-stdin"`. Useful for piping secrets without
echoing them in shell history.

### 6. Missing `--env` on deployment scope errors clearly

```bash
bitbottle variable list "$BB_TEST_CLOUD_REPO" --scope deployment
```

Exit code: non-zero. stderr mentions `--env` is required when
`--scope=deployment`. Not a raw 400.

### 7. Trigger a pipeline that runs the deploy step

```bash
bitbottle pipeline trigger "$BB_TEST_CLOUD_REPO" --branch main
export PIPE_UUID=$(bitbottle pipeline list "$BB_TEST_CLOUD_REPO" \
  --json uuid --limit 1 | jq -r '.[0].uuid')
bitbottle pipeline watch "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" --interval 5
```

The watch exits when the pipeline reaches a terminal state. Pipeline
should be `SUCCESSFUL` (or `FAILED` — investigate, but the manual-test
purpose is the deployment recording, not pipeline success).

### 8. `deployment list` includes the new deployment

```bash
bitbottle deployment list "$BB_TEST_CLOUD_REPO" --limit 5
```

Stdout is a table with columns including `UUID`, `ENVIRONMENT`, `STATE`,
`CREATED`. The most-recent row references `qa-manual` and a state in
`{IN_PROGRESS, COMPLETED, FAILED, STOPPED}`.

Capture the deployment UUID:

```bash
export DEPLOY_UUID=$(bitbottle deployment list "$BB_TEST_CLOUD_REPO" \
  --json uuid,environment --limit 5 \
  | jq -r '.[] | select(.environment.name=="qa-manual") | .uuid' | head -1)
echo "DEPLOY_UUID=$DEPLOY_UUID"
```

### 9. `deployment view`

```bash
bitbottle deployment view "$BB_TEST_CLOUD_REPO" "$DEPLOY_UUID"
```

Stdout shows the deployment's environment (`qa-manual`), commit, state, and
trigger. Exit code: `0`.

### 10. Pipeline logs prove the variables were injected

```bash
export STEP_UUID=$(bitbottle pipeline steps "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" \
  --json uuid,name | jq -r '.[] | select(.name=="deploy") | .uuid')
bitbottle pipeline logs "$BB_TEST_CLOUD_REPO" "$PIPE_UUID" "$STEP_UUID" | head -40
```

The log contains `deploying to qa-manual` (plain var was expanded) and
`secret=*****` or `secret=` with the value masked by Bitbucket
(secured vars are masked in pipeline output).

### 11. Bogus environment UUID gives a clear error

```bash
bitbottle deployment list "$BB_TEST_CLOUD_REPO" --json uuid \
  | jq '.[0]' >/dev/null  # sanity: list works
bitbottle variable list "$BB_TEST_CLOUD_REPO" \
  --scope deployment --env "{00000000-0000-0000-0000-000000000000}"
```

Exit code: non-zero. stderr mentions "not found" or "no such environment"
— typed `ErrNotFound`, not a raw 404.

### 12. `environment list` on Server/DC returns `host.unsupported`

```bash
bitbottle environment list "$BB_TEST_SERVER_REPO" \
  --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr mentions "unsupported on host". Skip if no
Server session.

## Cleanup

```bash
# Remove the two variables.
bitbottle variable delete "$BB_TEST_CLOUD_REPO" QA_LABEL \
  --scope deployment --env "$ENV_UUID" --confirm
bitbottle variable delete "$BB_TEST_CLOUD_REPO" QA_SECRET \
  --scope deployment --env "$ENV_UUID" --confirm

# Remove the environment.
bitbottle environment delete "$BB_TEST_CLOUD_REPO" "$ENV_UUID" --confirm

# Remove the repository/workspace-scope test variables.
bitbottle variable delete "$BB_TEST_CLOUD_REPO" QA_REPO_VAR --scope repository --confirm 2>/dev/null || true
bitbottle variable delete "$BB_TEST_CLOUD_REPO" QA_STDIN    --scope repository --confirm 2>/dev/null || true
bitbottle variable delete "$BB_TEST_CLOUD_REPO" QA_WS_VAR   --scope workspace  --confirm 2>/dev/null || true

cd - >/dev/null
rm -rf /tmp/bb-deploy
```
