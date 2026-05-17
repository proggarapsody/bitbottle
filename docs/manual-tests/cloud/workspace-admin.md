# Scenario: Cloud workspace admin (members, hooks, project list, code search, user SSH keys)

**Backend:** Cloud only. All of these surfaces are Cloud-exclusive
features; Server/DC has no workspace concept and returns
`host.unsupported` for the corresponding commands.

Covers **N** (workspace list, project list), **WORKSPACE-MEMBERS**,
**WORKSPACE-HOOKS**, **SR** (code search), and **SSH-KEYS** (user
account-level SSH keys, distinct from repo deploy keys).

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_WORKSPACE` set to the workspace slug. Typically the
  prefix of `BB_TEST_CLOUD_REPO`.
- A throwaway SSH public key — re-use `/tmp/bb-qa-key.pub` from the
  repo-settings scenario if available, otherwise:
  `ssh-keygen -t ed25519 -N "" -f /tmp/bb-qa-key`.

```bash
[ -f /tmp/bb-qa-key.pub ] || ssh-keygen -t ed25519 -N "" -f /tmp/bb-qa-key
export QA_KEY="$(cat /tmp/bb-qa-key.pub)"
```

## Steps — Workspace enumeration

### 1. `workspace list`

```bash
bitbottle workspace list --hostname "$BB_TEST_CLOUD_HOST"
```

Stdout is a table of workspaces you have access to (`SLUG`, `NAME`, `ROLE`).
`BB_TEST_CLOUD_WORKSPACE` is in the list. Exit code: `0`.

### 2. `project list WORKSPACE`

```bash
bitbottle project list "$BB_TEST_CLOUD_WORKSPACE" --hostname "$BB_TEST_CLOUD_HOST"
```

Stdout is a table of projects in the workspace. Cloud workspaces with no
explicit projects may emit an empty list — exit `0`. With `--json`:

```bash
bitbottle project list "$BB_TEST_CLOUD_WORKSPACE" --json key,name \
  --hostname "$BB_TEST_CLOUD_HOST" | jq 'length'
```

Output is a non-negative integer.

### 3. `workspace member list` (paginated)

```bash
bitbottle workspace member list "$BB_TEST_CLOUD_WORKSPACE" \
  --hostname "$BB_TEST_CLOUD_HOST" --limit 5
```

Stdout includes you (the current user) and up to 4 other members.
Columns: `USER`, `ROLE`, `UUID`. With `--json`:

```bash
bitbottle workspace member list "$BB_TEST_CLOUD_WORKSPACE" \
  --hostname "$BB_TEST_CLOUD_HOST" --json username,role | jq '.[0]'
```

The first member has both fields populated.

## Steps — Workspace webhooks

### 4. `workspace hook create`

```bash
bitbottle workspace hook create "$BB_TEST_CLOUD_WORKSPACE" \
  --url "https://example.test/ws-hook" \
  --events "repo:push,repo:created" \
  --active=true \
  --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: `0`. stderr prints the new hook UUID.

### 5. `workspace hook list` includes it

```bash
export WH_UUID=$(bitbottle workspace hook list "$BB_TEST_CLOUD_WORKSPACE" \
  --hostname "$BB_TEST_CLOUD_HOST" --json uuid,url \
  | jq -r '.[] | select(.url=="https://example.test/ws-hook") | .uuid')
echo "WH_UUID=$WH_UUID"
```

`WH_UUID` is a non-empty UUID.

### 6. `workspace hook delete`

```bash
bitbottle workspace hook delete "$BB_TEST_CLOUD_WORKSPACE" "$WH_UUID" \
  --hostname "$BB_TEST_CLOUD_HOST"
bitbottle workspace hook list "$BB_TEST_CLOUD_WORKSPACE" \
  --hostname "$BB_TEST_CLOUD_HOST" --json uuid \
  | jq 'map(.uuid) | index("'"$WH_UUID"'")'
```

Final `jq` prints `null`.

## Steps — Code search (SR)

### 7. `search code QUERY` (default workspace = current repo's)

Inside a clone of `$BB_TEST_CLOUD_REPO`:

```bash
rm -rf /tmp/bb-search
bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-search
cd /tmp/bb-search

bitbottle search code "manual test" --hostname "$BB_TEST_CLOUD_HOST"
```

Stdout is a table of matches with `FILE`, `LINE`, `MATCH`. If the
scratch repo contains no matches, the output is an empty result
(exit `0`).

### 8. `search code --workspace WORKSPACE`

```bash
bitbottle search code "README" --workspace "$BB_TEST_CLOUD_WORKSPACE" \
  --hostname "$BB_TEST_CLOUD_HOST" --limit 5
```

Stdout includes file matches from across the workspace. Exit code: `0`.

### 9. `--json` returns structured matches

```bash
bitbottle search code "README" --workspace "$BB_TEST_CLOUD_WORKSPACE" \
  --hostname "$BB_TEST_CLOUD_HOST" --json file,line,match --limit 5 \
  | jq '.[0] | .file,.line,.match'
```

Each `jq` output is populated.

## Steps — User SSH keys (SSH-KEYS, Cloud)

### 10. `ssh-key add` (current user account)

```bash
bitbottle ssh-key add --key "$QA_KEY" --label "qa-manual" \
  --hostname "$BB_TEST_CLOUD_HOST"
```

Exit code: `0`. stderr prints the new key ID. **This adds the key to your
Bitbucket account** — clean up in step 12.

### 11. `ssh-key list`

```bash
export SSH_ID=$(bitbottle ssh-key list --hostname "$BB_TEST_CLOUD_HOST" \
  --json id,label | jq -r '.[] | select(.label=="qa-manual") | .id' | head -1)
echo "SSH_ID=$SSH_ID"
```

`SSH_ID` is non-empty.

### 12. `ssh-key delete`

```bash
bitbottle ssh-key delete "$SSH_ID" --hostname "$BB_TEST_CLOUD_HOST"
bitbottle ssh-key list   --hostname "$BB_TEST_CLOUD_HOST" --json id \
  | jq 'map(.id) | index("'"$SSH_ID"'")'
```

Final `jq` prints `null`.

## Steps — `host.unsupported` parity (Server probe)

### 13. `workspace list` on Server returns the typed error

```bash
bitbottle workspace list --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr mentions "unsupported on host" — typed
`ErrUnsupportedOnHost`, not a raw 404. Skip if no Server session.

### 14. `search code` on Server returns the typed error

```bash
bitbottle search code "anything" --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr mentions "unsupported on host".

## Cleanup

```bash
[ -n "${WH_UUID:-}" ] && bitbottle workspace hook delete "$BB_TEST_CLOUD_WORKSPACE" "$WH_UUID" --hostname "$BB_TEST_CLOUD_HOST" 2>/dev/null || true
[ -n "${SSH_ID:-}"  ] && bitbottle ssh-key delete "$SSH_ID" --hostname "$BB_TEST_CLOUD_HOST" 2>/dev/null || true

rm -f /tmp/bb-qa-key /tmp/bb-qa-key.pub
cd - >/dev/null
rm -rf /tmp/bb-search
```
