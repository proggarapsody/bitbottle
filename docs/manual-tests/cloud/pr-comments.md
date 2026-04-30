# Scenario: Cloud PR comments

**Backend:** Cloud. Hits `/2.0/repositories/{ws}/{repo}/pullrequests/{id}/comments`
directly (the Server/DC variant uses the activities feed instead — see
`server/pr-comments.md`).

## Prerequisites

- An open PR on `$BB_TEST_CLOUD_REPO`. Easiest: run setup from
  `cloud/pr-happy-path.md` up through step 2, then capture `$PR_ID` here.

## Steps

### 1. `pr comment list` on a PR with no general comments

```bash
bitbottle pr comment list "$PR_ID"
```

Either an empty stdout + exit `0`, or a header line followed by zero rows.
Record which.

### 2. `pr comment add` (first comment)

```bash
bitbottle pr comment add "$PR_ID" --body "manual-test comment 1"
```

Exit code: `0`. **Verify in UI:** the comment appears.

### 3. `pr comment add` (second comment, multi-line)

```bash
bitbottle pr comment add "$PR_ID" --body "$(printf 'line one\nline two')"
```

Exit code: `0`. **Verify in UI:** the comment renders with two lines.

### 4. `pr comment list` returns both, newest-first or oldest-first

```bash
bitbottle pr comment list "$PR_ID"
```

Stdout has at least 2 data rows (or aligned-table entries). Both bodies are
visible. Record the ordering.

### 5. `pr comment list --json`

```bash
bitbottle pr comment list "$PR_ID" --json id,author,text | jq 'length'
bitbottle pr comment list "$PR_ID" --json id,author,text | jq '.[0] | keys'
```

First command prints a number ≥ 2. Second command prints `["author","id","text"]`.

### 6. `pr comment list --jq`

```bash
bitbottle pr comment list "$PR_ID" --json text --jq '.[].text'
```

Stdout is one body per line. The strings `manual-test comment 1` and the
two-line variant both appear.

### 7. `pr comment add` without `--body` is rejected

```bash
bitbottle pr comment add "$PR_ID"
```

Exit code: non-zero. stderr names `--body` as required.

### 8. `pr comment add` against a non-existent PR fails clearly

```bash
bitbottle pr comment add 999999999 --body "should fail"
```

Exit code: non-zero. stderr names the PR as not found.

## Cleanup

The added comments stay on the PR (Bitbucket's API does not let
non-authors delete; even authors can only edit/delete via the UI).
Optionally delete them by hand from the Bitbucket UI.
