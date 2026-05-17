# Scenario: Raw `api` passthrough (Cloud + Server/DC)

**Backend:** Both. Each variant is run against the matching host.

Exercises the `api` escape hatch: variable expansion, typed body fields,
header injection, pagination, and stdin-fed raw bodies. This command bypasses
the typed adapters and goes straight to the REST API — it's the surface most
likely to drift if the underlying httpx transport changes.

## Prerequisites

- Logged in to both `$BB_TEST_CLOUD_HOST` and `$BB_TEST_SERVER_HOST`.
- `BB_TEST_CLOUD_REPO` (`<workspace>/<repo>`) and `BB_TEST_SERVER_REPO`
  (`<projectKey>/<repo>`) exist as disposable scratch repos.
- `jq` available locally.

## Steps — Cloud

### 1. Basic GET with no expansion

```bash
bitbottle api user --hostname "$BB_TEST_CLOUD_HOST" | jq .uuid
```

Stdout is a UUID-shaped string (with braces). Exit code: `0`.

### 2. `{workspace}` / `{repo_slug}` variable expansion

Inside a clone of `$BB_TEST_CLOUD_REPO`:

```bash
rm -rf /tmp/bb-api-cloud
bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-api-cloud
cd /tmp/bb-api-cloud

bitbottle api 'repositories/{workspace}/{repo_slug}' | jq -r '.full_name'
```

Stdout equals `$BB_TEST_CLOUD_REPO`. Exit code: `0`.

### 3. `--jq` filter applies server-side response shaping

```bash
bitbottle api 'repositories/{workspace}/{repo_slug}' --jq .uuid
```

Stdout is a single UUID. Equivalent to piping through `jq`, but exercises
the embedded filter path.

### 4. `--paginate` follows `next` URLs

```bash
bitbottle api 'repositories/{workspace}?pagelen=2' --paginate \
  --jq '.values | length'
```

`--paginate` should follow Cloud's `next` cursor. The summed length is
≥ 2 and matches the workspace's total visible repo count.

### 5. `-X POST` with `-F` typed fields creates a resource

Create a repo issue (Cloud only; issues are the cheapest disposable
resource via raw API):

```bash
bitbottle api 'repositories/{workspace}/{repo_slug}/issues' \
  -X POST \
  -F 'title=qa: api passthrough' \
  -F 'priority=trivial' \
  -F 'kind=task' | jq -r '.id'
```

Stdout is the new issue's integer ID. Exit code: `0`. Capture it for
cleanup:

```bash
export ISSUE_ID=$(bitbottle api 'repositories/{workspace}/{repo_slug}/issues?q=title%3D%22qa%3A+api+passthrough%22' \
  --jq '.values[0].id')
echo "ISSUE_ID=$ISSUE_ID"
```

### 6. `-F` auto-detects boolean / numeric types

```bash
bitbottle api 'repositories/{workspace}/{repo_slug}/issues' \
  -X POST \
  -F 'title=qa: typed fields' \
  -F 'kind=bug' \
  -F 'priority=trivial' \
  -F 'votes=0' | jq '.votes,.kind'
```

`.votes` parses as a JSON number `0` (not the string `"0"`). `.kind` is
`"bug"`. Capture the ID and delete it in cleanup.

### 7. `-f` forces string typing (defeats auto-detection)

```bash
bitbottle api 'repositories/{workspace}/{repo_slug}/issues' \
  -X POST \
  -F 'title=qa: forced strings' \
  -f 'priority=trivial' \
  -f 'kind=task' | jq -r '.priority'
```

Exit code: `0`. The `-f` form transmitted `"trivial"` as a string (which
the server accepts here, since these enum fields ARE strings).

### 8. `--input -` reads stdin as the raw body

```bash
cat <<'EOF' | bitbottle api 'repositories/{workspace}/{repo_slug}/issues' \
  -X POST \
  -H 'Content-Type: application/json' \
  --input -
{"title":"qa: stdin body","kind":"task","priority":"trivial"}
EOF
```

Stdout is the created issue JSON. Exit code: `0`. Capture/delete in
cleanup.

### 9. `-H` injects an extra request header

```bash
bitbottle api user --hostname "$BB_TEST_CLOUD_HOST" \
  -H 'X-QA-Trace: manual-smoke' >/dev/null
```

Exit code: `0`. (No way to assert the header reached the server from the
CLI side; the test is that the flag does not break the request.)

### 10. Bad endpoint surfaces a clear error

```bash
bitbottle api 'this/does/not/exist' --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: non-zero. stderr mentions a 404 with the typed `ErrNotFound`
wording. Exit code is non-zero.

### 11. `--hostname` overrides the auto-detected host

```bash
cd /tmp  # outside a git repo
bitbottle api user --hostname "$BB_TEST_CLOUD_HOST" | jq -r '.account_id // .uuid'
```

Exit code: `0`. The response identifies the Cloud user. Run again with
`--hostname "$BB_TEST_SERVER_HOST"` and verify it returns a different
shape (Server: `{name, emailAddress, ...}`; Cloud: `{username, uuid, ...}`).

## Steps — Server / DC

### 12. Server: `{project}` / `{slug}` expansion

Inside a clone of `$BB_TEST_SERVER_REPO`:

```bash
rm -rf /tmp/bb-api-server
git clone "https://$BB_TEST_SERVER_HOST/scm/${BB_TEST_SERVER_REPO/\//\/}.git" /tmp/bb-api-server
cd /tmp/bb-api-server

bitbottle api 'rest/api/1.0/projects/{project}/repos/{slug}' \
  --hostname "$BB_TEST_SERVER_HOST" --jq '.slug'
```

Stdout equals the repo slug from `$BB_TEST_SERVER_REPO`. Exit code: `0`.

### 13. Server: pagination uses `start` / `limit` not `next`

```bash
bitbottle api 'rest/api/1.0/projects/{project}/repos/{slug}/branches?limit=2' \
  --hostname "$BB_TEST_SERVER_HOST" --paginate \
  --jq '.values | length'
```

The summed length is ≥ 2 and equals the actual branch count of the
scratch repo. Exit code: `0`. (Server APIs that don't support pagination
should still return exit `0` for one page.)

### 14. Server: typed POST against a real endpoint

Pick a low-risk write — adding a repo label is safe and reversible:

```bash
bitbottle api 'rest/api/1.0/projects/{project}/repos/{slug}/labels' \
  -X POST \
  --hostname "$BB_TEST_SERVER_HOST" \
  -F 'name=qa-manual'
```

Exit code: `0` (or a typed error if labels aren't enabled — record). The
JSON response includes `"name": "qa-manual"`. Cleanup deletes it.

## Cleanup

```bash
# Cloud: delete any test issues created above (loop over titles starting
# with "qa:").
for ID in $(bitbottle api 'repositories/{workspace}/{repo_slug}/issues?q=title~%22qa%3A%22' \
  --hostname "$BB_TEST_CLOUD_HOST" --jq '.values[].id' 2>/dev/null); do
  bitbottle api "repositories/{workspace}/{repo_slug}/issues/$ID" \
    -X DELETE --hostname "$BB_TEST_CLOUD_HOST" >/dev/null
done

# Server: delete the label.
bitbottle api 'rest/api/1.0/projects/{project}/repos/{slug}/labels/qa-manual' \
  -X DELETE --hostname "$BB_TEST_SERVER_HOST" 2>/dev/null || true

cd - >/dev/null
rm -rf /tmp/bb-api-cloud /tmp/bb-api-server
```
