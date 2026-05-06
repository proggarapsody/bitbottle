# Scenario: Server/DC Webhooks

**Backend:** Bitbucket Server / Data Center.

Create, list, view, and delete a repository webhook on a self-hosted
Bitbucket instance. Confirm Server-specific event keys
(`repo:refs_changed`, `pr:opened`) are accepted, secrets nest under
`configuration.secret`, and IDs are numeric strings.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_REPO` exists. Token has `REPO_ADMIN` permission on the
  repository (required to manage webhooks).
- A test delivery URL:

```bash
export HOOK_URL="https://example.com/bb-server-webhook-test"
```

## Steps

### 1. `webhook create` with Server event keys

```bash
bitbottle webhook create "$BB_TEST_SERVER_REPO" \
  --url "$HOOK_URL" \
  --events 'repo:refs_changed,pr:opened'
```

Exit code: `0`. Stdout matches `Created webhook <id> -> $HOOK_URL`. The
ID is a small integer (not a UUID).

**Verify in UI:** Repository settings → Webhooks shows the new entry
subscribed to `Push` and `Pull request opened`.

### 2. Capture the ID

```bash
export HOOK_ID=$(bitbottle webhook list "$BB_TEST_SERVER_REPO" \
  --json id,url --jq ".[] | select(.url==\"$HOOK_URL\") | .id" \
  | tr -d '"')
echo "HOOK_ID=$HOOK_ID"
```

`HOOK_ID` is a small integer (e.g. `42`).

### 3. `webhook list` (TTY)

```bash
bitbottle webhook list "$BB_TEST_SERVER_REPO"
```

TTY table header `ID  URL  ACTIVE  EVENTS`. The row for `$HOOK_URL` shows
`true` in ACTIVE and a comma-joined event list.

### 4. `webhook view` by numeric ID

```bash
bitbottle webhook view "$BB_TEST_SERVER_REPO" "$HOOK_ID"
```

Stdout shows the hook's id, url, active, events. `"id"` is the same numeric
string as `$HOOK_ID`.

### 5. `webhook create` with `--secret -` from stdin

```bash
echo "shared-secret-bytes" | bitbottle webhook create \
  "$BB_TEST_SERVER_REPO" \
  --url "https://example.com/bb-server-secret-test" \
  --events 'repo:refs_changed' \
  --secret -
```

Exit code: `0`. **Critical:** `shared-secret-bytes` must NOT appear in
stdout/stderr/logs.

**Verify in UI:** the new webhook has the secret set (Bitbucket Server
shows a "Secret" placeholder; the actual value is never re-displayed).

### 6. `webhook list` does not surface secrets

```bash
bitbottle webhook list "$BB_TEST_SERVER_REPO" --json id,url,events
```

No `secret` field appears in the JSON output. The CLI's `Webhook` type
has no Secret field; the Server REST API does not include secrets in
read responses.

### 7. `webhook create --active=false`

```bash
bitbottle webhook create "$BB_TEST_SERVER_REPO" \
  --url "https://example.com/bb-server-disabled-test" \
  --events 'repo:refs_changed' \
  --active=false
```

Exit code: `0`. UI shows the new webhook in disabled state.

### 8. Cloud event keys are rejected by Server

```bash
bitbottle webhook create "$BB_TEST_SERVER_REPO" \
  --url "$HOOK_URL" \
  --events 'repo:push'
```

Exit code: non-zero. stderr surfaces the Server validation error
(`repo:push` is a Cloud event key; Server uses `repo:refs_changed`).
This is a backend-side rejection — the CLI does not validate event keys
client-side.

### 9. `webhook delete` requires `--confirm` non-interactively

```bash
bitbottle webhook delete "$BB_TEST_SERVER_REPO" "$HOOK_ID" </dev/null
```

Exit code: non-zero. stderr says `--confirm required when not running
interactively`.

### 10. `webhook delete --confirm`

```bash
bitbottle webhook delete "$BB_TEST_SERVER_REPO" "$HOOK_ID" --confirm
bitbottle webhook list "$BB_TEST_SERVER_REPO" \
  --json id --jq ".[] | select(.id==\"$HOOK_ID\") | .id" \
  | grep . || echo "gone"
```

Last line prints `gone`.

### 11. Deleting a non-existent webhook fails clearly

```bash
bitbottle webhook delete "$BB_TEST_SERVER_REPO" 99999999 --confirm
```

Exit code: non-zero. stderr names the ID and says it was not found.

## Cleanup

```bash
bitbottle webhook list "$BB_TEST_SERVER_REPO" --json id,url \
  | jq -r '.[] | select(.url|test("bb-server-(webhook|secret|disabled)-test"))| .id' \
  | xargs -I{} bitbottle webhook delete "$BB_TEST_SERVER_REPO" {} --confirm \
      2>/dev/null || true
```
