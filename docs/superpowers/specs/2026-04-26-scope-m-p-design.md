# Scope M + P Design: Shell Completion & Auth Extras

**Date:** 2026-04-26
**Scopes:** M (Shell Completion), P (Auth Extras)
**Execution:** Sequential — M then P

---

## Philosophy

Both scopes are DX-tier with zero backend changes. They extend existing CLI
infrastructure only. Follow gh design exactly.

---

## Scope M — Shell Completion

### Goal

`bitbottle completion --shell bash|zsh|fish|powershell` — print shell completion
script to stdout. Mirrors `gh completion -s <shell>` exactly.

### Files

```
pkg/cmd/completion/completion.go       # new command
pkg/cmd/completion/completion_test.go  # tests
```

Wired into root command the same way every other subcommand is registered.

### Command

```
bitbottle completion --shell bash
bitbottle completion -s zsh
```

| Flag | Short | Required | Values |
|---|---|---|---|
| `--shell` | `-s` | yes | `bash`, `zsh`, `fish`, `powershell` |

`RunE` dispatches on the flag value:

```go
switch shell {
case "bash":
    return cmd.Root().GenBashCompletion(f.IOStreams().Out())
case "zsh":
    return cmd.Root().GenZshCompletion(f.IOStreams().Out())
case "fish":
    return cmd.Root().GenFishCompletion(f.IOStreams().Out(), true)
case "powershell":
    return cmd.Root().GenPowerShellCompletionWithDesc(f.IOStreams().Out())
default:
    return fmt.Errorf("unsupported shell %q: must be bash, zsh, fish, or powershell", shell)
}
```

No backend, no factory call beyond `f.IOStreams()`, no MCP tool.

### Tests

- Each shell variant produces non-empty output.
- Unknown shell value returns a non-nil error containing the shell name.

---

## Scope P — Auth Extras

### Goal

Add `auth token` and `auth refresh` to the existing `pkg/cmd/auth/` package.
Mirrors `gh auth token` and `gh auth refresh` behavior for a PAT-based CLI.

### Files

```
pkg/cmd/auth/token.go        # new
pkg/cmd/auth/token_test.go   # new
pkg/cmd/auth/refresh.go      # new
pkg/cmd/auth/refresh_test.go # new
```

Both commands wired into the existing `auth` cobra group alongside `login`,
`logout`, `status`.

### `auth token`

Reads `HostConfig.OAuthToken` from config and prints it to stdout. Exits 1 if
no token is stored for the resolved host.

```bash
bitbottle auth token
bitbottle auth token --hostname bitbucket.org
```

Output: raw token string, terminated by newline. No label. Matches `gh auth token`.

### `auth refresh`

Calls `client.GetCurrentUser()` against the resolved host. On success, if the
returned username differs from the stored `HostConfig.User`, updates the stored
value and calls `cfg.Save()`. Prints confirmation to stdout.

On any API error (including 401):

```
error: token validation failed — run 'bitbottle auth login --hostname HOST' to re-authenticate
```

Printed to stderr, exits 1. No interactive re-auth (PATs cannot be refreshed
programmatically; the user must generate a new token from the Bitbucket UI).

```bash
bitbottle auth refresh
bitbottle auth refresh --hostname bitbucket.example.com
```

No new backend interfaces. No MCP tools.

### Tests

**`auth token`:**
- Fake config with stored token → verify token printed to stdout.
- No token stored → verify exit 1 and error message.

**`auth refresh`:**
- `FakeClient.GetCurrentUserFn` returns a user → verify stored `User` updated
  in config and `cfg.Save()` called.
- `FakeClient.GetCurrentUserFn` returns error → verify stderr message contains
  host and exit 1.

---

## Definition of Done

- [ ] `go test ./pkg/cmd/completion/... ./pkg/cmd/auth/...` green
- [ ] `go test ./... -race` green
- [ ] README auth section updated with `auth token` and `auth refresh` examples
- [ ] BACKLOG.md marks M and P as ✅
