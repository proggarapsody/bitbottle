# Scenario: Server/DC auth lifecycle (incl. self-signed TLS)

**Backend:** Server / Data Center.

Login → status → token → refresh → logout. Also covers `--skip-tls-verify`
and `--git-protocol https`.

## Prerequisites

- `BB_TEST_SERVER_HOST`, `BB_TEST_SERVER_TOKEN`.
- `BB_TEST_SERVER_SKIP_TLS=true` if the server uses a self-signed cert.

## Setup

```bash
bitbottle auth logout --hostname "$BB_TEST_SERVER_HOST" 2>/dev/null || true
SKIP_TLS_FLAG=""
[ "$BB_TEST_SERVER_SKIP_TLS" = "true" ] && SKIP_TLS_FLAG="--skip-tls-verify"
```

## Steps

### 1. Login over HTTPS, optionally skipping TLS verify

```bash
echo "$BB_TEST_SERVER_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token \
  --git-protocol https $SKIP_TLS_FLAG
```

stderr ends with `Logged in as <username>`. Exit code: `0`.

`~/.config/bitbottle/hosts.yml` now contains the host with
`git_protocol: https`, and `skip_tls_verify: true` if the flag was used.

### 2. `auth status`

```bash
bitbottle auth status
```

Stdout includes:

```
$BB_TEST_SERVER_HOST: Logged in as <username> (Token in keyring: yes|no)
```

### 3. `auth token`

```bash
bitbottle auth token --hostname "$BB_TEST_SERVER_HOST"
```

Stdout equals `$BB_TEST_SERVER_TOKEN`. Exit code: `0`.

### 4. `auth refresh` succeeds

```bash
bitbottle auth refresh --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: `0`. stderr says the token is still valid.

### 5. (If `BB_TEST_SERVER_SKIP_TLS=true`) login WITHOUT the flag fails

```bash
bitbottle auth logout --hostname "$BB_TEST_SERVER_HOST"
echo "$BB_TEST_SERVER_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token --git-protocol https
```

If the cert is self-signed: exit code non-zero, stderr names a TLS error
(`x509: certificate signed by unknown authority` or similar) and suggests
`--skip-tls-verify`. If the cert is trusted by the system: exit code `0`
and this step is skipped.

### 6. Re-login with `--skip-tls-verify` recovers

```bash
echo "$BB_TEST_SERVER_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token \
  --git-protocol https $SKIP_TLS_FLAG
bitbottle auth status
```

Status line shows the host again.

### 7. Bad token + `auth refresh` fails clearly

```bash
echo "obviously-bad" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token \
  --git-protocol https $SKIP_TLS_FLAG
bitbottle auth refresh --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr names the host and instructs to generate a new
PAT and run `auth login`.

### 8. Logout

```bash
echo "$BB_TEST_SERVER_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token \
  --git-protocol https $SKIP_TLS_FLAG
bitbottle auth logout --hostname "$BB_TEST_SERVER_HOST"
bitbottle auth status
```

Status no longer lists the host.

## Cleanup

Re-login for downstream Server/DC scenarios:

```bash
echo "$BB_TEST_SERVER_TOKEN" | bitbottle auth login \
  --hostname "$BB_TEST_SERVER_HOST" --with-token \
  --git-protocol https $SKIP_TLS_FLAG
```
