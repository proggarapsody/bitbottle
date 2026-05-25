# admin — Bitbucket Server / DC Administration

> **Server / Data Center only.** All `admin` subcommands require **SYS_ADMIN** permission. Standard admin tokens do not include it — these actions must be performed by a system administrator.

---

## admin secrets rotate

Rotate the cluster's internal HTTPS secret. After rotation **all cluster nodes must be restarted** for the new secret to take effect.

```
bitbottle admin secrets rotate [--confirm] [--hostname HOST]
```

| Flag | Description |
|------|-------------|
| `--confirm` | Skip the confirmation prompt (required in non-TTY / CI mode) |
| `--hostname` | Override auto-detected Bitbucket hostname |

**403 hint:** Requires SYS_ADMIN permission. Standard admin tokens do not include it.

---

## admin logging get

Show the current log level and async-logging setting.

```
bitbottle admin logging get [--json] [--hostname HOST]
```

Default output:
```
Level: INFO
Async: false
```

`--json` outputs `{"level":"INFO","async":false}`.

---

## admin logging set

Change the log level and/or async-logging mode.

```
bitbottle admin logging set [--level LEVEL] [--async] [--persistent] [--hostname HOST]
```

| Flag | Description |
|------|-------------|
| `--level` | Log level: `DEBUG`, `INFO`, `WARN`, or `ERROR` (case-sensitive) |
| `--async` | Enable async logging |
| `--persistent` | Write to `log4j.properties` so the change survives restarts |

At least one of `--level` or `--async` must be provided.

Without `--persistent` the change is **runtime-only** and resets on restart.  
With `--persistent` the change survives restarts (writes `log4j.properties`).

**403 hint:** Requires SYS_ADMIN permission.

---

---

## admin user list

List users on the Bitbucket Server / DC instance.

```
bitbottle admin user list [--filter QUERY] [--limit N] [--json] [--hostname HOST]
```

| Flag | Description |
|------|-------------|
| `--filter` | Filter users by name or email prefix |
| `--limit` | Maximum number of users (default 50) |
| `--json` | Output as JSON array |

Default output columns: SLUG, DISPLAY_NAME, EMAIL, ACTIVE, TYPE.

---

## admin user activate

Activate a user account.

```
bitbottle admin user activate SLUG [--hostname HOST]
```

---

## admin user deactivate

Deactivate a user account.

```
bitbottle admin user deactivate SLUG [--hostname HOST]
```

---

## admin user rename

Rename a user (change their username / slug).

```
bitbottle admin user rename OLD_SLUG NEW_SLUG [--hostname HOST]
```

---

## admin license

Show license details for the instance.

```
bitbottle admin license [--json] [--hostname HOST]
```

Default output columns: TIER, USERS, SERVER_ID, EXPIRY, SUPPORT_EXPIRY.

`--json` outputs the full license struct.

---

## admin cluster

Show cluster node information.

```
bitbottle admin cluster [--json] [--hostname HOST]
```

Default output columns: NODE_ID, NAME, ADDRESS, STATE, LOCAL.

`--json` outputs the full node list.

---

## admin mail get

Show the current SMTP mail server configuration.

```
bitbottle admin mail get [--json] [--hostname HOST]
```

Default output columns: HOSTNAME, PORT, PROTOCOL, STARTTLS, USERNAME, SENDER.

`--json` outputs the full config struct (password is never returned).

---

## admin mail set

Update the SMTP mail server configuration.

```
bitbottle admin mail set --mail-hostname SMTP_HOST [--port N] [--protocol smtp|smtps]
    [--username USER] [--sender EMAIL] [--use-starttls] [--require-starttls]
    [--password PASS] [--hostname HOST]
```

| Flag | Description |
|------|-------------|
| `--mail-hostname` | Mail server hostname (required) |
| `--port` | SMTP port (default 25) |
| `--protocol` | `smtp` or `smtps` (default `smtp`) |
| `--use-starttls` | Use STARTTLS if available |
| `--require-starttls` | Require STARTTLS (fail if not available) |
| `--username` | SMTP authentication username |
| `--sender` | Sender address for outgoing email (From:) |
| `--password` | SMTP password (warning: visible in process list) |

---

## admin banner get

Show the current site-wide announcement banner.

```
bitbottle admin banner get [--json] [--hostname HOST]
```

Default output:
```
Message:  Maintenance on Friday
Audience: ALL
Enabled:  true
```

`--json` outputs `{"message":"...","audience":"ALL","enabled":true}`.

---

## admin banner set

Post or update the site-wide announcement banner.

```
bitbottle admin banner set MESSAGE [--audience all|authenticated|unauthenticated]
    [--enabled] [--hostname HOST]
```

| Flag | Description |
|------|-------------|
| `MESSAGE` | Banner text (positional argument, required) |
| `--audience` | `all`, `authenticated`, or `unauthenticated` (default `ALL`) |
| `--enabled` | Enable the banner immediately (default true) |

---

## admin banner clear

Remove the site-wide announcement banner.

```
bitbottle admin banner clear [--confirm] [--hostname HOST]
```

`--confirm` is required in non-TTY / CI mode.

---

## MCP tools

| Tool | Description |
|------|-------------|
| `rotate_secrets` | Rotate cluster HTTPS secret |
| `get_logging_config` | Get log level and async setting |
| `set_logging_config` | Set log level (`level`), async flag (`async`), persistence (`persistent`) |
| `list_admin_users` | List users (`filter`, `limit`) |
| `activate_user` | Activate a user (`slug`) |
| `deactivate_user` | Deactivate a user (`slug`) |
| `rename_user` | Rename a user (`slug`, `new_slug`) |
| `get_admin_license` | Get instance license details |
| `get_cluster_nodes` | Get cluster node list |
| `get_mail_server_config` | Get SMTP mail server config |
| `set_mail_server_config` | Update SMTP mail server config (`mail_hostname` required, optional `port`, `protocol`, `use_starttls`, `require_starttls`, `username`, `sender_address`, `password`) |
| `get_banner` | Get site-wide announcement banner |
| `set_banner` | Create/update banner (`message` required, optional `audience`, `enabled`) |
| `clear_banner` | Remove the site-wide announcement banner |
