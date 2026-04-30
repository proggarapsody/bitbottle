# Scenario: Server/DC branch lifecycle

**Backend:** Server / Data Center.

Mirrors `cloud/branch-lifecycle.md`; the user-facing CLI is identical, the
underlying API is `rest/api/1.0/projects/{p}/repos/{r}/branches`.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_REPO` exists with a default branch (`main` or `master`).
- Local clone:
  ```bash
  rm -rf /tmp/bb-server-qa
  bitbottle repo clone "$BB_TEST_SERVER_REPO" /tmp/bb-server-qa
  cd /tmp/bb-server-qa
  export DEFAULT_BRANCH=$(git symbolic-ref --short HEAD)
  export FB="qa/branch-lifecycle-$(date +%s)"
  ```

## Steps

### 1. `branch create --start-at <default>`

```bash
bitbottle branch create "$BB_TEST_SERVER_REPO" "$FB" --start-at "$DEFAULT_BRANCH"
```

Exit code: `0`.

### 2. `branch list` includes it

```bash
bitbottle branch list "$BB_TEST_SERVER_REPO" --limit 200 | grep -F "$FB"
```

`grep` exits `0`. The DEFAULT column on the default-branch row is `true`.

### 3. `branch list --json`

```bash
bitbottle branch list "$BB_TEST_SERVER_REPO" --json name,default,hash --limit 200 \
  | jq -r --arg fb "$FB" '.[] | select(.name==$fb) | .hash'
```

Stdout: a hex hash. Exit code: `0`.

### 4. `branch checkout`

```bash
cd /tmp/bb-server-qa
bitbottle branch checkout "$FB"
git rev-parse --abbrev-ref HEAD
```

Last command prints `$FB`.

### 5. Create from a commit hash

```bash
HASH=$(git rev-parse "$DEFAULT_BRANCH")
export FB2="qa/branch-from-hash-$(date +%s)"
bitbottle branch create "$BB_TEST_SERVER_REPO" "$FB2" --start-at "$HASH"
```

Exit code: `0`.

### 6. `branch delete`

```bash
bitbottle branch delete "$BB_TEST_SERVER_REPO" "$FB"
bitbottle branch delete "$BB_TEST_SERVER_REPO" "$FB2"
bitbottle branch list "$BB_TEST_SERVER_REPO" --limit 200 | grep -F "$FB" || echo "gone"
```

Final command prints `gone`.

### 7. Delete a non-existent branch

```bash
bitbottle branch delete "$BB_TEST_SERVER_REPO" qa/never-existed
```

Exit code: non-zero. stderr names the branch as not found.

## Cleanup

```bash
bitbottle branch delete "$BB_TEST_SERVER_REPO" "$FB"  2>/dev/null || true
bitbottle branch delete "$BB_TEST_SERVER_REPO" "$FB2" 2>/dev/null || true
```
