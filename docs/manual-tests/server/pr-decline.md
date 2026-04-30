# Scenario: Server/DC PR decline + unapprove

**Backend:** Server / Data Center.

Mirror of `cloud/pr-decline.md`. On Server/DC self-approve generally
works, so step 2 is expected to succeed.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- Local clone:
  ```bash
  rm -rf /tmp/bb-server-pr-decline
  bitbottle repo clone "$BB_TEST_SERVER_REPO" /tmp/bb-server-pr-decline
  cd /tmp/bb-server-pr-decline
  export DEFAULT_BRANCH=$(git symbolic-ref --short HEAD)
  export FB="qa/pr-decline-$(date +%s)"
  ```

## Setup

```bash
git checkout -b "$FB" "origin/$DEFAULT_BRANCH"
echo "decline $(date)" >> DECLINE_TEST.txt
git add DECLINE_TEST.txt
git commit -m "qa: pr-decline"
git push -u origin "$FB"
```

## Steps

### 1. Create the PR

```bash
bitbottle pr create --title "QA: pr-decline" --body "decline me" --base "$DEFAULT_BRANCH"
export PR_ID=$(bitbottle pr list --json id,title --limit 50 \
  | jq -r '.[] | select(.title=="QA: pr-decline") | .id' | head -1)
```

### 2. `pr approve` (self-approval works on Server/DC)

```bash
bitbottle pr approve "$PR_ID"
```

Exit code: `0`. UI shows approval mark.

### 3. `pr unapprove`

```bash
bitbottle pr unapprove "$PR_ID"
```

Exit code: `0`. Approval mark removed.

### 4. `pr decline`

```bash
bitbottle pr decline "$PR_ID"
```

Exit code: `0`. UI: PR is "Declined".

### 5. `pr view` reflects declined state

```bash
bitbottle pr view "$PR_ID" | grep -i declined
```

### 6. `pr list --state declined` (or `closed`) includes it

```bash
bitbottle pr list --state declined --limit 50 | grep -F "QA: pr-decline" \
  || bitbottle pr list --state closed --limit 50 | grep -F "QA: pr-decline"
```

At least one of the two `grep`s exits `0`. Record which state name the CLI
accepts on Server/DC.

### 7. Decline an already-declined PR fails clearly

```bash
bitbottle pr decline "$PR_ID"
```

Exit code: non-zero. stderr says PR is not in an open state.

## Cleanup

```bash
bitbottle branch delete "$BB_TEST_SERVER_REPO" "$FB" 2>/dev/null || true
```
