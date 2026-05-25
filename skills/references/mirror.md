# Mirror Servers (Server/DC only)

Read Smart Mirror server configuration from a Bitbucket Server / Data Center instance.
Requires the Bitbucket Server Mirror module to be enabled.
All three commands return `host.unsupported` on Bitbucket Cloud.

## Commands

### mirror list

List all configured Smart Mirror servers.

```bash
bitbottle mirror list --hostname git.example.com -k
bitbottle mirror list --hostname git.example.com -k --json
bitbottle mirror list --hostname git.example.com -k --limit 50
```

Output columns: ID, NAME, BASE URL, ENABLED

### mirror view

View details of a specific mirror server by ID.

```bash
bitbottle mirror view MIRROR-ID --hostname git.example.com -k
bitbottle mirror view MIRROR-ID --hostname git.example.com -k --json
```

Output: key-value summary (ID, Name, Base URL, Enabled).

### mirror repo list

List repositories mirrored by a specific mirror server.

```bash
bitbottle mirror repo list MIRROR-ID --hostname git.example.com -k
bitbottle mirror repo list MIRROR-ID --hostname git.example.com -k --json
bitbottle mirror repo list MIRROR-ID --hostname git.example.com -k --limit 100
```

Output columns: SLUG, STATUS, LAST SYNC

## Notes

- Mirror IDs are returned by `mirror list` in the ID column.
- `--limit` defaults to 30; use higher values to fetch more mirrors/repos.
- All commands require `--hostname` (or `BB_HOST`) pointing to the Server/DC host.
- Use `-k` / `skip_tls_verify: true` for hosts with self-signed certificates.
