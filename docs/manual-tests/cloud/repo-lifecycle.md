# Scenario: Cloud repo lifecycle

**Backend:** Cloud.

Create a fresh repo, view it, clone it, then delete it.

## Prerequisites

- Logged in to `$BB_TEST_CLOUD_HOST`.
- `BB_TEST_CLOUD_WORKSPACE` set (the workspace under which to create the
  scratch repo) — defaults to the workspace half of `BB_TEST_CLOUD_REPO`.
- Token has `repository:write` and `repository:admin` scope.
- A unique scratch slug:
  ```bash
  export SCRATCH_SLUG="bb-qa-$(date +%s)"
  export SCRATCH_FQN="$BB_TEST_CLOUD_WORKSPACE/$SCRATCH_SLUG"
  ```

## Steps

### 1. `repo create`

```bash
bitbottle repo create "$SCRATCH_SLUG" \
  --project "$BB_TEST_CLOUD_WORKSPACE" \
  --description "bitbottle manual test" \
  --private=true
```

Exit code: `0`. stderr/stdout confirms creation and prints the web URL.

**Verify in UI:** repo appears at `https://bitbucket.org/$SCRATCH_FQN`.

### 2. `repo view` shows it

```bash
bitbottle repo view "$SCRATCH_FQN"
```

Stdout includes the slug, project/workspace, "bitbottle manual test"
description, and a web URL line. Exit code: `0`.

### 3. `repo view --web` opens the browser

```bash
bitbottle repo view "$SCRATCH_FQN" --web
```

Browser tab opens at the repo page. Exit code: `0`.

### 4. `repo list` includes the new repo

```bash
bitbottle repo list --limit 100 | grep -F "$SCRATCH_SLUG"
```

`grep` exits `0`; line shows `<slug>\t<workspace>\tgit`.

### 5. `repo clone` fetches it

```bash
cd /tmp
bitbottle repo clone "$SCRATCH_FQN"
ls "$SCRATCH_SLUG/.git" >/dev/null
```

Working tree exists at `/tmp/$SCRATCH_SLUG`. New empty repo (Bitbucket
shows the initial-commit landing page) — `git log` may exit `128` (no
commits yet) which is fine.

### 6. `repo delete` (with `--confirm`)

```bash
bitbottle repo delete "$SCRATCH_FQN" --confirm
```

Exit code: `0`.

**Verify in UI:** the repo page now 404s.

### 7. `repo delete` without `--confirm` on a non-TTY refuses

```bash
# Recreate so we can re-test
bitbottle repo create "$SCRATCH_SLUG" --project "$BB_TEST_CLOUD_WORKSPACE" --private=true
echo "" | bitbottle repo delete "$SCRATCH_FQN"
```

Exit code: non-zero. stderr says confirmation is required (use `--confirm`
or run interactively).

### 8. `repo view` on a deleted repo returns a clean 404

```bash
bitbottle repo delete "$SCRATCH_FQN" --confirm
bitbottle repo view "$SCRATCH_FQN"
```

Exit code: non-zero. stderr names the repo and says it was not found (no
stack trace, no panic).

## Cleanup

```bash
bitbottle repo delete "$SCRATCH_FQN" --confirm 2>/dev/null || true
rm -rf "/tmp/$SCRATCH_SLUG"
```
