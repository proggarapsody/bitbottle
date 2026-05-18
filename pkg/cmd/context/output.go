package context

// Context is the orientation snapshot returned by `bitbottle context` and the
// MCP `get_context` tool. It collapses auth status, repo info, and git state
// into one structured response so callers can orient in a single round-trip.
//
// Ahead and Behind use *int with omitempty so that "unknown" is encoded as
// absent JSON keys, never as 0/0 — which would falsely signal "in sync".
// Both pointers are populated as a pair: either both non-nil, or both nil.
//
// This type lives in the cmd layer, not the backend domain, because it
// carries json tags purely for CLI --json / MCP output. No backend adapter
// returns or accepts it; it is assembled in Build().
type Context struct {
	Host          string      `json:"host"`
	Project       string      `json:"project"`
	Slug          string      `json:"slug"`
	Branch        string      `json:"branch"`
	DefaultBranch string      `json:"default_branch"`
	Ahead         *int        `json:"ahead,omitempty"`
	Behind        *int        `json:"behind,omitempty"`
	User          ContextUser `json:"user"`
	Backend       string      `json:"backend"`
}

// ContextUser is the user shape embedded in Context. It mirrors backend.User
// but carries json tags so the output contract is stable regardless of future
// backend.User reshaping.
type ContextUser struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
}
