# Scenario: TTY vs pipe output, `--json`, `--jq`, `--web`

**Backend:** Both (run once on Cloud, once on Server/DC).

Verifies that list commands switch between aligned-table and tab-separated
output based on TTY detection, and that `--json` / `--jq` / `--web` work.

## Prerequisites

- Logged in to the target host (`bitbottle auth status` shows it).
- `BB_TEST_REPO` set to `BB_TEST_CLOUD_REPO` or `BB_TEST_SERVER_REPO`.
- Repo has ≥ 2 branches, ≥ 1 tag, ≥ 1 commit, ≥ 1 open PR. If not, run
  `cloud/branch-lifecycle.md`, `cloud/tag-lifecycle.md`, and
  `cloud/pr-happy-path.md` first (or the Server/DC equivalents).

## Steps

### 1. `repo list` — TTY shows aligned table with header

```bash
bitbottle repo list --limit 3
```

Expected stdout (column widths vary, header line is exact):

```
SLUG          PROJECT     TYPE
…             …           git
…             …           git
…             …           git
```

Exit code: `0`.

### 2. `repo list` — pipe drops the header and uses tabs

```bash
bitbottle repo list --limit 3 | cat -A | head -3
```

Each line ends in `$` (no trailing space) and fields are separated by `^I`
(tab). No `SLUG` header line. Exit code: `0`.

### 3. `pr list` — same TTY/pipe contract

```bash
bitbottle pr list "$BB_TEST_REPO"
bitbottle pr list "$BB_TEST_REPO" | head -1
```

TTY: header `TITLE … AUTHOR … STATE`. Pipe: no header, tab-separated.

### 4. `branch list` — TTY

```bash
bitbottle branch list "$BB_TEST_REPO" --limit 3
```

Header: `NAME    DEFAULT   HASH`. The default branch row has `true` in the
DEFAULT column.

### 5. `commit log` — pipe gives full hash + RFC3339 date

```bash
bitbottle commit log "$BB_TEST_REPO" --limit 1 | awk -F'\t' '{print $1, $4}'
```

First field is a 40-char hex hash. Last field parses as RFC3339
(`2026-01-02T15:04:05Z` or with offset).

### 6. `--json` returns JSON array

```bash
bitbottle pr list "$BB_TEST_REPO" --json id,title,state --limit 3
```

Stdout is valid JSON: an array of objects, each with exactly the keys `id`,
`title`, `state`. Pipe through `jq .` to confirm parse.

### 7. `--jq` filter

```bash
bitbottle pr list "$BB_TEST_REPO" --json id,state --jq '.[].id'
```

One numeric ID per line. No JSON brackets. Exit code: `0`.

### 8. `--web` opens the browser

```bash
bitbottle repo view "$BB_TEST_REPO" --web
```

Browser tab opens at the repo's web URL. CLI prints `Opening … in your
browser.` to stderr and exits `0`.

> If running in CI / headless, set `BROWSER=echo` to capture the URL on
> stdout instead of launching a browser; expect a single URL line.

### 9. Unknown `--json` field is rejected

```bash
bitbottle pr list "$BB_TEST_REPO" --json id,bogus
```

Exit code: non-zero. stderr names `bogus` and lists valid fields.

## Cleanup

None — read-only scenario.
