# profile — Named Credential Profiles

Manage named sets of Bitbucket credentials. Profiles let you keep multiple
Bitbucket instances (work Server + personal Cloud) and switch between them
in one command, similar to kubectl contexts.

## Commands

### `profile create NAME`

```
bitbottle profile create work \
  --hostname git.work.com \
  --token BBDC-... \
  --user alice \
  --auth-user alice@work.com \
  --skip-tls \
  --backend server \
  --git-protocol https
```

Required flags: `--hostname`, `--token`.

Optional flags:

| Flag | Description |
|---|---|
| `--user USER` | Username for the account |
| `--auth-user USER` | HTTP Basic auth username (email for Cloud App Passwords) |
| `--skip-tls` | Skip TLS certificate verification |
| `--backend server\|cloud` | Force backend routing (default: auto-detect) |
| `--git-protocol https\|ssh` | Preferred git protocol |

A profile with an existing name is silently overwritten.

### `profile use NAME`

```
bitbottle profile use work
```

Writes the profile's credentials into `hosts.yml` for the profile's
hostname. All subsequent commands targeting that hostname use the
profile's token, user, and settings.

### `profile list`

```
bitbottle profile list
bitbottle profile list --json name,hostname,backend_type,skip_tls_verify
bitbottle profile list --json name,hostname --jq '.[].hostname'
```

Token is **never** printed. Supported `--json` fields:
`name`, `hostname`, `user`, `backend_type`, `skip_tls_verify`.

### `profile delete NAME`

```
bitbottle profile delete work --confirm
```

Without `--confirm` on a non-TTY, the command errors. On a TTY, prompts
`[y/N]`.

## Storage

Profiles are stored in `~/.config/bitbottle/profiles.yml`
(respects `$XDG_CONFIG_HOME` and `$BB_CONFIG_DIR`).
