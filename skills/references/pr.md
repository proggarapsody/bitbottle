# bitbottle pr — pull request commands

## Command matrix

```bash
bitbottle pr list   [PROJ/repo] [--state open|closed|merged]   # default: open
bitbottle pr view   42 [--web]
bitbottle pr create --title "x" --base main [--body "x"] [--draft] [--head BRANCH]
bitbottle pr merge  42 [--merge|--squash] [--delete-branch]
bitbottle pr approve   42
bitbottle pr unapprove 42
bitbottle pr diff      42                       # unified diff; pipes to pager on TTY
bitbottle pr checkout  42
bitbottle pr edit      42 [--title "x"] [--body "x"]
bitbottle pr decline   42
bitbottle pr ready     42                       # draft → ready
bitbottle pr request-review  42 --reviewer alice [--reviewer bob]
bitbottle pr request-changes 42                 # Cloud only
bitbottle pr comment list 42 [--inline]              # --inline filters to file:line review comments
bitbottle pr comment add  42 --body "x"
```

The `pr comment list` output includes inline review comments (file:line
anchored) alongside general comments. Use `--inline` to filter to only
inline comments, or `--json inline,parentId,resolved,updatedAt` to read
thread structure and resolution state. On Server/DC `resolved` is always
`false` (resolution lives on tasks; out of scope until RV3).

## Flag reality check

- `pr list --state` accepts only `open`, `closed`, `merged` (no
  `all`). All `list` commands support `--limit N` (default 30) plus
  `--json`/`--jq`. `pr create` also has `--json`/`--jq`.
- No `--author`, `--mine`, or `--reviewer @me` filter — Bitbucket's
  REST API doesn't expose those.
- `pr create --head` defaults to the current local branch.
- `pr merge` requires exactly one of `--merge` or `--squash`.
  `--delete-branch` removes the source branch on the remote (Cloud
  only auto-deletes locally; Server/DC needs a separate `git push`).
- `pr request-changes` is Cloud-only — Server/DC has no API for it.

## Automation pattern

Whenever the agent is feeding PR data into another step, prefer JSON:

```bash
# All open PRs by IDs and titles, as a stream of objects
bitbottle pr list --json id,title --jq '.[]'

# A single PR's body for inspection
bitbottle pr view 42 --json id,title,body --jq '.body'

# Conditional: merge if mergeable
bitbottle pr view 42 --json mergeable --jq '.mergeable' \
  | grep -q true && bitbottle pr merge 42 --squash --delete-branch
```

Field discovery applies to every command, not just PR — see SKILL.md
safety rule 4 (pass a bogus `--json X` to list supported fields).

## Destructive ops

`pr merge` and `pr decline` follow the canonical destructive-op rule
in SKILL.md (safety rule 2). State the irreversible effect explicitly
("merges and deletes the source branch", "declines and cannot be
undone via the API") before asking for confirmation.

## Common failures

- *PR ID exists but `pr approve` errors* → you may already be the
  author. Bitbucket forbids self-approval. Use a different account
  or skip approval.
- *`pr merge` rejects with "not mergeable"* → unresolved comments,
  failing checks, or required reviewers. Run `pr view ID --web` to
  inspect.
- *`pr request-review` adds a reviewer but they can't see it* → on
  Server/DC the username is the slug, not the display name; on Cloud
  it's the workspace member nickname.
