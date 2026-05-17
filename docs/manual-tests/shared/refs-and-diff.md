# Scenario: Branches, tags, commits, and `diff` between refs

**Backend:** Both. Each step is run against both `$BB_TEST_CLOUD_REPO` and
`$BB_TEST_SERVER_REPO`; only deviations between backends are called out.

Exercises the **L** (branch create/checkout), **E** (tag CRUD), **F**
(commit log/view), **K** + **COMMIT-STATUS-REPORT** (commit status read +
write), **COMMIT-FILE** (commit files), **DIFF** (REF1..REF2), and the
**RV6** commit-comment + **REACT-COMMIT** reactions surface — all things
that hit real git refs and real repo state.

## Prerequisites

- Logged in to both backends.
- Scratch repos exist on both with at least one commit on `main`.
- `jq` available locally.

## Steps

Per backend, set the host/repo and clone:

```bash
# Cloud variant
export BB_HOST="$BB_TEST_CLOUD_HOST"; export BB_REPO="$BB_TEST_CLOUD_REPO"
# Server variant (re-run below from here)
# export BB_HOST="$BB_TEST_SERVER_HOST"; export BB_REPO="$BB_TEST_SERVER_REPO"

rm -rf /tmp/bb-refs
bitbottle repo clone "$BB_REPO" /tmp/bb-refs --hostname "$BB_HOST"
cd /tmp/bb-refs
git checkout main
export FB="qa/refs-$(date +%s)"
```

### 1. `branch create --start-at`

```bash
bitbottle branch create "$BB_REPO" "$FB" --start-at main --hostname "$BB_HOST"
```

Exit code: `0`. stderr confirms creation. The branch exists on the server
(`git fetch` then `git branch -r` shows `origin/$FB`).

### 2. `branch list` includes the new branch

```bash
bitbottle branch list "$BB_REPO" --hostname "$BB_HOST" | grep -F "$FB"
```

`grep` exits `0`. With `--json`:

```bash
bitbottle branch list "$BB_REPO" --hostname "$BB_HOST" --json name \
  | jq -r '.[].name' | grep -Fx "$FB"
```

### 3. `branch checkout` switches local working tree

```bash
bitbottle branch checkout "$FB"
git rev-parse --abbrev-ref HEAD
```

Stdout of `git rev-parse` is `$FB`. Exit code: `0`.

### 4. Push two commits on the new branch

```bash
echo "alpha $(date)" >> MANUAL_TEST.txt
git add MANUAL_TEST.txt && git commit -m "qa: alpha"
echo "beta $(date)"  >> MANUAL_TEST.txt
git add MANUAL_TEST.txt && git commit -m "qa: beta"
git push -u origin "$FB"

export HEAD_HASH=$(git rev-parse HEAD)
export PREV_HASH=$(git rev-parse HEAD~1)
echo "HEAD=$HEAD_HASH PREV=$PREV_HASH"
```

### 5. `commit log` enumerates recent commits

```bash
bitbottle commit log "$BB_REPO" --branch "$FB" --hostname "$BB_HOST" --limit 5
```

Stdout table includes the two `qa:` subjects with hashes, authors, dates.
JSON variant exits `0` and parses:

```bash
bitbottle commit log "$BB_REPO" --branch "$FB" --hostname "$BB_HOST" \
  --json hash,message --limit 5 | jq '.[0].message' | grep -F "qa: beta"
```

### 6. `commit view` on `$HEAD_HASH`

```bash
bitbottle commit view "$BB_REPO" "$HEAD_HASH" --hostname "$BB_HOST"
```

Stdout shows hash, author, message (`qa: beta`), and the parent ref.
Exit code: `0`.

### 7. `commit files` lists changed paths

```bash
bitbottle commit files "$HEAD_HASH" "$BB_REPO" --hostname "$BB_HOST"
```

Stdout lists `MANUAL_TEST.txt` with a status column (`M`/`A`/`D`).
Exit code: `0`.

### 8. `diff REF1..REF2` (full unified diff)

```bash
bitbottle diff "$PREV_HASH..$HEAD_HASH" "$BB_REPO" --hostname "$BB_HOST" | head -10
```

Output begins with `diff --git ` and includes `+++ b/MANUAL_TEST.txt`.
Exit code: `0`.

### 9. `diff --stat` summary

```bash
bitbottle diff "$PREV_HASH..$HEAD_HASH" "$BB_REPO" --stat --hostname "$BB_HOST"
```

Stdout is a one-line stat (`MANUAL_TEST.txt | 1 +`) plus a totals line.
Exit code: `0`.

### 10. `tag create` (lightweight)

```bash
export TAG_LIGHT="qa/tag-light-$(date +%s)"
bitbottle tag create "$BB_REPO" "$TAG_LIGHT" --start-at "$HEAD_HASH" \
  --hostname "$BB_HOST"
```

Exit code: `0`. `git fetch --tags && git tag -l "$TAG_LIGHT"` prints
the tag.

### 11. `tag create --message` (annotated)

```bash
export TAG_ANNO="qa/tag-anno-$(date +%s)"
bitbottle tag create "$BB_REPO" "$TAG_ANNO" --start-at "$HEAD_HASH" \
  --message "QA: annotated tag" --hostname "$BB_HOST"
```

Exit code: `0`. Verify it's annotated:

```bash
git fetch --tags
git for-each-ref "refs/tags/$TAG_ANNO" --format='%(objecttype)'  # → "tag"
```

Server: annotated tags are supported via REST when `message` is provided.
Cloud: annotated tag support is API-limited — record exit code and stderr
if Cloud falls back to lightweight.

### 12. `tag list` includes both tags

```bash
bitbottle tag list "$BB_REPO" --hostname "$BB_HOST" | grep -E "$TAG_LIGHT|$TAG_ANNO"
```

`grep` exits `0` and prints two rows.

### 13. Bogus `--start-at` gives a clear error

```bash
bitbottle branch create "$BB_REPO" "qa/bogus-$$" --start-at "nope-not-a-ref" \
  --hostname "$BB_HOST"
```

Exit code: non-zero. stderr mentions the bad ref (typed `ErrNotFound`),
not a raw 404.

### 14. `commit status report` writes a build status (COMMIT-STATUS-REPORT)

```bash
bitbottle commit status report "$BB_REPO" "$HEAD_HASH" \
  --key qa-manual --state INPROGRESS \
  --name "QA manual smoke" --description "running" \
  --url "https://example.test/build/1" \
  --hostname "$BB_HOST"
```

Exit code: `0`. Re-run with `--state SUCCESSFUL` to verify update path:

```bash
bitbottle commit status report "$BB_REPO" "$HEAD_HASH" \
  --key qa-manual --state SUCCESSFUL --hostname "$BB_HOST"
```

### 15. `commit status` lists the build status

```bash
bitbottle commit status "$BB_REPO" "$HEAD_HASH" --hostname "$BB_HOST"
```

Stdout lists a row with `KEY=qa-manual`, `STATE=SUCCESSFUL`.
Exit code: `0`.

### 16. `commit comment` lifecycle (RV6)

```bash
bitbottle commit comment add "$BB_REPO" "$HEAD_HASH" \
  --body "QA: top-level commit comment" --hostname "$BB_HOST"
export CCID=$(bitbottle commit comment list "$BB_REPO" "$HEAD_HASH" \
  --hostname "$BB_HOST" --json id,text \
  | jq -r '.[] | select(.text=="QA: top-level commit comment") | .id')
echo "CCID=$CCID"

bitbottle commit comment edit "$BB_REPO" "$HEAD_HASH" "$CCID" \
  --body "QA: edited commit comment" --hostname "$BB_HOST"
bitbottle commit comment list "$BB_REPO" "$HEAD_HASH" --hostname "$BB_HOST" \
  --json id,text | jq -r '.[] | select(.id=='"$CCID"').text'
```

The final `jq` prints `"QA: edited commit comment"`.

### 17. `commit comment react` / `unreact` (REACT-COMMIT — Server only)

```bash
if [ "$BB_HOST" = "$BB_TEST_SERVER_HOST" ]; then
  bitbottle commit comment react   "$BB_REPO" "$HEAD_HASH" "$CCID" \
    --emoji thumbs-up --hostname "$BB_HOST"
  bitbottle commit comment list "$BB_REPO" "$HEAD_HASH" --reactions \
    --hostname "$BB_HOST" | grep -F ":thumbsup:"
  bitbottle commit comment unreact "$BB_REPO" "$HEAD_HASH" "$CCID" \
    --emoji thumbs-up --hostname "$BB_HOST"
fi
```

Cloud: reactions are not supported; the commands return typed
`ErrUnsupportedOnHost` ("unsupported on host"). Skip this step for Cloud
and record the typed error.

### 18. `commit comment delete`

```bash
bitbottle commit comment delete "$BB_REPO" "$HEAD_HASH" "$CCID" --hostname "$BB_HOST"
bitbottle commit comment list   "$BB_REPO" "$HEAD_HASH" --hostname "$BB_HOST" \
  --json id | jq 'map(.id) | index('"$CCID"')'
```

Final `jq` prints `null`.

### 19. `tag delete` removes both tags

```bash
bitbottle tag delete "$BB_REPO" "$TAG_LIGHT" --hostname "$BB_HOST"
bitbottle tag delete "$BB_REPO" "$TAG_ANNO"  --hostname "$BB_HOST"
git fetch --prune --tags
git tag -l "$TAG_LIGHT" "$TAG_ANNO"
```

Final command prints nothing. Exit code: `0`.

### 20. `branch delete` removes the branch

```bash
git checkout main
bitbottle branch delete "$BB_REPO" "$FB" --hostname "$BB_HOST"
git fetch --prune
git branch -r | grep -F "origin/$FB" || echo "deleted"
```

Final line prints `deleted`.

## Cleanup

```bash
cd - >/dev/null
rm -rf /tmp/bb-refs
```

Re-run this entire scenario from the top with `BB_HOST`/`BB_REPO` switched
to the Server variant.
