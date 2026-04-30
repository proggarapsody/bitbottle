# Scenario: Server/DC tag lifecycle

**Backend:** Server / Data Center.

Mirror of `cloud/tag-lifecycle.md`. On Server/DC tags use the
`rest/api/1.0/projects/{p}/repos/{r}/tags` endpoint, and annotated tags
are created via the `git/tags` (or `branch-utils`) extension depending on
server version — record any version-specific behavior.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_REPO` exists with a default branch.

```bash
export DEFAULT_BRANCH=${DEFAULT_BRANCH:-main}
export TAG_LW="qa-lw-$(date +%s)"
export TAG_AN="qa-an-$(date +%s)"
```

## Steps

### 1. Create lightweight tag

```bash
bitbottle tag create "$BB_TEST_SERVER_REPO" "$TAG_LW" --start-at "$DEFAULT_BRANCH"
```

Exit code: `0`. **Verify in UI:** tag appears.

### 2. Create annotated tag

```bash
bitbottle tag create "$BB_TEST_SERVER_REPO" "$TAG_AN" \
  --start-at "$DEFAULT_BRANCH" --message "Release $TAG_AN"
```

Exit code: `0`. **Verify in UI:** tag appears with the message. (Some
older Server versions silently drop the message — record if so.)

### 3. `tag list` includes both

```bash
bitbottle tag list "$BB_TEST_SERVER_REPO" --limit 200 \
  | awk -v lw="$TAG_LW" -v an="$TAG_AN" '$1==lw||$1==an'
```

Two lines.

### 4. `tag list --json`

```bash
bitbottle tag list "$BB_TEST_SERVER_REPO" --json name,hash --limit 200 \
  | jq -r --arg t "$TAG_AN" '.[] | select(.name==$t) | .hash'
```

Stdout: a hex hash.

### 5. Missing `--start-at` rejected

```bash
bitbottle tag create "$BB_TEST_SERVER_REPO" qa-bad
```

Exit code: non-zero.

### 6. Bogus ref fails clearly

```bash
bitbottle tag create "$BB_TEST_SERVER_REPO" qa-bad --start-at does-not-exist
```

Exit code: non-zero. stderr says ref not found.

### 7. `tag delete`

```bash
bitbottle tag delete "$BB_TEST_SERVER_REPO" "$TAG_LW"
bitbottle tag delete "$BB_TEST_SERVER_REPO" "$TAG_AN"
bitbottle tag list "$BB_TEST_SERVER_REPO" --limit 200 | grep -F "$TAG_LW" || echo "gone"
```

Last command prints `gone`.

## Cleanup

```bash
bitbottle tag delete "$BB_TEST_SERVER_REPO" "$TAG_LW" 2>/dev/null || true
bitbottle tag delete "$BB_TEST_SERVER_REPO" "$TAG_AN" 2>/dev/null || true
```
