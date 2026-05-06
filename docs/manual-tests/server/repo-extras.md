# Scenario: Server repo extras (rename + fork-rejected)

**Backend:** Bitbucket Server / Data Center.

Exercise the scope-Q additions: `repo rename` (Server) and the
unsupported-capability surface for `repo fork` on Server.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_PROJECT` set; token has `REPO_ADMIN` permission.
- A unique scratch slug:
  ```bash
  export SCRATCH_SLUG="bb-qa-rx-$(date +%s)"
  export SCRATCH_FQN="$BB_TEST_SERVER_PROJECT/$SCRATCH_SLUG"
  ```

## Steps

### 1. Set up — create scratch repo

```bash
bitbottle repo create "$SCRATCH_SLUG" \
  --project "$BB_TEST_SERVER_PROJECT" \
  --description "bitbottle scope-Q server test" \
  --private=true
```

Exit code: `0`.

### 2. `repo rename` updates the slug

```bash
export RENAMED_SLUG="${SCRATCH_SLUG}-v2"
bitbottle repo rename "$SCRATCH_FQN" "$RENAMED_SLUG"
```

Exit code: `0`. stdout reports the rename and prints the new
`PROJECT/slug` coordinate.

**Verify:**
```bash
bitbottle repo view "$BB_TEST_SERVER_PROJECT/$RENAMED_SLUG"
bitbottle repo view "$SCRATCH_FQN"
```

The first call exits `0`; the second exits non-zero with a clean
not-found error.

### 3. `repo fork` returns a typed unsupported-capability error

```bash
bitbottle repo fork "$BB_TEST_SERVER_PROJECT/$RENAMED_SLUG" --into anywhere
```

Exit code: non-zero. stderr clearly says fork is not supported on this
host (Cloud only). **Critical:** no panic, no stack trace — just the
domain-error message naming the host and the missing capability.

## Cleanup

```bash
bitbottle repo delete "$BB_TEST_SERVER_PROJECT/$RENAMED_SLUG" --confirm 2>/dev/null || true
```
