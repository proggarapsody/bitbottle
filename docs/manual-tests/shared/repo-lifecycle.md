# Scenario: Repo lifecycle (create → clone → rename → visibility → fork → watcher → transfer → delete)

**Backend:** Both. Each step is annotated where Cloud and Server diverge
(e.g. `--project` only valid on Server). Run twice — once with Cloud env
vars, once with Server.

Exercises **Q** (rename/fork/set-default), **REPO-TRANSFER**,
**REPO-VISIBILITY**, **REPO-FORKS**, **REPO-WATCHER**, and the
create/clone/delete primitives. All of these persist on the real backend
and aren't covered by automation.

## Prerequisites

- Logged in to both backends.
- For Cloud: `BB_TEST_CLOUD_WORKSPACE` (the workspace slug — typically the
  prefix of `BB_TEST_CLOUD_REPO`). Used as `--project` on Cloud.
- For Server: `BB_TEST_SERVER_PROJECT` (the project key — typically the
  prefix of `BB_TEST_SERVER_REPO`).
- For transfer (step 9): a **second** target project on Server (`BB_TEST_SERVER_PROJECT_ALT`)
  or workspace on Cloud (`BB_TEST_CLOUD_WORKSPACE_ALT`). Skip step 9 if
  unavailable.

## Setup

```bash
# Cloud variant
export BB_HOST="$BB_TEST_CLOUD_HOST"
export BB_NS="$BB_TEST_CLOUD_WORKSPACE"
export BB_NS_ALT="${BB_TEST_CLOUD_WORKSPACE_ALT:-}"
# Server variant — re-run with these instead:
# export BB_HOST="$BB_TEST_SERVER_HOST"
# export BB_NS="$BB_TEST_SERVER_PROJECT"
# export BB_NS_ALT="${BB_TEST_SERVER_PROJECT_ALT:-}"

export NAME1="bitbottle-qa-$(date +%s)"
export NAME2="bitbottle-qa-renamed-$(date +%s)"
export REPO="$BB_NS/$NAME1"
```

## Steps

### 1. `repo create`

Cloud:

```bash
bitbottle repo create "$NAME1" \
  --project "$BB_NS" --description "qa: lifecycle test" \
  --hostname "$BB_HOST"
```

Server:

```bash
bitbottle repo create "$NAME1" \
  --project "$BB_NS" --description "qa: lifecycle test" \
  --hostname "$BB_HOST"
```

Exit code: `0`. stdout/stderr prints the new repo URL.

### 2. `repo list` includes the new repo

```bash
bitbottle repo list "$BB_NS" --hostname "$BB_HOST" | grep -F "$NAME1"
```

`grep` exits `0`.

### 3. `repo view` shows metadata

```bash
bitbottle repo view "$REPO" --hostname "$BB_HOST"
```

Stdout shows full slug, project, description ("qa: lifecycle test"),
default branch, visibility. Exit code: `0`.

### 4. `repo clone` into a temp dir

```bash
rm -rf /tmp/bb-repo-life
bitbottle repo clone "$REPO" /tmp/bb-repo-life --hostname "$BB_HOST"
cd /tmp/bb-repo-life
git log --oneline | head -5
```

`git log` runs (the repo is initialized with a default branch). Exit
code: `0`. Bootstrap an initial commit so subsequent steps have content:

```bash
echo "qa init" > README.md
git add README.md
git commit -m "qa: init"
git push -u origin main 2>/dev/null || git push -u origin master
```

### 5. `repo set-default` pins this repo for the cwd

```bash
bitbottle repo set-default "$REPO" --hostname "$BB_HOST"
git config bitbottle.host    # → $BB_HOST
git config bitbottle.project # → $BB_NS
git config bitbottle.slug    # → $NAME1
```

Subsequent commands run without `--hostname` and pick this repo up:

```bash
bitbottle repo view
```

Output matches the explicit `repo view "$REPO" --hostname "$BB_HOST"`.

### 6. `repo rename`

```bash
bitbottle repo rename "$REPO" "$NAME2" --confirm --hostname "$BB_HOST"
export REPO="$BB_NS/$NAME2"
bitbottle repo view "$REPO" --hostname "$BB_HOST" | grep -F "$NAME2"
```

Exit code: `0`. `grep` exits `0`. The local git remote may need
re-pointing:

```bash
git remote set-url origin "$(bitbottle repo view "$REPO" --hostname "$BB_HOST" --json clone_url 2>/dev/null | jq -r '.clone_url // empty' || \
  echo "https://$BB_HOST/$BB_NS/$NAME2.git")"
```

### 7. `repo visibility` — read

```bash
bitbottle repo visibility "$REPO" --hostname "$BB_HOST"
```

Stdout is `private` (default for newly-created repos) or `public`.
Exit code: `0`.

### 8. `repo visibility public` then back to `private`

```bash
bitbottle repo visibility "$REPO" public  --hostname "$BB_HOST"
bitbottle repo visibility "$REPO"          --hostname "$BB_HOST"  # → "public"
bitbottle repo visibility "$REPO" private --hostname "$BB_HOST"
bitbottle repo visibility "$REPO"          --hostname "$BB_HOST"  # → "private"
```

Each toggle exits `0`. Re-read confirms the new state.

### 9. `repo fork create` (Cloud only)

Cloud: forks are a first-class API.

```bash
if [ "$BB_HOST" = "$BB_TEST_CLOUD_HOST" ]; then
  bitbottle repo fork create "$REPO" --into "$BB_NS" \
    --name "${NAME2}-fork" --hostname "$BB_HOST"
  export FORK="$BB_NS/${NAME2}-fork"
fi
```

Exit code: `0`. The new repo appears in `repo list`.

Server: fork via API isn't supported by bitbottle; record the typed
`UnsupportedOnHost` error:

```bash
if [ "$BB_HOST" = "$BB_TEST_SERVER_HOST" ]; then
  bitbottle repo fork create "$REPO" --into "$BB_NS" --hostname "$BB_HOST"
fi
```

Exit code: non-zero. stderr mentions "unsupported on host".

### 10. `repo fork list` (both backends)

```bash
bitbottle repo fork list "$REPO" --hostname "$BB_HOST"
```

Cloud: the fork from step 9 appears. Server: empty list is acceptable
(forks may not be enumerable through the same API).

### 11. `repo watcher list`

```bash
bitbottle repo watcher list "$REPO" --hostname "$BB_HOST"
```

Stdout is a table of watchers (at least the creator). With `--json`:

```bash
bitbottle repo watcher list "$REPO" --hostname "$BB_HOST" --json username,uuid \
  | jq 'length'
```

Output is ≥ 1.

### 12. `repo transfer --to` (skip if no alt project/workspace)

```bash
if [ -n "$BB_NS_ALT" ]; then
  bitbottle repo transfer "$REPO" --to "$BB_NS_ALT" --hostname "$BB_HOST"
  bitbottle repo view "$BB_NS_ALT/$NAME2" --hostname "$BB_HOST"
  # Transfer back so cleanup works:
  bitbottle repo transfer "$BB_NS_ALT/$NAME2" --to "$BB_NS" --hostname "$BB_HOST"
fi
```

Exit code: `0` per step. The view after transfer shows the new project
and the same slug.

### 13. Bogus name on create gives a clear error

```bash
bitbottle repo create "this/has/slashes" --project "$BB_NS" --hostname "$BB_HOST"
```

Exit code: non-zero. stderr mentions invalid name — not a raw 400.

## Cleanup

```bash
# Delete the fork (Cloud).
if [ "$BB_HOST" = "$BB_TEST_CLOUD_HOST" ] && [ -n "${FORK:-}" ]; then
  bitbottle repo delete "$FORK" --confirm --hostname "$BB_HOST"
fi

# Delete the main scratch repo.
bitbottle repo delete "$REPO" --confirm --hostname "$BB_HOST"

# Verify gone.
bitbottle repo view "$REPO" --hostname "$BB_HOST" 2>&1 | grep -i "not found" >/dev/null && echo "deleted"

cd - >/dev/null
rm -rf /tmp/bb-repo-life
```

Re-run the scenario with the Server variant.
