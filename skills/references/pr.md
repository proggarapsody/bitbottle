# bitbottle pr — pull request commands

## Command matrix

```bash
bitbottle pr list   [PROJ/repo] [--state open|closed|merged]   # default: open
bitbottle pr view   42 [--web]
bitbottle pr create --title "x" --base main [--body "x"] [--draft] [--head BRANCH]
bitbottle pr merge  42 [--merge|--squash] [--delete-branch]
bitbottle pr merge  42 --auto [--squash|--rebase]          # queue for auto-merge when checks pass
bitbottle pr merge  42 --auto-off                          # cancel a queued auto-merge
bitbottle pr approve   42
bitbottle pr unapprove 42
bitbottle pr diff      42                       # unified diff; pipes to pager on TTY
bitbottle pr checkout  42
bitbottle pr edit      42 [--title "x"] [--body "x"]
bitbottle pr decline   42
bitbottle pr ready     42                       # draft → ready
bitbottle pr request-review  42 --reviewer alice [--reviewer bob]
bitbottle pr request-changes 42                 # Cloud only
bitbottle pr review 42 --approve                     # approve + optional body/inline comments
bitbottle pr review 42 --request-changes             # Cloud only; --body "see comments"
bitbottle pr review 42 --comment --body "x" --inline pkg/foo.go:42:nit    # PATH:LINE:BODY
bitbottle pr comment list 42 [--inline]              # --inline filters to file:line review comments
bitbottle pr comment add  42 --body "x"
bitbottle pr comment add  42 --body "nit" --inline pkg/foo.go:88 [--side new|old]
bitbottle pr comment add  42 --body "agreed" --parent 1234     # reply to thread
bitbottle pr comment edit 42 1234 --body "..."
bitbottle pr comment delete 42 1234
bitbottle pr comment resolve 42 1234                          # Cloud only
bitbottle pr comment react   42 1234 --emoji thumbs_up        # Server/DC only; add emoji reaction
bitbottle pr comment unreact 42 1234 --emoji thumbs_up        # Server/DC only; remove emoji reaction
bitbottle pr comment list 42 --reactions                      # Server/DC only; fetch & show reactions column
bitbottle pr task list   42 [--state open|resolved|all]      # Server/DC only; BLOCKER comments
bitbottle pr task create 42 --body "TEXT" [--parent COMMENT_ID]
bitbottle pr task resolve 42 TASK_ID
bitbottle pr task reopen  42 TASK_ID
bitbottle pr activity 42                                      # PR event stream
bitbottle pr activity 42 --limit 20
bitbottle pr activity 42 --json type,actor,createdAt,detail   # structured output
bitbottle pr checks    42                                      # list CI statuses for the PR head commit
bitbottle pr checks    42 --watch [--interval 10]             # poll until all checks settle
bitbottle pr commits   42                                      # list commits in a pull request
bitbottle pr commits   42 --json hash,message,author,date     # structured output
bitbottle pr files        42                                         # list files changed in a pull request
bitbottle pr files        42 --json status,path,additions,deletions  # structured output
bitbottle pr participants 42                                         # list participants (author, reviewers, observers)
bitbottle pr participants 42 --json role,display_name,username,approved # structured output
bitbottle pr update-branch 42                                  # rebase/sync PR branch onto target (Cloud merge commit; Server rebase)
bitbottle pr status [PROJ/repo]                               # show your open PRs split by role (AUTHOR / REVIEWER)
bitbottle pr reopen 42                                        # reopen a declined/closed PR
bitbottle pr default-reviewer list [PROJ/repo]               # list configured default reviewers
bitbottle pr default-reviewer add  [PROJ/repo] USER          # add a default reviewer (slug on Server, account ID/nickname on Cloud)
bitbottle pr default-reviewer remove [PROJ/repo] USER        # remove a default reviewer
```

The `pr comment list` output includes inline review comments (file:line
anchored) alongside general comments. Use `--inline` to filter to only
inline comments, or `--json inline,parentId,resolved,updatedAt` to read
thread structure and resolution state. On Server/DC `resolved` is always
`false` (resolution lives on tasks; out of scope).

`pr comment add --inline` accepts `path:line` or `path:start-end` (multi-
line ranges are Cloud-only — Server/DC anchors are single-line and the
command rejects ranges with a typed error). `--side` defaults to `new`;
pass `--side old` to comment on the removed/old side of the diff.

`pr comment resolve` is Cloud-only — Server/DC returns a typed
`host.unsupported` because resolution lives on tasks, not regular
comments.

`pr comment react` / `pr comment unreact` are Server/DC-only. Supported
emoji shortcodes: `thumbs_up`, `thumbs_down`, `heart`, `laugh`, `hooray`,
`confused`. Colon-wrapped forms (`:thumbsup:`, `:heart:`) are automatically
normalised. `pr comment list --reactions` fetches each comment's reactions
concurrently (worker pool of 4) and renders them as emoji glyphs in the
REACTIONS column (`👍×2 ❤️×1`). MCP tools: `list_comment_reactions`,
`add_comment_reaction`, `remove_comment_reaction`; `list_pr_comments` also
accepts `include_reactions: true`.

`pr task` is Server/DC-only — Bitbucket Cloud has no task (BLOCKER comment)
primitive. `pr task list` returns BLOCKER-severity comments and accepts
`--state open` (default), `resolved`, or `all`. `pr task create` creates a
BLOCKER comment (optionally threaded under `--parent`). `pr task resolve` and
`pr task reopen` toggle the comment state between RESOLVED and OPEN.
Cloud returns a typed `host.unsupported` error for resolve/reopen. MCP tools:
`list_pr_tasks`, `create_pr_task`, `resolve_pr_task`, `reopen_pr_task`.

`pr activity` streams all PR events (approvals, unapprovals, comments,
updates, merges, declines, rescopes) from both backends. The TTY table
shows TIME (relative), TYPE, ACTOR. Use `--json type,actor,createdAt,detail`
for structured output; `detail` carries the raw backend sub-object.
`--limit N` caps results (default: no limit). MCP tool: `get_pr_activity`.

`pr commits` lists all commits included in a pull request, paginated from both
backends. Output table: HASH (first 8 chars on TTY) | MESSAGE | AUTHOR | DATE.
Supports `--json hash,message,author,date`. MCP tool: `list_pr_commits`.

`pr files` lists all files changed in a pull request, paginated from both
backends. Cloud uses `GET /repositories/{ws}/{slug}/pullrequests/{id}/diffstat`;
Server/DC uses `GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{id}/changes`.
Output table: STATUS | PATH | +ADDITIONS | -DELETIONS. Supports
`--json status,path,additions,deletions`. MCP tool: `list_pr_files`.

`pr participants` lists all users involved in a pull request (author, reviewers,
observers) from both backends. Cloud uses `GET /repositories/{ws}/{slug}/pullrequests/{id}/participants`;
Server/DC uses `GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{id}/participants`.
Output table: ROLE | DISPLAY_NAME | USERNAME | APPROVED. Cloud `state` values
(`approved`, `changes_requested`) are normalised to uppercase; Server `status`
`NEEDS_WORK` maps to `CHANGES_REQUESTED`, `UNAPPROVED` maps to `""`. Supports
`--json role,display_name,username,approved`. MCP tool: `list_pr_participants`.

## Flag reality check

- `pr list --state` accepts only `open`, `closed`, `merged` (no
  `all`). All `list` commands support `--limit N` (default 30) plus
  `--json`/`--jq`. `pr create` also has `--json`/`--jq`.
- No `--author`, `--mine`, or `--reviewer @me` filter — Bitbucket's
  REST API doesn't expose those.
- `pr create --head` defaults to the current local branch.
- `pr merge` supports `--merge`, `--squash`, and (with `--auto`) `--rebase`.
  `--delete-branch` removes the source branch on the remote after an
  immediate merge (Cloud only auto-deletes locally; Server/DC needs a
  separate `git push`).
- `pr merge --auto` queues the PR for auto-merge; strategy defaults to
  `merge`. Pass `--squash` or `--rebase` to override. On Bitbucket Cloud
  the feature is currently in beta — if the workspace hasn't opted in you
  get a `pr.automerge.beta_disabled` error with a hint to ask the admin.
- `pr merge --auto-off` cancels a queued auto-merge without merging.
- Running `pr merge` (immediate) when the PR already has auto-merge queued
  prompts "PR is queued for auto-merge. Cancel and merge now? [y/N]" on a
  TTY; on non-TTY the auto-merge is silently cancelled first.
- `pr view` shows `Auto-merge: enabled (strategy)` when auto-merge is queued.
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

## PR default reviewers

`pr default-reviewer` manages the per-repository list of users automatically
added as reviewers when a PR is created. Both Cloud and Server/DC are
supported.

```bash
bitbottle pr default-reviewer list MYPROJ/my-service
bitbottle pr default-reviewer add  MYPROJ/my-service jsmith
bitbottle pr default-reviewer remove MYPROJ/my-service jsmith
```

- **Server/DC**: USER is the user slug (e.g. `jsmith`).
- **Cloud**: USER is the account ID (UUID) or nickname.

MCP tools: `list_default_reviewers`, `add_default_reviewer`, `remove_default_reviewer`.
Each accepts `repo` (required) and `hostname` (optional).

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
