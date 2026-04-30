# Scenario: Cloud tag lifecycle

**Backend:** Cloud.

Create a lightweight tag and an annotated tag, list them, delete them.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_REPO` exists with a `main` branch.

```bash
export TAG_LW="qa-lw-$(date +%s)"
export TAG_AN="qa-an-$(date +%s)"
```

## Steps

### 1. Create lightweight tag

```bash
bitbottle tag create "$BB_TEST_CLOUD_REPO" "$TAG_LW" --start-at main
```

Exit code: `0`.

**Verify in UI:** Tags page shows `$TAG_LW` pointing at the tip of `main`,
no annotation message.

### 2. Create annotated tag

```bash
bitbottle tag create "$BB_TEST_CLOUD_REPO" "$TAG_AN" \
  --start-at main --message "Release $TAG_AN"
```

Exit code: `0`.

**Verify in UI:** Tags page shows `$TAG_AN` with the message
`Release $TAG_AN`.

### 3. `tag list` includes both

```bash
bitbottle tag list "$BB_TEST_CLOUD_REPO" --limit 100 \
  | awk -v lw="$TAG_LW" -v an="$TAG_AN" '$1==lw||$1==an'
```

Two lines, one per tag, each with an 8-char hash in column 2.

### 4. `tag list --json`

```bash
bitbottle tag list "$BB_TEST_CLOUD_REPO" --json name,hash --limit 100 \
  | jq -r --arg t "$TAG_AN" '.[] | select(.name==$t) | .hash'
```

Stdout: a hex hash. Exit code: `0`.

### 5. `tag create` with no `--start-at` is rejected

```bash
bitbottle tag create "$BB_TEST_CLOUD_REPO" qa-bad
```

Exit code: non-zero. stderr names `--start-at` as required.

### 6. `tag create` against a bogus ref fails clearly

```bash
bitbottle tag create "$BB_TEST_CLOUD_REPO" qa-bad --start-at does-not-exist-branch
```

Exit code: non-zero. stderr says the ref was not found.

### 7. `tag delete`

```bash
bitbottle tag delete "$BB_TEST_CLOUD_REPO" "$TAG_LW" --confirm
bitbottle tag delete "$BB_TEST_CLOUD_REPO" "$TAG_AN" --confirm
bitbottle tag list "$BB_TEST_CLOUD_REPO" --limit 100 | grep -F "$TAG_LW" || echo "gone"
```

Last command prints `gone`. Exit code: `0` for the deletes.

> **Note:** `--confirm` is required when stdin is not a TTY (e.g. scripts). In
> an interactive terminal you can omit it and confirm the prompt instead.

### 8. Deleting a non-existent tag fails clearly

```bash
bitbottle tag delete "$BB_TEST_CLOUD_REPO" qa-never-existed
```

Exit code: non-zero. stderr names the tag and says it was not found.

## Cleanup

```bash
bitbottle tag delete "$BB_TEST_CLOUD_REPO" "$TAG_LW" --confirm 2>/dev/null || true
bitbottle tag delete "$BB_TEST_CLOUD_REPO" "$TAG_AN" --confirm 2>/dev/null || true
```
