# Scenario: Cloud Webhooks

**Backend:** Cloud.

Create, list, view, and delete a repository webhook on Bitbucket Cloud.
Confirm event keys, secret stdin path, and secret redaction on read.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_REPO` exists. Token has `webhook` scope (and `repository`).
- A test delivery URL that won't actually do anything dangerous when hit:

```bash
export HOOK_URL="https://example.com/bb-webhook-test"
```

## Steps

### 1. `webhook create` with required flags

```bash
HOOK_ID=$(bitbottle webhook create "$BB_TEST_CLOUD_REPO" \
  --url "$HOOK_URL" \
  --events 'repo:push,pullrequest:created' \
  --json id --jq '.id' 2>/dev/null)
# fallback: parse from text output if --json not honored on create
[ -z "$HOOK_ID" ] && bitbottle webhook create "$BB_TEST_CLOUD_REPO" \
  --url "$HOOK_URL" --events 'repo:push,pullrequest:created'
```

Exit code: `0`. Stdout matches `Created webhook <id> -> $HOOK_URL`.

**Verify in UI:** Settings → Webhooks shows the new entry pointing at
`$HOOK_URL`, subscribed to `Repository push` and `Pull request: Created`.

### 2. `webhook list` shows the new hook (TTY)

```bash
bitbottle webhook list "$BB_TEST_CLOUD_REPO"
```

TTY table header `ID  URL  ACTIVE  EVENTS`. The row for `$HOOK_URL` shows
`true` in the ACTIVE column and a comma-joined event list.

### 3. `webhook list --json` returns event array

```bash
bitbottle webhook list "$BB_TEST_CLOUD_REPO" \
  --json id,url,active,events \
  --jq ".[] | select(.url==\"$HOOK_URL\")"
```

Stdout includes `"events":["repo:push","pullrequest:created"]` (in some
order — Cloud may reorder). `"active":true`.

Capture the ID for subsequent steps:

```bash
export HOOK_ID=$(bitbottle webhook list "$BB_TEST_CLOUD_REPO" \
  --json id,url --jq ".[] | select(.url==\"$HOOK_URL\") | .id" -r 2>/dev/null \
  || bitbottle webhook list "$BB_TEST_CLOUD_REPO" --json id,url \
       | jq -r --arg u "$HOOK_URL" '.[] | select(.url==$u) | .id')
echo "HOOK_ID=$HOOK_ID"
```

### 4. `webhook view` by ID

```bash
bitbottle webhook view "$BB_TEST_CLOUD_REPO" "$HOOK_ID"
```

TTY output shows ID, URL, active, events for the single hook.

### 5. `webhook view --json`

```bash
bitbottle webhook view "$BB_TEST_CLOUD_REPO" "$HOOK_ID" \
  --json id,url,events,active \
  | jq '.[0] | keys | sort'
```

Stdout: `["active","events","id","url"]`.

### 6. `webhook create` with `--secret -` reads from stdin

```bash
echo "super-secret-token" | bitbottle webhook create "$BB_TEST_CLOUD_REPO" \
  --url "https://example.com/bb-webhook-test-secret" \
  --events 'repo:push' \
  --secret -
```

Exit code: `0`. **Critical:** the literal string `super-secret-token`
must NOT appear in stdout/stderr, shell history, or any log file.

**Verify in UI:** new webhook entry exists; secret field is set (Bitbucket
will not display it but emits an HMAC header on delivery).

### 7. Webhook secrets are write-only on read

```bash
bitbottle webhook list "$BB_TEST_CLOUD_REPO" --json id,url,events
```

Stdout has NO `secret` field. The Cloud API does not return secrets on
read, and the CLI's `Webhook` type has no Secret field — so there is no
chokepoint to bypass.

### 8. `webhook create --active=false`

```bash
bitbottle webhook create "$BB_TEST_CLOUD_REPO" \
  --url "https://example.com/bb-disabled-test" \
  --events 'repo:push' \
  --active=false
```

Exit code: `0`. Verify in UI that the hook is created in disabled state.

### 9. `webhook create` rejects whitespace-only `--events`

```bash
bitbottle webhook create "$BB_TEST_CLOUD_REPO" \
  --url "$HOOK_URL" \
  --events ' , ,  '
```

Exit code: non-zero. stderr names `--events` and says at least one event
key is required.

### 10. `webhook create` requires `--url` and `--events`

```bash
bitbottle webhook create "$BB_TEST_CLOUD_REPO"
```

Exit code: non-zero. stderr names both `--url` and `--events` as required.

### 11. `webhook delete` requires `--confirm` non-interactively

```bash
bitbottle webhook delete "$BB_TEST_CLOUD_REPO" "$HOOK_ID" </dev/null
```

Exit code: non-zero. stderr says `--confirm required when not running
interactively`.

### 12. `webhook delete --confirm` removes it

```bash
bitbottle webhook delete "$BB_TEST_CLOUD_REPO" "$HOOK_ID" --confirm
bitbottle webhook list "$BB_TEST_CLOUD_REPO" \
  --json id --jq ".[] | select(.id==\"$HOOK_ID\") | .id" \
  | grep . || echo "gone"
```

Last line prints `gone`. Exit code: `0` for the delete.

### 13. Deleting a non-existent webhook fails clearly

```bash
bitbottle webhook delete "$BB_TEST_CLOUD_REPO" \
  00000000-0000-0000-0000-000000000000 --confirm
```

Exit code: non-zero. stderr names the ID and says it was not found.

## Cleanup

```bash
# Remove any test webhooks left behind by failures above.
bitbottle webhook list "$BB_TEST_CLOUD_REPO" --json id,url \
  | jq -r '.[] | select(.url|test("bb-(webhook|disabled)-test"))| .id' \
  | xargs -I{} bitbottle webhook delete "$BB_TEST_CLOUD_REPO" {} --confirm \
      2>/dev/null || true
```
