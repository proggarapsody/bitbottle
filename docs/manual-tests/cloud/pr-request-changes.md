# Scenario: Cloud PR request-changes

**Backend:** Cloud only (`request-changes` is not supported on Server/DC).

Open a PR, request changes on it, verify the state, then clean up.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_REPO` exists with a `main` branch.
- Local clone:

  ```bash
  rm -rf /tmp/bb-pr-rc
  bitbottle repo clone "$BB_TEST_CLOUD_REPO" /tmp/bb-pr-rc
  cd /tmp/bb-pr-rc
  export FB="qa/pr-rc-$(date +%s)"
  ```

## Setup

```bash
git checkout -b "$FB" origin/main
echo "request-changes test $(date)" >> RC_TEST.txt
git add RC_TEST.txt
git commit -m "qa: pr-request-changes"
git push -u origin "$FB"
```

## Steps

### 1. Create the PR

```bash
bitbottle pr create --title "QA: pr-request-changes" --body "request changes on me" --base main
export PR_ID=$(bitbottle pr list --json id,title --limit 50 \
  | jq -r '.[] | select(.title=="QA: pr-request-changes") | .id' | head -1)
echo "PR_ID=$PR_ID"
```

`PR_ID` is a positive integer.

### 2. `pr view` shows the PR is open

```bash
bitbottle pr view "$PR_ID"
```

Stdout includes the title and a state line showing `open` (not `merged`,
`declined`, or `superseded`).

### 3. `pr request-changes`

```bash
bitbottle pr request-changes "$PR_ID"
```

Exit code: `0`. Stdout is:

```
Requested changes on pull request #<PR_ID>
```

**Verify in UI:** the PR on bitbucket.org shows your account in the reviewers
section with a "Changes requested" badge.

### 4. `pr request-changes` on a non-existent PR fails clearly

```bash
bitbottle pr request-changes 999999999
```

Exit code: non-zero. stderr includes an error message (e.g. "not found" or
HTTP 404).

### 5. Decline the PR (cleanup gate)

```bash
bitbottle pr decline "$PR_ID"
```

Exit code: `0`. This validates that the PR is still open and actionable
after request-changes (i.e. it was not auto-merged or auto-closed).

## Cleanup

```bash
bitbottle branch delete "$BB_TEST_CLOUD_REPO" "$FB" 2>/dev/null || true
```
