# Scenario: Cloud — check out someone else's PR locally

**Backend:** Cloud.

Verifies `pr checkout` against a PR you did not create.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- An existing **open** PR on `$BB_TEST_CLOUD_REPO` opened by a different
  user. If none exists, first run `cloud/pr-happy-path.md` from another
  user account up through step 2 only (do not merge), then come back here.
  Capture its ID into `$PR_ID`.
- Local clone:
  ```bash
  rm -rf /tmp/bb-pr-checkout
  bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-pr-checkout
  cd /tmp/bb-pr-checkout
  ```

## Steps

### 1. `pr view` shows the PR you don't own

```bash
bitbottle pr view "$PR_ID"
```

Exit code: `0`. Author is NOT your token's username.

### 2. `pr checkout` fetches and switches

```bash
bitbottle pr checkout "$PR_ID"
git rev-parse --abbrev-ref HEAD
```

`HEAD` is now the PR's source branch (matches the `source.branch.name` from
`bitbottle api 2.0/.../pullrequests/$PR_ID --jq .source.branch.name`).
Working tree contains the PR's changes (compare a known file).

### 3. `pr diff` matches local diff against base

```bash
BASE=$(bitbottle pr view "$PR_ID" --json id,toBranch 2>/dev/null \
  | jq -r '.toBranch // "main"')
bitbottle pr diff "$PR_ID" | head -1
git diff "origin/$BASE"...HEAD | head -1
```

Both first lines start with `diff --git ` and reference the same first
file. (Whitespace-level equivalence is enough; exact byte-equality is not
required.)

### 4. Checking out a non-existent PR fails clearly

```bash
bitbottle pr checkout 999999999
```

Exit code: non-zero. stderr names the PR ID as not found.

## Cleanup

```bash
cd /tmp && rm -rf /tmp/bb-pr-checkout
```
