# Scenario: Server/DC commit inspection (log, view, status)

**Backend:** Server / Data Center.

`commit log` / `view` use `rest/api/1.0/...`. **`commit status` is the
divergent path** — Server/DC uses the dedicated
`rest/build-status/1.0/commits/{hash}` endpoint, not the Cloud-style
`/commit/{hash}/statuses` URL. This scenario must be run separately from
the Cloud version.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_REPO` has ≥ 3 commits on the default branch.
- Local clone:
  ```bash
  cd /tmp/bb-server-qa
  git fetch origin
  export DEFAULT_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null || echo main)
  export HASH=$(git rev-parse origin/$DEFAULT_BRANCH)
  export SHORT=${HASH:0:7}
  ```

## Steps

### 1. `commit log` defaults

```bash
bitbottle commit log "$BB_TEST_SERVER_REPO" --limit 3
```

TTY header `HASH … MESSAGE … AUTHOR … DATE`. 3 data rows. 7-char hashes.

### 2. `commit log --branch`

```bash
bitbottle commit log "$BB_TEST_SERVER_REPO" --branch "$DEFAULT_BRANCH" --limit 1
```

One row. Hash equals `$SHORT`.

### 3. Pipe gives full hash + RFC3339

```bash
bitbottle commit log "$BB_TEST_SERVER_REPO" --limit 1 \
  | awk -F'\t' '{print length($1), $4}'
```

First column: `40`. Last: parses as RFC3339.

### 4. `commit log --json`

```bash
bitbottle commit log "$BB_TEST_SERVER_REPO" --limit 1 --json hash,message,author \
  | jq '.[0] | keys | sort'
```

Stdout: `["author","hash","message"]`.

### 5. `commit view` by full hash

```bash
bitbottle commit view "$BB_TEST_SERVER_REPO" "$HASH"
```

Stdout includes `commit $HASH`, the message, `Author:`, `Date:`, and a
`Web:` URL pointing at
`https://$BB_TEST_SERVER_HOST/projects/.../commits/$HASH`.

### 6. `commit view --web`

```bash
bitbottle commit view "$BB_TEST_SERVER_REPO" "$HASH" --web
```

Browser opens at the commit page.

### 7. `commit view` of a bogus hash

```bash
bitbottle commit view "$BB_TEST_SERVER_REPO" 0000000000000000000000000000000000000000
```

Exit code: non-zero. stderr says not found.

### 8. `commit status` against `build-status/1.0`

```bash
bitbottle commit status "$BB_TEST_SERVER_REPO" "$HASH"
```

If CI has reported statuses against this commit (Bamboo, Jenkins, etc.):
TTY table with `KEY … STATE` columns. Else: empty stdout, exit `0`.

### 9. `commit status --json`

```bash
bitbottle commit status "$BB_TEST_SERVER_REPO" "$HASH" --json key,state
```

Stdout: a JSON array (possibly empty). Each entry has exactly `key` and
`state`. Exit code: `0`.

### 10. `commit status` does NOT 404 silently for a bogus hash

```bash
bitbottle commit status "$BB_TEST_SERVER_REPO" 0000000000000000000000000000000000000000
```

Either: empty array `[]` + exit `0` (the build-status endpoint returns no
results for unknown hashes), OR exit non-zero with a clear "not found"
error. Record which.

## Cleanup

None.
