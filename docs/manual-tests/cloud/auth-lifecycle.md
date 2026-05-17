# Scenario: Cloud auth lifecycle (login → status → token → refresh → migrate → logout)

**Backend:** Cloud.

End-to-end of the `auth` command group against a real Cloud host. Exercises
the OS keyring path, plaintext-token migration, and the `doctor` health
check — all things automated tests can't fake.

## Prerequisites

- `BB_TEST_CLOUD_HOST`, `BB_TEST_CLOUD_TOKEN` set.
- An OS keyring is available (macOS Keychain, Secret Service on Linux,
  Credential Manager on Windows). On headless Linux without a keyring,
  bitbottle falls back to plaintext storage — record this and continue.
- No active session for `$BB_TEST_CLOUD_HOST` (the Setup step ensures this).

## Setup

Start clean:

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST" 2>/dev/null || true
```

## Steps

### 1. `auth login --with-token` stores the token

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --with-token
```

Exit code: `0`. stderr ends with `Logged in as <username>`.

### 2. `auth status` confirms the session

```bash
bitbottle auth status
```

Stdout contains a line of the form:

```
<host>: Logged in as <user> (Token in keyring: yes|no)
```

If the keyring is available, `Token in keyring: yes`. On headless Linux
without a keyring, `Token in keyring: no` is acceptable — record it.

### 3. `auth token` prints the active token

```bash
bitbottle auth token --hostname "$BB_TEST_CLOUD_HOST" | head -c 8
echo
```

Stdout is the first 8 chars of the PAT. Exit code: `0`. Verify it matches
the prefix of `$BB_TEST_CLOUD_TOKEN`.

### 4. `auth refresh` re-validates the token against the API

```bash
bitbottle auth refresh --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: `0`. stderr indicates the refresh succeeded. (For PATs this
revalidates rather than rotating; OAuth tokens are actually rotated. Record
which path was taken.)

### 5. Bad token surfaces a clear error

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST"
echo "not-a-real-token" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --with-token
```

Exit code: non-zero. stderr mentions authentication failure / 401 — not a
generic stack trace. Re-login with the real token before continuing:

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --with-token
```

### 6. `auth migrate` moves plaintext → keyring (when applicable)

```bash
bitbottle auth migrate
```

If a plaintext token exists in `hosts.yml`, stderr reports it was moved to
the keyring and the YAML is rewritten without the token. If no plaintext
token exists, stderr reports nothing to migrate and exits `0`. Either is a
pass — the critical bit is the messaging is clear.

### 7. `auth doctor` reports a healthy environment

```bash
bitbottle auth doctor
```

Stdout enumerates checks (config dir, hosts.yml, keyring availability, host
reachability) with PASS/FAIL/WARN per row. Exit code: `0` if all PASS or
WARN; non-zero if any FAIL. Investigate any FAIL and record it.

### 8. `auth logout` clears the session

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST"
bitbottle auth status
```

After logout, `auth status` reports no session for `$BB_TEST_CLOUD_HOST`
(either an empty list or the host is omitted). Exit code: non-zero for
`auth status` when there are no sessions at all; `0` if other hosts remain
logged in.

### 9. Commands requiring auth fail with a clear hint

```bash
bitbottle repo list --hostname "$BB_TEST_CLOUD_HOST" --limit 1
```

Exit code: non-zero. stderr mentions "not logged in" and suggests
`auth login`.

## Cleanup

```bash
# Re-login so downstream scenarios have a session.
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --with-token
```
