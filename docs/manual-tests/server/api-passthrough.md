# Scenario: Server/DC `api` passthrough

**Backend:** Server / Data Center. Path prefix: `rest/api/1.0/`.

Mirror of `cloud/api-passthrough.md`. Critical differences:

- Path prefix is `rest/api/1.0/`, not `2.0/`.
- Pagination uses `nextPageStart` / `isLastPage`, not `next` URLs.
- Variable expansion: `{project}` and `{slug}` (Cloud's `{workspace}` /
  `{repo_slug}` are aliases — both must work).

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- Inside a local clone of `$BB_TEST_SERVER_REPO`:
  ```bash
  cd /tmp/bb-server-qa
  ```

## Steps

### 1. GET current user

```bash
bitbottle api 'rest/api/1.0/users/{username}' --jq .name 2>/dev/null \
  || bitbottle api 'rest/api/1.0/application-properties' --jq .version
```

Stdout is a single non-empty line. Exit `0`.

> The exact "current user" endpoint differs by server version. The
> `application-properties` endpoint is universally available and serves as
> a connectivity check.

### 2. Variable expansion `{project}` / `{slug}`

```bash
bitbottle api 'rest/api/1.0/projects/{project}/repos/{slug}' --jq .slug
```

Stdout: the slug from `$BB_TEST_SERVER_REPO`.

### 3. Cloud-style aliases `{workspace}` / `{repo_slug}` also work

```bash
bitbottle api 'rest/api/1.0/projects/{workspace}/repos/{repo_slug}' --jq .slug
```

Same output as step 2 — confirms aliases are honored on Server/DC.

### 4. `--paginate` walks `nextPageStart`

```bash
bitbottle api --paginate \
  --jq '.[].name' \
  'rest/api/1.0/projects/{project}/repos' | wc -l
```

Numeric count > 0. If the project has > 25 repos (default page size),
expect to cross at least one page boundary.

### 5. `-X POST` with `-F` body

```bash
git checkout -b qa/api-passthrough "origin/$(git symbolic-ref --short HEAD)"
echo "$(date)" >> API_TEST.txt
git add API_TEST.txt
git commit -m "qa: api passthrough"
git push -u origin qa/api-passthrough

bitbottle api -X POST \
  -F 'title=QA: api-passthrough' \
  -F 'fromRef.id=refs/heads/qa/api-passthrough' \
  -F "toRef.id=refs/heads/$(git symbolic-ref --short HEAD)" \
  'rest/api/1.0/projects/{project}/repos/{slug}/pull-requests' \
  --jq .id
```

Stdout: a positive integer PR id. Capture as `$API_PR_ID`. Exit `0`.

### 6. `-f` raw string field

```bash
bitbottle api -X PUT \
  -f 'title=QA: api passthrough (raw string)' \
  -F "version=$(bitbottle api "rest/api/1.0/projects/{project}/repos/{slug}/pull-requests/$API_PR_ID" --jq .version)" \
  -F "id=$API_PR_ID" \
  "rest/api/1.0/projects/{project}/repos/{slug}/pull-requests/$API_PR_ID" \
  --jq .title
```

Stdout: `QA: api passthrough (raw string)`. (Server/DC PR-update payloads
must include the current `version` — that is why the inline `api` call
fetches it; this also re-verifies GET works.)

### 7. `--input -` from stdin

```bash
VERSION=$(bitbottle api "rest/api/1.0/projects/{project}/repos/{slug}/pull-requests/$API_PR_ID" --jq .version)
printf '{"id":%s,"version":%s,"title":"QA: api stdin body"}' "$API_PR_ID" "$VERSION" \
  | bitbottle api -X PUT --input - \
      "rest/api/1.0/projects/{project}/repos/{slug}/pull-requests/$API_PR_ID" \
      --jq .title
```

Stdout: `QA: api stdin body`.

### 8. `--header`

```bash
bitbottle api -H 'Accept: application/json' \
  'rest/api/1.0/projects/{project}/repos/{slug}' --jq .slug
```

Same as step 2.

### 9. Bad endpoint returns clean error

```bash
bitbottle api 'rest/api/1.0/this/path/does/not/exist'
```

Exit non-zero. stderr includes HTTP status (`404`) and error body.

### 10. `--hostname` override

```bash
bitbottle api --hostname "$BB_TEST_SERVER_HOST" \
  'rest/api/1.0/projects/{project}/repos/{slug}' --jq .slug
```

Same as step 2.

## Cleanup

```bash
bitbottle pr decline "$API_PR_ID" 2>/dev/null || true
bitbottle branch delete "$BB_TEST_SERVER_REPO" qa/api-passthrough 2>/dev/null || true
```
