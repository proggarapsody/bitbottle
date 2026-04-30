# Scenario: Cloud PR decline + unapprove

**Backend:** Cloud.

Open a PR, approve it, unapprove it, then decline it.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- Local clone of `$BB_TEST_CLOUD_REPO` at `/tmp/bb-pr-decline`.
- Token user can approve PRs in this workspace (otherwise skip step 3 and
  record).

## Setup

```bash
rm -rf /tmp/bb-pr-decline
bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-pr-decline
cd /tmp/bb-pr-decline
export FB="qa/pr-decline-$(date +%s)"
git checkout -b "$FB" origin/main
echo "decline test $(date)" >> DECLINE_TEST.txt
git add DECLINE_TEST.txt
git commit -m "qa: pr-decline"
git push -u origin "$FB"
```

## Steps

### 1. Create the PR

```bash
bitbottle pr create --title "QA: pr-decline" --body "decline me" --base main
export PR_ID=$(bitbottle pr list --json id,title --limit 50 \
  | jq -r '.[] | select(.title=="QA: pr-decline") | .id' | head -1)
echo "PR_ID=$PR_ID"
```

`PR_ID` is a positive integer.

### 2. `pr approve`

```bash
bitbottle pr approve "$PR_ID"
```

Exit code: `0` (or non-zero with clear "self-approval" wording on Cloud —
record).

### 3. `pr unapprove`

```bash
bitbottle pr unapprove "$PR_ID"
```

Exit code: `0`. **Verify in UI:** approval mark is removed.

### 4. `pr decline`

```bash
bitbottle pr decline "$PR_ID"
```

Exit code: `0`. **Verify in UI:** PR state is "Declined".

### 5. `pr view` shows declined state

```bash
bitbottle pr view "$PR_ID" | grep -i -E 'declined|closed'
```

`grep` exits `0`.

### 6. `pr list --state closed` includes it

```bash
bitbottle pr list --state closed --limit 50 | grep -F "QA: pr-decline"
```

`grep` exits `0`. (Bitbucket bundles "declined" under closed-states; if the
exact label differs, record.)

### 7. Decline an already-declined PR fails clearly

```bash
bitbottle pr decline "$PR_ID"
```

Exit code: non-zero. stderr explains the PR is not in an open state.

## Cleanup

```bash
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB" 2>/dev/null || true
```
