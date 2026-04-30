# Scenario: Cloud PR happy path (full lifecycle)

**Backend:** Cloud.

End-to-end: branch → commit → push → `pr create --draft` → `pr view` →
`pr diff` → `pr edit` → `pr request-review` → `pr approve` → `pr ready` →
`pr merge --squash --delete-branch`.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_REPO` exists with `main` branch.
- A second user account on the same workspace whose username is in
  `BB_TEST_CLOUD_REVIEWER` (used for `request-review`). If unavailable, skip
  step 7 and record that.
- Local clone:
  ```bash
  rm -rf /tmp/bb-pr-happy
  bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-pr-happy
  cd /tmp/bb-pr-happy
  export FB="qa/pr-happy-$(date +%s)"
  ```

## Steps

### 1. Create branch + commit + push

```bash
git checkout -b "$FB" origin/main
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
  --base  main \
  --draft
```

Exit code: `0`. stdout/stderr prints the PR URL. Capture the PR ID:

```bash
export PR_ID=$(bitbottle pr list --json id,title --limit 50 \
  | jq -r '.[] | select(.title=="QA: pr-happy-path") | .id' | head -1)
echo "PR_ID=$PR_ID"
```

`PR_ID` is a positive integer.

### 3. `pr list --state open` includes it

```bash
bitbottle pr list --state open | grep -F "QA: pr-happy-path"
```

`grep` exits `0`.

### 4. `pr view`

```bash
bitbottle pr view "$PR_ID"
```

Stdout includes the title, the body, the source branch (`$FB`), the target
(`main`), the author, and a state line indicating draft.

### 5. `pr diff` streams a unified diff

```bash
bitbottle pr diff "$PR_ID" | head -5
```

Output begins with `diff --git ` and includes `+++ b/MANUAL_TEST.txt`. Pipe
exits `0`.

### 6. `pr edit` updates title + body

```bash
bitbottle pr edit "$PR_ID" \
  --title "QA: pr-happy-path (edited)" \
  --body  "Edited body."
bitbottle pr view "$PR_ID" | grep -F "QA: pr-happy-path (edited)"
```

`grep` exits `0`.

### 7. `pr request-review`

```bash
bitbottle pr request-review "$PR_ID" --reviewer "$BB_TEST_CLOUD_REVIEWER"
```

Exit code: `0`.

**Verify in UI:** the reviewer is listed on the PR.

### 8. `pr approve` (run as a different user, OR self-approve if your token
allows it; on Cloud self-approval is typically disabled — record the result)

```bash
bitbottle pr approve "$PR_ID"
```

Either exit `0` (approval recorded), or non-zero with stderr explaining
self-approval is not permitted. Either is acceptable for this test; the
critical bit is the error wording is clear.

### 9. `pr ready` promotes draft → open

```bash
bitbottle pr ready "$PR_ID"
bitbottle pr view "$PR_ID" | grep -i -E 'state|draft'
```

State line no longer says "draft".

### 10. `pr merge --squash --delete-branch`

```bash
bitbottle pr merge "$PR_ID" --squash --delete-branch
```

Exit code: `0`.

**Verify in UI:** PR shows "Merged"; the source branch `$FB` no longer
appears under Branches.

### 11. `pr view` on the merged PR still works

```bash
bitbottle pr view "$PR_ID" | grep -i merged
```

`grep` exits `0`.

### 12. `pr list --state merged` includes it

```bash
bitbottle pr list --state merged --limit 50 | grep -F "QA: pr-happy-path"
```

`grep` exits `0`.

## Cleanup

```bash
git checkout main && git branch -D "$FB" 2>/dev/null || true
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB" 2>/dev/null || true
# MANUAL_TEST.txt is now on main via the squash merge — remove via a follow-up
# commit if you want a clean repo:
echo "(optional) revert the squash commit on main if needed"
```
