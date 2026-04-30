# Scenario: Server/DC — `pipeline` commands rejected

**Backend:** Server / Data Center. Negative test.

Pipelines are a Bitbucket Cloud feature. Every `pipeline` subcommand must
exit non-zero with a clear message when the resolved host is Server/DC,
and must NOT make any HTTP request.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- Logged out of Cloud (or use `--hostname` explicitly to force Server/DC):

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST" 2>/dev/null || true
```

## Steps

### 1. `pipeline list`

```bash
bitbottle pipeline list "$BB_TEST_SERVER_REPO"
```

Exit code: non-zero. stderr says pipelines are Cloud-only and names the
host as a Server/DC instance. No HTTP error message — this is a
client-side gate.

### 2. `pipeline view`

```bash
bitbottle pipeline view "$BB_TEST_SERVER_REPO" 00000000-0000-0000-0000-000000000000
```

Same: exit non-zero, clear "Cloud only" message.

### 3. `pipeline run`

```bash
bitbottle pipeline run "$BB_TEST_SERVER_REPO" --branch main
```

Same: exit non-zero, clear "Cloud only" message.

### 4. With `--hostname` overriding to Server/DC explicitly

```bash
bitbottle pipeline list "$BB_TEST_SERVER_REPO" --hostname "$BB_TEST_SERVER_HOST"
```

Same: exit non-zero, clear "Cloud only" message. Confirms the gate fires
even when host is supplied explicitly.

### 5. With `--hostname` pointing at Cloud, but no Cloud auth

```bash
bitbottle pipeline list "$BB_TEST_SERVER_REPO" --hostname "$BB_TEST_CLOUD_HOST"
```

Exit non-zero. stderr says "not authenticated" for the Cloud host (NOT a
"Cloud only" error — the gate is host-type, the auth error is what should
surface here).

## Cleanup

Re-login Cloud if other scenarios need it:

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --with-token
```
