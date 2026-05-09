# Scenario: Server/DC PR happy path

**Backend:** Server / Data Center.

Full PR lifecycle. Same CLI shape as Cloud, but the underlying API is
`rest/api/1.0/projects/.../pull-requests`.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_REPO` exists; default branch known.
- `BB_TEST_SERVER_REVIEWER` set to a username that exists on the server
  (otherwise skip step 7).
- Local clone:
  ```bash
  rm -rf /tmp/bb-server-pr-happy
  bitbottle repo clone "$BB_TEST_SERVER_REPO" /tmp/bb-server-pr-happy
  cd /tmp/bb-server-pr-happy
  export DEFAULT_BRANCH=$(git symbolic-ref --short HEAD)
  export FB="qa/pr-happy-$(date +%s)"
  ```

## Steps

### 1. Branch + commit + push

```bash
git checkout -b "$FB" "origin/$DEFAULT_BRANCH"
echo "manual test $(date)" >> MANUAL_TEST.txt
git add MANUAL_TEST.txt
git commit -m "qa: pr-happy-path"
git push -u origin "$FB"
```

### 2. `pr create --draft`

```bash
bitbottle pr create \
  --title "QA: pr-happy-path" \
  --body  "Manual test PR; safe to ignore." \
  --base  "$DEFAULT_BRANCH" \
  --draft
export PR_ID=$(bitbottle pr list --json id,title --limit 50 \
  | jq -r '.[] | select(.title=="QA: pr-happy-path") | .id' | head -1)
echo "PR_ID=$PR_ID"
```

`PR_ID` is a positive integer. (On Server/DC, "draft" maps to PR
description prefix or a flag depending on version — record the actual
representation.)

### 3. `pr list --state open` includes it

```bash
bitbottle pr list --state open | grep -F "QA: pr-happy-path"
```

### 4. `pr view`

Stdout shows title, body, source `$FB`, target `$DEFAULT_BRANCH`, author.

### 5. `pr diff`

```bash
bitbottle pr diff "$PR_ID" | head -5
```

Output includes `+++ b/MANUAL_TEST.txt`.

### 6. `pr edit`

```bash
bitbottle pr edit "$PR_ID" --title "QA: pr-happy-path (edited)" --body "Edited."
bitbottle pr view "$PR_ID" | grep -F "(edited)"
```

### 7. `pr request-review`

```bash
bitbottle pr request-review "$PR_ID" --reviewer "$BB_TEST_SERVER_REVIEWER"
```

Exit code: `0`. UI shows the reviewer.

### 8. `pr approve`

```bash
bitbottle pr approve "$PR_ID"
```

On Server/DC self-approval is permitted by default — exit `0`. If the
project disallows it: non-zero with a clear message.

### 8.5. `pr comment list` exposes the new inline-aware surface

Server activities flatten nested reply trees with `parentId` set; inline
anchors come from `commentAnchor` (FROM = old side, TO = new side).
`resolved` is always `false` on Server (Bitbucket Server's resolution
lives on tasks; out of scope until RV3).

```bash
bitbottle pr comment add  "$PR_ID" --body "QA: testing pr comment list"
bitbottle pr comment list "$PR_ID"
bitbottle pr comment list "$PR_ID" --inline
bitbottle pr comment list "$PR_ID" \
  --json id,parentId,resolved,updatedAt,inline | jq 'length >= 1'
```

Each command exits `0`. The last `jq` prints `true`.

### 8.6. `pr comment add --inline` posts an anchored review comment

Server requires `fromHash`/`toHash` from the PR diff envelope. The CLI
fetches them transparently — the user only supplies `path:line`.

```bash
bitbottle pr comment add "$PR_ID" \
  --inline MANUAL_TEST.txt:1 \
  --body "QA: inline new-side"
INLINE_ID=$(bitbottle pr comment list "$PR_ID" --inline --json id,inline,text \
  | jq '.[] | select(.text=="QA: inline new-side") | .id')

# Multi-line on Server is rejected with a typed error — verify wording.
bitbottle pr comment add "$PR_ID" \
  --inline MANUAL_TEST.txt:1-3 \
  --body "QA: should fail" || echo "rejected (expected)"

# Reply nested under the first inline comment
bitbottle pr comment add "$PR_ID" --parent "$INLINE_ID" --body "QA: reply"
```

The first command exits `0` and the new inline comment shows up under
`pr comment list --inline`. The multi-line attempt prints
"rejected (expected)" and a "multi-line inline comments are not
supported on Bitbucket Server / Data Center" message on stderr. The
reply appears under the inline thread in the Bitbucket Server UI.

### 8.7. `pr comment edit / delete` (Server)

```bash
bitbottle pr comment edit   "$PR_ID" "$INLINE_ID" --body "QA: edited"
bitbottle pr comment list   "$PR_ID" --json id,text | jq '.[] | select(.id=='"$INLINE_ID"').text'
bitbottle pr comment delete "$PR_ID" "$INLINE_ID"

# Resolve is unavailable on Server — verify the typed error.
bitbottle pr comment resolve "$PR_ID" "$INLINE_ID" 2>&1 \
  | grep -i "not supported"
```

The `text` jq prints `"QA: edited"`. `delete` exits `0`. The `resolve`
attempt non-zero and the grep matches — Server returns the typed
`host.unsupported` error because comment resolution lives on tasks, not
regular comments.

### 9. `pr ready`

```bash
bitbottle pr ready "$PR_ID"
```

Exit code: `0`. PR no longer shows draft state.

### 10. `pr merge --squash --delete-branch`

```bash
bitbottle pr merge "$PR_ID" --squash --delete-branch
```

Exit code: `0`. UI: PR merged, branch `$FB` removed.

> Note: Server/DC may require an additional permission for squash merges
> depending on project config; if non-zero, retry with the default
> strategy:
> ```bash
> bitbottle pr merge "$PR_ID" --merge --delete-branch
> ```

### 11. `pr list --state merged` includes it

```bash
bitbottle pr list --state merged --limit 50 | grep -F "QA: pr-happy-path"
```

## Cleanup

```bash
git checkout "$DEFAULT_BRANCH" 2>/dev/null || true
bitbottle branch delete "$BB_TEST_SERVER_REPO" "$FB" 2>/dev/null || true
```
