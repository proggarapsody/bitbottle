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

// Repository is the domain representation of a Bitbucket repository.
//
// ID is the backend's numeric identifier (Server / Data Center). Cloud
// uses opaque UUIDs surfaced through other channels; for Cloud repos
// this field stays zero. ID is exposed so the BBS default-reviewers
// endpoint (which requires source/target repo IDs as query params) has
// what it needs without an adapter-internal cache.
type Repository struct {
	Slug        string
	Name        string
	Namespace   string
	SCM         string
	WebURL      string
	Description string
	ID          int
}

// PullRequest is the domain representation of a Bitbucket pull request.
type PullRequest struct {
	ID             int
	Title          string
	Description    string
	State          string
	Draft          bool
	Author         User
	FromBranch     string
	ToBranch       string
	WebURL         string
	HeadCommitHash string
	AutoMerge      *AutoMergeState // nil when auto-merge is not enabled
}

// AutoMergeState records the auto-merge configuration for a pull request.
// Strategy is one of "merge", "squash", or "rebase" (the CLI vocabulary).
type AutoMergeState struct {
	Enabled  bool
	Strategy string // "merge" | "squash" | "rebase"
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

// MyPREntry is a cross-repo PR summary for the status dashboard.
type MyPREntry struct {
	PullRequest        // embed — carries ID, Title, State, Author, WebURL, HeadCommitHash
	Repo        string // "PROJECT/REPO" (Server) or "workspace/slug" (Cloud)
	Role        string // "AUTHOR" | "REVIEWER"
}

// User is the domain representation of a Bitbucket user.
type User struct {
	Slug        string
	DisplayName string
}

// CreateRepoInput carries the parameters for creating a repository.
type CreateRepoInput struct {
	Name        string
	SCM         string
	Public      bool
	Description string
}

// ForkRepoInput carries the parameters for forking a Bitbucket Cloud
// repository. Workspace is required (the fork's destination workspace).
// Name is optional — when empty, Bitbucket Cloud reuses the source name.
type ForkRepoInput struct {
	Workspace string
	Name      string
}

// CreatePRInput carries the parameters for creating a pull request.
//
// Reviewers are user slugs (Server) or usernames / account IDs (Cloud).
// Both adapters serialise the same logical list into their backend's wire
// shape. A nil or empty slice creates a PR with no reviewers — Bitbucket
// does NOT auto-apply repo "default reviewers" on the create endpoint, so
// callers that want that behaviour must fetch them via
// DefaultReviewersResolver and merge the result here.
type CreatePRInput struct {
	Title       string
	Description string
	Draft       bool
	FromBranch  string
	ToBranch    string
	Reviewers   []string
}

// MergePRInput carries the parameters for merging a pull request.
type MergePRInput struct {
	Message  string
	Strategy string
}

// Branch is the domain representation of a repository branch.
type Branch struct {
	Name       string
	IsDefault  bool
	LatestHash string
}

// Pipeline is the domain representation of a Bitbucket Cloud pipeline run.
type Pipeline struct {
	UUID        string
	BuildNumber int
	State       string // PENDING, IN_PROGRESS, SUCCESSFUL, FAILED, ERROR, STOPPED
	RefType     string // "branch", "tag", "commit"
	RefName     string
	CreatedOn   string
	Duration    int // seconds
	WebURL      string
}

// RunPipelineInput carries the parameters for triggering a pipeline run.
type RunPipelineInput struct {
	Branch string
}

// PipelineStep is the domain representation of a single step within a
// Bitbucket Cloud pipeline run.
type PipelineStep struct {
	UUID     string
	Name     string
	State    string // PENDING, IN_PROGRESS, SUCCESSFUL, FAILED, ERROR, STOPPED
	Result   string // populated when State has flattened from COMPLETED
	Duration int    // seconds
}

// PipelineVariable is a repository-level pipeline variable on Bitbucket Cloud.
// Value is empty when Secured is true (the API never returns secured values).
type PipelineVariable struct {
	UUID    string
	Key     string
	Value   string
	Secured bool
}

// PipelineVariableInput carries the parameters for upserting a pipeline
// variable by Key.
type PipelineVariableInput struct {
	Key     string
	Value   string
	Secured bool
}

// PipelineTriggerInput carries the parameters for triggering a pipeline via
// the PipelineTriggerClient interface. Variables supplements the per-run
// environment; each entry maps to a Bitbucket pipeline variable object.
type PipelineTriggerInput struct {
	Branch    string
	Variables []PipelineVariable
}

// PipelineTriggerResult is returned by TriggerPipeline on success.
type PipelineTriggerResult struct {
	UUID  string `json:"uuid"`
	State string `json:"state"`
	Link  string `json:"link"`
}

// PipelineSchedule is the domain representation of a Bitbucket Cloud pipeline
// schedule.
type PipelineSchedule struct {
	UUID           string `json:"uuid"`
	Enabled        bool   `json:"enabled"`
	CronExpression string `json:"cronExpression"`
	Branch         string `json:"branch"`
}

// PipelineScheduleInput carries the parameters for creating a pipeline
// schedule.
type PipelineScheduleInput struct {
	CronExpression string
	Branch         string
	Enabled        bool
}

// Tag is the domain representation of a repository tag.
type Tag struct {
	Name    string
	Hash    string
	Message string // empty for lightweight tags; first line for annotated
	WebURL  string
}

// CreateBranchInput carries the parameters for creating a branch.
type CreateBranchInput struct {
	Name    string
	StartAt string // branch name or commit hash
}

// CreateTagInput carries the parameters for creating a tag.
type CreateTagInput struct {
	Name    string
	StartAt string // branch name or commit hash
	Message string // empty = lightweight tag; non-empty = annotated tag
}

// UpdatePRInput carries the parameters for editing a pull request.
type UpdatePRInput struct {
	Title       string // empty = no change
	Description string // empty = no change
}

// Commit is the domain representation of a single repository commit.
type Commit struct {
	Hash      string
	Message   string // subject line only (first line of commit message)
	Author    User
	Timestamp time.Time
	WebURL    string
}

// CommentReaction is the domain representation of an emoji reaction on a
// pull-request comment. Emoji is the canonical shortcode (underscore form):
// "thumbs_up", "thumbs_down", "heart", "laugh", "hooray", "confused".
// Users lists every user who reacted with that emoji.
type CommentReaction struct {
	Emoji string // canonical shortcode: thumbs_up | thumbs_down | heart | laugh | hooray | confused
	Users []User // all users who reacted with this emoji
}

// PRComment is the domain representation of a comment on a pull request.
// Inline is non-nil for inline (file:line) review comments and nil for
// general PR comments. ParentID is 0 for top-level comments and the parent
// comment's ID for replies. Resolved reflects backend-native resolution
// state (Cloud `resolution`); on Server it is always false because the
// equivalent concept lives on tasks (RV3 scope).
//
// Severity, State, and Version are populated on Bitbucket Server / Data Center
// for task-like comments (comments with severity="BLOCKER"). Severity is
// "BLOCKER" for tasks and "" for regular comments. State is "OPEN" or
// "RESOLVED" for tasks and "" for regular comments. Version is the
// optimistic-lock token required by SetPRCommentState on Server.
//
// Reactions is only populated when explicitly requested (e.g. --reactions flag
// or include_reactions MCP parameter). Server/DC only.
type PRComment struct {
	ID        int
	Author    User
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Inline    *PRCommentInline
	ParentID  int
	Resolved  bool
	Severity  string            // "" | "BLOCKER" (Server task comments)
	State     string            // "" | "OPEN" | "RESOLVED" (Server task comments)
	Version   int               // optimistic-lock token (Server only)
	Reactions []CommentReaction // only populated when explicitly requested; Server/DC only
}

// PRCommentInline anchors a PR comment to a file and line range in the diff.
// Side is "old" (LHS / from-side) or "new" (RHS / to-side). StartLine is 0
// for single-line comments and set to the first line of the range for
// multi-line comments; Line is the last (or only) line.
type PRCommentInline struct {
	Path      string
	Side      string // "old" | "new"
	Line      int
	StartLine int // 0 = single-line
}

// AddPRCommentInput carries the parameters for adding a comment to a PR.
// Inline is non-nil for inline (file:line) review comments and nil for
// general PR comments. Parent is non-nil to post a reply nested under an
// existing comment thread.
//
// Severity controls the comment type on Bitbucket Server / Data Center:
// set to "BLOCKER" to create a task comment. Cloud ignores this field and
// always creates a regular comment.
type AddPRCommentInput struct {
	Text     string
	Inline   *PRCommentInline
	Parent   *int
	Severity string // "BLOCKER" to create a task; "" for normal comment (Server only)
}

// SubmitReviewInput bundles the review action, optional top-level body, and
// optional inline comments for a compound `pr review` call. The adapter
// posts the body comment first (if any), then each inline comment in order,
// then applies the action — so partial failures surface at the first
// failing step rather than corrupting the review state silently.
//
// Action is one of:
//   - "approve"          — calls ApprovePR after comments
//   - "request_changes"  — calls RequestChangesPR (Cloud only; Server returns
//     a typed *DomainError with Kind=ErrUnsupportedOnHost)
//   - "comment"          — comment-only review; no action call after the
//     comments are posted
type SubmitReviewInput struct {
	Action string
	Body   string
	Inline []SubmitReviewInline
}

// SubmitReviewInline is one inline comment in a compound review. The shape
// mirrors PRCommentInline + the comment body, so the adapter can splat each
// entry into AddPRComment without an intermediate translation.
type SubmitReviewInline struct {
	Path      string
	Line      int
	StartLine int    // 0 = single-line
	Side      string // "new" or "old"; defaults to "new" when empty
	Body      string
}

// CommitStatus is a build / CI status reported against a commit hash.
type CommitStatus struct {
	Key         string
	State       string // SUCCESSFUL | FAILED | INPROGRESS | STOPPED
	Name        string
	Description string
	URL         string
}

// Webhook is the domain representation of a repository webhook.
// Both Bitbucket Cloud and Server/DC expose a similar shape — a remote URL,
// a list of subscribed events, and an active flag. ID is the backend's
// stable identifier (UUID on Cloud, integer-as-string on Server/DC).
type Webhook struct {
	ID     string
	URL    string
	Events []string
	Active bool
}

// Workspace is a Bitbucket Cloud workspace (the top-level ownership unit
// containing repositories and projects). Bitbucket Server / Data Center has
// no analogous concept — projects sit directly under the instance — so
// Workspace is Cloud-only and surfaced via the WorkspaceClient optional
// interface.
type Workspace struct {
	Slug   string
	Name   string
	UUID   string
	WebURL string
}

// Project is a Bitbucket Cloud project (a logical group of repositories
// inside a workspace). The naming clashes with Server/DC's "project" — which
// is the namespace itself — but Cloud's project sits one level deeper.
// Listed via WorkspaceClient.ListProjects(workspace).
type Project struct {
	Key    string
	Name   string
	UUID   string
	WebURL string
}

// Issue is a Bitbucket Cloud issue. Issues are a Cloud-only feature gated
// by per-repository "issue tracker" enablement; the API returns 404 on
// repositories where the tracker is disabled. We surface that as the
// adapter's standard ErrNotFound.
//
// Assignee is a pointer so the zero value cleanly distinguishes
// "unassigned" from "assigned to user with empty username". Reporter is
// always present on Cloud and uses the value type.
type Issue struct {
	ID        int
	Title     string
	State     string // new, open, on hold, resolved, duplicate, invalid, wontfix, closed
	Kind      string // bug, enhancement, proposal, task
	Priority  string // trivial, minor, major, critical, blocker
	Reporter  User
	Assignee  *User // nil when unassigned
	CreatedOn time.Time
	UpdatedOn time.Time
	WebURL    string
	Content   string // raw markdown body
}

// CreateIssueInput carries the parameters for opening a new issue. Bitbucket
// Cloud applies sane defaults (kind=bug, priority=major) when fields are
// empty, so callers can omit everything but Title.
type CreateIssueInput struct {
	Title    string
	Content  string
	Kind     string
	Priority string
}

// UpdateIssueInput carries the parameters for changing an issue. Empty
// strings mean "no change". `issue close` sets State="closed" and leaves
// the rest untouched. Assignee is the Bitbucket Cloud username to assign to;
// set to the special sentinel AssigneeNone to explicitly clear the assignee.
type UpdateIssueInput struct {
	Title    string
	Content  string
	State    string
	Kind     string
	Priority string
	Assignee string // "" = no change; AssigneeNone = clear
}

// AssigneeNone is a sentinel value for UpdateIssueInput.Assignee that
// signals "unassign" (set assignee to null on the wire).
const AssigneeNone = "__none__"

// IssueComment is a comment on a Bitbucket Cloud issue.
type IssueComment struct {
	ID        int
	Author    User
	Content   string
	CreatedOn time.Time
	UpdatedOn time.Time
}

// CreateWebhookInput carries the parameters for creating a webhook.
// Secret is write-only — neither backend returns it on read.
type CreateWebhookInput struct {
	URL    string
	Events []string
	Active bool
	Secret string
}

// BranchProtection is a single branch-restriction record on Bitbucket
// Server / Data Center. Cloud has a similar concept under a different,
// non-trivial wire shape that we don't model yet — BranchProtection is
// surfaced only via the BranchProtector optional interface.
//
// Type values match BBS exactly (lowercased, dash-separated):
//
//	read-only           — disallow all writes
//	no-deletes          — disallow branch deletion
//	fast-forward-only   — disallow non-fast-forward writes
//	pull-request-only   — only PR merges may write
//
// MatcherKind is the BBS matcher type id ("BRANCH", "PATTERN",
// "MODEL_BRANCH", "MODEL_CATEGORY") and MatcherID is the corresponding
// value (a branch name, glob, or model id). Users / Groups list slugs
// that are exempted from the restriction.
type BranchProtection struct {
	ID          int
	Type        string
	MatcherID   string
	MatcherKind string
	Users       []string
	Groups      []string
}

// CreateBranchProtectionInput carries the parameters for adding a
// branch-restriction. All fields except Type and MatcherID are optional.
// MatcherKind defaults to "BRANCH" when empty (the most common case).
type CreateBranchProtectionInput struct {
	Type        string
	MatcherID   string
	MatcherKind string
	Users       []string
	Groups      []string
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

// TreeEntry is a single child of a directory listing under a ref. Type is
// the normalised "file" or "dir" — adapters fold their backend-specific
// vocabularies (Cloud "commit_file"/"commit_directory", Server
// "FILE"/"DIRECTORY"/"SUBMODULE") to one of these two values. Submodules
// surface as "dir" with the submodule pointer in Hash so renderers treat
// them as recursable.
//
// Path is the full repo-relative path (including the parent prefix), so a
// renderer doesn't need to know what path was requested to display the
// listing. Size is 0 for directories.
type TreeEntry struct {
	Path string
	Type string
	Size int64
	Hash string
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

// CodeInsightsReport is a Bitbucket Server / Data Center Code Insights report
// attached to a specific commit. Reports aggregate quality/security/build
// annotations and roll up to a single PASS/FAIL/PENDING result. This type is
// used for both API responses and the input shape (CodeInsightsReportInput).
//
// Result values: "PASS", "FAIL", "PENDING" (empty = PENDING).
// ReportType values: "TESTING", "COVERAGE", "BUG", "SECURITY", "DUPLICATION",
// "DEPENDENCY" (empty uses TESTING as default on Server).
type CodeInsightsReport struct {
	Key        string                    `json:"key"`
	Title      string                    `json:"title"`
	Result     string                    `json:"result"`
	ReportType string                    `json:"report_type,omitempty"`
	Details    string                    `json:"details,omitempty"`
	Reporter   string                    `json:"reporter,omitempty"`
	Link       string                    `json:"link,omitempty"`
	LogoURL    string                    `json:"logo_url,omitempty"`
	Data       []CodeInsightsReportDatum `json:"data,omitempty"`
	CreatedAt  string                    `json:"created_at,omitempty"`
	UpdatedAt  string                    `json:"updated_at,omitempty"`
}

// CodeInsightsReportDatum is a single key/value data point attached to a
// Code Insights report. Type values: "BOOLEAN", "DATE", "DURATION",
// "LINK", "NUMBER", "PERCENTAGE", "TEXT".
type CodeInsightsReportDatum struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// CodeInsightsReportInput is the upsert payload for SetReport. All fields map
// 1:1 to CodeInsightsReport; the separation exists only so the input shape
// stays explicit in the interface signature.
type CodeInsightsReportInput struct {
	Title      string
	Result     string // "PASS", "FAIL", "PENDING"
	ReportType string
	Details    string
	Reporter   string
	Link       string
	LogoURL    string
	Data       []CodeInsightsReportDatum
}

// CodeInsightsAnnotation is a single file/line annotation posted under a
// Code Insights report. This type serves as both the API response shape and
// the input payload (CodeInsightsAnnotationInput is an alias).
//
// Severity values: "LOW", "MEDIUM", "HIGH", "CRITICAL".
// Type values: "VULNERABILITY", "CODE_SMELL", "BUG".
type CodeInsightsAnnotation struct {
	ExternalID string `json:"external_id,omitempty"`
	Path       string `json:"path"`
	Line       int    `json:"line,omitempty"`
	Message    string `json:"message"`
	Severity   string `json:"severity,omitempty"`
	Type       string `json:"type,omitempty"`
	Link       string `json:"link,omitempty"`
}

// CodeInsightsAnnotationInput is an alias of CodeInsightsAnnotation used on
// the write side of the interface to keep the signature intent clear.
type CodeInsightsAnnotationInput = CodeInsightsAnnotation

// MergeCheck represents a Code Insights merge-check configuration. Merge
// checks block PR merges when report criteria are not met. This feature is
// partly undocumented in the Bitbucket Server REST API — treat output as
// best-effort.
//
// MinSeverity values: "LOW", "MEDIUM", "HIGH", "CRITICAL" (empty = no threshold).
type MergeCheck struct {
	Key         string `json:"key"`
	ReportKey   string `json:"report_key"`
	MustPass    bool   `json:"must_pass"`
	MinSeverity string `json:"min_severity,omitempty"`
}

// MergeCheckInput is an alias of MergeCheck used on the write side.
type MergeCheckInput = MergeCheck

// CommitComment is a comment attached to a specific commit.
// Reactions is only populated when explicitly requested (e.g. --reactions flag
// or include_reactions MCP parameter). Server/DC only.
type CommitComment struct {
	ID        int
	Author    User
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Reactions []CommentReaction // only populated when explicitly requested; Server/DC only
}

// AddCommitCommentInput carries parameters for creating a commit comment.
type AddCommitCommentInput struct {
	Body string
}

// PRActivityEvent is one event in the pull-request activity stream.
// Type is a normalised event kind: "approval", "unapproval", "comment",
// "update", "merge", "declined", "rescoped". Detail carries the raw
// sub-object so callers can surface backend-specific fields via --json.
type PRActivityEvent struct {
	Type      string
	Actor     User
	CreatedAt time.Time
	Detail    map[string]any
}

// Deployment is the domain representation of a Bitbucket Cloud deployment.
// State values: PENDING, IN_PROGRESS, COMPLETED, STOPPED, FAILED.
type Deployment struct {
	UUID        string
	State       string
	Environment Environment
	Release     struct {
		Name       string
		URL        string
		CommitHash string
	}
}

// Environment is a Bitbucket Cloud deployment environment (e.g. Test, Staging, Production).
// Rank is the numeric ordering position.
type Environment struct {
	UUID string
	Name string
	Type string // Test | Staging | Production
	Rank int
}

// CreateEnvironmentInput carries the parameters for creating a deployment environment.
type CreateEnvironmentInput struct {
	Name string
	Type string
	Rank int
}

// EnvVariable is a deployment-environment variable on Bitbucket Cloud.
// Value is empty when Secured is true (the API never returns secured values).
type EnvVariable struct {
	UUID    string
	Key     string
	Value   string // empty if Secured
	Secured bool
}

// EnvVariableInput carries the parameters for creating or updating an
// environment variable.
type EnvVariableInput struct {
	Key     string
	Value   string
	Secured bool
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

// PermissionGrant pairs a subject (user or group) with a permission level.
// Permission values:
//
//	PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN
//	REPO_READ, REPO_WRITE, REPO_ADMIN
type PermissionGrant struct {
	Subject    PermissionSubject
	Permission string
}

// DefaultReviewer is the domain representation of a repository default reviewer.
type DefaultReviewer struct {
	UserSlug     string `json:"userSlug"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// DeployKey is the domain representation of a repository deploy key (SSH public key).
type DeployKey struct {
	ID       int    `json:"id"`
	Label    string `json:"label"`
	Key      string `json:"key"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

// DeployKeyInput carries the parameters for adding a deploy key.
type DeployKeyInput struct {
	Label string
	Key   string
}

// BranchRule is the domain representation of a Bitbucket Cloud branch restriction rule.
type BranchRule struct {
	ID      int    `json:"id"`
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Value   int    `json:"value,omitempty"`
}

// BranchRuleInput carries the parameters for adding a branch restriction rule.
type BranchRuleInput struct {
	Kind    string
	Pattern string
	Value   int
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

// DiffStat is the domain representation of a repository diff summary.
type DiffStat struct {
	FilesChanged int
	Additions    int
	Deletions    int
	Files        []DiffStatEntry
}

// DiffStatEntry is the domain representation of a single file's diff summary.
type DiffStatEntry struct {
	Path      string
	Status    string // "added", "modified", "deleted", "renamed"
	Additions int
	Deletions int
}
