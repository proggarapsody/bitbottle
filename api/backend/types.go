package backend

// Options overrides the stored config when constructing a backend client.
// Used by auth login to validate a new token before it is persisted.
type Options struct {
	Token         string
	SkipTLSVerify bool
	// Email is the Atlassian account email address used as the HTTP Basic auth
	// identity for Bitbucket Cloud when authenticating with an Atlassian API
	// token. Leave empty when using a Bearer / OAuth2 token.
	Email string
	// Username is the Bitbucket Server/DC username used as the HTTP Basic auth
	// identity. Not used for Bitbucket Cloud.
	Username string
}

// User is the domain representation of a Bitbucket user.
type User struct {
	Slug        string
	DisplayName string
}

// CommentReaction is the domain representation of an emoji reaction on a
// pull-request comment. Emoji is the canonical shortcode (underscore form):
// "thumbs_up", "thumbs_down", "heart", "laugh", "hooray", "confused".
// Users lists every user who reacted with that emoji.
type CommentReaction struct {
	Emoji string // canonical shortcode: thumbs_up | thumbs_down | heart | laugh | hooray | confused
	Users []User // all users who reacted with this emoji
}

// Context is the one-call orientation primitive returned by `bitbottle
// context` and the MCP `get_context` tool. It collapses three previously
// independent calls (auth status / repo view / git status) into a single
// structured response so AI agents can orient themselves in one round-trip.
//
// Zero-valued Project / Slug / Branch / DefaultBranch indicate "outside a
// git repo" — the rest of the shape (Host, User, Backend) still resolves
// through config + the backend's current-user endpoint.
//
// Ahead and Behind are *int with omitempty so that "unknown" (git failed,
// no upstream, base ref missing, outside a repo) is encoded as the keys
// being absent from JSON — never as 0/0, which would lie to agents that
// would otherwise conclude "in sync". Both pointers are populated as a
// pair: either both non-nil, or both nil.
//
// Backend is the literal "cloud" or "server" string matching the
// bbinstance backend-type vocabulary, so consumers can branch on
// host shape without re-deriving it.
//
// User is a Context-local shape with JSON tags so the public contract
// emits {"slug": ..., "display_name": ...} without bleeding tags onto
// backend.User (which is reused across many wire surfaces and would
// regress existing JSON outputs if tagged here).
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

// ContextUser is the user shape carried inside Context. It mirrors User but
// stamps explicit JSON tags so the documented contract is stable
// (`{"slug": ..., "display_name": ...}`) regardless of how backend.User
// is later reshaped.
type ContextUser struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
}

// CodeSearchHit is one result row from Bitbucket Cloud's workspace-scoped
// code search. The hit may match on the file path (PathMatches non-empty),
// on file content (ContentMatches non-empty), or both. Renderers use the
// SearchSegment.Match flag to bold the matched runs.
//
// Repository is "workspace/slug" — the Cloud "full_name" form, kept verbatim
// so JSON consumers can split it themselves and the table renderer needs no
// extra column.
type CodeSearchHit struct {
	Repository        string
	Path              string
	PathMatches       []SearchSegment
	ContentMatches    []ContentMatch
	ContentMatchCount int
	FileURL           string
}

// ContentMatch is a single matched line within a file: the 1-based line
// number plus a sequence of segments. Cloud groups consecutive matched
// lines into "content_matches" objects each with a "lines" array; bitbottle
// flattens those to a single ContentMatch slice in arrival order.
type ContentMatch struct {
	Line     int
	Segments []SearchSegment
}

// SearchSegment is one run of text in a path or content match. Match=true
// marks the substring that triggered the hit so renderers can highlight it.
type SearchSegment struct {
	Text  string
	Match bool
}

// PermissionSubject identifies a user or group in a permission grant.
// Kind is "user" or "group". For users, Slug is the login slug. For groups,
// Name is the group name (may contain spaces). DisplayName is populated on
// read from the API and ignored on write.
type PermissionSubject struct {
	Kind        string // "user" | "group"
	Slug        string // user slug (Kind=user)
	Name        string // group name (Kind=group)
	DisplayName string // populated on read; ignored on write
}

// PermissionGrant pairs a subject (user or group) with a permission level.
// Permission values:
//
//	PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN
//	REPO_READ, REPO_WRITE, REPO_ADMIN
type PermissionGrant struct {
	Subject    PermissionSubject
	Permission string
}

// LoggingConfig is the domain representation of the Bitbucket Server / Data
// Center logging configuration. Level is one of DEBUG, INFO, WARN, ERROR.
// Async controls whether log events are written asynchronously.
// Persistent, when true, directs SetLoggingConfig to write to
// log4j.properties so the change survives restarts; false = runtime-only.
type LoggingConfig struct {
	Level      string // DEBUG | INFO | WARN | ERROR
	Async      bool
	Persistent bool // write-only flag; ignored in GetLoggingConfig response
}

// LoggingConfigInput is the upsert payload for SetLoggingConfig. It mirrors
// LoggingConfig; the alias exists so the interface signature stays explicit.
type LoggingConfigInput = LoggingConfig

// SSHKey is the domain representation of a user SSH key on Bitbucket Cloud.
type SSHKey struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Key   string `json:"key"`
}

// SSHKeyInput carries the parameters for adding a user SSH key.
type SSHKeyInput struct {
	Label string
	Key   string
}

// ToCloudMergeStrategy translates CLI strategy names to Bitbucket Cloud API
// values used on the auto-merge endpoint.
func ToCloudMergeStrategy(s string) string {
	switch s {
	case "squash":
		return "squash"
	case "rebase":
		return "fast_forward"
	default:
		return "merge_commit"
	}
}

// ToServerMergeStrategy translates CLI strategy names to Bitbucket Server /
// Data Center API values used on the auto-merge endpoint.
func ToServerMergeStrategy(s string) string {
	switch s {
	case "squash":
		return "squash"
	case "rebase":
		return "fast-forward"
	default:
		return "merge-commit"
	}
}
