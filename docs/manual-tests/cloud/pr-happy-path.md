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

### 5.5. `pr commits` lists the PR's commits (PR-COMMITS)

```bash
bitbottle pr commits "$PR_ID"
```

Stdout is a table with columns `HASH`, `AUTHOR`, `DATE`, `MESSAGE`. The
single `qa: pr-happy-path` commit is present. With `--json`:

```bash
bitbottle pr commits "$PR_ID" --json hash,message | jq 'length'
```

Output is `1`.

### 5.6. `pr files` lists changed files (PR-FILES)

```bash
bitbottle pr files "$PR_ID"
```

Stdout is a table with `STATUS`, `PATH`, `+`, `-`. `MANUAL_TEST.txt`
appears with status `A` (added) or `M` (modified) and a non-zero `+`.
Exit code: `0`.

### 5.7. `pr participant list` (PR-PARTICIPANTS)

```bash
bitbottle pr participant list "$PR_ID"
```

Stdout enumerates the PR author + any current reviewers with their
`ROLE` (`PARTICIPANT`/`REVIEWER`) and `APPROVED` flag. At this point only
the author is present. Exit code: `0`.

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

### 7.1. `pr participant list` now includes the reviewer

```bash
bitbottle pr participant list "$PR_ID" | grep -F "$BB_TEST_CLOUD_REVIEWER"
```

`grep` exits `0`. With `--json`:

```bash
bitbottle pr participant list "$PR_ID" --json username,role,approved \
  | jq '.[] | select(.username=="'"$BB_TEST_CLOUD_REVIEWER"'") | .role'
```

Output is `"REVIEWER"` (or the Cloud equivalent — record exact value).

### 7.2. `pr default-reviewer list` (read-only smoke)

```bash
bitbottle pr default-reviewer list
```

Stdout lists the per-repo default reviewers (may be empty). Exit code:
`0`. The full add/remove cycle is covered in
[`shared/repo-settings.md`](../shared/repo-settings.md).

### 8. `pr approve` (run as a different user, OR self-approve if your token
allows it; on Cloud self-approval is typically disabled — record the result)

```bash
bitbottle pr approve "$PR_ID"
```

Either exit `0` (approval recorded), or non-zero with stderr explaining
self-approval is not permitted. Either is acceptable for this test; the
critical bit is the error wording is clear.

### 8.5. `pr comment list` exposes the new inline-aware surface

```bash
# Post a top-level comment so there's something to list.
bitbottle pr comment add "$PR_ID" --body "QA: testing pr comment list"

# Default list — the new LOCATION column appears only when an inline
# comment is present, so on a fresh PR with no inline reviews the
# columns remain ID/AUTHOR/CREATED/TEXT.
bitbottle pr comment list "$PR_ID"

# --inline filter on a PR with no inline comments yields an empty list.
bitbottle pr comment list "$PR_ID" --inline

# Verify the new JSON fields parse end-to-end.
bitbottle pr comment list "$PR_ID" \
  --json id,parentId,resolved,updatedAt,inline | jq 'length >= 1'
```

Each command exits `0`. The last `jq` prints `true`.

### 8.6. `pr comment add --inline` posts an anchored review comment

```bash
bitbottle pr comment add "$PR_ID" \
  --inline MANUAL_TEST.txt:1 \
  --body "QA: inline new-side"
INLINE_ID=$(bitbottle pr comment list "$PR_ID" --inline --json id,inline,text \
  | jq '.[] | select(.text=="QA: inline new-side") | .id')

# Multi-line range (Cloud only)
bitbottle pr comment add "$PR_ID" \
  --inline MANUAL_TEST.txt:1-3 \
  --body "QA: inline range"

# Reply nested under the first inline comment
bitbottle pr comment add "$PR_ID" --parent "$INLINE_ID" --body "QA: reply"
```

Exit `0` on each. `pr comment list "$PR_ID" --inline` now shows the new
inline comment with the LOCATION column populated as
`MANUAL_TEST.txt:1`. Verify in the Bitbucket UI that the comment appears
anchored at the right line, and the reply nests under it.

### 8.7. `pr comment edit / delete / resolve`

```bash
bitbottle pr comment edit    "$PR_ID" "$INLINE_ID" --body "QA: edited"
bitbottle pr comment list    "$PR_ID" --json id,text | jq '.[] | select(.id=='"$INLINE_ID"').text'
bitbottle pr comment resolve "$PR_ID" "$INLINE_ID"
bitbottle pr comment list    "$PR_ID" --json id,resolved | jq '.[] | select(.id=='"$INLINE_ID"').resolved'
bitbottle pr comment delete  "$PR_ID" "$INLINE_ID"
```

The `text` jq prints `"QA: edited"`. The `resolved` jq prints `true`.
The `delete` exits `0` and the comment disappears from subsequent
`pr comment list` output.

### 8.8. `pr checks` lists build status (GHP)

```bash
bitbottle pr checks "$PR_ID"
```

Stdout is a table of build/check rows for the PR's source commit, with
columns `CONTEXT`, `STATE`, `URL`. If the repo has no CI configured, the
output reports "no checks found" and exits `0`. With `--watch`, the
command blocks until every check reaches a terminal state:

```bash
bitbottle pr checks "$PR_ID" --watch
```

Press Ctrl-C if no CI is configured; otherwise the command exits `0` once
all checks land on `SUCCESSFUL`/`FAILED`/`STOPPED`.

### 8.9. `pr update-branch` rebases/merges main into the PR branch (GHP)

```bash
bitbottle pr update-branch "$PR_ID"
```

If `main` has diverged since the PR was created, the source branch is
updated (Cloud: merges `main` in via the API). If not, stderr reports
"already up to date" and exits `0`. Either is acceptable.

**Verify in UI:** the PR's commit list either gains a merge commit from
`main` or is unchanged.

### 8.10. `pr status` shows the active PR for the current branch (GHP)

Still inside `/tmp/bb-pr-happy` on branch `$FB`:

```bash
bitbottle pr status
```

Stdout shows the current branch, its active PR (`$PR_ID` with title),
review state, and merge readiness. Exit code: `0`.

### 8.11. Root `status` summarises the repo at-a-glance (GHP)

```bash
bitbottle status
```

Stdout shows the current host/repo/branch, the active PR for the branch
(if any), open PRs you're reviewing, and recent pipeline runs. Exit
code: `0`. This is the cross-cutting "what's happening here" view.

### 8.12. Root `browse` opens the repo in the UI (GHP)

```bash
bitbottle browse
```

A browser tab opens to `https://$BB_TEST_CLOUD_HOST/<workspace>/<repo>`.
Exit code: `0`. Try a target shortcut:

```bash
bitbottle browse "" pulls
```

The browser navigates to the repo's PR list. (Empty `[PROJECT/REPO]` means
"current repo".)

### 9. `pr ready` promotes draft → open

```bash
bitbottle pr ready "$PR_ID"
bitbottle pr view "$PR_ID" | grep -i -E 'state|draft'
```

State line no longer says "draft".

### 9.5. `pr merge --auto` queues auto-merge (AUTOMERGE — Cloud beta)

Auto-merge queues the PR to merge once all required checks pass. The Cloud
implementation is beta; record any divergence from the expected wording.

```bash
bitbottle pr merge "$PR_ID" --auto --squash
```

Exit code: `0`. stdout reports the PR is queued for auto-merge with the
chosen strategy.

### 9.6. `pr merge --auto-off` cancels the queued auto-merge

```bash
bitbottle pr merge "$PR_ID" --auto-off
```

Exit code: `0`. stdout reports the auto-merge was cancelled. The PR is
not merged.

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
