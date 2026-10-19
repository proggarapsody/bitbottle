# Public operation QA matrix

The matrix separates the complete runtime catalog from executed public test
scenarios. A registered command, a private handler test, and a passing suite
alone do not establish a verified operation.

## Run and inspect

```sh
make qa-matrix
make qa-critical-check
make qa-matrix-check
```

`qa-matrix` discovers the current Cobra tree and MCP SDK `tools/list`, then
runs Go tests once with JSON events, race detection, shuffle, no cache, and a
coverage profile. It removes ambient Bitbucket credentials and live opt-ins
and uses an isolated home. `QA_MATRIX_DIR` selects the evidence directory;
the default `.qa-matrix/` is ignored by Git.

Read `report.md` for progress and `report.json` for each operation/provider
gap. `catalog.json` preserves flags, input schemas, and declared backend
metadata. Application `--version` is probed separately through `app.Run` so
root-constructor inventory cannot silently omit it.

`qa-critical-check` verifies the saved full `./...` run against its source
fingerprint, contracts, raw test events, and hashes, then requires every pair
and named public test in the frozen
[`release-critical.json`](../../test/qamatrix/release-critical.json) profile.
The profile freezes the 59 reviewed operations, provider statuses, and important
inputs. Cloud `--auto` and `--auto-off` remain required release blockers until
their provider contracts and public behavior are verified. A missing, failed, skipped,
or stale critical proof fails this gate. Run `make qa-matrix` first to create
fresh saved evidence, or use the release workflow's single-run critical check.
`qa-matrix-check` applies the same evidence validation and additionally
requires completion of every catalog row. It remains the exhaustive QA
completion check.
Narrow package runs are useful during development but cannot satisfy the
full `./...` gate. Missing, skipped, stale, unsupported-without-proof, and
unknown cases remain incomplete.
Packages without tests are listed separately. The exact live opt-in
`acceptance/TestAcceptance` skip is reported as `not_run` in this hermetic lane;
it never proves Cloud or DC acceptance. Other skipped test cases block the run.

## Add a public scenario

1. Review the operation/provider applicability in
   [`test/qamatrix/contracts.json`](../../test/qamatrix/contracts.json).
   Supported operations need success and error scenarios. Unsupported
   providers need a reason and a public rejection scenario. Local-only
   operations need an explicit classification; omitted entries stay unknown.
   A reason-only `not_applicable` provider remains incomplete. Missing MCP
   backend metadata is reported separately; an explicitly reviewed provider
   contract supplies applicability evidence, while absence alone supplies none.
   For a local command, use `scope: local`, reviewed reasons for both remote
   providers, and a supported `local` pair. Only a complete local success/error
   proof (including important inputs) establishes remote non-applicability.
   Contradictory remote scenarios still block it. Reports count these derived
   classifications separately from supported and unsupported provider outcomes;
   each local command adds one local proof row to the two default remote rows.
2. Assert known output, request inputs or observable effects, error behavior,
   important flags/parameters, and pagination where applicable. Record only
   assertions actually made by the test, after those assertions.
3. Emit `qamatrix.Record(t, qamatrix.Scenario{...})` from the public test.
   The runner binds the claim to the exact Go package/test pass event; a
   declaration or log line without a passing test is not accepted evidence.
4. Run the matrix again. Review the test's assertions as well as its labels:
   labels describe evidence and cannot prove assertion strength themselves.

Important inputs must be exercised across the provider's passing scenarios;
every flag combination is unnecessary. A per-scenario non-applicability
comment cannot clear an operation's reviewed important input by itself.

CLI scenarios currently use the registered command tree and concrete
adapters against local HTTP fixtures. MCP scenarios use the public SDK and
a real stdio child with stateful synthetic backend clients. The latter
prove the MCP/domain contract, not provider HTTP compatibility or native
release artifact installation. Preserve those distinctions in QA claims.
For reads, an effect assertion identifies the observed backend request and
its independently expected routing/inputs; it does not imply mutated state.
For generated local output, the effect is the validated artifact (for example,
the selected shell's completion program), with settings preserved on errors.

## CI and release behavior

Ordinary PR CI publishes the complete catalog report and fails on test or
evidence errors; valid but incomplete coverage is reported as incomplete.
Release Please PRs and the publishing workflow additionally require
`qa-critical-check` before publication. Their critical pass does not mark the
remaining catalog rows complete. Keep the unresolved rows in the QA backlog
and use `qa-matrix-check` to verify eventual exhaustive completion. A green
ordinary PR or a critical-only pass therefore does not mean release-ready.

The full matrix remains work in progress. The current release profile and its
two Cloud auto-merge blockers are documented in
[`2026-10-19-critical-release-scenarios.md`](../quality/2026-10-19-critical-release-scenarios.md).
Final-SHA CI and Cloud live acceptance, native/npm/Docker candidate checks,
artifact digest/publication binding, and Release Please credentials remain
separate release requirements. Live acceptance must prove provisioning,
readback, and independently verified cleanup on the final SHA.
**DC не проверен live** must remain visible in QA reports and release notes;
the exception does not waive hermetic Server/DC failures.
