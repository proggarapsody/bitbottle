# bitbottle issue — Cloud-only issue tracker

Issues are gated behind the issue tracker being enabled on the repo. All
commands accept `[PROJECT/REPO]` as an optional first arg; if omitted the
repo is inferred from the current checkout.

```bash
# List / view
bitbottle issue list [PROJECT/REPO] [--state open|new|on-hold|…|all] [--limit N] [--json] [--jq EXPR]
bitbottle issue view [PROJECT/REPO] ID [--json] [--jq EXPR]

# Create / close / reopen
bitbottle issue create [PROJECT/REPO] --title "T" [--body "B"] [--kind bug|enhancement|proposal|task] [--priority trivial|minor|major|critical|blocker]
bitbottle issue close  [PROJECT/REPO] ID
bitbottle issue reopen [PROJECT/REPO] ID

# Edit (all flags optional; supply only what you want to change)
bitbottle issue edit [PROJECT/REPO] ID [--title "T"] [--body "B"] [--kind …] [--priority …] [--assignee USER] [--state …]

# Assign
bitbottle issue assign [PROJECT/REPO] ID USER

# Comments
bitbottle issue comment list   [PROJECT/REPO] ISSUE_ID [--json] [--jq EXPR]
bitbottle issue comment add    [PROJECT/REPO] ISSUE_ID --body "text"
bitbottle issue comment edit   [PROJECT/REPO] ISSUE_ID COMMENT_ID --body "new text"
bitbottle issue comment delete [PROJECT/REPO] ISSUE_ID COMMENT_ID
```

Valid states: `new`, `open`, `resolved`, `on hold`, `invalid`, `duplicate`, `wontfix`, `closed`.
Use `--state on-hold` on the CLI (the hyphen is normalized; the API uses a space).

MCP tools: `list_issues`, `get_issue`, `create_issue`, `close_issue`,
`update_issue`, `reopen_issue`, `assign_issue`, `list_issue_comments`,
`add_issue_comment`, `edit_issue_comment`, `delete_issue_comment`.

## Attachments
```bash
bitbottle issue attachment list   [PROJECT/REPO] ISSUE_ID [--json]
bitbottle issue attachment delete [PROJECT/REPO] ISSUE_ID FILENAME
```

## Vote / watch
```bash
bitbottle issue vote    [PROJECT/REPO] ISSUE_ID
bitbottle issue unvote  [PROJECT/REPO] ISSUE_ID
bitbottle issue watch   [PROJECT/REPO] ISSUE_ID
bitbottle issue unwatch [PROJECT/REPO] ISSUE_ID
```

MCP tools: `list_issue_attachments`, `delete_issue_attachment`, `vote_issue`,
`unvote_issue`, `watch_issue`, `unwatch_issue`.
