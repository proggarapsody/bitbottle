# Release process

Release Please manages version PRs and creates the version tag. The release
workflow verifies a sealed candidate before publishing it. PR verification and
production preparation share `.github/workflows/release-candidate.yml`.

## Verify a proposed version

Run the read-only wrapper against the intended source revision, explicitly naming
the real proposed version:

```sh
gh workflow run release-dryrun.yml --ref main -f tag=v1.136.0
```

For a branch awaiting merge, replace `main` with that branch. Relevant PR changes
also trigger this wrapper. An omitted tag uses `.release-please-manifest.json`;
use the explicit tag when checking a future release. This workflow never
publishes, creates remote tags, or receives publication credentials.

The preparation job pins the source SHA, runs candidate-validator tests and the
full hermetic public-operation suite with `--require-critical`, then invokes
GoReleaser once using the real version and `--skip=publish`. A local temporary tag
supplies the proposed version; this is not a snapshot build. Changelog generation
uses local Git history so a proposed tag need not exist remotely. Build hooks
must leave tracked source files unchanged.

The candidate contains the original GoReleaser archives, checksums, metadata,
`.deb`/`.rpm`/`.apk` packages, one packed npm tarball, both Docker architectures in
an OCI layout, an SPDX SBOM of the GoReleaser files, and release notes including `docs/quality/release-limitations.md`.
`candidate-manifest.json` binds all those bytes to the full source SHA and tag.
The immutable workflow artifact is downloaded by every verification job.

All five native targets must pass: Linux amd64/arm64, macOS amd64/arm64, and
Windows amd64. Each checks the archive identity, command behavior, a local
read-only provider fixture, and MCP startup. Both Linux targets additionally
install the packed npm wrapper and run their sealed Docker image. Reports bind
the archive and other tested digests to the root candidate manifest hash. A final
job rejects missing or inconsistent receipts and produces a publication plan
without executing it.

A green local test suite validates the checking code. Only completed candidate CI
at the intended SHA validates native execution; cross-compilation, a prior SHA,
and skipped jobs do not count as runtime proof.

## Publish the verified files

A trusted version-tag push may publish. A manual `release.yml` dispatch runs
verification by default; publication requires `publish=true` and dispatching the
workflow from `main`. Its `tag` input names an existing remote tag, whose commit
must match the candidate SHA and belong to `main` history.

Before any registry authentication or publication, the publisher requires a
successful trusted live Cloud acceptance run from `main` at exactly that SHA,
including passing acceptance reports and independently verified cleanup. If
that proof is absent, publication fails: run the existing Cloud acceptance
workflow on `main` at that SHA, then retry. The existing main-only environment
and independent-reviewer policy remain in force. This gate does not provide new
credentials, approve reviews, or treat the actor as an independent reviewer.

The publisher rechecks the candidate manifest and every receipt, then uploads
the existing release assets, publishes the existing npm tarball, and copies the
sealed OCI images while preserving their digests. It does not rebuild or repack.
Source or byte changes require a new candidate and new verification receipts.

The SBOM is generated before sealing and uploaded unchanged with the release.
After successful publication, the SLSA workflow attaches provenance whose subjects
are the candidate manifest, checksums, and SBOM. These asset digests bind the
remaining sealed files through the manifest. Dry runs perform no provenance
release upload.

The sealed candidate notes carry the current Cloud auto-merge limitation and
`DC не проверен live`. Hermetic fixtures and candidate runtime checks do not
establish live Server/DC compatibility.
