# Issue versions reference

Issue versions are a Bitbucket Cloud issue-tracker feature. **Cloud only** —
invocations against Server/DC return a typed `host.unsupported` error.

## Commands

```bash
bitbottle version list [WORKSPACE/REPO] [--limit N] [--json]
bitbottle version view ID [WORKSPACE/REPO] [--json]
bitbottle version create [WORKSPACE/REPO] --name NAME
bitbottle version delete ID [WORKSPACE/REPO] [--confirm]
```

`list` returns version ID and name for all versions in the repository.
`view` returns a single version by its numeric ID.
`create` creates a new version with the given name.
`delete` deletes a version; requires `--confirm` or TTY confirmation.

All read commands support `--json`, `--jq expr`, `--template`, `--hostname`.
`list` supports `--limit N` (default 30).

## MCP tools

- `list_issue_versions(workspace, slug, [limit, hostname])` — returns JSON array of versions
- `view_issue_version(workspace, slug, id, [hostname])` — returns a single version
- `create_issue_version(workspace, slug, name, [hostname])` — creates a version
- `delete_issue_version(workspace, slug, id, [hostname])` — deletes a version

## JSON output shape

```json
[{"id": 1, "name": "1.0"}, {"id": 2, "name": "2.0"}]
```
