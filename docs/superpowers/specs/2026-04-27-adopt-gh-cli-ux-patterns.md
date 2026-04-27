# Adopt gh CLI UX patterns: BaseRepo factory, -R/--repo flag, ARGUMENTS help section

## Problem Statement

Bitbottle commands have inconsistent and rough UX in three places that gh CLI solved long ago:

1. **`--help` doesn't document positional arguments.** `bitbottle pr list --help` shows `bitbottle pr list [PROJECT/REPO]` in the usage line but no flag/section explains what `PROJECT/REPO` means or that it's optional with auto-detection.
2. **Errors leak raw exec internals.** Outside a git repo, `bitbottle pr list` returns `could not detect repo: exit status 128; pass PROJECT/REPO as an argument` — the `exit status 128` is meaningless to users.
3. **Repo resolution logic is duplicated** across every command (`pr list`, `pr view`, `commit view`, `repo view`, …). Each command re-implements: parse arg → run `git remote get-url origin` → infer host/project/slug → apply `--hostname` override → fall back to single-host config. New commands routinely diverge (e.g. `repo view` originally ignored `--hostname` entirely).
4. **No global `--repo` flag.** Users who aren't sitting inside a git checkout have no fluent way to target a repo across many commands; they must pass `PROJECT/REPO` as a positional everywhere it's accepted.

The pattern of small, similar bugs we keep finding (UUID slugs in author lines, empty `Web:` lines, port suffixes in hostnames, missing `--hostname`) is symptomatic of these missing shared abstractions.

## Solution

Port the four concrete pieces of gh CLI's command-construction toolkit into bitbottle:

1. **A custom root help function** that renders an `ARGUMENTS` section from a Cobra `Annotations["help:arguments"]` entry on each parent command, plus standard sections (USAGE, FLAGS, EXAMPLES, ENVIRONMENT VARIABLES, LEARN MORE).
2. **A `Factory.BaseRepo func() (bbrepo.RepoRef, error)`** that centralises arg-parse / git-remote-detect / single-host-fallback once, replacing the duplicated `resolveRepoRef`/`resolvePRTarget` helpers in every command.
3. **A `cmdutil.EnableRepoOverride(cmd, f)` helper** that registers a persistent `-R/--repo [HOST/]PROJECT/REPO` flag plus a `BB_REPO` env-var fallback, layered over `f.BaseRepo`.
4. **`Long:` and `Example:` heredocs** on every leaf command, plus angle-bracket `Use:` placeholders matching gh's idiom (`view [<project>/<repo>]`).

Friendly error messages drop on top: detect the "not a git repo" case (exit 128 from `git remote get-url`) and emit a clean `no git remotes found; pass [HOST/]PROJECT/REPO or use --repo` instead.

## User Stories

1. As a user inspecting a command, I want `bitbottle pr list --help` to include an `ARGUMENTS` section that describes `[PROJECT/REPO]` and that it's optional, so that I don't have to read source code to use the CLI.
2. As a user inspecting a command, I want each leaf command's `--help` to include `EXAMPLES` showing common invocations, so that I can copy-paste a working command.
3. As a user running a command outside a git repo, I want a one-line, action-oriented error (no `exit status 128`), so that I immediately know how to fix the invocation.
4. As a user with no Bitbucket remote in the current directory, I want `bitbottle pr list -R MYPROJ/myrepo` to work, so that I don't have to `cd` into a checkout.
5. As a user, I want `BB_REPO=MYPROJ/myrepo` to be honoured by every repo-scoped command, so that I can persist a default for a shell session.
6. As a user with multiple Bitbucket hosts, I want `-R bb.example.com/MYPROJ/myrepo` to disambiguate without needing a separate `--hostname` flag, so that one flag does the job.
7. As a user, I want `--hostname` to keep working as today for backward compatibility, so that scripts don't break.
8. As a user inside a git repo with a Bitbucket remote, I want every repo-scoped command to auto-detect the repo without arguments, so that the common case stays terse.
9. As a contributor adding a new command, I want a single-line `opts.BaseRepo = f.BaseRepo` integration, so that I cannot accidentally re-introduce the per-command divergence bugs we keep fixing.
10. As a contributor, I want a `runF func(*XxxOptions) error` test seam on each command, so that I can unit-test command behaviour without spinning up an httptest server.
11. As a user, I want angle-bracket placeholders (`view [<project>/<repo>]`) in usage lines, so that the syntax matches gh and other modern CLIs I already know.
12. As a user, I want consistent error wording across commands when no host is configured (`run \`bitbottle auth login\` first`), so that the CLI feels coherent.
13. As a user running `bitbottle repo view --help`, I want to see that the argument is optional (it isn't today — `cobra.ExactArgs(1)`), so that I can use auto-detection like every sibling command.
14. As a user, I want the `ARGUMENTS` section content to be defined once on the parent command (`pr`, `repo`) and inherited by every leaf, so that documentation stays consistent.
15. As a maintainer reviewing a PR that adds a new command, I want the absence of `Long:`/`Example:` to be a review red-flag, so that we don't ship undocumented commands.
16. As a user, I want `bitbottle pr view 42` from inside a checkout and `bitbottle pr view 42 -R PROJ/repo` from outside to behave identically, so that there's no special case to remember.
17. As a user, I want shell completion of `-R` to suggest configured hosts and known remotes, so that I don't have to retype them.

## Implementation Decisions

**Modules to add or reshape:**

- **`pkg/cmd/root/help.go` (new)** — port a trimmed version of gh's `rootHelpFunc`. Drop GitHub-specific concepts (help topics, accessibility flags). Render sections: USAGE, ALIASES, COMMANDS, FLAGS, INHERITED FLAGS, JSON FIELDS, **ARGUMENTS**, EXAMPLES, ENVIRONMENT VARIABLES, LEARN MORE. Wired via `cmd.SetHelpFunc(rootHelpFunc)` on the root command. The `ARGUMENTS` section reads `Annotations["help:arguments"]` from the **command itself or any ancestor** (so `pr` can define it once for all `pr *` leaves).

- **`pkg/cmd/factory.Factory` — add `BaseRepo func() (bbrepo.RepoRef, error)`** alongside the existing `Backend`. Default implementation:
  1. If a `--repo`/`-R` override (or `BB_REPO`) was set, parse it (supports `[HOST/]PROJECT/REPO`).
  2. Else read `git remote get-url origin` via `f.GitRunner()`.
  3. Else if exactly one host is configured and `--repo` was set as bare `PROJECT/REPO`, use that host.
  4. Else error with a single-line, action-oriented message.

- **`pkg/cmdutil.EnableRepoOverride(cmd, f)` (new)** — registers a persistent `-R/--repo` flag (description ``Select another repository using the `[HOST/]PROJECT/REPO` format``) and a `PersistentPreRunE` that wraps `f.BaseRepo` with an override function. Honours `BB_REPO`. Called once per top-level group (`pr`, `repo`, `commit`, `branch`, `pipeline`, `tag`).

- **Each leaf command** — replace the duplicated `resolveRepoRef`/`resolvePRTarget` block with `ref, err := opts.BaseRepo()`. Add `Long:` (heredoc explaining the argument and auto-detection) and `Example:` (heredoc of common invocations). Switch `Use:` to angle-bracket placeholders.

- **Each parent command (`pr.go`, `repo.go`, `commit.go`, …)** — set `Annotations["help:arguments"]` describing the shared `[<project>/<repo>]` argument once. Call `cmdutil.EnableRepoOverride(cmd, f)`.

- **Existing `--hostname` flag** — keep for back-compat; document as superseded by `--repo`. When both are set, `--repo`'s host component wins.

**Error message style:**
- Detect git exit 128 specifically (no remotes / not a repo) and translate to `no git remotes found; pass [HOST/]PROJECT/REPO or use --repo`.
- Drop `fmt.Errorf("%w", execErr)` wrapping where the wrapped error is just `exit status N`.

**Out-of-scope refactors (deferred):**
- Splitting `pkg/cmd/pr/*.go` into `pkg/cmd/pr/<verb>/<verb>.go` subpackages. Note as a future migration; do not block this work on it. When migrating, adopt the `runF func(*XxxOptions) error` test seam at the same time.

## Testing Decisions

**What makes a good test here:** assert observable behaviour through `--help` output, command exit codes, and stdout/stderr text. Do not assert on Cobra internals (flag registration order, help template structure). Treat the `ARGUMENTS` section as a contract — a snapshot test on the exact text for a representative command is fine; per-flag white-box tests are not.

**Modules to test:**

- **`rootHelpFunc`** — golden-file test on `bitbottle pr list --help`, `bitbottle repo view --help`, `bitbottle pr --help`. Asserts presence of `ARGUMENTS`, `EXAMPLES`, and that the argument description is inherited from the parent annotation.
- **`Factory.BaseRepo`** — table-driven tests covering: explicit `--repo HOST/PROJ/REPO`, explicit `--repo PROJ/REPO` with single host, `BB_REPO` env var, git-remote auto-detection, no-git-repo error message exact text, multiple-hosts-bare-arg error message exact text. Use the existing `TestFactoryOpts` (`InitialConfig`, `GitRunner`) — no new test seam required.
- **`cmdutil.EnableRepoOverride`** — integration test that runs `pr list --repo X/Y` with a stub backend and asserts `X/Y` was passed to `ListPRs`.
- **Leaf commands** — replace the per-command "resolve repo" tests (currently scattered) with a single assertion that the command uses `f.BaseRepo`. Same pattern as the existing `TestRepoView_ExplicitHostname_UsesProvidedHost` smoke test.
- **Error message regression** — assert that `exit status 128` does not appear anywhere in command stderr when running outside a git repo.

**Prior art in the codebase:**
- `pkg/cmd/repo/view_smoke_test.go::TestRepoView_ExplicitHostname_UsesProvidedHost` — uses `httptest.NewTLSServer` + captured `BaseURL` to assert which host the command resolved to. Reuse this pattern for `-R` tests.
- `pkg/cmd/pr/diff_smoke_test.go::TestPRDiff_SendsAcceptTextPlain` — asserts an HTTP-level contract (the `Accept` header). Same shape works for asserting `BaseRepo` resolution.
- `factory.NewTestFactory` — already supports `InitialConfig`, `GitRunner`, `HTTPClient`, `BaseURL`. Add `RepoOverride string` and `EnvRepo string` fields here for the new tests.

## Out of Scope

- Migrating `pkg/cmd/pr/*.go` to per-command subpackages.
- Adopting gh's `runF` test seam universally (only add it where a new test would otherwise be awkward).
- Shell completion for `-R` against remote repository lists (only complete from configured hosts and the current directory's remotes).
- Replacing `--hostname` — it stays for back-compat.
- Reworking the `--json`/`--jq` rendering — `JSON FIELDS` section in the help template is a follow-up.
- Help-topics (`bitbottle help environment`, etc.) — not needed yet at this command-surface size.

## Further Notes

- The licence on cli/cli is MIT, so verbatim adoption of `rootHelpFunc` and `repo_override.go` is straightforward; preserve attribution in a header comment on each ported file.
- Source citations for the side-by-side analysis (paths in the gh repo): `pkg/cmd/root/help.go:90,168-194`, `pkg/cmdutil/repo_override.go:22-70`, `pkg/cmd/factory/default.go:39-168`, `pkg/cmd/factory/remote_resolver.go:28-95`, `pkg/cmd/pr/pr.go:27-47`, `pkg/cmd/pr/list/list.go:57-110`, `pkg/cmd/pr/view/view.go:47-75`, `pkg/cmd/repo/view/view.go:45-67`.
- Suggested PR sequencing (one PR each, mergeable independently):
  1. Add `rootHelpFunc` + wire `Annotations["help:arguments"]` on parents. No behaviour change, only `--help` output.
  2. Add `f.BaseRepo` + migrate one command (`pr list`) as the canary.
  3. Add `EnableRepoOverride` + wire on `pr`, `repo`. Other groups follow.
  4. Migrate remaining commands off `resolveRepoRef`/`resolvePRTarget`.
  5. Friendlier error messages (small, isolated PR).
- After step 4, `internal/bbrepo` becomes the only place that touches git remotes. The class of bug we keep fixing in this codebase (one command's resolution differs from another's) becomes structurally impossible.
