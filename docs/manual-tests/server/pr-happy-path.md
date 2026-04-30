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
