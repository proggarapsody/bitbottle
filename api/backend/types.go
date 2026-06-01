package backend

import "time"

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
//
// AccountID, UUID, CreatedOn and HTMLURL are Cloud-stable machine-readable
// identifiers. They are empty on Bitbucket Server/DC (which has no equivalent)
// and on read paths that don't request them; the omitempty tags keep the wire
// shape clean. MCP's get_current_user marshals this struct directly, so the
// json tags here are the public contract: callers get a durable handle
// (account_id / uuid) that survives nickname/display-name changes, plus the
// profile URL under links.html.href to mirror Bitbucket Cloud's own shape.
type User struct {
	Slug        string     `json:"slug,omitempty"`
	DisplayName string     `json:"name,omitempty"`
	AccountID   string     `json:"account_id,omitempty"`
	UUID        string     `json:"uuid,omitempty"`
	CreatedOn   string     `json:"created_on,omitempty"`
	HTMLURL     string     `json:"-"`
	Links       *UserLinks `json:"links,omitempty"`
}

// UserLinks mirrors the subset of Bitbucket Cloud's user "links" object that
// bitbottle exposes. Only the html.href profile URL is carried today.
type UserLinks struct {
	HTML *Link `json:"html,omitempty"`
}

// Link is a single {href} entry, the atom of Bitbucket Cloud's HAL-style
// links objects.
type Link struct {
	Href string `json:"href,omitempty"`
}

// CommentReaction is the domain representation of an emoji reaction on a
// pull-request comment. Emoji is the canonical shortcode (underscore form):
// "thumbs_up", "thumbs_down", "heart", "laugh", "hooray", "confused".
// Users lists every user who reacted with that emoji.
type CommentReaction struct {
	Emoji string // canonical shortcode: thumbs_up | thumbs_down | heart | laugh | hooray | confused
	Users []User // all users who reacted with this emoji
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

// RateLimitConfig holds Bitbucket Server/DC REST API rate-limit settings.
type RateLimitConfig struct {
	Enabled         bool `json:"enabled"`
	RequestsPerHour int  `json:"requests_per_hour"`
	ThrottleWaitMS  int  `json:"throttle_wait_ms"`
}

// PipelineConfig is the domain representation of a repository's Bitbucket Cloud
// pipeline configuration.
type PipelineConfig struct {
	Enabled bool `json:"enabled"`
}

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

// RepoDownload is the domain representation of a Bitbucket Cloud repository
// download artifact.
type RepoDownload struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Downloads int       `json:"downloads"`
	CreatedOn time.Time `json:"created_on"`
}

// Milestone is the domain representation of a Bitbucket Cloud issue milestone.
type Milestone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IssueVersion is the domain representation of a Bitbucket Cloud issue version.
type IssueVersion struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WorkspaceProject is the domain representation of a Bitbucket Cloud workspace project.
type WorkspaceProject struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
}

// CreateWorkspaceProjectInput carries the parameters for creating a workspace project.
type CreateWorkspaceProjectInput struct {
	Key         string
	Name        string
	Description string
	IsPrivate   bool
}

// UpdateWorkspaceProjectInput carries the parameters for updating a workspace project.
// Nil pointer fields are left unchanged.
type UpdateWorkspaceProjectInput struct {
	Name        *string
	Description *string
	IsPrivate   *bool
}

// MirrorServer represents a Bitbucket Smart Mirror server (Server/DC only).
type MirrorServer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Enabled bool   `json:"enabled"`
}

// MirroredRepo represents a repository mirrored by a Smart Mirror server.
type MirroredRepo struct {
	Slug       string    `json:"slug"`
	MirrorID   string    `json:"mirror_id"`
	LastSyncAt time.Time `json:"last_sync_at"`
	Status     string    `json:"status"`
}

// WorkspaceMemberPerm represents a workspace-level permission for a user
// (Bitbucket Cloud only).
type WorkspaceMemberPerm struct {
	User       string `json:"user"`
	Permission string `json:"permission"`
}

// WorkspaceRepoPerm represents a repository-level workspace permission for a
// user (Bitbucket Cloud only).
type WorkspaceRepoPerm struct {
	Repo       string `json:"repo"`
	User       string `json:"user"`
	Permission string `json:"permission"`
}

// IssueChange represents one change event in a Cloud issue's audit trail.
// Kind is the field that changed: "status", "priority", "assignee",
// "component", "milestone", "title", "content", or "comment".
type IssueChange struct {
	ID        int       `json:"id"`
	Kind      string    `json:"kind"`
	OldVal    string    `json:"old_val"`
	NewVal    string    `json:"new_val"`
	CreatedOn time.Time `json:"created_on"`
	User      User      `json:"user"`
}

// HostInfo describes the connected Bitbucket host: its backend type, base URL,
// optional version information (Server/DC only), and the Feature constants it
// supports. Used by `host info` and the `get_host_info` MCP tool as a
// single-call orientation primitive.
type HostInfo struct {
	BackendType       string   `json:"backend_type"` // "cloud" | "server" | "datacenter"
	BaseURL           string   `json:"base_url"`
	Version           string   `json:"version,omitempty"`      // Server/DC only
	BuildNumber       string   `json:"build_number,omitempty"` // Server/DC only
	DisplayName       string   `json:"display_name"`
	SupportedFeatures []string `json:"supported_features"`
}
