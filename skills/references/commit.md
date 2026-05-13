# bitbottle commit comment — commit comment commands

## Command matrix

```bash
bitbottle commit comment list    PROJECT/REPO HASH [--reactions] [--json fields] [--jq expr]
bitbottle commit comment add     PROJECT/REPO HASH --body "text"
bitbottle commit comment edit    PROJECT/REPO HASH COMMENT_ID --body "text"
bitbottle commit comment delete  PROJECT/REPO HASH COMMENT_ID
bitbottle commit comment react   PROJECT/REPO HASH COMMENT_ID --emoji EMOJI   # Server/DC only
bitbottle commit comment unreact PROJECT/REPO HASH COMMENT_ID --emoji EMOJI   # Server/DC only
```

Both Bitbucket Cloud and Server / Data Center are supported for list/add/edit/delete.
`react`, `unreact`, and `--reactions` are Bitbucket Server / Data Center only.

## commit comment list

List all comments on a commit.

```bash
bitbottle commit comment list myproject/myrepo abc123
bitbottle commit comment list myproject/myrepo abc123 --json id,author,body
bitbottle commit comment list myproject/myrepo abc123 --jq '.[].body'
bitbottle commit comment list myproject/myrepo abc123 --reactions   # adds REACTIONS column (Server/DC only)
```

TTY output columns: `ID`, `AUTHOR`, `CREATED`, `BODY` (truncated to 70 chars
on TTY). Use `--json` for structured output — full body, `updatedAt`, etc.
Add `--reactions` to also fetch and display emoji reactions (Server/DC only).

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

## commit comment react / unreact _(Server / DC only)_

Add or remove an emoji reaction on a commit comment.

```bash
bitbottle commit comment react   myproject/myrepo abc123 1234 --emoji thumbs_up
bitbottle commit comment unreact myproject/myrepo abc123 1234 --emoji thumbs_up
```

`--emoji` is required. Accepted shortcodes: `thumbs_up`, `thumbs_down`, `heart`,
`laugh`, `hooray`, `confused`. Colon-wrapped aliases (`:thumbsup:`, `:heart:`) and
bare shortcodes are normalised automatically.

## MCP tools

| Tool | Required params | Notes |
|---|---|---|
| `list_commit_comments` | project, slug, hash | Returns JSON array; accepts `include_reactions: true` (Server/DC) |
| `add_commit_comment` | project, slug, hash, body | Returns `{id, author, body}` |
| `edit_commit_comment` | project, slug, hash, comment_id, body | Returns `{id}` |
| `delete_commit_comment` | project, slug, hash, comment_id | Returns `{deleted: true}` |
| `list_commit_comment_reactions` | project, slug, hash, comment_id | Server/DC only; returns grouped reaction array |
| `add_commit_comment_reaction` | project, slug, hash, comment_id, emoji | Server/DC only |
| `remove_commit_comment_reaction` | project, slug, hash, comment_id, emoji | Server/DC only |

All tools accept an optional `hostname` parameter (omit when only one host is
configured). `comment_id` is an integer.
