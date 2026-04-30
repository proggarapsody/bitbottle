# Scenario: multi-host (Cloud + Server/DC simultaneously)

**Backend:** Both, at the same time.

Verifies that bitbottle can hold credentials for two hosts at once and that
`--hostname` correctly disambiguates.

## Prerequisites

- `BB_TEST_CLOUD_HOST`, `BB_TEST_CLOUD_TOKEN`, `BB_TEST_CLOUD_REPO` set.
- `BB_TEST_SERVER_HOST`, `BB_TEST_SERVER_TOKEN`, `BB_TEST_SERVER_REPO` set.
- Optional: `BB_TEST_SERVER_SKIP_TLS=true` for self-signed servers.

## Setup

Start clean:

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST" 2>/dev/null || true
bitbottle auth logout --hostname "$BB_TEST_SERVER_HOST" 2>/dev/null || true
```

## Steps

### 1. Log in to Cloud

```bash
echo "$BB_TEST_CLOUD_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_CLOUD_HOST" --with-token
```

stderr ends with `Logged in as <username>`. Exit code: `0`.

### 2. Log in to Server/DC

```bash
SKIP_TLS_FLAG=""
[ "$BB_TEST_SERVER_SKIP_TLS" = "true" ] && SKIP_TLS_FLAG="--skip-tls-verify"
echo "$BB_TEST_SERVER_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token --git-protocol https $SKIP_TLS_FLAG
```

stderr ends with `Logged in as <username>`. Exit code: `0`.

### 3. `auth status` lists both

```bash
bitbottle auth status
```

Stdout contains two lines, one per host, each of the form:

```
<host>: Logged in as <user> (Token in keyring: yes|no)
```

### 4. `--hostname` disambiguates `repo list`

```bash
bitbottle repo list --hostname "$BB_TEST_CLOUD_HOST"  --limit 1
bitbottle repo list --hostname "$BB_TEST_SERVER_HOST" --limit 1
```

Each emits the table for its respective host. Slugs/projects must reflect
the corresponding host (cross-check in the Bitbucket UI if uncertain).

### 5. Without `--hostname` and outside a git repo, default host applies

```bash
cd /tmp && bitbottle repo list --limit 1
```

If `bitbottle config get default_host` is set, results come from that host.
Otherwise behavior is host-resolution-fallback per `bbinstance` rules — record
what is observed.

### 6. Inside a git repo with a Server/DC remote, host auto-resolves

```bash
git clone "https://$BB_TEST_SERVER_HOST/scm/${BB_TEST_SERVER_REPO/\//\/}.git" /tmp/bb-multi-host
cd /tmp/bb-multi-host
bitbottle repo view
```

Output reflects the Server/DC repo (project + slug match `BB_TEST_SERVER_REPO`).

### 7. Logout one host leaves the other intact

```bash
bitbottle auth logout --hostname "$BB_TEST_CLOUD_HOST"
bitbottle auth status
```

Output now shows only the Server/DC host.

## Cleanup

```bash
bitbottle auth logout --hostname "$BB_TEST_SERVER_HOST" 2>/dev/null || true
rm -rf /tmp/bb-multi-host
# Re-login if other scenarios need it.
```
