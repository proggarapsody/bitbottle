# bitbottle auth, hosts.yml, env vars

## Login flows

Cloud needs an **email**; Server/DC needs a **username**. The token
comes from stdin via `--with-token`.

```bash
# Bitbucket Cloud (App Password or API token)
echo "$APP_PASSWORD" | bitbottle auth login \
  --hostname bitbucket.org \
  --email you@example.com \
  --with-token

# Bitbucket Server / Data Center (PAT, often "BBDC-…")
echo "$BBDC_PAT" | bitbottle auth login \
  --hostname git.example.com \
  --username your.user \
  --with-token \
  --git-protocol https \
  --skip-tls-verify             # only for self-signed certs
```

`--skip-tls-verify` is set once at login and remembered per host in
`hosts.yml` as `skip_tls_verify: true`.

For Server/DC hosts, `auth login` also runs a TLS pre-flight: if the
host's certificate is not signed by a CA your OS already trusts,
bitbottle prints the cert (Subject CN, Issuer CN, validity window,
SHA-256 fingerprint) and asks `Trust this certificate? [y/N]`. On
confirm it sets `skip_tls_verify: true` automatically; on decline
(or non-TTY / `BB_PROMPT_DISABLED=1`) it surfaces the `x509` error
and expects you to re-run with `--skip-tls-verify`. The probe is
skipped when `--skip-tls-verify` is already passed and for Cloud hosts.

For a one-off override (e.g. recovering from a new self-signed cert on
an already-configured host) pass the persistent root flag `-k` /
`--skip-tls-verify` to any command:

```bash
bitbottle -k pr approve 42 -R git.example.com/PROJ/repo
bitbottle --skip-tls-verify pr list -R git.example.com/PROJ/repo
```

The root flag only affects the current invocation — it does not write
to `hosts.yml`.

## Lifecycle commands

```bash
bitbottle auth status
bitbottle auth token   [--hostname HOST]   # print stored token to stdout
bitbottle auth refresh [--hostname HOST]   # re-validate, refresh user slug
bitbottle auth logout  --hostname HOST
bitbottle auth migrate [--hostname HOST]   # move config-file token into the OS keyring
```

## Migrating tokens from hosts.yml to the keyring

If `bitbottle auth login` was run on an older version, the token may be stored
in plain text inside `hosts.yml`. Run `auth migrate` once to move it into the
OS keyring (macOS Keychain, GNOME Keyring, Windows Credential Manager):

```bash
bitbottle auth migrate                  # all configured hosts
bitbottle auth migrate --hostname HOST  # single host only
```

After migration the token is removed from `hosts.yml`. In headless environments
(CI, Docker, SSH) set `BITBOTTLE_ALLOW_INSECURE_STORE=1` to fall back to an
AES-256-GCM encrypted file store instead of the OS keyring.

## hosts.yml reference

Stored at `~/.config/bitbottle/hosts.yml` (or `$BB_CONFIG_DIR/hosts.yml`).

```yaml
bitbucket.org:
  oauth_token: <app-password-or-api-token>
  auth_user: you@example.com
  user: your-username
  git_protocol: ssh

git.example.com:
  oauth_token: BBDC-…
  user: your.user
  git_protocol: https
  skip_tls_verify: true
  backend_type: server          # or "cloud" — see below
```

`backend_type` forces cloud-vs-server dispatch when the hostname
doesn't tell the truth. Set it for **self-hosted Bitbucket Cloud
Data Center** (a server-style hostname that speaks the v2/Cloud
API). For ordinary `bitbucket.org` and standard Server/DC, omit it
— the hostname-based default works.

## Full env-var reference

| Var | Effect |
|---|---|
| `BB_TOKEN` | Token override for API calls (CI use). Backend interpretation depends on the resolved host. |
| `BB_HOST` | Default `--hostname`. |
| `BB_REPO` | Default `-R [HOST/]PROJECT/REPO`. |
| `BB_EDITOR` | Editor for interactive prompts (`pr edit`). Falls back to config → `$EDITOR` → `vi`. |
| `BB_PAGER` | Pager for paginated TTY output. Falls back to config → `$PAGER`. Empty string disables. |
| `BB_BROWSER` | Browser launched by `--web` flags. Falls back to config → platform default. |
| `BB_FORCE_TTY` | Force aligned/colored output even when stdout is piped (mirrors `GH_FORCE_TTY`). |
| `BB_PROMPT_DISABLED` | Fail rather than prompt; required for non-interactive scripts. |
| `BB_CONFIG_DIR` | Override config dir (default `$XDG_CONFIG_HOME/bitbottle`). |
| `NO_COLOR` | Standard cross-tool convention; disables color. |
| `XDG_CONFIG_HOME` | Respected when `BB_CONFIG_DIR` is unset. |

## Multi-host setup

When two or more hosts are authenticated, every command needs to
know which one. Three ways, in precedence order:

1. **Explicit flag** — `--hostname HOST` on the command, or
   `-R HOST/PROJ/repo`.
2. **Env var** — `BB_HOST=git.example.com bitbottle pr list …`.
3. **Auto-detect from `origin` remote** — works inside a checkout.

If none of those resolves a single host, you'll see *"multiple hosts
configured; specify hostname"*. Pick one of the three.
