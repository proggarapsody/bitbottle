# Changelog

## [1.77.1](https://github.com/proggarapsody/bitbottle/compare/v1.77.0...v1.77.1) (2026-05-18)


### Bug Fixes

* **errfmt:** remove phantom --debug hint from transport.timeout ([#394](https://github.com/proggarapsody/bitbottle/issues/394)) ([c24f964](https://github.com/proggarapsody/bitbottle/commit/c24f964ea9e2c2842140029cc5a4ecef9a259887))

## [1.77.0](https://github.com/proggarapsody/bitbottle/compare/v1.76.2...v1.77.0) (2026-05-18)


### Features

* **auth:** auto-detect self-signed CA and prompt to trust on login ([#390](https://github.com/proggarapsody/bitbottle/issues/390)) ([b30ec09](https://github.com/proggarapsody/bitbottle/commit/b30ec09fe0f7f73f113ae25c48ee1a4bc3308f28))


### Bug Fixes

* **root:** wire global -k / --skip-tls-verify flag ([#387](https://github.com/proggarapsody/bitbottle/issues/387)) ([4e5b753](https://github.com/proggarapsody/bitbottle/commit/4e5b75357165716c88a0aef56b30b1e4b14dd6d4))

## [1.76.2](https://github.com/proggarapsody/bitbottle/compare/v1.76.1...v1.76.2) (2026-05-18)


### Bug Fixes

* **npm:** add repository.url to package.json for provenance verification ([#391](https://github.com/proggarapsody/bitbottle/issues/391)) ([46ece44](https://github.com/proggarapsody/bitbottle/commit/46ece44ec866ed945db9b9b87e481b0e08d83b5a))

## [1.76.1](https://github.com/proggarapsody/bitbottle/compare/v1.76.0...v1.76.1) (2026-05-18)


### Bug Fixes

* **release:** remove goreleaser signs block to fix cosign PATH issue ([#388](https://github.com/proggarapsody/bitbottle/issues/388)) ([0f0a7e2](https://github.com/proggarapsody/bitbottle/commit/0f0a7e2b166411e2e5c92c0d8a90b2978f375d81))

## [1.76.0](https://github.com/proggarapsody/bitbottle/compare/v1.75.0...v1.76.0) (2026-05-18)


### Features

* pr edit --remove-reviewer removes reviewers from a PR ([1add15f](https://github.com/proggarapsody/bitbottle/commit/1add15fa73adf3ab042fa6de5bec342dca020585))


### Bug Fixes

* **release:** remove cosign-installer to fix goreleaser bundle verification ([3136c80](https://github.com/proggarapsody/bitbottle/commit/3136c80bdf641402f2e67b9e5ffafbd1be492855))

## [1.75.0](https://github.com/proggarapsody/bitbottle/compare/v1.74.0...v1.75.0) (2026-05-18)


### Features

* **pr:** add pr unready command to convert PR back to draft ([a80700b](https://github.com/proggarapsody/bitbottle/commit/a80700b5fa8b3038e1ab3552209eb60b9d989750)), closes [#380](https://github.com/proggarapsody/bitbottle/issues/380)

## [1.74.0](https://github.com/proggarapsody/bitbottle/compare/v1.73.1...v1.74.0) (2026-05-18)


### Features

* **security:** OpenSSF Scorecard score improvements (token-perms, SAST, signed releases) ([f7eb1bf](https://github.com/proggarapsody/bitbottle/commit/f7eb1bf749b9eb159a41e81787b783f9f085828b)), closes [#377](https://github.com/proggarapsody/bitbottle/issues/377)

## [1.73.1](https://github.com/proggarapsody/bitbottle/compare/v1.73.0...v1.73.1) (2026-05-18)


### Bug Fixes

* **auth:** post-migrate auth + hostname scheme normalization + workflow rule ([08b3d9a](https://github.com/proggarapsody/bitbottle/commit/08b3d9a0a577dd9aa44d286b2603ff88d997b177))

## [1.73.0](https://github.com/proggarapsody/bitbottle/compare/v1.72.1...v1.73.0) (2026-05-17)


### Features

* **backend:** add AllFeatureSpecs registry + capability contract tests ([0078124](https://github.com/proggarapsody/bitbottle/commit/0078124fb8725eb03abb68c1d7a19567dbf0fcac))
* **backend:** stamp ErrInvalidRequest+CodeInvalidRequest on bare 400/422 (ERR-EMPTY-400) ([e58ee30](https://github.com/proggarapsody/bitbottle/commit/e58ee30bf94100b5e9538370a3c8c8bb65a5b521))
* **cmd:** add --json golden-file tests for field stability (JSON-STABILITY) ([ce115e5](https://github.com/proggarapsody/bitbottle/commit/ce115e500911910bd999236278a06f15b4dba006))
* **cmd:** surface partial list results on mid-pagination error ([74e0804](https://github.com/proggarapsody/bitbottle/commit/74e08049c3c210e09749aeae22a001c2674b00f3))
* **mcp:** add validateEnum + validateRange input validation (MCP-VALIDATION) ([8f53e29](https://github.com/proggarapsody/bitbottle/commit/8f53e290b55971155465a41e4764f340a373e4c5))
* **profile:** validate --backend type at profile create time (BACKEND-TYPE-STRICT) ([1dbc66d](https://github.com/proggarapsody/bitbottle/commit/1dbc66d99e461fbf6ab17d41012ed34f9461f29f))

## [1.72.1](https://github.com/proggarapsody/bitbottle/compare/v1.72.0...v1.72.1) (2026-05-17)


### Bug Fixes

* **cloud:** omit empty variables in pipeline trigger body ([7a25c13](https://github.com/proggarapsody/bitbottle/commit/7a25c136c263fe44b01364232d3d680d669f717e))

## [1.72.0](https://github.com/proggarapsody/bitbottle/compare/v1.71.1...v1.72.0) (2026-05-15)


### Features

* **openapi-types:** add gen/ infrastructure and migrate server PR wire types ([6dbbbec](https://github.com/proggarapsody/bitbottle/commit/6dbbbec011c73879d42ee2797afdfc71304ee61c))

## [1.71.1](https://github.com/proggarapsody/bitbottle/compare/v1.71.0...v1.71.1) (2026-05-15)


### Bug Fixes

* **cmd:** validate --limit must be &gt;= 1 in list commands ([38b31dd](https://github.com/proggarapsody/bitbottle/commit/38b31dd2262caff11b3a14f5da0f6ced5c27350c))

## [1.71.0](https://github.com/proggarapsody/bitbottle/compare/v1.70.2...v1.71.0) (2026-05-15)


### Features

* **ci:** enforce layer boundary via depguard + smell-scan rule 5 ([358de5e](https://github.com/proggarapsody/bitbottle/commit/358de5e2d563e8462e408391cd9a2053d49669b6))

## [1.70.2](https://github.com/proggarapsody/bitbottle/compare/v1.70.1...v1.70.2) (2026-05-15)


### Bug Fixes

* **reactions:** surface concurrent fetch errors and DRY into shared helper ([8e88260](https://github.com/proggarapsody/bitbottle/commit/8e88260ce18516f188bc9898de145d1b17b895cd))

## [1.70.1](https://github.com/proggarapsody/bitbottle/compare/v1.70.0...v1.70.1) (2026-05-15)


### Bug Fixes

* **test:** add compile-time interface assertions to FakeClient ([5ead6e2](https://github.com/proggarapsody/bitbottle/commit/5ead6e2a9328d04bc45c2a454db4a6fc980ceab7))

## [1.70.0](https://github.com/proggarapsody/bitbottle/compare/v1.69.0...v1.70.0) (2026-05-15)


### Features

* **auth-doctor:** add auth doctor diagnostics command ([e75dc74](https://github.com/proggarapsody/bitbottle/commit/e75dc74b29cead46a7e2d555db1dc06958c2a6d7))

## [1.69.0](https://github.com/proggarapsody/bitbottle/compare/v1.68.0...v1.69.0) (2026-05-15)


### Features

* **pr-reviewer-group:** add pr reviewer-group list/add/remove commands ([4870d20](https://github.com/proggarapsody/bitbottle/commit/4870d201e71828faab6f8fd042200c8c688f1143))

## [1.68.0](https://github.com/proggarapsody/bitbottle/compare/v1.67.0...v1.68.0) (2026-05-15)


### Features

* **pr-suggestion:** add pr suggestion apply command ([b78aad5](https://github.com/proggarapsody/bitbottle/commit/b78aad5ce7f4d60b6e531d8bee17cfe132d2df5b))

## [1.67.0](https://github.com/proggarapsody/bitbottle/compare/v1.66.0...v1.67.0) (2026-05-15)


### Features

* **pipeline-cache:** add pipeline cache list and delete commands ([6b995e5](https://github.com/proggarapsody/bitbottle/commit/6b995e5f9bf8613304100896419dfe0b92e582c3))

## [1.66.0](https://github.com/proggarapsody/bitbottle/compare/v1.65.0...v1.66.0) (2026-05-15)


### Features

* **workspace-hooks:** add workspace hook list/create/delete (Cloud only) ([178e4de](https://github.com/proggarapsody/bitbottle/commit/178e4de479724075f334b3ca0a3ea68cda90e0bc)), closes [#318](https://github.com/proggarapsody/bitbottle/issues/318)

## [1.65.0](https://github.com/proggarapsody/bitbottle/compare/v1.64.0...v1.65.0) (2026-05-15)


### Features

* **repo-visibility:** add repo visibility get/set command ([5b089f2](https://github.com/proggarapsody/bitbottle/commit/5b089f224bf2dba9afd82c534fc767fdcc54d713)), closes [#315](https://github.com/proggarapsody/bitbottle/issues/315)

## [1.64.0](https://github.com/proggarapsody/bitbottle/compare/v1.63.0...v1.64.0) (2026-05-15)


### Features

* **repo-forks:** add repo fork list command (both backends) ([d4c5e16](https://github.com/proggarapsody/bitbottle/commit/d4c5e16cad89aac48270a14383b185b1836e21ef))

## [1.63.0](https://github.com/proggarapsody/bitbottle/compare/v1.62.0...v1.63.0) (2026-05-15)


### Features

* **user-view:** add user view command (both backends) ([d7efe70](https://github.com/proggarapsody/bitbottle/commit/d7efe706d21748e4861793b4128533af3fb4268a))

## [1.62.0](https://github.com/proggarapsody/bitbottle/compare/v1.61.0...v1.62.0) (2026-05-15)


### Features

* **auto-iter:** tier 1 + tier 2 scripts — pick/await/overlap/worktree/pre-merge ([#302](https://github.com/proggarapsody/bitbottle/issues/302)) ([30bca81](https://github.com/proggarapsody/bitbottle/commit/30bca8140b341e9102a6d98fca788069c44d91e7))
* **workspace-members:** add workspace member list command (Cloud only) ([78d8439](https://github.com/proggarapsody/bitbottle/commit/78d843997dfc0b0c59b4c3361d92478192403ba6))

## [1.61.0](https://github.com/proggarapsody/bitbottle/compare/v1.60.0...v1.61.0) (2026-05-14)


### Features

* **auto-iter:** scripts foundation for mechanical pipeline steps ([#298](https://github.com/proggarapsody/bitbottle/issues/298)) ([890fd7c](https://github.com/proggarapsody/bitbottle/commit/890fd7cb3c7de2031280de22f9227f45f1cad122))

## [1.60.0](https://github.com/proggarapsody/bitbottle/compare/v1.59.0...v1.60.0) (2026-05-13)


### Features

* **pr:** add pr participant list command ([8ba3366](https://github.com/proggarapsody/bitbottle/commit/8ba336651b417aa197b61b33763cdaf9e5389e6d))

## [1.59.0](https://github.com/proggarapsody/bitbottle/compare/v1.58.0...v1.59.0) (2026-05-13)


### Features

* **commit:** add commit status report command ([dc4ca70](https://github.com/proggarapsody/bitbottle/commit/dc4ca7035ba3db45eddbdf9c529fc78423cbd727))

## [1.58.0](https://github.com/proggarapsody/bitbottle/compare/v1.57.0...v1.58.0) (2026-05-13)


### Features

* **repo:** add repo watcher list command ([0b9160f](https://github.com/proggarapsody/bitbottle/commit/0b9160f1c095435030ca3ad1495be6f115c89d6a))

## [1.57.0](https://github.com/proggarapsody/bitbottle/compare/v1.56.1...v1.57.0) (2026-05-13)


### Features

* **pr:** add pr files command ([983862d](https://github.com/proggarapsody/bitbottle/commit/983862dccd7e16ba73f6e36c3701668b09013973))

## [1.56.1](https://github.com/proggarapsody/bitbottle/compare/v1.56.0...v1.56.1) (2026-05-13)


### Bug Fixes

* **api/cloud:** paginate ListWebhooks, ListPipelineSteps, ListPipelineVariables ([f888e97](https://github.com/proggarapsody/bitbottle/commit/f888e9792eb0a3fb13ce89e0502f1de2f2c148e9))

## [1.56.0](https://github.com/proggarapsody/bitbottle/compare/v1.55.0...v1.56.0) (2026-05-13)


### Features

* **pr:** add pr commits command (list commits in a PR) ([8f77d91](https://github.com/proggarapsody/bitbottle/commit/8f77d918bea9c09bb2a75f685ba0fe1b36386ca7))

## [1.55.0](https://github.com/proggarapsody/bitbottle/compare/v1.54.0...v1.55.0) (2026-05-13)


### Features

* **commit:** add commit files command (list changed files) ([71f5001](https://github.com/proggarapsody/bitbottle/commit/71f500168eedd4e50d1b4cdaa63ee3d4fa245e75))

## [1.54.0](https://github.com/proggarapsody/bitbottle/compare/v1.53.0...v1.54.0) (2026-05-13)


### Features

* **pipeline:** add pipeline schedule list/create/delete ([88bd006](https://github.com/proggarapsody/bitbottle/commit/88bd006960a605c9345733010f86425d5b98f530))

## [1.53.0](https://github.com/proggarapsody/bitbottle/compare/v1.52.0...v1.53.0) (2026-05-13)


### Features

* **branch-rule:** add branch-rule list/add/delete commands ([7581354](https://github.com/proggarapsody/bitbottle/commit/758135423a0a3ed799dcd0971f1ef0e1093a72e3))

## [1.52.0](https://github.com/proggarapsody/bitbottle/compare/v1.51.0...v1.52.0) (2026-05-13)


### Features

* **repo:** add repo transfer command ([cfdc0ab](https://github.com/proggarapsody/bitbottle/commit/cfdc0ab36b51a81c961ff2b7b518fae6fa5cde39))

## [1.51.0](https://github.com/proggarapsody/bitbottle/compare/v1.50.0...v1.51.0) (2026-05-13)


### Features

* **ssh-key:** add user SSH key management (Cloud only) ([6418d9b](https://github.com/proggarapsody/bitbottle/commit/6418d9b44a3b8dc204a3fbc440bd2ec2cb1ca7cb))

## [1.50.0](https://github.com/proggarapsody/bitbottle/compare/v1.49.1...v1.50.0) (2026-05-13)


### Features

* **pr:** add default-reviewer list/add/remove commands ([2981ef4](https://github.com/proggarapsody/bitbottle/commit/2981ef4ab0d28cdf1c67c47792947880dc6ffe81))

## [1.49.1](https://github.com/proggarapsody/bitbottle/compare/v1.49.0...v1.49.1) (2026-05-13)


### Bug Fixes

* **api/cloud:** paginate GetDiffStat via paging.Collect ([c36ec4f](https://github.com/proggarapsody/bitbottle/commit/c36ec4fe430c878358090f6d3e72cd1917bc901d))

## [1.49.0](https://github.com/proggarapsody/bitbottle/compare/v1.48.0...v1.49.0) (2026-05-13)


### Features

* **diff:** add diff command to compare two refs (Cloud + Server/DC) ([8c520fa](https://github.com/proggarapsody/bitbottle/commit/8c520fa2b23de4ef2fdb08c0cf2acce44ab298b7))

## [1.48.0](https://github.com/proggarapsody/bitbottle/compare/v1.47.0...v1.48.0) (2026-05-13)


### Features

* **pipe-trigger:** add pipeline trigger command (Cloud only) ([561a46c](https://github.com/proggarapsody/bitbottle/commit/561a46c407dfb9b6293cf00810e13682e892d5cc))

## [1.47.0](https://github.com/proggarapsody/bitbottle/compare/v1.46.0...v1.47.0) (2026-05-13)


### Features

* **deploy-key:** add deploy-key list/add/delete for Cloud and Server/DC ([5ee450b](https://github.com/proggarapsody/bitbottle/commit/5ee450be8b3ab2f440175dd708b57ed6d338289c))

## [1.46.0](https://github.com/proggarapsody/bitbottle/compare/v1.45.0...v1.46.0) (2026-05-13)


### Features

* **srvver:** server version detection helper with AtLeast and per-client caching ([fa44d89](https://github.com/proggarapsody/bitbottle/commit/fa44d899525ff679167b416e4f6a2b4119dbd7cf))

## [1.45.0](https://github.com/proggarapsody/bitbottle/compare/v1.44.0...v1.45.0) (2026-05-13)


### Features

* **ext-runtime:** extension exec command with env sanitization ([f51d123](https://github.com/proggarapsody/bitbottle/commit/f51d123722ebe398d31d47e8a19ade7b57f10554))

## [1.44.0](https://github.com/proggarapsody/bitbottle/compare/v1.43.0...v1.44.0) (2026-05-13)


### Features

* **ext-mgmt:** extension upgrade and remove commands ([0e4bb6c](https://github.com/proggarapsody/bitbottle/commit/0e4bb6cbc7a0e7beab283cf12e1e1ecec2693644))

## [1.43.0](https://github.com/proggarapsody/bitbottle/compare/v1.42.0...v1.43.0) (2026-05-13)


### Features

* **ext-core:** extension install and list commands ([b81a445](https://github.com/proggarapsody/bitbottle/commit/b81a44585714c8efbeca4e43143f897062da20f7))


### Bug Fixes

* **nix:** open PR for flake SHA update instead of direct push to main ([eb9ba0b](https://github.com/proggarapsody/bitbottle/commit/eb9ba0b88f2e866f27513a234302de9c87f1156d))

## [1.42.0](https://github.com/proggarapsody/bitbottle/compare/v1.41.0...v1.42.0) (2026-05-13)


### Features

* **nix:** Nix flake packaging for nix run github:proggarapsody/bitbottle ([a31a40d](https://github.com/proggarapsody/bitbottle/commit/a31a40d4ea9d7509932e610d62f3da09626600a3)), closes [#211](https://github.com/proggarapsody/bitbottle/issues/211)

## [1.41.0](https://github.com/proggarapsody/bitbottle/compare/v1.40.0...v1.41.0) (2026-05-13)


### Features

* **react-commit:** commit comment emoji reactions (Server/DC only) ([d3e5597](https://github.com/proggarapsody/bitbottle/commit/d3e5597b2faa7c6bb3fc28130db93e39f69db120))


### Bug Fixes

* **pipeline:** preserve required status checks; pin actions/cache ([88ec7b7](https://github.com/proggarapsody/bitbottle/commit/88ec7b7f3b66942c694d78d392166ae81a0b8f77))

## [1.40.0](https://github.com/proggarapsody/bitbottle/compare/v1.39.0...v1.40.0) (2026-05-13)


### Features

* **react-pr:** pr comment react/unreact + reactions column (Server/DC) ([#205](https://github.com/proggarapsody/bitbottle/issues/205)) ([94172f6](https://github.com/proggarapsody/bitbottle/commit/94172f653c0211bffab12b9bc76d9a864cc41c3f))

## [1.39.0](https://github.com/proggarapsody/bitbottle/compare/v1.38.0...v1.39.0) (2026-05-13)


### Features

* **task:** pr task list/create/resolve/reopen (Server/DC) ([1768fbf](https://github.com/proggarapsody/bitbottle/commit/1768fbf66abb259e218a6649b958ebefb0250dd2))


### Bug Fixes

* **task:** fix Cloud interface gate, add MCP handler tests, update skills docs ([#202](https://github.com/proggarapsody/bitbottle/issues/202)) ([da54fb9](https://github.com/proggarapsody/bitbottle/commit/da54fb998fb513fdb8d55545310ec921ed9cd88a))

## [1.38.0](https://github.com/proggarapsody/bitbottle/compare/v1.37.0...v1.38.0) (2026-05-13)


### Features

* **admin:** add admin secrets rotate / logging get/set (Server/DC) ([9f62e44](https://github.com/proggarapsody/bitbottle/commit/9f62e4474665a5f38328f9c430897e55a72dc18d))

## [1.37.0](https://github.com/proggarapsody/bitbottle/compare/v1.36.0...v1.37.0) (2026-05-13)


### Features

* **perms:** add perms project/repo list/grant/revoke (Server/DC) ([#196](https://github.com/proggarapsody/bitbottle/issues/196)) ([62ee06d](https://github.com/proggarapsody/bitbottle/commit/62ee06d22f7837e3ae0ea904924f1e29af233587))

## [1.36.0](https://github.com/proggarapsody/bitbottle/compare/v1.35.0...v1.36.0) (2026-05-12)


### Features

* **cis+varops:** SHA-pin Actions, supply-chain tooling & VariableOps interface ([#181](https://github.com/proggarapsody/bitbottle/issues/181)) ([7fddae2](https://github.com/proggarapsody/bitbottle/commit/7fddae2bcafb9471830438bdc3a67b32c25c05f1))

## [1.35.0](https://github.com/proggarapsody/bitbottle/compare/v1.34.0...v1.35.0) (2026-05-12)


### Features

* **out2:** extended output formats — YAML, Go templates, global flags ([#177](https://github.com/proggarapsody/bitbottle/issues/177)) ([9084c59](https://github.com/proggarapsody/bitbottle/commit/9084c5985c9dd2c7c7e22120510f31f9b821e26b))

## [1.34.0](https://github.com/proggarapsody/bitbottle/compare/v1.33.0...v1.34.0) (2026-05-12)


### Features

* **httph:** HTTP client hardening — retry, rate-limit, ETag cache ([#174](https://github.com/proggarapsody/bitbottle/issues/174)) ([c9eeb90](https://github.com/proggarapsody/bitbottle/commit/c9eeb90d5692a8b818fe65023f4137e51c71f122)), closes [#173](https://github.com/proggarapsody/bitbottle/issues/173)

## [1.33.0](https://github.com/proggarapsody/bitbottle/compare/v1.32.0...v1.33.0) (2026-05-12)


### Features

* **sec:** token-never-in-config, keyring hardening, auth migrate ([#171](https://github.com/proggarapsody/bitbottle/issues/171)) ([b4b1b47](https://github.com/proggarapsody/bitbottle/commit/b4b1b47cea5ca980624f6fb819fcfb375e6b335e))

## [1.32.0](https://github.com/proggarapsody/bitbottle/compare/v1.31.0...v1.32.0) (2026-05-12)


### Features

* **prof:** named credential profiles (profile create/list/use/delete) ([#166](https://github.com/proggarapsody/bitbottle/issues/166)) ([e065e46](https://github.com/proggarapsody/bitbottle/commit/e065e4665447e45bda48eaf2d1a481e490a44e23))

## [1.31.0](https://github.com/proggarapsody/bitbottle/compare/v1.30.0...v1.31.0) (2026-05-12)


### Features

* **var:** variable list/set/delete with --scope repository|workspace|deployment ([#164](https://github.com/proggarapsody/bitbottle/issues/164)) ([8c21b56](https://github.com/proggarapsody/bitbottle/commit/8c21b56762f247649034db2b14f20ebe1d5145a4))

## [1.30.0](https://github.com/proggarapsody/bitbottle/compare/v1.29.0...v1.30.0) (2026-05-12)


### Features

* **automerge:** pr merge --auto / --auto-off and pr view auto-merge status ([#162](https://github.com/proggarapsody/bitbottle/issues/162)) ([8a0034a](https://github.com/proggarapsody/bitbottle/commit/8a0034a3b5049212011f4421aa4861218cd08221))

## [1.29.0](https://github.com/proggarapsody/bitbottle/compare/v1.28.0...v1.29.0) (2026-05-12)


### Features

* **dep:** deployment list/view, environment CRUD, environment variable CRUD ([#159](https://github.com/proggarapsody/bitbottle/issues/159)) ([0da519a](https://github.com/proggarapsody/bitbottle/commit/0da519a906cee99dcd933f4e49ad328effd94963))

## [1.28.0](https://github.com/proggarapsody/bitbottle/compare/v1.27.0...v1.28.0) (2026-05-11)


### Features

* **ghp:** pr checks / pr update-branch / pr status / status / browse / pipeline watch ([#155](https://github.com/proggarapsody/bitbottle/issues/155)) ([f6dbb49](https://github.com/proggarapsody/bitbottle/commit/f6dbb49a5511433f46b2a9fd9ed9080ca36ec5cf))

## [1.27.0](https://github.com/proggarapsody/bitbottle/compare/v1.26.0...v1.27.0) (2026-05-11)


### Features

* **rv6:** commit comment list/add/edit/delete ([#152](https://github.com/proggarapsody/bitbottle/issues/152)) ([5e9ddfd](https://github.com/proggarapsody/bitbottle/commit/5e9ddfdec65dd208602f9e2e479b1c971e23f048))

## [1.26.0](https://github.com/proggarapsody/bitbottle/compare/v1.25.0...v1.26.0) (2026-05-11)


### Features

* **rv5:** pr activity event stream ([#149](https://github.com/proggarapsody/bitbottle/issues/149)) ([66e2487](https://github.com/proggarapsody/bitbottle/commit/66e24870f73351fc726a5bc4b75050b31b88fe47))

## [1.25.0](https://github.com/proggarapsody/bitbottle/compare/v1.24.0...v1.25.0) (2026-05-11)


### Features

* **rv4:** pr review compound command (approve/request-changes/comment + inline) ([#146](https://github.com/proggarapsody/bitbottle/issues/146)) ([28b752f](https://github.com/proggarapsody/bitbottle/commit/28b752f0ca6e994f7c85ce0ec107fed946c89805))

## [1.24.0](https://github.com/proggarapsody/bitbottle/compare/v1.23.0...v1.24.0) (2026-05-09)


### Features

* **rv3:** inline PR comment write path (add inline / edit / delete / resolve) ([#132](https://github.com/proggarapsody/bitbottle/issues/132)) ([3d15050](https://github.com/proggarapsody/bitbottle/commit/3d150500a57b5bc1367433743fd2f1fb946bc3a6)), closes [#131](https://github.com/proggarapsody/bitbottle/issues/131)

## [1.23.0](https://github.com/proggarapsody/bitbottle/compare/v1.22.0...v1.23.0) (2026-05-08)


### Features

* **code-insights:** reports, annotations, and merge-check for Bitbucket Server/DC ([#137](https://github.com/proggarapsody/bitbottle/issues/137)) ([babf073](https://github.com/proggarapsody/bitbottle/commit/babf073e7ed0b7747d2e9c70afd969ff8e8f8494))
* **issue:** finish issue lifecycle (edit/reopen/assign + comments) ([#136](https://github.com/proggarapsody/bitbottle/issues/136)) ([61c4d11](https://github.com/proggarapsody/bitbottle/commit/61c4d11c211154e808ee49e50b597f24126d8ef5))

## [1.22.0](https://github.com/proggarapsody/bitbottle/compare/v1.21.0...v1.22.0) (2026-05-08)


### Features

* **rv1:** repo file get + repo tree (SourceReader primitives) ([#127](https://github.com/proggarapsody/bitbottle/issues/127)) ([676220e](https://github.com/proggarapsody/bitbottle/commit/676220e05613a7540b8d4c6d4239b5f659b33054)), closes [#126](https://github.com/proggarapsody/bitbottle/issues/126)
* **rv2:** surface inline PR review comments on pr comment list ([#130](https://github.com/proggarapsody/bitbottle/issues/130)) ([cdcb7a5](https://github.com/proggarapsody/bitbottle/commit/cdcb7a59e1561061f806592171d5072394e1e046))

## [1.21.0](https://github.com/proggarapsody/bitbottle/compare/v1.20.0...v1.21.0) (2026-05-08)


### Features

* add bitbottle context one-call orientation primitive ([#110](https://github.com/proggarapsody/bitbottle/issues/110)) ([3b376ea](https://github.com/proggarapsody/bitbottle/commit/3b376ea7311a6cca2674ca99c2aa15551813660a))
* add bitbottle search code (Cloud only) ([#111](https://github.com/proggarapsody/bitbottle/issues/111)) ([2a761d9](https://github.com/proggarapsody/bitbottle/commit/2a761d9a5e92057a173696e7827addceafdc12b6))
* **pr:** add reopen command (Server/DC only) ([#112](https://github.com/proggarapsody/bitbottle/issues/112)) ([c7a669e](https://github.com/proggarapsody/bitbottle/commit/c7a669ebcd7957fd608a6066d7da3341af0f01aa))

## [1.20.0](https://github.com/proggarapsody/bitbottle/compare/v1.19.0...v1.20.0) (2026-05-07)


### Features

* **t:** pager opt-in coverage on long-output read-only commands ([#103](https://github.com/proggarapsody/bitbottle/issues/103)) ([e1beb17](https://github.com/proggarapsody/bitbottle/commit/e1beb1766413f2c0bcf5fc275cfef92a816dd9a0))

## [1.19.0](https://github.com/proggarapsody/bitbottle/compare/v1.18.1...v1.19.0) (2026-05-07)


### Features

* scope EX2-EX6 — error UX clusters (repo, pr, branch, network, MCP) ([#100](https://github.com/proggarapsody/bitbottle/issues/100)) ([c88e643](https://github.com/proggarapsody/bitbottle/commit/c88e6436a998db948e23666a0c1c287d44689bc4))

## [1.18.1](https://github.com/proggarapsody/bitbottle/compare/v1.18.0...v1.18.1) (2026-05-07)


### Bug Fixes

* errfmt polish — symmetry, docs, catalogue gate ([#98](https://github.com/proggarapsody/bitbottle/issues/98)) ([efb43c5](https://github.com/proggarapsody/bitbottle/commit/efb43c583a5681a990205c51471f4cbc4e244f61))
* **security:** sanitise terminal control bytes in error rendering ([#95](https://github.com/proggarapsody/bitbottle/issues/95)) ([043331a](https://github.com/proggarapsody/bitbottle/commit/043331a4844ed11e7e0b11954251072a9a211c6a))

## [1.18.0](https://github.com/proggarapsody/bitbottle/compare/v1.17.0...v1.18.0) (2026-05-07)


### Features

* scope EX1 — error UX foundation + auth cluster ([#92](https://github.com/proggarapsody/bitbottle/issues/92)) ([d2c92f8](https://github.com/proggarapsody/bitbottle/commit/d2c92f8ceab62e6e053577178de1b928253fe291))

## [1.17.0](https://github.com/proggarapsody/bitbottle/compare/v1.16.0...v1.17.0) (2026-05-07)


### Features

* scope BP — branch protect (list, create, delete) on Server/DC ([#91](https://github.com/proggarapsody/bitbottle/issues/91)) ([2da93ab](https://github.com/proggarapsody/bitbottle/commit/2da93aba6733972e77814c0d9367e31ec66cdb68))
* scope N — workspace list, project list (Bitbucket Cloud) ([#86](https://github.com/proggarapsody/bitbottle/issues/86)) ([47cb4be](https://github.com/proggarapsody/bitbottle/commit/47cb4be47c2ef20b31cb2a7ece1cafa6af072700))
* scope O — issues (list, view, create, close) on Cloud ([#87](https://github.com/proggarapsody/bitbottle/issues/87)) ([0e3e937](https://github.com/proggarapsody/bitbottle/commit/0e3e937be51cce072633294c65fa2323a0a83b41))
* scope T — colourise state columns + add --no-color global flag ([#85](https://github.com/proggarapsody/bitbottle/issues/85)) ([70d7037](https://github.com/proggarapsody/bitbottle/commit/70d70378dc620c756d8047f0bf7e7530bc0507c6))


### Bug Fixes

* apply default reviewers automatically on pr create ([#89](https://github.com/proggarapsody/bitbottle/issues/89)) ([d16bb40](https://github.com/proggarapsody/bitbottle/commit/d16bb40f331390d4833e8dda7f53bdb9f428d1e6))

## [1.16.0](https://github.com/proggarapsody/bitbottle/compare/v1.15.0...v1.16.0) (2026-05-06)


### Features

* scope Q — Repo Extras (rename, fork) + set-default flip ([#83](https://github.com/proggarapsody/bitbottle/issues/83)) ([bab3c93](https://github.com/proggarapsody/bitbottle/commit/bab3c93a9cf861bf021af1354cc4514fbe5a2622))

## [1.15.0](https://github.com/proggarapsody/bitbottle/compare/v1.14.3...v1.15.0) (2026-05-06)


### Features

* scope H — Pipeline Depth (steps, logs, variables) + gh-style refactor ([#80](https://github.com/proggarapsody/bitbottle/issues/80)) ([6fffd61](https://github.com/proggarapsody/bitbottle/commit/6fffd612510d26d2d6dfe8bc5ed4b601fa91b21d))
* scope I — Webhooks (list, view, create, delete) on both backends ([#82](https://github.com/proggarapsody/bitbottle/issues/82)) ([48026e1](https://github.com/proggarapsody/bitbottle/commit/48026e11b05310631be4170e30d98eb78444de3b))

## [1.14.3](https://github.com/proggarapsody/bitbottle/compare/v1.14.2...v1.14.3) (2026-05-06)


### Bug Fixes

* repair 4 Cloud API bugs found during manual testing ([1f19891](https://github.com/proggarapsody/bitbottle/commit/1f19891e22d51de934c71d10cd5b1a822e5ec18c))

## [1.14.2](https://github.com/proggarapsody/bitbottle/compare/v1.14.1...v1.14.2) (2026-05-05)


### Bug Fixes

* **skill:** auto-bump version labels via release-please ([#76](https://github.com/proggarapsody/bitbottle/issues/76)) ([2881aaf](https://github.com/proggarapsody/bitbottle/commit/2881aaf21ce7313299b23c8ff1265d4decbd29c1))

## [1.14.1](https://github.com/proggarapsody/bitbottle/compare/v1.14.0...v1.14.1) (2026-05-05)


### Bug Fixes

* **skill:** audit round-3 + sync_help drift detector ([#74](https://github.com/proggarapsody/bitbottle/issues/74)) ([3a5f49c](https://github.com/proggarapsody/bitbottle/commit/3a5f49ca59dbd9db3258fa69ef0140d676be1523))

## [1.14.0](https://github.com/proggarapsody/bitbottle/compare/v1.13.1...v1.14.0) (2026-05-05)


### Features

* add `bitbottle skill` subcommand ([#72](https://github.com/proggarapsody/bitbottle/issues/72)) ([adbd131](https://github.com/proggarapsody/bitbottle/commit/adbd1311cfb042cf8af6487dfffa4b87d8916235))

## [1.13.1](https://github.com/proggarapsody/bitbottle/compare/v1.13.0...v1.13.1) (2026-05-05)


### Bug Fixes

* correct skill content for 1.13.0 + refresh on reinstall ([#69](https://github.com/proggarapsody/bitbottle/issues/69)) ([6ceb4dc](https://github.com/proggarapsody/bitbottle/commit/6ceb4dc597bc8f36f02d27ebc087875542b27018))

## [1.13.0](https://github.com/proggarapsody/bitbottle/compare/v1.12.0...v1.13.0) (2026-05-05)


### Features

* auto-register agent skill on npm postinstall ([#67](https://github.com/proggarapsody/bitbottle/issues/67)) ([105fe02](https://github.com/proggarapsody/bitbottle/commit/105fe02ee0eeda783d440318c41bf161e3b918de))

## [1.12.0](https://github.com/proggarapsody/bitbottle/compare/v1.11.0...v1.12.0) (2026-05-04)


### Features

* unified repo resolution via factory.ResolveTarget (gh-style) ([#63](https://github.com/proggarapsody/bitbottle/issues/63)) ([c2bf0f2](https://github.com/proggarapsody/bitbottle/commit/c2bf0f2eaca2c93c7c9de5538fd077bd9947a440))

## [1.11.0](https://github.com/proggarapsody/bitbottle/compare/v1.10.0...v1.11.0) (2026-05-02)


### Features

* **npm:** include README in published npm bundle ([#59](https://github.com/proggarapsody/bitbottle/issues/59)) ([11e3264](https://github.com/proggarapsody/bitbottle/commit/11e32641b3068401ecdc03b9a0d37ab0e420252a))

## [1.10.0](https://github.com/proggarapsody/bitbottle/compare/v1.9.1...v1.10.0) (2026-05-02)


### Features

* add Claude skill for bitbottle CLI ([#57](https://github.com/proggarapsody/bitbottle/issues/57)) ([7bd5bb9](https://github.com/proggarapsody/bitbottle/commit/7bd5bb98231058c017b728392908d221e1761da7))

## [1.9.1](https://github.com/proggarapsody/bitbottle/compare/v1.9.0...v1.9.1) (2026-05-02)


### Bug Fixes

* **ci:** untrack dist/bitbottle and ignore /dist/ ([#55](https://github.com/proggarapsody/bitbottle/issues/55)) ([f8d07bb](https://github.com/proggarapsody/bitbottle/commit/f8d07bb3ea24d747ad503784fb334400fdd480e4))

## [1.9.0](https://github.com/proggarapsody/bitbottle/compare/v1.8.0...v1.9.0) (2026-05-01)


### Features

* add pager and ANSI color output (Scope T) ([#53](https://github.com/proggarapsody/bitbottle/issues/53)) ([b64f8f2](https://github.com/proggarapsody/bitbottle/commit/b64f8f27a2bb19c482e1e2d328928f915886de83))

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
