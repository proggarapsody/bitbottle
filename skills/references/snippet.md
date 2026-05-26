# snippet reference

Snippets are the Bitbucket Cloud analogue of GitHub Gists: shareable, optionally-private
collections of one or more named files. **Cloud only** — Server/DC returns
`host.unsupported`.

## Commands

### `snippet list`

```
bitbottle snippet list [--workspace SLUG] [--limit N] [--json] [--jq EXPR]
                        [--hostname HOST]
```

Lists snippets in the authenticated user's workspace (or the workspace
specified with `--workspace`).

| Flag | Default | Notes |
|---|---|---|
| `--workspace` | user slug from config | Bitbucket workspace slug |
| `--limit` | 30 | Max results (0 = no cap) |
| `--json` | — | Output full objects as JSON |
| `--jq EXPR` | — | Filter JSON output |

**Examples**

```bash
bitbottle snippet list
bitbottle snippet list --workspace myteam --limit 10
bitbottle snippet list --json | jq '.[].title'
```

---

### `snippet view SNIPPET_ID`

```
bitbottle snippet view SNIPPET_ID [--workspace SLUG] [--json] [--jq EXPR]
                                    [--hostname HOST]
```

Shows details for a single snippet, including its file list and creation date.

**Examples**

```bash
bitbottle snippet view Xqjyp1GV
bitbottle snippet view Xqjyp1GV --workspace myteam
bitbottle snippet view Xqjyp1GV --json
```

---

### `snippet create`

```
bitbottle snippet create --title TITLE [--file PATH ...] [--private]
                          [--workspace SLUG] [--json] [--jq EXPR]
                          [--hostname HOST]
```

Creates a new snippet. Pass `--file` once per file to include (the filename
is inferred from the path's base name). Omitting `--file` creates an empty
snippet.

| Flag | Notes |
|---|---|
| `--title` | Required. Snippet title |
| `--file PATH` | Repeatable. Local file to include in the snippet |
| `--private` | Make the snippet private (default: public) |

**Examples**

```bash
bitbottle snippet create --title "Hello world" --file hello.go
bitbottle snippet create --title "Multi-file" --file a.go --file b.go --private
bitbottle snippet create --title "Config" --file ~/.vimrc --workspace myteam
```

---

### `snippet delete SNIPPET_ID`

```
bitbottle snippet delete SNIPPET_ID [--confirm] [--workspace SLUG]
                                     [--hostname HOST]
```

Deletes a snippet. In non-interactive mode (piped output), `--confirm` is
required to prevent accidental deletion.

**Examples**

```bash
bitbottle snippet delete Xqjyp1GV --confirm
bitbottle snippet delete Xqjyp1GV --workspace myteam --confirm
```

---

### `snippet comment list SNIPPET_ID [WORKSPACE]`

```
bitbottle snippet comment list SNIPPET_ID [WORKSPACE] [--limit N] [--json] [--jq EXPR]
                                            [--hostname HOST]
```

Lists comments on a snippet. `WORKSPACE` defaults to the authenticated user's
workspace slug when omitted.

| Flag | Default | Notes |
|---|---|---|
| `--limit` | 50 | Max results (0 = no cap) |
| `--json` | — | Output full objects as JSON |

**Examples**

```bash
bitbottle snippet comment list Xqjyp1GV
bitbottle snippet comment list Xqjyp1GV myteam --limit 10
bitbottle snippet comment list Xqjyp1GV --json id,author,body
```

---

### `snippet comment add SNIPPET_ID [WORKSPACE] --body TEXT`

```
bitbottle snippet comment add SNIPPET_ID [WORKSPACE] --body TEXT
                               [--hostname HOST]
```

Posts a new comment on a snippet.

**Examples**

```bash
bitbottle snippet comment add Xqjyp1GV --body "Nice snippet!"
bitbottle snippet comment add Xqjyp1GV myteam --body "Thanks!"
```

---

### `snippet comment delete SNIPPET_ID COMMENT_ID [WORKSPACE]`

```
bitbottle snippet comment delete SNIPPET_ID COMMENT_ID [WORKSPACE] [--confirm]
                                  [--hostname HOST]
```

Deletes a comment from a snippet. `--confirm` is required in non-interactive mode.

**Examples**

```bash
bitbottle snippet comment delete Xqjyp1GV 17 --confirm
bitbottle snippet comment delete Xqjyp1GV 17 myteam --confirm
```

---

## MCP tools

| Tool | Description |
|---|---|
| `list_snippets` | List snippets in a workspace |
| `view_snippet` | Get a single snippet by ID |
| `create_snippet` | Create a new snippet |
| `delete_snippet` | Delete a snippet |
| `list_snippet_comments` | List comments on a snippet |
| `add_snippet_comment` | Post a comment on a snippet |
| `delete_snippet_comment` | Delete a comment from a snippet |

All tools require `workspace` and accept an optional `hostname` parameter.
`create_snippet` accepts `title`, `private` (bool), and `files` (JSON object
mapping filename to content, e.g. `{"hello.go": "package main"}`).
`list_snippet_comments` and `add_snippet_comment` / `delete_snippet_comment`
also require `snippet_id`. `add_snippet_comment` requires `body`.
`delete_snippet_comment` requires `comment_id` (number).
