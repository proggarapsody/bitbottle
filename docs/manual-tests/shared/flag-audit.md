# Flag / Hint Audit

Systematic cross-check of every flag mentioned in `pkg/errfmt/errfmt.go` hints and
`pkg/cmd/**` error-message strings against actually-registered Cobra flags.

**Scope**: `pkg/errfmt/errfmt.go` catalogue entries + `fmt.Errorf` / `errors.New` calls
in `pkg/cmd/**` (excluding `*_test.go`). Run against commit where UX-FLAG-AUDIT shipped.

## Findings

| Source | Flag / Shorthand | Actually Wired? | Fix |
|--------|------------------|-----------------|-----|
| `pkg/errfmt/errfmt.go:93` — `CodeTransportTimeout` hint | `--debug` | ❌ no such flag registered anywhere (at UX-FLAG-AUDIT time) | Removed stale hint in UX-FLAG-AUDIT; wired real `--debug` persistent flag in DEBUG-TRANSPORT-FLAG (v1.78.0) |
| `pkg/errfmt/errfmt.go:89` — `CodeNetworkTLSUnknownAuthority` hint | `-k` / `--skip-tls-verify` | ✅ `PersistentFlags().BoolP("skip-tls-verify","k",…)` on root | TLS-VERIFY-GLOBAL (shipped v1.77.0) |
| `pkg/errfmt/errfmt.go:41` — `CodeAuthNoToken` hint | `--hostname` | ✅ registered on every subcommand + root | — |
| `pkg/errfmt/errfmt.go:45` — `CodeAuthInvalidToken` hint | `--hostname` | ✅ same | — |
| `pkg/cmd/auth/login.go:33` | `--hostname` | ✅ | — |
| `pkg/cmd/auth/login.go:46` | `--username` | ✅ | — |
| `pkg/cmd/webhook/delete/delete.go:51` | `--confirm` | ✅ | — |
| `pkg/cmd/webhook/create/create.go:67` | `--events` | ✅ | — |
| `pkg/cmd/commit/comment_edit.go:22` | `--body` | ✅ | — |
| `pkg/cmd/commit/comment_add.go:22` | `--body` | ✅ | — |
| `pkg/cmd/commit/status_report.go:31` | `--state` | ✅ | — |
| `pkg/cmd/commit/comment_react.go:23` | `--emoji` | ✅ | — |
| `pkg/cmd/commit/comment_unreact.go:23` | `--emoji` | ✅ | — |
| `pkg/cmd/pipeline/trigger/trigger.go:58` | `--branch` | ✅ | — |
| `pkg/cmd/pipeline/trigger/trigger.go:109` | `--variable` | ✅ | — |
| `pkg/cmd/branch/protect/create.go:70` | `--type` | ✅ | — |
| `pkg/cmd/branch/protect/create.go:73` | `--branch` / `--pattern` | ✅ | — |
| `pkg/cmd/extension/upgrade.go:39` | `--all` | ✅ | — |
| `pkg/cmd/extension/install.go:35` | `--local` | ✅ | — |
| `pkg/cmd/variable/delete/delete.go:57` | `--confirm` | ✅ | — |
| `pkg/cmd/variable/shared/ops.go:56` | `--env` / `--scope` | ✅ | — |
| `pkg/cmd/variable/set/set.go:115` | `--body` | ✅ | — |

## Summary

One finding: `--debug` was referenced in the `CodeTransportTimeout` error hint but had
never been registered as a flag. The stale hint was removed in UX-FLAG-AUDIT (v1.77.1),
then a real `--debug` persistent flag that logs HTTP method/URL/status/latency to stderr
was wired in DEBUG-TRANSPORT-FLAG (v1.78.0), and the hint was restored.

All other hinted flags verified as correctly wired at the point of this audit.
