# Scenario: Cloud repo extras (rename + fork)

**Backend:** Cloud.

Exercise the scope-Q additions: `repo rename` (both backends) and
`repo fork` (Cloud only).

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_WORKSPACE` set; token has `repository:admin` scope.
- A second workspace `BB_TEST_CLOUD_FORK_WORKSPACE` you can write to (for
  the fork step). If you don't have one, skip section 3.
- A unique scratch slug:
  ```bash
  export SCRATCH_SLUG="bb-qa-rx-$(date +%s)"
  export SCRATCH_FQN="$BB_TEST_CLOUD_WORKSPACE/$SCRATCH_SLUG"
  ```

## Steps

### 1. Set up — create scratch repo

```bash
bitbottle repo create "$SCRATCH_SLUG" \
  --project "$BB_TEST_CLOUD_WORKSPACE" \
  --description "bitbottle scope-Q rename + fork test" \
  --private=true
```

Exit code: `0`.

### 2. `repo rename` updates name and slug

```bash
export RENAMED_SLUG="${SCRATCH_SLUG}-v2"
bitbottle repo rename "$SCRATCH_FQN" "$RENAMED_SLUG"
```

Exit code: `0`. stdout reports the rename and prints the new
`workspace/slug` coordinate.

**Verify:**
```bash
bitbottle repo view "$BB_TEST_CLOUD_WORKSPACE/$RENAMED_SLUG"
# old slug should now 404:
bitbottle repo view "$SCRATCH_FQN"
```

The first call exits `0`; the second exits non-zero with a clean
not-found error (no panic, no stack).

### 3. `repo fork` into another workspace _(Cloud only)_

Skip this section if `BB_TEST_CLOUD_FORK_WORKSPACE` is unset.

```bash
bitbottle repo fork "$BB_TEST_CLOUD_WORKSPACE/$RENAMED_SLUG" \
  --into "$BB_TEST_CLOUD_FORK_WORKSPACE"
```

Exit code: `0`. stdout reports the new fork's `workspace/slug` and the
fork web URL.

**Verify in UI:** new repo appears under
`https://bitbucket.org/$BB_TEST_CLOUD_FORK_WORKSPACE/$RENAMED_SLUG`,
labelled as a fork of the source.

### 4. `repo fork --name` overrides the fork's slug

```bash
export FORK_NAME="${RENAMED_SLUG}-fork"
bitbottle repo fork "$BB_TEST_CLOUD_WORKSPACE/$RENAMED_SLUG" \
  --into "$BB_TEST_CLOUD_FORK_WORKSPACE" \
  --name "$FORK_NAME"
```

Exit code: `0`. stdout shows `$BB_TEST_CLOUD_FORK_WORKSPACE/$FORK_NAME`.

### 5. `repo fork` without `--into` errors with a clear message

```bash
bitbottle repo fork "$BB_TEST_CLOUD_WORKSPACE/$RENAMED_SLUG"
```

Exit code: non-zero. stderr mentions `--into` is required.

## Cleanup

```bash
bitbottle repo delete "$BB_TEST_CLOUD_WORKSPACE/$RENAMED_SLUG" --confirm 2>/dev/null || true
bitbottle repo delete "$BB_TEST_CLOUD_FORK_WORKSPACE/$RENAMED_SLUG" --confirm 2>/dev/null || true
bitbottle repo delete "$BB_TEST_CLOUD_FORK_WORKSPACE/$FORK_NAME" --confirm 2>/dev/null || true
```
