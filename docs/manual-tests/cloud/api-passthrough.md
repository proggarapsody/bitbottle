# Scenario: Cloud `api` passthrough

**Backend:** Cloud. Path prefix: `2.0/`.

Verifies the generic REST escape hatch — GET, `--paginate`, `--jq`, `-F`,
`-X`, variable expansion.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- Inside a local clone of `$BB_TEST_CLOUD_REPO` (so `{workspace}` /
  `{repo_slug}` expand from the git remote):
  ```bash
  cd /tmp/bb-cloud-qa  # or any clone
  ```

## Steps

### 1. GET current user

```bash
bitbottle api 2.0/user --jq .username
```

Stdout: a single line with your username. Exit code: `0`.

### 2. Variable expansion `{workspace}` / `{repo_slug}`

```bash
bitbottle api '2.0/repositories/{workspace}/{repo_slug}' --jq .full_name
```

Stdout: `<workspace>/<slug>` matching `$BB_TEST_CLOUD_REPO`. Exit code: `0`.

### 3. `--paginate` walks `next` URLs

```bash
bitbottle api --paginate --jq '.[].full_name' '2.0/repositories/{workspace}' \
  | wc -l
```

Numeric count > 0. Higher than `pagelen` default (i.e. you crossed at
least one page boundary if the workspace has > 10 repos).

### 4. `-X POST` with `-F` body fields creates a PR

```bash
cd /tmp/bb-cloud-qa
git checkout -b qa/api-passthrough origin/main
echo "$(date)" >> API_TEST.txt
git add API_TEST.txt
git commit -m "qa: api passthrough"
git push -u origin qa/api-passthrough

bitbottle api -X POST \
  -F 'title=QA: api-passthrough' \
  -F 'source.branch.name=qa/api-passthrough' \
  -F 'destination.branch.name=main' \
  '2.0/repositories/{workspace}/{repo_slug}/pullrequests' \
  --jq .id
```

Stdout: a positive integer PR id. Exit code: `0`. Capture as `$API_PR_ID`.

### 5. `-F` auto-types booleans / numbers

```bash
bitbottle api -X POST \
  -F 'title=QA: api typed' \
  -F 'source.branch.name=qa/api-passthrough' \
  -F 'destination.branch.name=main' \
  -F 'close_source_branch=true' \
  '2.0/repositories/{workspace}/{repo_slug}/pullrequests' --jq .close_source_branch \
  || echo "second-PR creation failed (expected if a PR for this branch is already open)"
```

If it succeeds, stdout shows `true` (boolean) — confirming `-F` did not
quote it as the string `"true"`.

### 6. `-f` (raw-field) sends as string

```bash
bitbottle api -X PUT \
  -f 'title=QA: api passthrough (raw string)' \
  "2.0/repositories/{workspace}/{repo_slug}/pullrequests/$API_PR_ID" \
  --jq .title
```

Stdout exactly `QA: api passthrough (raw string)`. Exit code: `0`.

### 7. `--input -` streams a body from stdin

```bash
printf '{"title":"QA: api stdin body"}' \
  | bitbottle api -X PUT --input - \
      "2.0/repositories/{workspace}/{repo_slug}/pullrequests/$API_PR_ID" \
      --jq .title
```

Stdout: `QA: api stdin body`. Exit code: `0`.

### 8. `--header`

```bash
bitbottle api -H 'Accept: application/json' 2.0/user --jq .username
```

Same output as step 1. Exit code: `0`.

### 9. Bad endpoint returns a clean error

```bash
bitbottle api 2.0/this/path/does/not/exist
```

Exit code: non-zero. stderr includes the HTTP status (e.g. `404`) and the
error body.

### 10. `--hostname` overrides default-host

```bash
bitbottle api --hostname "$BB_TEST_CLOUD_HOST" 2.0/user --jq .username
```

Same output as step 1.

## Cleanup

```bash
bitbottle pr decline "$API_PR_ID" 2>/dev/null || true
bitbottle branch delete "$BB_TEST_CLOUD_REPO" qa/api-passthrough 2>/dev/null || true
```
