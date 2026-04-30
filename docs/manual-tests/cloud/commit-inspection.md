# Scenario: Cloud commit inspection (log, view, status)

**Backend:** Cloud. Hits `/2.0/repositories/{ws}/{repo}/commit/{hash}/statuses`.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_REPO` has at least 3 commits on `main`.
- Local clone (any). Set up and capture a commit hash:
  ```bash
  # Create a fresh clone if one doesn't already exist
  [ -d /tmp/bb-cloud-qa ] || bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-cloud-qa
  cd /tmp/bb-cloud-qa
  git fetch origin main
  export HASH=$(git rev-parse origin/main)
  export SHORT=${HASH:0:7}
  ```

## Steps

### 1. `commit log` defaults to current branch → main

```bash
bitbottle commit log "$BB_TEST_CLOUD_REPO" --limit 3
```

TTY output: `HASH … MESSAGE … AUTHOR … DATE` header followed by 3 rows.
Hashes are 7-char short. Exit code: `0`.

### 2. `commit log --branch`

```bash
bitbottle commit log "$BB_TEST_CLOUD_REPO" --branch main --limit 1
```

Exactly one data row. The hash matches `$SHORT`.

### 3. `commit log` pipe gives full hash + RFC3339

```bash
bitbottle commit log "$BB_TEST_CLOUD_REPO" --limit 1 \
  | awk -F'\t' '{print length($1), $4}'
```

First field is `40` (full hash length), last field parses as RFC3339.

### 4. `commit log --json`

```bash
bitbottle commit log "$BB_TEST_CLOUD_REPO" --limit 1 --json hash,message,author \
  | jq '.[0] | keys | sort'
```

Stdout: `["author","hash","message"]`. Exit code: `0`.

### 5. `commit view` by full hash

```bash
bitbottle commit view "$BB_TEST_CLOUD_REPO" "$HASH"
```

Stdout includes lines:

- `commit <full-hash>`
- a blank line
- the commit message
- `Author:  …`
- `Date:    …`
- `Web:     https://bitbucket.org/…/commits/$HASH`

Exit code: `0`.

### 6. `commit view --web`

```bash
bitbottle commit view "$BB_TEST_CLOUD_REPO" "$HASH" --web
```

Browser opens at the commit page. Exit code: `0`.

### 7. `commit view` of a bogus hash fails clearly

```bash
bitbottle commit view "$BB_TEST_CLOUD_REPO" 0000000000000000000000000000000000000000
```

Exit code: non-zero. stderr says the commit was not found.

### 8. `commit status`

```bash
bitbottle commit status "$BB_TEST_CLOUD_REPO" "$HASH"
```

If the repo has Pipelines / external CI configured against this commit:
TTY table with at least `KEY … STATE` columns. Else: stdout is empty (or
"No statuses." stderr) and exit `0`.

### 9. `commit status --json`

```bash
bitbottle commit status "$BB_TEST_CLOUD_REPO" "$HASH" --json key,state
```

Stdout: a JSON array (possibly empty). Each entry has exactly `key` and
`state`. Exit code: `0`.

## Cleanup

None — read-only.
