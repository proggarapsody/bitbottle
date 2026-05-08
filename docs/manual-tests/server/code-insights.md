# Server: Code Insights smoke test

End-to-end smoke for the `code-insights` command group against a real
Bitbucket Server / Data Center instance. Requires an authenticated session
(`bitbottle auth login --hostname <HOST>`) and a repository with at least
one commit.

## Prerequisites

```bash
export BB_HOST=git.example.com   # your BBS host
export BB_PROJ=MYPROJ
export BB_REPO=my-service
# Get a recent commit hash from the repo
export BB_HASH=$(git -C /path/to/clone rev-parse HEAD)
export REPORT_KEY=smoke-test-report
export CHECK_KEY=smoke-test-check
```

## 1 — Report: create (set)

```bash
bitbottle --hostname $BB_HOST code-insights report set \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY \
  --title "Smoke Test Report" \
  --result PASS \
  --report-type TESTING \
  --reporter "bitbottle-smoke" \
  --details "Manual smoke test run"
```

Expected: table row printed with key=`$REPORT_KEY`, result=`PASS`.

## 2 — Report: list

```bash
bitbottle --hostname $BB_HOST code-insights report list \
  $BB_PROJ/$BB_REPO $BB_HASH
```

Expected: table includes the row for `$REPORT_KEY`.

## 3 — Report: view

```bash
bitbottle --hostname $BB_HOST code-insights report view \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY
```

Expected: single-row table with correct title and result.

## 4 — Report: view as JSON

```bash
bitbottle --hostname $BB_HOST code-insights report view \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY \
  --json key,title,result,details
```

Expected: JSON object with all four fields.

## 5 — Annotation: add (single)

```bash
bitbottle --hostname $BB_HOST code-insights annotation add \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY \
  --path src/main.go \
  --line 42 \
  --severity HIGH \
  --type BUG \
  --message "Potential null dereference"
```

Expected: `Added 1 annotation(s) to report "smoke-test-report"`.

## 6 — Annotation: add (bulk from JSON)

Create a file `annotations.json`:

```json
[
  {"path": "src/util.go", "line": 10, "severity": "LOW", "type": "CODE_SMELL", "message": "unused variable"},
  {"path": "src/util.go", "line": 55, "severity": "MEDIUM", "type": "BUG", "message": "error not checked"}
]
```

```bash
bitbottle --hostname $BB_HOST code-insights annotation add \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY \
  --from-json @annotations.json
```

Expected: `Added 2 annotation(s) to report "smoke-test-report"`.

## 7 — Annotation: list

```bash
bitbottle --hostname $BB_HOST code-insights annotation list \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY
```

Expected: table with at least the annotations added above.

## 8 — Annotation: delete

```bash
bitbottle --hostname $BB_HOST code-insights annotation delete \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY
```

Expected: `Deleted annotations for report "smoke-test-report"`.

Verify list is now empty:

```bash
bitbottle --hostname $BB_HOST code-insights annotation list \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY --json path
```

Expected: empty JSON array `[]`.

## 9 — Merge check: set (EXPERIMENTAL)

```bash
bitbottle --hostname $BB_HOST code-insights merge-check set \
  $BB_PROJ/$BB_REPO $CHECK_KEY \
  --report-key $REPORT_KEY \
  --must-pass \
  --min-severity MEDIUM
```

Expected: `Set merge check "smoke-test-check"`.

## 10 — Merge check: get (EXPERIMENTAL)

```bash
bitbottle --hostname $BB_HOST code-insights merge-check get \
  $BB_PROJ/$BB_REPO $CHECK_KEY
```

Expected: JSON object with `key`, `report_key`, `must_pass=true`,
`min_severity="MEDIUM"`.

## 11 — Merge check: delete (EXPERIMENTAL)

```bash
bitbottle --hostname $BB_HOST code-insights merge-check delete \
  $BB_PROJ/$BB_REPO $CHECK_KEY
```

Expected: `Deleted merge check "smoke-test-check"`.

## 12 — Report: delete

```bash
bitbottle --hostname $BB_HOST code-insights report delete \
  $BB_PROJ/$BB_REPO $BB_HASH $REPORT_KEY
```

Expected: `Deleted report "smoke-test-report"`.

Verify gone:

```bash
bitbottle --hostname $BB_HOST code-insights report list \
  $BB_PROJ/$BB_REPO $BB_HASH
```

Expected: the row for `$REPORT_KEY` no longer appears.

## 13 — Cloud unsupported

If a Cloud host is also configured, verify the error envelope:

```bash
bitbottle --hostname cloud.bitbucket.org code-insights report list \
  myworkspace/myrepo abc123 2>&1
```

Expected: error message containing "not supported" / `host.unsupported`.
