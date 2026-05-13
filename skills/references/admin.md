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

## MCP tools

| Tool | Description |
|------|-------------|
| `rotate_secrets` | Rotate cluster HTTPS secret |
| `get_logging_config` | Get log level and async setting |
| `set_logging_config` | Set log level (`level`), async flag (`async`), persistence (`persistent`) |
