# Scenario: Server/DC repo lifecycle

**Backend:** Server / Data Center. Operates against
`rest/api/1.0/projects/{projectKey}/repos`.

## Prerequisites

- Logged in to `$BB_TEST_SERVER_HOST`.
- `BB_TEST_SERVER_PROJECT` set (project key, e.g. `MYPROJ`) — defaults to
  the project half of `BB_TEST_SERVER_REPO`.
- Token has `PROJECT_ADMIN` (needed to create + delete repos in that project).

```bash
export SCRATCH_SLUG="bb-qa-$(date +%s)"
export SCRATCH_FQN="$BB_TEST_SERVER_PROJECT/$SCRATCH_SLUG"
```

## Steps

### 1. `repo create`

```bash
bitbottle repo create "$SCRATCH_SLUG" \
  --project "$BB_TEST_SERVER_PROJECT" \
  --description "bitbottle manual test" \
  --private=true
```

Exit code: `0`. **Verify in UI:** repo appears under the project.

### 2. `repo view`

```bash
bitbottle repo view "$SCRATCH_FQN"
```

Stdout shows slug, project key, description, and a web URL pointing to
`https://$BB_TEST_SERVER_HOST/projects/$BB_TEST_SERVER_PROJECT/repos/$SCRATCH_SLUG/...`.

### 3. `repo list` includes it

```bash
bitbottle repo list --limit 200 | grep -F "$SCRATCH_SLUG"
```

`grep` exits `0`.

### 4. `repo clone`

```bash
cd /tmp
bitbottle repo clone "$SCRATCH_FQN"
test -d "$SCRATCH_SLUG/.git"
```

Working tree exists. Exit code: `0`.

### 5. `repo delete --confirm`

```bash
bitbottle repo delete "$SCRATCH_FQN" --confirm
bitbottle repo view "$SCRATCH_FQN"
```

First command exits `0`. Second exits non-zero with a clear "not found"
error (no panic). On Server/DC, repo deletion can be asynchronous — if
`view` still finds it, retry after 10s.

### 6. `repo delete` without `--confirm` on a non-TTY refuses

```bash
bitbottle repo create "$SCRATCH_SLUG" --project "$BB_TEST_SERVER_PROJECT" --private=true
echo "" | bitbottle repo delete "$SCRATCH_FQN"
```

Exit code: non-zero. stderr says confirmation is required.

## Cleanup

```bash
bitbottle repo delete "$SCRATCH_FQN" --confirm 2>/dev/null || true
rm -rf "/tmp/$SCRATCH_SLUG"
```
