# JavaScript style guide — bitbottle

JavaScript is an install-shim-only language in this project. The total
surface is two files in `packages/mcp-npm/`. This guide exists to keep
that shim small, predictable, and free of supply-chain risk.

## When JavaScript is appropriate

Only the npm postinstall path. The Go binary is the product; npm is one
distribution channel. The shim's job is:

1. Detect the user's platform (`os` / `arch`).
2. Download the matching binary from GitHub Releases.
3. Place it on `$PATH` so `npx @proggarapsody/bitbottle` works.

If it's not strictly that, it doesn't belong in the npm package.

## Versions and dependencies

- **Node**: any LTS. The shim must work on whatever Node version a
  developer or CI runner already has — never pin a specific version in
  `package.json` `engines`.
- **Zero runtime dependencies.** `package.json` has no `dependencies`
  field; everything uses Node's standard library (`https`, `fs`, `path`,
  `os`, `child_process`).
- **Zero build step.** The published files are exactly the source files.

This zero-deps stance is deliberate: every transitive dependency is
supply-chain risk, and the only consumers of this script are users
running it once at install time. Keep it boring.

## File layout

```
packages/mcp-npm/
├── package.json       — name, version, files, scripts.postinstall
├── install.js         — the shim itself
├── bin/run.js         — exec wrapper that runs the downloaded binary
└── README.md          — copied from repo root by the release workflow
```

`files` in `package.json` pins exactly `bin/run.js`, `install.js`, and
`README.md`. Never broaden it — published packages are public forever.

## Style

- **Module system**: CommonJS (`require` / `module.exports`). Node has
  always supported it; ESM adds friction for no upside in a postinstall
  script.
- **Strict mode**: `'use strict';` at the top of every file.
- **No semicolons-or-not religion**: just be consistent within the file.
- **Two-space indent.**
- **`const` everywhere**; `let` only when a variable is genuinely
  reassigned. Never `var`.
- **Async**: prefer `async/await` over callback chains. Wrap legacy
  callback APIs (`https.get`) in `new Promise(...)` rather than nesting.
- **Errors**: throw `Error` (or a subclass) with a clear message. Always
  exit with a non-zero code on failure so npm install fails loudly.
- **No console spam**: print one line per significant action
  (`Downloading bitbottle vX.Y.Z…`, `Installed at …`). Verbose output is
  noise during install.

## Security

- **Verify downloads.** Compute and check the binary's checksum against
  the value GoReleaser publishes in the GitHub release.
- **Use `https` only.** Never fall back to `http`.
- **Validate platform/arch strings** before interpolating them into
  paths or URLs — don't trust `process.arch` to be a known value.
- **No `child_process.exec` with shell strings.** Use
  `spawnSync(file, args, { stdio: 'inherit' })`.

## Tests

There are intentionally none. The shim is exercised by the release
workflow itself: each release publishes to npm, then the workflow
installs the just-published version on a clean runner and runs
`bitbottle --version`. If the smoke test passes, the shim works.

## Formatting

`prettier` if you have it installed locally; otherwise just match the
surrounding style. No formatter is wired into CI for these files.
