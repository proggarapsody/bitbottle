# Public operation QA matrix

The matrix separates the complete runtime catalog from executed public test
scenarios. A registered command, a private handler test, and a passing suite
alone do not establish a verified operation.

## Run and inspect

```sh
make qa-matrix
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

`qa-matrix-check` verifies the saved run against its source fingerprint,
contracts, raw test events, and hashes before requiring complete evidence.
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

Ordinary PR CI publishes the matrix and fails on test or evidence errors;
valid but incomplete coverage is reported as incomplete. Release Please
PRs additionally run the complete-matrix check. The release workflow runs
the same strict gate before registry login or publication. A green ordinary
PR therefore does not mean release-ready.

The full matrix remains work in progress. This gate intentionally blocks
publication while operation/provider pairs or important input checks are
missing. Cloud live acceptance, native/npm/Docker candidate checks, and
artifact digest/publication binding remain separate release requirements.
**DC не проверен live** must remain visible in QA reports and release notes;
the exception does not waive hermetic Server/DC failures.
