# bitbottle code-insights — Server/DC only

Code Insights attaches structured CI/scanner output to commits (reports,
annotations, optional merge gates). Cloud returns `host.unsupported`.

```bash
# Upsert a report on a commit
bitbottle code-insights report set PROJ/REPO HASH KEY \
  --title "Tool" --result PASS --report-type SECURITY

# List reports / report details
bitbottle code-insights report list PROJ/REPO HASH [--json] [--jq EXPR]
bitbottle code-insights report view PROJ/REPO HASH KEY

# Bulk-add annotations (JSON array with path/line/severity/type/message)
bitbottle code-insights annotation add PROJ/REPO HASH KEY \
  --from-json @findings.json

# Single annotation
bitbottle code-insights annotation add PROJ/REPO HASH KEY \
  --path src/main.go --line 42 --severity HIGH --type BUG \
  --message "null ptr"

# Merge check (experimental — API partly undocumented)
bitbottle code-insights merge-check set PROJ/REPO CHECK_KEY \
  --report-key REPORT_KEY --must-pass --min-severity MEDIUM
```

All subcommands accept `--hostname` and `--json` / `--jq` where applicable.
On merge-check errors, verify the `--report-key` matches an existing
report on the same repo — that's the most common failure mode.
