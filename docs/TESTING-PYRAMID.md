# Testing Pyramid — bitbottle

Six-tier test strategy. Each tier owns a distinct seam; together they give
confidence that the tool works for users, not just that individual functions
return expected values.

| Tier | What it owns | Where it lives | Catches | Status |
|---|---|---|---|---|
| 1. **Unit** | one function / one struct method | `api/cloud/**/*_test.go`, `api/server/**/*_test.go`, `internal/**/*_test.go` | logic bugs inside a leaf module | shipped |
| 2. **Adapter / integration** | one cobra command against an `httptest` fake | `pkg/cmd/**/*_integration_test.go` | wire compatibility for one route + handler | shipped |
| 3. **Script (TESTSCRIPT)** | whole binary, real argv/env/exit codes, hermetic by default | `test/script/testdata/*.txtar` | flag wiring, alias expansion, errfmt rendering, `bitbottle.host` defaulting, capability gaps, env handling, upgrade paths | shipped |
| 4. **Contract** | cross-file invariants asserted in Go | one focused `*_contract_test.go` per invariant | "every flag named in an errfmt hint is registered on the failing command", "every backend exposes every feature in `AllFeatures`", `--json` field stability | shipped |
| 5. **System / pipeline** | release tooling outside Go (`goreleaser`, `npm publish`, `cosign`, `slsa-generator`) | PR-gated dry-run GHA workflow on changes to `.goreleaser.yaml`, `.github/workflows/release.yml`, `packages/mcp-npm/` | release rework cascades | shipped |
| 6. **Live wire** | real Bitbucket Server + Cloud sandbox | nightly GHA gated by `BITBOTTLE_E2E=1`; reuses tier-3 `.txtar` corpus | wire-level drift the hermetic fake by construction cannot detect | nightly (cron — `nightly-e2e` workflow) |

---

## Priority order

Tiers 3–5 were the highest-leverage gaps. Tier 6 closes the live-wire seam.

1. **TESTSCRIPT** (tier 3) — single biggest leverage; covers ~60% of recent escapes.
2. **RELEASE-DRY-RUN** (tier 5) — covers the 30% release-pipeline rework cluster.
3. **HINT-FLAG-CONTRACT** (tier 4) — five-line fixture, catches the entire #387/#394 class.
4. **UPGRADE-PATH-TESTS** (tier 3 sub-corpus) — drops pre-migration state on disk inside `.txtar`, asserts the upgrade.
5. **NIGHTLY-E2E** (tier 6) — `.github/workflows/nightly-e2e.yml`; catches real-backend drift invisible to hermetic fakes.

---

## Principles

- **No new feature ships without at least one `.txtar` script** once TESTSCRIPT lands.
- **Every error hint is a contract.** If a hint says "pass `--foo`", a contract test must prove `--foo` is reachable from the failing command.
- **Every release-pipeline change runs the pipeline.**
- **Upgrade is a tested operation.**
- **Coverage % is not the metric.** The metric is "what fraction of seams are observed at composition."
