# Scenario: Server/DC PR comments

**Backend:** Server / Data Center. **`pr comment list` walks the
`activities` feed and filters for `COMMENTED` events** — this is the
divergent path from Cloud, which hits a direct `comments` endpoint.

This scenario must be run separately from `cloud/pr-comments.md`; the goal
is to exercise the activities-feed walker.

## Prerequisites

- An open PR on `$BB_TEST_SERVER_REPO`. Easiest: run setup from
  `server/pr-happy-path.md` through step 2 only, then capture `$PR_ID`.

## Steps

### 1. `pr comment list` on a PR with no top-level comments

```bash
bitbottle pr comment list "$PR_ID"
```

Empty stdout (or header + zero rows), exit `0`. The activities feed almost
always contains non-`COMMENTED` events (open, approve, etc.) — verify
those are correctly filtered out.

### 2. Add a top-level comment

```bash
bitbottle pr comment add "$PR_ID" --body "manual-test top-level comment"
```

Exit `0`. UI shows it as a general comment, not an inline comment.

### 3. Add a multi-line comment

```bash
bitbottle pr comment add "$PR_ID" --body "$(printf 'line one\nline two')"
```

Exit `0`. UI renders both lines.

### 4. Add an inline / file comment via the UI (manual step)

In the Bitbucket UI, leave an **inline** comment on a specific line of the
PR's diff. This MUST NOT appear in `pr comment list` output (which is
top-level only).

### 5. `pr comment list` returns top-level only

```bash
bitbottle pr comment list "$PR_ID"
```

Stdout contains the two top-level comments from steps 2-3. The inline
comment from step 4 is NOT present. This is the key behavioral guarantee
for Server/DC.

### 6. `pr comment list --json`

```bash
bitbottle pr comment list "$PR_ID" --json id,author,text \
  | jq 'length'
```

`length` is `2` (only top-level — inline excluded).

### 7. `pr comment list --jq`

```bash
bitbottle pr comment list "$PR_ID" --json text --jq '.[].text'
```

One body per line; both top-level bodies appear.

### 8. `pr comment add` without `--body` rejected

```bash
bitbottle pr comment add "$PR_ID"
```

Exit code: non-zero. stderr names `--body` as required.

### 9. Add against a non-existent PR fails clearly

```bash
bitbottle pr comment add 999999999 --body "should fail"
```

Exit code: non-zero. stderr names the PR as not found.

## Cleanup

Comments persist; delete from the UI if desired.
