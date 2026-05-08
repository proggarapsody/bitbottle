# Scenario: Issue lifecycle (Cloud only)

**Backend:** Bitbucket Cloud (`bitbucket.org`)

Covers `issue list`, `issue view`, `issue create`, `issue close`,
`issue edit`, `issue reopen`, `issue assign`, and `issue comment {list|add|edit|delete}`.

---

## Prerequisites

- `BB_TEST_CLOUD_REPO` set to `<workspace>/<repo>` with issue tracker enabled.
- `BB_TEST_CLOUD_TOKEN` set (scopes: `account` read, `issue` read+write).
- `BB_TEST_USER` set to the authenticated user's Bitbucket username (for assign tests).

```bash
export R=$BB_TEST_CLOUD_REPO
export H=bitbucket.org
```

## Setup

```bash
make build
export PATH="$PWD/dist:$PATH"
bitbottle auth login --hostname $H --with-token <<< "$BB_TEST_CLOUD_TOKEN"
```

---

## Steps

### 1 — Create an issue

```bash
bitbottle issue create $R \
  --title "Manual test issue $(date +%s)" \
  --body "Created by issue-lifecycle smoke test." \
  --kind task \
  --priority minor
```

Expected: prints `Created issue #<N>: Manual test issue …`

Set `ISSUE_ID` to the printed `<N>`.

---

### 2 — View the issue

```bash
bitbottle issue view $R $ISSUE_ID
```

Expected: shows `ISSUE_ID`, title, `state=new`, `kind=task`, `priority=minor`.

```bash
bitbottle issue view $R $ISSUE_ID --json title,state,kind,priority
```

Expected: JSON with `"kind":"task"` and `"priority":"minor"`.

---

### 3 — List issues

```bash
bitbottle issue list $R --state new --limit 5
```

Expected: row for `$ISSUE_ID` visible.

---

### 4 — Edit the issue

```bash
bitbottle issue edit $R $ISSUE_ID \
  --title "Edited smoke issue" \
  --priority major \
  --state open
```

Expected: `Updated issue #<N>: Edited smoke issue`

Verify: `bitbottle issue view $R $ISSUE_ID --json title,priority,state`
→ `"priority":"major"`, `"state":"open"`.

---

### 5 — Assign the issue

```bash
bitbottle issue assign $R $ISSUE_ID $BB_TEST_USER
```

Expected: `Assigned issue #<N> to <user>`

Verify: `bitbottle issue view $R $ISSUE_ID --json assignee` → `{"assignee":"<user>"}`.

---

### 6 — Add a comment

```bash
bitbottle issue comment add $R $ISSUE_ID --body "Smoke test comment"
```

Expected: `Added comment #<M> to issue #<N>`.

Set `COMMENT_ID` to `<M>`.

---

### 7 — List comments

```bash
bitbottle issue comment list $R $ISSUE_ID
```

Expected: row for `$COMMENT_ID` with content `Smoke test comment`.

```bash
bitbottle issue comment list $R $ISSUE_ID --json id,author,content
```

Expected: JSON array with one entry, `"content":"Smoke test comment"`.

---

### 8 — Edit the comment

```bash
bitbottle issue comment edit $R $ISSUE_ID $COMMENT_ID --body "Updated smoke comment"
```

Expected: `Updated comment #<M> on issue #<N>`.

Verify: `bitbottle issue comment list $R $ISSUE_ID --json id,content`
→ comment `$COMMENT_ID` has `"content":"Updated smoke comment"`.

---

### 9 — Close the issue

```bash
bitbottle issue close $R $ISSUE_ID
```

Expected: `Closed issue #<N>`.

Verify: `bitbottle issue view $R $ISSUE_ID --json state` → `"state":"closed"`.

---

### 10 — Reopen the issue

```bash
bitbottle issue reopen $R $ISSUE_ID
```

Expected: `Reopened issue #<N>`.

Verify: `bitbottle issue view $R $ISSUE_ID --json state` → `"state":"open"`.

---

### 11 — Delete the comment

```bash
bitbottle issue comment delete $R $ISSUE_ID $COMMENT_ID
```

Expected: `Deleted comment #<M> from issue #<N>`.

Verify: `bitbottle issue comment list $R $ISSUE_ID` → empty or no row for `$COMMENT_ID`.

---

### 12 — Server backend returns host.unsupported

If you also have a Server/DC host configured:

```bash
bitbottle issue list --hostname <server-host> PROJ/repo
```

Expected exit code 1, error contains `not supported` or `host.unsupported`.

---

## Cleanup

```bash
bitbottle issue close $R $ISSUE_ID
```

(Issues cannot be deleted via the API; closing is sufficient to remove them
from active views.)
