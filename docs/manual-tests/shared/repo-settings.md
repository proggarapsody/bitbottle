# Scenario: Repo settings CRUD (webhooks, branch protect/rule, deploy keys, default reviewers, reviewer groups)

**Backend:** Both. Each subsection notes Cloud / Server divergence.
Branch governance has two parallel surfaces: `branch protect` (Server-only)
and `branch-rule` (Cloud-only) — they are exercised in their respective
backend subsections only.

Covers **I** (webhooks), **BP** (branch protect, Server), **BRANCH-RULE**
(Cloud), **DEPLOY-KEY**, **DEFAULT-REVIEWERS**, and **PR-REVIEWER-GROUP**.
All settings persist on the real backend.

## Prerequisites

- Logged in to both backends.
- Scratch repos exist on both (`$BB_TEST_CLOUD_REPO`, `$BB_TEST_SERVER_REPO`).
- For default reviewers / reviewer groups: at least one **other** user
  on the same host whose slug/username is in `BB_TEST_CLOUD_REVIEWER` /
  `BB_TEST_SERVER_REVIEWER`. If unavailable, record and skip those steps.
- A throwaway SSH **public** key (or generate one for the test):
  `ssh-keygen -t ed25519 -N "" -f /tmp/bb-qa-key`. Its `.pub` body is the
  `--key` argument for `deploy-key add` and `ssh-key add` (Cloud
  workspace-admin scenario).

## Setup

```bash
# Cloud variant
export BB_HOST="$BB_TEST_CLOUD_HOST"; export BB_REPO="$BB_TEST_CLOUD_REPO"
export BB_REVIEWER="${BB_TEST_CLOUD_REVIEWER:-}"
# Server variant — re-run with these instead:
# export BB_HOST="$BB_TEST_SERVER_HOST"; export BB_REPO="$BB_TEST_SERVER_REPO"
# export BB_REVIEWER="${BB_TEST_SERVER_REVIEWER:-}"

[ -f /tmp/bb-qa-key.pub ] || ssh-keygen -t ed25519 -N "" -f /tmp/bb-qa-key
export QA_KEY="$(cat /tmp/bb-qa-key.pub)"
```

## Steps — Webhooks (both backends)

### 1. `webhook create`

```bash
bitbottle webhook create "$BB_REPO" \
  --url "https://example.test/qa-hook" \
  --events "repo:push,pullrequest:created" \
  --secret qa-secret \
  --active=true \
  --hostname "$BB_HOST"
```

Exit code: `0`. stderr/stdout prints the new webhook ID/UUID.

### 2. `webhook list` includes it

```bash
export WHID=$(bitbottle webhook list "$BB_REPO" --hostname "$BB_HOST" --json id,url \
  | jq -r '.[] | select(.url=="https://example.test/qa-hook") | .id' | head -1)
echo "WHID=$WHID"
```

`WHID` is a non-empty ID. The tabular `webhook list` includes the URL.

### 3. `webhook view` shows configuration

```bash
bitbottle webhook view "$BB_REPO" "$WHID" --hostname "$BB_HOST"
```

Stdout shows URL, events list, active flag, and a created-at timestamp.
The secret is **not** echoed — verify the output never contains
`qa-secret`. Exit code: `0`.

### 4. `webhook delete` removes it

```bash
bitbottle webhook delete "$BB_REPO" "$WHID" --hostname "$BB_HOST"
bitbottle webhook list "$BB_REPO" --hostname "$BB_HOST" \
  --json id | jq 'map(.id) | index("'"$WHID"'")'
```

Final `jq` prints `null`. Exit code: `0`.

## Steps — Deploy keys (both backends)

### 5. `deploy-key add`

```bash
bitbottle deploy-key add "$BB_REPO" \
  --key "$QA_KEY" --label "qa-manual" \
  --hostname "$BB_HOST"
```

Exit code: `0`. stderr prints the new key ID.

### 6. `deploy-key list`

```bash
export DK_ID=$(bitbottle deploy-key list "$BB_REPO" --hostname "$BB_HOST" \
  --json id,label | jq -r '.[] | select(.label=="qa-manual") | .id' | head -1)
echo "DK_ID=$DK_ID"
```

`DK_ID` is non-empty. Tabular output includes label and a fingerprint/comment.
The **private** key body is never echoed (we never sent it).

### 7. `deploy-key delete`

```bash
bitbottle deploy-key delete "$BB_REPO" "$DK_ID" --hostname "$BB_HOST"
bitbottle deploy-key list   "$BB_REPO" --hostname "$BB_HOST" \
  --json id | jq 'map(.id) | index("'"$DK_ID"'")'
```

Final `jq` prints `null`.

## Steps — Default reviewers (both backends)

Skip if `BB_REVIEWER` is empty.

### 8. `pr default-reviewer add`

```bash
bitbottle pr default-reviewer add "$BB_REPO" "$BB_REVIEWER" --hostname "$BB_HOST"
```

Exit code: `0`. (Cloud and Server use different API shapes but both
succeed if the user exists.)

### 9. `pr default-reviewer list` includes them

```bash
bitbottle pr default-reviewer list "$BB_REPO" --hostname "$BB_HOST" \
  | grep -F "$BB_REVIEWER"
```

`grep` exits `0`.

### 10. `pr default-reviewer remove`

```bash
bitbottle pr default-reviewer remove "$BB_REPO" "$BB_REVIEWER" --hostname "$BB_HOST"
bitbottle pr default-reviewer list   "$BB_REPO" --hostname "$BB_HOST" \
  | grep -F "$BB_REVIEWER" || echo "removed"
```

Final line prints `removed`.

## Steps — Reviewer groups (Server only — Cloud returns `host.unsupported`)

Skip on Cloud (the gated implementation returns
`ErrUnsupportedOnHost`; record the typed error message and continue).

### 11. `pr reviewer-group add` (Server)

```bash
if [ "$BB_HOST" = "$BB_TEST_SERVER_HOST" ] && [ -n "$BB_REVIEWER" ]; then
  bitbottle pr reviewer-group add "$BB_REPO" \
    --name qa-reviewers \
    --users "$BB_REVIEWER" \
    --hostname "$BB_HOST"
fi
```

Exit code: `0`.

### 12. `pr reviewer-group list`

```bash
bitbottle pr reviewer-group list "$BB_REPO" --hostname "$BB_HOST" | grep qa-reviewers
```

`grep` exits `0`.

### 13. `pr reviewer-group remove`

```bash
bitbottle pr reviewer-group remove "$BB_REPO" qa-reviewers --hostname "$BB_HOST"
bitbottle pr reviewer-group list   "$BB_REPO" --hostname "$BB_HOST" \
  | grep qa-reviewers || echo "removed"
```

Final line prints `removed`.

## Steps — Branch governance (Cloud: `branch-rule`)

Skip on Server.

### 14. `branch-rule add` (Cloud)

```bash
if [ "$BB_HOST" = "$BB_TEST_CLOUD_HOST" ]; then
  bitbottle branch-rule add "$BB_REPO" \
    --kind require_approvals_to_merge \
    --pattern "qa/*" \
    --hostname "$BB_HOST"
fi
```

Exit code: `0`. stderr prints the new rule ID.

### 15. `branch-rule list`

```bash
export RULE_ID=$(bitbottle branch-rule list "$BB_REPO" --hostname "$BB_HOST" \
  --json id,pattern | jq -r '.[] | select(.pattern=="qa/*") | .id' | head -1)
echo "RULE_ID=$RULE_ID"
```

`RULE_ID` is a non-empty integer (Cloud rule IDs).

### 16. `branch-rule delete`

```bash
bitbottle branch-rule delete "$BB_REPO" "$RULE_ID" --hostname "$BB_HOST"
```

Exit code: `0`. Re-listing no longer includes the rule.

## Steps — Branch governance (Server: `branch protect`)

Skip on Cloud — `branch protect` returns `host.unsupported`.

### 17. `branch protect create` (Server)

```bash
if [ "$BB_HOST" = "$BB_TEST_SERVER_HOST" ]; then
  bitbottle branch protect create "$BB_REPO" \
    --pattern "qa/*" --type fast-forward-only \
    --hostname "$BB_HOST"
fi
```

Exit code: `0`. stderr prints the new restriction ID.

(Use whichever `--type` your Server build supports —
`fast-forward-only`/`no-deletes`/`pull-request-only`. Record which works.)

### 18. `branch protect list`

```bash
export BP_ID=$(bitbottle branch protect list "$BB_REPO" --hostname "$BB_HOST" \
  --json id,matcher | jq -r '.[] | select(.matcher.displayId=="qa/*") | .id' | head -1)
echo "BP_ID=$BP_ID"
```

`BP_ID` is non-empty.

### 19. `branch protect delete`

```bash
bitbottle branch protect delete "$BB_REPO" "$BP_ID" --hostname "$BB_HOST"
```

Exit code: `0`. Re-listing no longer includes the restriction.

## Cleanup

```bash
# Most CRUD above is self-cleaning. Catch-all:
[ -n "${WHID:-}"     ] && bitbottle webhook delete    "$BB_REPO" "$WHID"     --hostname "$BB_HOST" 2>/dev/null || true
[ -n "${DK_ID:-}"    ] && bitbottle deploy-key delete "$BB_REPO" "$DK_ID"    --hostname "$BB_HOST" 2>/dev/null || true
[ -n "${RULE_ID:-}"  ] && bitbottle branch-rule delete "$BB_REPO" "$RULE_ID" --hostname "$BB_HOST" 2>/dev/null || true
[ -n "${BP_ID:-}"    ] && bitbottle branch protect delete "$BB_REPO" "$BP_ID" --hostname "$BB_HOST" 2>/dev/null || true
rm -f /tmp/bb-qa-key /tmp/bb-qa-key.pub
```

Re-run with the other backend's env vars.
