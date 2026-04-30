# Changelog

## [1.8.0](https://github.com/proggarapsody/bitbottle/compare/v1.7.3...v1.8.0) (2026-04-30)


### Features

* seamless audit — bug-class fixes for v1.8.0 ([#49](https://github.com/proggarapsody/bitbottle/issues/49)) ([5eb221f](https://github.com/proggarapsody/bitbottle/commit/5eb221fb801254dd304ae065d2c37c0e2102dfbe))

## [1.7.3](https://github.com/proggarapsody/bitbottle/compare/v1.7.2...v1.7.3) (2026-04-30)


### Bug Fixes

* **server:** fetch PR version before merge to prevent HTTP 409 ([#45](https://github.com/proggarapsody/bitbottle/issues/45)) ([fce1b91](https://github.com/proggarapsody/bitbottle/commit/fce1b91134607e332c024b7c14cd75230259ea89))

## [1.7.2](https://github.com/proggarapsody/bitbottle/compare/v1.7.1...v1.7.2) (2026-04-30)


### Bug Fixes

* **cloud:** 4 bugs found during Cloud manual-test run ([dd1c223](https://github.com/proggarapsody/bitbottle/commit/dd1c2236d061ea3c078cc47b246f5791d591ab60))

## [1.7.1](https://github.com/proggarapsody/bitbottle/compare/v1.7.0...v1.7.1) (2026-04-29)


### Bug Fixes

* **cloud:** multiple Bitbucket Cloud adapter fixes + auth improvements ([#41](https://github.com/proggarapsody/bitbottle/issues/41)) ([4d4be63](https://github.com/proggarapsody/bitbottle/commit/4d4be638f2013d85a4f5226ab1d9ed449be1f28f))

## [1.6.4](https://github.com/proggarapsody/bitbottle/compare/v1.6.3...v1.6.4) (2026-04-29)


### Bug Fixes

* **cloud:** multiple Bitbucket Cloud adapter fixes + auth improvements ([#41](https://github.com/proggarapsody/bitbottle/issues/41)) ([4d4be63](https://github.com/proggarapsody/bitbottle/commit/4d4be638f2013d85a4f5226ab1d9ed449be1f28f))

## [1.6.3](https://github.com/proggarapsody/bitbottle/compare/v1.6.2...v1.6.3) (2026-04-29)


### Bug Fixes

* set Content-Type on bodyless POST/DELETE to pass Bitbucket Server CSRF check ([#32](https://github.com/proggarapsody/bitbottle/issues/32)) ([35a55f6](https://github.com/proggarapsody/bitbottle/commit/35a55f674f5a88edc0ac339db613dab3167041ff))

## [1.6.2](https://github.com/proggarapsody/bitbottle/compare/v1.6.1...v1.6.2) (2026-04-28)


### Bug Fixes

* **server:** use POST/DELETE .../approve for PR approval ([#30](https://github.com/proggarapsody/bitbottle/issues/30)) ([fc4e62a](https://github.com/proggarapsody/bitbottle/commit/fc4e62ad3c030261788860ebea6714648b41134c))

## [1.6.1](https://github.com/proggarapsody/bitbottle/compare/v1.6.0...v1.6.1) (2026-04-28)


### Bug Fixes

* **pr:** add --head flag to pr create ([#28](https://github.com/proggarapsody/bitbottle/issues/28)) ([b1f798e](https://github.com/proggarapsody/bitbottle/commit/b1f798e0b23e595ca1d0f42ebc72aac97976730d))

## [1.6.0](https://github.com/proggarapsody/bitbottle/compare/v1.5.0...v1.6.0) (2026-04-28)


### Features

* add `pr comment` and `commit status` commands ([#26](https://github.com/proggarapsody/bitbottle/issues/26)) ([f4e18ff](https://github.com/proggarapsody/bitbottle/commit/f4e18fff8df9d5b5d91519980f8a88ab18e1ceef))

## [1.5.0](https://github.com/proggarapsody/bitbottle/compare/v1.4.0...v1.5.0) (2026-04-28)


### Features

* add api/config/alias commands modeled on gh CLI ([#24](https://github.com/proggarapsody/bitbottle/issues/24)) ([6e5679a](https://github.com/proggarapsody/bitbottle/commit/6e5679a5f059da9b5a30621259ca619fccd2e266))

## [1.4.0](https://github.com/proggarapsody/bitbottle/compare/v1.3.0...v1.4.0) (2026-04-27)


### Features

* gh CLI UX patterns + 5 bug fixes ([#22](https://github.com/proggarapsody/bitbottle/issues/22)) ([3edf123](https://github.com/proggarapsody/bitbottle/commit/3edf123bccaa27bb31edb142a89fe357978adf03))

## [1.3.0](https://github.com/proggarapsody/bitbottle/compare/v1.2.0...v1.3.0) (2026-04-27)


### Features

* **auth:** probe PAT management URL before opening browser on Server/DC ([b6e7701](https://github.com/proggarapsody/bitbottle/commit/b6e7701ddf44fb8fc025b92ea16b06b3fdc787d3))


### Bug Fixes

* **auth:** use NewRequestWithContext in PAT URL probe (noctx) ([#20](https://github.com/proggarapsody/bitbottle/issues/20)) ([287dec9](https://github.com/proggarapsody/bitbottle/commit/287dec91f5bc178d03bad75788949dd04fc87220))

## [1.2.0](https://github.com/proggarapsody/bitbottle/compare/v1.1.3...v1.2.0) (2026-04-27)


### Features

* **auth:** interactive guided login flow ([d95262b](https://github.com/proggarapsody/bitbottle/commit/d95262b5dd150bfcbfe20ac1848722d384ba5b67))
* **auth:** open browser to PAT page during interactive login ([31f85e2](https://github.com/proggarapsody/bitbottle/commit/31f85e2d0c2b22544aeec3f7f6f90f14bfcb0127))


### Bug Fixes

* **auth:** implement OS keyring and always print PAT URL ([685e65c](https://github.com/proggarapsody/bitbottle/commit/685e65c91d8b3db2a651fbea2f949d4632bdea25))

## [1.1.3](https://github.com/proggarapsody/bitbottle/compare/v1.1.2...v1.1.3) (2026-04-27)


### Bug Fixes

* auth login fails with HTTP 404 on Bitbucket Server (GET /users/~ unsupported) ([#16](https://github.com/proggarapsody/bitbottle/issues/16)) ([01aa25b](https://github.com/proggarapsody/bitbottle/commit/01aa25b0b656e836619ddbabc56ca57b5c1c6bbd))

## [1.1.2](https://github.com/proggarapsody/bitbottle/compare/v1.1.1...v1.1.2) (2026-04-27)


### Bug Fixes

* keyring stub panics crash auth login, status, and logout ([#15](https://github.com/proggarapsody/bitbottle/issues/15)) ([2b64f3a](https://github.com/proggarapsody/bitbottle/commit/2b64f3ac865a31133683ff61702209dbd185a938))

## [1.1.1](https://github.com/proggarapsody/bitbottle/compare/v1.1.0...v1.1.1) (2026-04-26)


### Bug Fixes

* auth login interactive prompt, error visibility, skip-tls-verify race ([#11](https://github.com/proggarapsody/bitbottle/issues/11)) ([ba22eeb](https://github.com/proggarapsody/bitbottle/commit/ba22eeb105c0033231c9eaeb7f11f71a89b20eeb))

## [1.1.0](https://github.com/proggarapsody/bitbottle/compare/v1.0.1...v1.1.0) (2026-04-26)


### Features

* expose full CLI via npm wrapper, not just mcp subcommand ([0b0695e](https://github.com/proggarapsody/bitbottle/commit/0b0695ec1581506cdec3489595fde70721ecdf7d))
* expose full CLI via npm wrapper, not just mcp subcommand ([fee9a10](https://github.com/proggarapsody/bitbottle/commit/fee9a107934cc46f8693ae22dd17c9a70493a944))

## [1.0.1](https://github.com/proggarapsody/bitbottle/compare/v1.0.0...v1.0.1) (2026-04-26)


### Bug Fixes

* document NPM_TOKEN granular access token requirement ([9664b0f](https://github.com/proggarapsody/bitbottle/commit/9664b0fdd2c58960d1b6417a0e89d384cd4f720c))
* document NPM_TOKEN must be granular access token with 2FA bypass ([0bac29d](https://github.com/proggarapsody/bitbottle/commit/0bac29d37d62a1fa4637d21937a4cf8f507aa1b4))

## 1.0.0 (2026-04-26)


### Features

* add auth token and auth refresh commands (scope P) ([a0dc1e7](https://github.com/proggarapsody/bitbottle/commit/a0dc1e7039cdf9511307676aa7242199b3a65dfc))
* add Bitbucket Cloud (api.bitbucket.org) backend support ([baca9cb](https://github.com/proggarapsody/bitbottle/commit/baca9cb15ed091047333732904242e032f696786))
* add cloud commit adapter (scope F) ([3accf67](https://github.com/proggarapsody/bitbottle/commit/3accf676e75942dc404077fcc44c7bd96b89b3e8))
* add Commit domain type and interfaces (scope F) ([207b73f](https://github.com/proggarapsody/bitbottle/commit/207b73f7664c4e8ff5502a6a712305a1dda7ff0d))
* add commit log and commit view commands (scope F) ([285e8ba](https://github.com/proggarapsody/bitbottle/commit/285e8ba998a73b8fd1151c03e00c9f17f3ae12a1))
* add list_commits and get_commit MCP tools (scope F) ([8969462](https://github.com/proggarapsody/bitbottle/commit/896946219b8b48907a45f9bba30fd5ea3d72bc3e))
* add pipeline and branch commands with MCP tools ([fbcfff8](https://github.com/proggarapsody/bitbottle/commit/fbcfff8bf63b3c6d976af8a25f5821df89992307))
* add server commit adapter (scope F) ([d67369f](https://github.com/proggarapsody/bitbottle/commit/d67369fb500579d61abe06e83601a5b6290a15e6))
* add shell completion command (scope M) ([f99c554](https://github.com/proggarapsody/bitbottle/commit/f99c554a3376c37e4a3a5903c55bdee68e6fa88a))
* **api:** Bitbucket REST client with typed error handling ([bab6092](https://github.com/proggarapsody/bitbottle/commit/bab6092f50438a6bfdd1a83064f44e36d858c64c))
* **auth:** login, logout, and status commands ([15c26af](https://github.com/proggarapsody/bitbottle/commit/15c26afa18e90121dfe5cbedaa47e5dddcb1b790))
* branch create and checkout commands (scope-l) ([701d5d2](https://github.com/proggarapsody/bitbottle/commit/701d5d213489e06c223df495f408ca5ffd0732c6))
* **git:** git wrapper around pluggable Runner interface ([a090bc1](https://github.com/proggarapsody/bitbottle/commit/a090bc1be88acb2c68f3d0f4be777242d93ad187))
* implement --json/--jq output for repo and pr commands ([aaaf856](https://github.com/proggarapsody/bitbottle/commit/aaaf8568b2e378c5ea374785d420b0cef4061c6b))
* implement auth, repo, and pr commands ([480a4f4](https://github.com/proggarapsody/bitbottle/commit/480a4f49b8069f0e19e2c632d679043baea3ccbf))
* implement MCP server (bitbottle mcp serve) ([2418db5](https://github.com/proggarapsody/bitbottle/commit/2418db593e84cc8af38ad305cc7ec35207a04087))
* **internal:** bbrepo parsing and bbinstance URL builders ([3e44de9](https://github.com/proggarapsody/bitbottle/commit/3e44de93c7be45438dae4a54dffcb6c233309c86))
* **internal:** config, keyring, run, and text packages ([70ab853](https://github.com/proggarapsody/bitbottle/commit/70ab85339003967a281991410260acd36d9e7ee6))
* PR lifecycle commands (scope-g) ([08e191b](https://github.com/proggarapsody/bitbottle/commit/08e191bf20e3fac7c5294db29095532d515d114d))
* **pr:** pr list command with integration tests ([dd4f7b9](https://github.com/proggarapsody/bitbottle/commit/dd4f7b9a9c930ef0f5cb51596a2619baf4b7f61b))
* **repo:** repo list command with integration tests ([9f8ffa4](https://github.com/proggarapsody/bitbottle/commit/9f8ffa4648e70a503da109aafb13e655a4a29104))
* **scope-e:** tag list, create, and delete commands ([09da190](https://github.com/proggarapsody/bitbottle/commit/09da190f158c859066859e7c9829799ed95fccf8))
* **scope-g:** PR lifecycle commands (edit, decline, unapprove, ready, request-review, request-changes) ([80666b1](https://github.com/proggarapsody/bitbottle/commit/80666b15712047f4dadec0ba7a080b67deb85b20))
* **scope-l:** branch create and checkout commands ([9c04c1a](https://github.com/proggarapsody/bitbottle/commit/9c04c1a189ee6db955c7806c1cc9afa52a9076e9))
* **tableprinter:** TTY-aware table printer with headers and UTF-8 support ([7f6288c](https://github.com/proggarapsody/bitbottle/commit/7f6288c1066ba1d55168f06ff536f0cae7b281b8))
* tag list, create, and delete commands (scope-e) ([f17887b](https://github.com/proggarapsody/bitbottle/commit/f17887bd7223a61c9ac089615fb001a679f99bb9))


### Bug Fixes

* add missing cmd/bitbottle entrypoint and fix golangci-lint config ([d6ab188](https://github.com/proggarapsody/bitbottle/commit/d6ab1883ab4d3207002ff79f8e9f5ac91a6d22b4))
* downgrade mcp-go to v0.48.0, pin go 1.23 for golangci-lint compat ([a9c2cc9](https://github.com/proggarapsody/bitbottle/commit/a9c2cc971e9f763b3d640f53286a43a4df18f2ba))
* fix goimports grouping across all packages ([7b14012](https://github.com/proggarapsody/bitbottle/commit/7b1401255f41012e5502048640b6e0f504db976d))
* gofmt formatting across new files ([4fd5b57](https://github.com/proggarapsody/bitbottle/commit/4fd5b57052f5b0620f19d5e99cdca93de636d504))
* resolve golangci-lint failures (noctx, gofmt, goimports) ([b94e1aa](https://github.com/proggarapsody/bitbottle/commit/b94e1aa34b2f3dcf43dc75146376aacf07b155c7))
* **scope-e:** add --web to tag list; gate delete prompt on TTY ([f17f0b6](https://github.com/proggarapsody/bitbottle/commit/f17f0b62f388f1d676f71806b79a50975ddcd51d))
* **scope-e:** use MarkFlagRequired for start-at, remove dead test writes ([c4781dc](https://github.com/proggarapsody/bitbottle/commit/c4781dcb2b9b3e91fe3f460c7432c484511cb92f))
* **scope-g:** pr ready prints URL via GetPR after success ([5539f8a](https://github.com/proggarapsody/bitbottle/commit/5539f8a5f9413d35251c45ad7a41758ec0f74732))
* **scope-g:** server ReadyPR GET-then-PUT, MCP readyPR returns PR data, minor guards ([c00c347](https://github.com/proggarapsody/bitbottle/commit/c00c34742132bb13847b33f7e1d0fe1deb7fbd9a))
* **scope-l:** use MarkFlagRequired for start-at, fix test hash length ([1aaa441](https://github.com/proggarapsody/bitbottle/commit/1aaa441533185488be78eae433c4c335092082d9))
* unblock cmd/bitbottle from gitignore and add entrypoint ([7b427ad](https://github.com/proggarapsody/bitbottle/commit/7b427ad2a48b7f23df20907460c893f8cfd5db6f))
