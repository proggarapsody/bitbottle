# Scenario: Server review extras (PR tasks, suggestions, comment reactions)

**Backend:** Server / Data Center only. Cloud has no `pr task` API and
the suggestion/reaction endpoints used here are Server-specific (Cloud
returns typed `host.unsupported`).

Covers **TASK** (`pr task list/create/resolve/reopen`), **PR-SUGGESTION**
(`pr suggestion apply [--preview]`), and **REACT-PR** (`pr comment
react/unreact`, `--reactions` listing).

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_REPO` exists with `main`.
- An open scratch PR. The simplest way: run the existing
  [`server/pr-happy-path.md`](pr-happy-path.md) up through step 6
  (`pr edit`), then come back here. Capture:
  - `PR_ID` — the open PR's ID
  - `FB`    — the source branch name
  - The PR contains at least one modifiable line (e.g. `MANUAL_TEST.txt`).

## Setup

```bash
cd /tmp/bb-pr-happy   # or wherever the PR's clone lives
git pull --ff-only

# Sanity: confirm PR is still open.
bitbottle pr view "$PR_ID" --hostname "$BB_TEST_SERVER_HOST" | grep -i state
```

## Steps — PR tasks (TASK)

Server's "tasks" map to severity-`BLOCKER` comments on Server ≥ 7.2 (the
`SRVVER` helper gates this).

### 1. `pr task create`

```bash
bitbottle pr task create "$PR_ID" \
  --body "QA: fix this before merge" \
  --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: `0`. stderr prints the new task ID. Capture it:

```bash
export TASK_ID=$(bitbottle pr task list "$PR_ID" --hostname "$BB_TEST_SERVER_HOST" \
  --json id,text | jq -r '.[] | select(.text=="QA: fix this before merge") | .id')
echo "TASK_ID=$TASK_ID"
```

### 2. `pr task list` (default `--state open`)

```bash
bitbottle pr task list "$PR_ID" --hostname "$BB_TEST_SERVER_HOST" \
  | grep -F "QA: fix this before merge"
```

`grep` exits `0`. Columns: `ID`, `STATE`, `AUTHOR`, `TEXT`. The new task
is in state `OPEN`.

### 3. `pr task list --state all` includes resolved tasks too

```bash
bitbottle pr task list "$PR_ID" --state all --hostname "$BB_TEST_SERVER_HOST"
```

Same row appears with `STATE=OPEN`. Resolved tasks (none yet) would show
`STATE=RESOLVED`.

### 4. `pr task resolve`

```bash
bitbottle pr task resolve "$PR_ID" "$TASK_ID" --hostname "$BB_TEST_SERVER_HOST"
bitbottle pr task list    "$PR_ID" --state all --hostname "$BB_TEST_SERVER_HOST" \
  --json id,state | jq '.[] | select(.id=='"$TASK_ID"').state'
```

Final `jq` prints `"RESOLVED"`.

### 5. `pr task reopen`

```bash
bitbottle pr task reopen "$PR_ID" "$TASK_ID" --hostname "$BB_TEST_SERVER_HOST"
bitbottle pr task list   "$PR_ID" --state open --hostname "$BB_TEST_SERVER_HOST" \
  --json id,state | jq '.[] | select(.id=='"$TASK_ID"').state'
```

Final `jq` prints `"OPEN"`.

### 6. Bogus task ID gives a typed error

```bash
bitbottle pr task resolve "$PR_ID" 99999999 --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr mentions "not found" — typed `ErrNotFound`,
not a raw 404.

### 7. `pr task` on Cloud returns `host.unsupported`

```bash
bitbottle pr task list 1 --hostname "$BB_TEST_CLOUD_HOST" 2>&1 | grep -i unsupported
```

`grep` exits `0` (typed error message present). Exit code non-zero.

## Steps — PR comment reactions (REACT-PR, Server only)

### 8. Add a top-level comment to react to

```bash
bitbottle pr comment add "$PR_ID" --body "QA: comment for reactions" \
  --hostname "$BB_TEST_SERVER_HOST"
export RX_CID=$(bitbottle pr comment list "$PR_ID" --hostname "$BB_TEST_SERVER_HOST" \
  --json id,text | jq -r '.[] | select(.text=="QA: comment for reactions") | .id')
echo "RX_CID=$RX_CID"
```

### 9. `pr comment react`

```bash
bitbottle pr comment react "$PR_ID" "$RX_CID" --emoji thumbs-up \
  --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: `0`. Also try a shortcode form:

```bash
bitbottle pr comment react "$PR_ID" "$RX_CID" --emoji ":heart:" \
  --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: `0`. The CLI normalises both forms.

### 10. `pr comment list --reactions` shows them

```bash
bitbottle pr comment list "$PR_ID" --reactions \
  --hostname "$BB_TEST_SERVER_HOST" | grep "$RX_CID"
```

`grep` exits `0`. Output includes a reactions column with `:thumbsup:`
and `:heart:` (or counted equivalents).

### 11. `pr comment unreact`

```bash
bitbottle pr comment unreact "$PR_ID" "$RX_CID" --emoji thumbs-up \
  --hostname "$BB_TEST_SERVER_HOST"
bitbottle pr comment list    "$PR_ID" --reactions \
  --hostname "$BB_TEST_SERVER_HOST" --json id,reactions \
  | jq '.[] | select(.id=='"$RX_CID"').reactions | map(.emoji) | index("thumbsup")'
```

Final `jq` prints `null` (no thumbs-up; only `:heart:` remains).

### 12. `pr comment list --reactions` on Cloud returns typed `host.unsupported`

```bash
bitbottle pr comment list 1 --reactions --hostname "$BB_TEST_CLOUD_HOST" 2>&1 \
  | grep -i unsupported
```

`grep` exits `0`. Exit code non-zero. (The base `pr comment list` works on
Cloud — only the `--reactions` flag is gated.)

## Steps — PR suggestions (PR-SUGGESTION, Server only)

Suggestions are inline review comments that contain a fenced
```` ```suggestion ```` block. Applying one commits the replacement to
the PR's source branch.

### 13. Post an inline `suggestion` comment

Pick a line in `MANUAL_TEST.txt` and write a replacement using the
suggestion fence:

```bash
bitbottle pr comment add "$PR_ID" \
  --inline MANUAL_TEST.txt:1 \
  --body $'Consider this rewrite:\n\n```suggestion\nManual test (qa: suggestion)\n```' \
  --hostname "$BB_TEST_SERVER_HOST"

export SG_CID=$(bitbottle pr comment list "$PR_ID" --inline \
  --hostname "$BB_TEST_SERVER_HOST" --json id,text \
  | jq -r '.[] | select(.text|test("```suggestion")) | .id' | head -1)
echo "SG_CID=$SG_CID"
```

Server APIs use commentId + suggestionId (often `1` for single-suggestion
blocks). If your Server build exposes a `suggestionId` field in the
comment JSON, use that — otherwise default to `1`:

```bash
export SG_ID="${SG_ID:-1}"
```

### 14. `pr suggestion apply --preview`

```bash
bitbottle pr suggestion apply "$PR_ID" "$SG_CID" "$SG_ID" --preview \
  --hostname "$BB_TEST_SERVER_HOST"
```

Stdout prints the suggestion body (no commit is made). Exit code: `0`.

### 15. `pr suggestion apply` (real)

```bash
bitbottle pr suggestion apply "$PR_ID" "$SG_CID" "$SG_ID" \
  --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: `0`. **Verify in UI:** the PR's source branch now contains
a new commit replacing `MANUAL_TEST.txt:1` with the suggestion body. The
comment is marked applied.

### 16. Applying an already-applied suggestion errors clearly

```bash
bitbottle pr suggestion apply "$PR_ID" "$SG_CID" "$SG_ID" \
  --hostname "$BB_TEST_SERVER_HOST"
```

Exit code: non-zero. stderr mentions "already applied" or similar
(record exact wording).

### 17. `pr suggestion apply` on Cloud returns typed `host.unsupported`

```bash
bitbottle pr suggestion apply 1 1 1 --hostname "$BB_TEST_CLOUD_HOST" 2>&1 \
  | grep -i unsupported
```

`grep` exits `0`. Exit code non-zero.

## Cleanup

```bash
# Remove the reaction-test comment and any remaining tasks.
[ -n "${RX_CID:-}" ] && bitbottle pr comment delete "$PR_ID" "$RX_CID" \
  --hostname "$BB_TEST_SERVER_HOST" 2>/dev/null || true

[ -n "${TASK_ID:-}" ] && bitbottle pr task resolve "$PR_ID" "$TASK_ID" \
  --hostname "$BB_TEST_SERVER_HOST" 2>/dev/null || true

# Don't merge or decline the scratch PR here — the parent
# server/pr-happy-path.md scenario owns its lifecycle.
```
