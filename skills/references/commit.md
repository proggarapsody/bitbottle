# bitbottle commit comment — commit comment commands

## Command matrix

```bash
bitbottle commit comment list   PROJECT/REPO HASH [--json fields] [--jq expr]
bitbottle commit comment add    PROJECT/REPO HASH --body "text"
bitbottle commit comment edit   PROJECT/REPO HASH COMMENT_ID --body "text"
bitbottle commit comment delete PROJECT/REPO HASH COMMENT_ID
```

Both Bitbucket Cloud and Server / Data Center are supported for all four
operations.

## commit comment list

List all comments on a commit.

```bash
bitbottle commit comment list myproject/myrepo abc123
bitbottle commit comment list myproject/myrepo abc123 --json id,author,body
bitbottle commit comment list myproject/myrepo abc123 --jq '.[].body'
```

TTY output columns: `ID`, `AUTHOR`, `CREATED`, `BODY` (truncated to 70 chars
on TTY). Use `--json` for structured output — full body, `updatedAt`, etc.

## commit comment add

Add a comment to a commit.

```bash
bitbottle commit comment add myproject/myrepo abc123 --body "Looks good"
# → Created comment 1234
```

`--body` is required. The created comment ID is printed on success.

## commit comment edit

Edit the body of an existing commit comment. Requires the numeric comment ID
(from `commit comment list --json id`).

```bash
bitbottle commit comment edit myproject/myrepo abc123 1234 --body "Updated text"
```

On Bitbucket Server / Data Center the edit performs an optimistic-concurrency
GET + PUT (the current `version` is fetched before the write). A 409 Conflict
means the comment was edited by another user between the two calls — retry.

`--body` is required.

## commit comment delete

Delete an existing commit comment.

```bash
bitbottle commit comment delete myproject/myrepo abc123 1234
```

On Bitbucket Server / Data Center the delete performs a GET (to fetch the
current `version`) followed by a DELETE with `?version=N`. No output on
success.

## MCP tools

| Tool | Required params | Notes |
|---|---|---|
| `list_commit_comments` | project, slug, hash | Returns JSON array |
| `add_commit_comment` | project, slug, hash, body | Returns `{id, author, body}` |
| `edit_commit_comment` | project, slug, hash, comment_id, body | Returns `{id}` |
| `delete_commit_comment` | project, slug, hash, comment_id | Returns `{deleted: true}` |

All tools accept an optional `hostname` parameter (omit when only one host is
configured). `comment_id` is an integer.
