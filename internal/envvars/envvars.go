// Package envvars is the single source of truth for environment variables that
// influence bitbottle's behavior. Each constant is the actual variable name
// (suitable for direct use in os.Getenv), and the comment above it documents
// what consumers do with the value.
//
// New env vars MUST be declared here and used via the constant — not as a
// string literal scattered through the code. This keeps the public contract
// (which the user can rely on across releases) reviewable in one file and
// future-proofs scripts and extensions that lean on these names.
package envvars

const (
	// Token is the personal access token used for API requests, overriding
	// any token stored in hosts.yml or the OS keyring. Useful in CI.
	// Format: opaque string. Backend-specific (Cloud app password vs.
	// Server PAT) is determined by the resolved host.
	Token = "BB_TOKEN"

	// Host overrides the default hostname. Equivalent to passing
	// --hostname=<value> to every command. Useful when multiple hosts are
	// configured and a single environment is dedicated to one of them.
	Host = "BB_HOST"

	// Repo overrides the resolved [HOST/]PROJECT/REPO. Equivalent to
	// --repo on commands that accept it. Already in use by the -R/--repo
	// flag — see pkg/cmd/factory/repo_override.go.
	Repo = "BB_REPO"

	// Editor overrides the editor invoked for interactive prompts (e.g.
	// `pr edit`). Falls back to `editor` from config.yml, then $EDITOR,
	// then "vi". Bitbottle-specific name takes precedence so users can
	// scope a different editor to bitbottle without touching shell config.
	Editor = "BB_EDITOR"

	// Pager overrides the pager used for paginated TTY output. Falls back
	// to `pager` from config.yml, then $PAGER. Set to empty string to
	// disable paging.
	Pager = "BB_PAGER"

	// Browser overrides the browser launched by `--web` flags. Falls back
	// to `browser` from config.yml, then platform default.
	Browser = "BB_BROWSER"

	// ForceTTY forces TTY-style (aligned, colored) output even when stdout
	// is a pipe. Set to "1", "true", or any non-empty value to enable.
	// Mirrors gh's GH_FORCE_TTY.
	ForceTTY = "BB_FORCE_TTY"

	// PromptDisabled, when non-empty, suppresses every interactive prompt
	// and forces commands to fail rather than block. Required for use in
	// non-interactive scripts.
	PromptDisabled = "BB_PROMPT_DISABLED"

	// ConfigDir overrides the bitbottle config directory (default
	// $XDG_CONFIG_HOME/bitbottle). Affects hosts.yml, config.yml, and
	// aliases.yml.
	ConfigDir = "BB_CONFIG_DIR"

	// NoColor disables colored output. Standard cross-tool convention; we
	// honor it instead of inventing our own variable.
	// Reference: https://no-color.org
	NoColor = "NO_COLOR"

	// XDGConfigHome is the platform-standard config root. We respect it
	// when BB_CONFIG_DIR is unset.
	XDGConfigHome = "XDG_CONFIG_HOME"

	// UpdateGolden, when set to "1" or "true", causes test golden files to
	// be rewritten in-place from the current output. Test-only.
	UpdateGolden = "BITBOTTLE_UPDATE_GOLDEN"
)
