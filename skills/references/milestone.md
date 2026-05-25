# Milestones reference

Milestones are a Bitbucket Cloud issue-tracker feature. **Cloud only** —
invocations against Server/DC return a typed `host.unsupported` error.

## Commands

```bash
bitbottle milestone list [WORKSPACE/REPO] [--limit N] [--json]
bitbottle milestone view [WORKSPACE/REPO] ID [--json]
```

`list` returns milestone ID and name for all milestones in the repository.
`view` returns a single milestone by its numeric ID.

Both commands support `--json`, `--jq expr`, `--template`, `--hostname`.

## MCP tools

- `list_milestones(project, slug, [limit, hostname])` — returns JSON array of milestones
- `view_milestone(project, slug, id, [hostname])` — returns a single milestone JSON object
