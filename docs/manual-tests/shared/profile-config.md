# Scenario: Named profiles, local config, user view, and context

**Backend:** Both. The local-state pieces (`profile`, `config`, `context`)
are host-agnostic; `user view` is run against each backend.

Covers **PROF** (named profiles), **U** (config get/set/list), **USER-VIEW**,
and **CTX** (context primitive). All exercise real filesystem state
(`~/.config/bitbottle/` plus the OS keyring for profile tokens) and the
auth fallback chain.

## Prerequisites

- Logged in to both `$BB_TEST_CLOUD_HOST` and `$BB_TEST_SERVER_HOST`.
- A second Cloud API token in `BB_TEST_CLOUD_TOKEN_ALT` for the named-profile
  smoke (can be the same value if you don't want to provision a second
  token — the test still validates the profile-switching plumbing).

## Steps — Named profiles (PROF)

### 1. `profile create` (Cloud, "work")

```bash
bitbottle profile create work \
  --hostname "$BB_TEST_CLOUD_HOST" \
  --token "$BB_TEST_CLOUD_TOKEN" \
  --backend cloud
```

Exit code: `0`. The profile is written to
`~/.config/bitbottle/profiles/work.yml` (or platform equivalent). The
token is stored in the OS keyring under the profile name.

### 2. `profile create` (Cloud, "side") with an alternate token

```bash
bitbottle profile create side \
  --hostname "$BB_TEST_CLOUD_HOST" \
  --token "${BB_TEST_CLOUD_TOKEN_ALT:-$BB_TEST_CLOUD_TOKEN}" \
  --backend cloud
```

Exit code: `0`.

### 3. `profile list` enumerates both

```bash
bitbottle profile list
```

Stdout includes rows for `work` and `side`, an `ACTIVE` indicator, and the
host. Exit code: `0`.

### 4. `profile use work` activates it

```bash
bitbottle profile use work
bitbottle profile list | grep -E '^\* *work'
```

`grep` exits `0` (asterisk or similar marker on the active row).
`auth status` now reflects the work profile's host:

```bash
bitbottle auth status
```

### 5. Switch to the other profile

```bash
bitbottle profile use side
bitbottle profile list | grep -E '^\* *side'
```

`grep` exits `0`.

### 6. Bogus profile name errors clearly

```bash
bitbottle profile use does-not-exist
```

Exit code: non-zero. stderr mentions the profile is not found.

### 7. `profile delete` removes one

```bash
bitbottle profile delete side --confirm
bitbottle profile list | grep -E '^\* *side' || echo "removed"
```

Final line prints `removed`. The keyring entry for `side` is also
removed (verify in Keychain Access / Secret Service / Credential Manager
that the entry no longer exists).

## Steps — Local config (U)

### 8. `config set` writes a value

```bash
bitbottle config set editor "vim"
```

Exit code: `0`. The change is persisted to `~/.config/bitbottle/config.yml`.

### 9. `config get` reads it back

```bash
bitbottle config get editor
```

Stdout is `vim`. Exit code: `0`.

### 10. `config list` shows all keys

```bash
bitbottle config list
```

Stdout includes `editor=vim` (or `editor: vim` depending on output format).
Exit code: `0`.

### 11. Bogus key on get returns empty / non-zero

```bash
bitbottle config get nope-not-a-key
```

Exit code: non-zero OR exit `0` with empty stdout — record the actual
behaviour and verify it's consistent across runs.

### 12. Unset via `config set <key> ""` (or equivalent)

```bash
bitbottle config set editor ""
bitbottle config get editor
```

Stdout is empty. (If `config set` rejects empty strings, the alternate
unset path is documented elsewhere — record what happened.)

## Steps — user view (USER-VIEW)

### 13. `user view` (no arg — current user)

```bash
bitbottle user view --hostname "$BB_TEST_CLOUD_HOST"
```

Stdout shows username, display name, account UUID/ID (Cloud) or slug
(Server), email (where exposed). Exit code: `0`.

### 14. `user view USERNAME` (other user)

Cloud:

```bash
bitbottle user view "$BB_TEST_CLOUD_REVIEWER" --hostname "$BB_TEST_CLOUD_HOST"
```

Server:

```bash
bitbottle user view "$BB_TEST_SERVER_REVIEWER" --hostname "$BB_TEST_SERVER_HOST"
```

Each shows that user's profile. Exit code: `0`. Skip if no reviewer env
var is set.

### 15. Bogus username gives a typed not-found

```bash
bitbottle user view "no-such-user-12345" --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: non-zero. stderr mentions "not found" — typed `ErrNotFound`,
not a raw 404.

## Steps — context primitive (CTX)

### 16. `context` inside a repo

Inside a clone of `$BB_TEST_CLOUD_REPO` (or any tracked repo):

```bash
rm -rf /tmp/bb-ctx
bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-ctx
cd /tmp/bb-ctx
bitbottle context
```

Stdout lists host, project/workspace, slug, current branch, current user,
default branch, ahead/behind counts. Exit code: `0`.

### 17. `context --json` emits the same data structured

```bash
bitbottle context --json | jq '.host,.repo.slug,.branch.current,.user.username'
```

Each `jq` output is non-null. Exit code: `0`. This is the "agent
orientation" primitive — verify it works without a connected backend by
checking that `--json` returns even partial data when offline:

```bash
# Simulate offline by pointing at a nonexistent host:
bitbottle context --json --hostname "no-such-host.invalid" 2>/dev/null \
  | jq '.repo.slug,.branch.current'
```

The structural fields are still populated (local git state), even though
backend-dependent fields are null.

## Cleanup

```bash
# Remove the test profile.
bitbottle profile delete work --confirm 2>/dev/null || true

# Unset the editor key.
bitbottle config set editor "" 2>/dev/null || true

cd - >/dev/null
rm -rf /tmp/bb-ctx
```
