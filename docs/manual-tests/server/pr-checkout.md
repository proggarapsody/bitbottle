# Scenario: Server/DC — check out someone else's PR

**Backend:** Server / Data Center.

Mirror of `cloud/pr-checkout.md`.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- An open PR on `$BB_TEST_SERVER_REPO` opened by a different user.
  Capture its ID into `$PR_ID`.
- Local clone:
  ```bash
  rm -rf /tmp/bb-server-pr-checkout
  bitbottle repo clone "$BB_TEST_SERVER_REPO" /tmp/bb-server-pr-checkout
  cd /tmp/bb-server-pr-checkout
  ```

## Steps

### 1. `pr view`

```bash
bitbottle pr view "$PR_ID"
```

Exit `0`. Author is NOT the current token's user.

### 2. `pr checkout`

```bash
bitbottle pr checkout "$PR_ID"
git rev-parse --abbrev-ref HEAD
```

`HEAD` matches the PR's source branch.

### 3. Diff equivalence

```bash
BASE=$(bitbottle pr view "$PR_ID" --json destinationBranch --jq .destinationBranch \
  2>/dev/null || echo main)
bitbottle pr diff "$PR_ID" | head -1
git diff "origin/$BASE"...HEAD | head -1
```

Both lines start with `diff --git` and reference the same first file.

### 4. `pr checkout` on a non-existent PR

```bash
bitbottle pr checkout 999999999
```

Exit code: non-zero. stderr names the PR ID as not found.

## Cleanup

```bash
cd /tmp && rm -rf /tmp/bb-server-pr-checkout
```
