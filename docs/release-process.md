# Release Process

Releases are automated via release-please + GoReleaser + npm publish.

## PR-gated dry-run

Any PR touching `.goreleaser.yaml`, `.github/workflows/release.yml`, `.github/workflows/scorecard.yml`, or `packages/mcp-npm/**` triggers `.github/workflows/release-dryrun.yml`:

- **goreleaser-dryrun**: builds snapshot binaries (`--snapshot --clean --skip=publish`) to verify the build matrix works
- **npm-dryrun**: runs `npm publish --dry-run` to verify the npm package config is correct

Both jobs must pass before the PR can merge. This is the canonical gate for release pipeline changes.
