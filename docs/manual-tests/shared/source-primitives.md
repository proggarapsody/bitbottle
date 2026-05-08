# Scenario: source primitives (`repo file get`, `repo tree`)

**Backend:** both Cloud and Server / DC.

Read-only smoke for the RV1 source-reading commands. Run against
whichever backend you have configured locally — both are wired through
the same `SourceReader` interface and either is sufficient to catch
adapter regressions.

## Prerequisites

- One Bitbucket host configured: `bitbottle auth status` lists exactly
  one host. (If two are configured, pass `--hostname HOST` on every
  command below.)
- A repository with at least one tagged release and a non-trivial
  directory tree. The `bitbottle-qa` test repo from the PR happy-path
  scenarios is fine.
- `BB_TEST_REPO=PROJECT/repo` (Server) or `BB_TEST_REPO=workspace/repo`
  (Cloud) — whichever your configured host expects. `BB_TEST_REF=main`
  (or any branch / tag the repo carries).

## Setup

```bash
bitbottle --version
bitbottle auth status
```

## Steps

1. **Read a text file at the default branch — straight to stdout.**
   ```bash
   bitbottle repo file get "$BB_TEST_REPO" README.md --ref "$BB_TEST_REF"
   ```
   Expect: file content streams to stdout verbatim, exit 0. No envelope,
   no JSON wrapping, no trailing summary line.

2. **Read the same file via `--out` and diff against the inline read.**
   ```bash
   bitbottle repo file get "$BB_TEST_REPO" README.md --ref "$BB_TEST_REF" --out /tmp/readme.md
   bitbottle repo file get "$BB_TEST_REPO" README.md --ref "$BB_TEST_REF" | diff - /tmp/readme.md
   ```
   Expect: empty `diff` output, exit 0. (`--out` must round-trip
   byte-for-byte against the stdout path — this is the binary-safety
   guard.)

3. **Read a binary file — confirm bytes survive the round-trip.**
   Use any binary in the repo (logo PNG, compiled wasm, fixture
   tarball). If none exists, skip this step but leave a note.
   ```bash
   bitbottle repo file get "$BB_TEST_REPO" path/to/binary.png --ref "$BB_TEST_REF" --out /tmp/binary.png
   file /tmp/binary.png   # confirms it's a real binary, not a UTF-8 mangle
   ```
   Expect: `file` reports the original media type, not "ASCII text".

4. **Read a missing file.**
   ```bash
   bitbottle repo file get "$BB_TEST_REPO" definitely-not-a-real-file.go --ref "$BB_TEST_REF"
   echo "exit=$?"
   ```
   Expect: a typed `repo.not_found` error with a hint on stderr,
   non-zero exit. The error must NOT be a raw HTTP envelope — it goes
   through the `errfmt` catalogue.

5. **Read a directory path with `repo file get` — should refuse.**
   ```bash
   bitbottle repo file get "$BB_TEST_REPO" cmd --ref "$BB_TEST_REF"
   ```
   Expect: error mentioning "use `repo tree`" (the adapter detects the
   listing-vs-content case and surfaces a typed error), non-zero exit.

6. **List the repository root.**
   ```bash
   bitbottle repo tree "$BB_TEST_REPO" --ref "$BB_TEST_REF"
   ```
   Expect: TTY-aligned table with `TYPE`, `PATH`, `SIZE`, `HASH`
   columns. `TYPE` values are exactly `file` or `dir` (lowercase) —
   never `commit_file` (Cloud raw) or `FILE` (Server raw).

7. **List a nested directory.**
   ```bash
   bitbottle repo tree "$BB_TEST_REPO" cmd --ref "$BB_TEST_REF"
   ```
   Expect: every `PATH` value is full repo-relative
   (e.g. `cmd/main.go`), not just the basename.

8. **JSON output + jq filter.**
   ```bash
   bitbottle repo tree "$BB_TEST_REPO" --ref "$BB_TEST_REF" --json path,type,size --jq '.[]|select(.type=="dir").path'
   ```
   Expect: one path per line, only directory paths, no header, no JSON
   envelope.

9. **Read at a tag, then at a commit hash, to confirm `--ref` accepts
   non-branch refs.**
   ```bash
   bitbottle tag list "$BB_TEST_REPO" --limit 1 --json name --jq '.[0].name'
   # use the printed tag as --ref
   bitbottle repo tree "$BB_TEST_REPO" --ref <tag-name>

   bitbottle commit log "$BB_TEST_REPO" --limit 1 --json hash --jq '.[0].hash'
   bitbottle repo tree "$BB_TEST_REPO" --ref <hash>
   ```
   Expect: both calls succeed and return listings.

## Cleanup

Nothing to clean up — every command in this scenario is read-only.
