# Scenario: Cloud branch lifecycle

**Backend:** Cloud.

Create a branch from `main`, list branches, check it out locally, delete it.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_REPO` exists and has a `main` branch with at least one
  commit.
- Local clone of the repo at `/tmp/bb-cloud-qa`:
  ```bash
  rm -rf /tmp/bb-cloud-qa
  bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-cloud-qa
  cd /tmp/bb-cloud-qa
  ```
- Branch name unique to this run:
  ```bash
  export FB="qa/branch-lifecycle-$(date +%s)"
  ```

## Steps

### 1. `branch create --start-at main`

```bash
bitbottle branch create "$BB_TEST_CLOUD_REPO" "$FB" --start-at main
```

Exit code: `0`. stderr/stdout confirms creation.

**Verify in UI:** the branch appears under Branches.

### 2. `branch list` shows it

```bash
bitbottle branch list "$BB_TEST_CLOUD_REPO" --limit 100 | grep -F "$FB"
```

`grep` exits `0`. Row shows `<name>\tfalse\t<8-char-hash>` (DEFAULT=false).

### 3. `branch list --json`

```bash
bitbottle branch list "$BB_TEST_CLOUD_REPO" --json name,default,hash --limit 100 \
  | jq -r --arg fb "$FB" '.[] | select(.name==$fb) | .hash' | head -1
```

Stdout: a hex hash (≥ 8 chars). Exit code: `0`.

### 4. `branch checkout` fetches and switches

```bash
cd /tmp/bb-cloud-qa
bitbottle branch checkout "$FB"
git rev-parse --abbrev-ref HEAD
```

Last command prints `$FB`. Exit code: `0` for both.

### 5. Create a branch from a specific commit hash

```bash
HASH=$(git rev-parse main)
export FB2="qa/branch-from-hash-$(date +%s)"
bitbottle branch create "$BB_TEST_CLOUD_REPO" "$FB2" --start-at "$HASH"
bitbottle branch list "$BB_TEST_CLOUD_REPO" --json name,hash --limit 100 \
  | jq -r --arg fb "$FB2" '.[] | select(.name==$fb) | .hash'
```

Returned hash starts with `${HASH:0:8}`.

### 6. `branch create` with a missing `--start-at` is rejected

```bash
bitbottle branch create "$BB_TEST_CLOUD_REPO" "qa/no-start"
```

Exit code: non-zero. stderr names `--start-at` as required.

### 7. `branch delete` removes both branches

```bash
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB"
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB2"
bitbottle branch list "$BB_TEST_CLOUD_REPO" --limit 100 | grep -F "$FB" || echo "gone"
```

Final command prints `gone`. UI no longer shows the branches.

### 8. Deleting a non-existent branch fails clearly

```bash
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "qa/never-existed"
```

Exit code: non-zero. stderr names the branch and says it was not found.

## Cleanup

```bash
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB"  2>/dev/null || true
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB2" 2>/dev/null || true
```
