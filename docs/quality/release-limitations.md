## QA limitations

**DC не проверен live**. Server/DC is covered by hermetic tests; this release has no live Data Center acceptance result.

Cloud auto-merge is deferred. Cloud `--auto`, `--auto-off`, and MCP `auto=true` return `pr.automerge.unavailable` with guidance instead of making an auto-merge request. Ordinary merge, squash, and rebase remain available.
