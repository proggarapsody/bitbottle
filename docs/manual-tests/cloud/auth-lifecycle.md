# Scenario: Cloud auth lifecycle

**Backend:** Cloud.

Login → status → token → refresh → logout.

## Prerequisites

- `BB_TEST_CLOUD_HOST`, `BB_TEST_CLOUD_TOKEN`, `BB_TEST_CLOUD_EMAIL`.
  `BB_TEST_CLOUD_EMAIL` is the Atlassian account email associated with the
  token — required by `auth login --with-token` for Bitbucket Cloud.
- Not currently logged in to that host (or you do not mind being logged out).
- If multiple Bitbucket hosts are configured, all commands below must include
  `--hostname "$BB_TEST_CLOUD_HOST"` explicitly.

## Setup

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST" 2>/dev/null || true
```

## Steps

### 1. Login via stdin

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --email "$BB_TEST_CLOUD_EMAIL" --with-token
```

Expected stderr (last line):

```
Logged in as <username>
```

Exit code: `0`. `~/.config/bitbottle/hosts.yml` now has a `bitbucket.org:`
entry with `oauth_token` (or empty `oauth_token` if stored in the keyring).

### 2. Status shows the host

```bash
bitbottle auth status
```

Stdout includes:

```
bitbucket.org: Logged in as <username> (Token in keyring: yes|no)
```

Exit code: `0`.

### 3. Print stored token

```bash
bitbottle auth token --hostname "$BB_TEST_CLOUD_HOST"
```

Stdout: the raw PAT, one line, no trailing whitespace beyond the newline.
Equals `$BB_TEST_CLOUD_TOKEN`. Exit code: `0`.

### 4. Refresh validates the token

```bash
bitbottle auth refresh --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: `0`. stderr says the token is still valid (wording varies but
mentions "valid" or the username).

### 5. Refresh with a known-bad token fails clearly

```bash
echo "obviously-not-a-real-token" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --email "$BB_TEST_CLOUD_EMAIL" --with-token || true
bitbottle auth refresh --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: non-zero. stderr names the host and tells the user to generate a
new token in the Bitbucket UI and run `auth login`.

### 6. Re-login with the good token recovers

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --email "$BB_TEST_CLOUD_EMAIL" --with-token
bitbottle auth status
```

Status line again shows `Logged in as <username>`. Exit code: `0`.

### 7. Logout removes the host

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST"
bitbottle auth status
```

Status no longer lists `bitbucket.org`. Exit code of `status`: `0` (or
non-zero if no hosts remain — record which).

### 8. Token after logout is unavailable

```bash
bitbottle auth token --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: non-zero. stderr names the host as not authenticated.

## Cleanup

Re-login for downstream scenarios:

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --email "$BB_TEST_CLOUD_EMAIL" --with-token
```
