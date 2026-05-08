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
	ID          int
	Title       string
	Description string
	State       string
	Draft       bool
	Author      User
	FromBranch  string
	ToBranch    string
	WebURL      string
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

// AppProperties holds Bitbucket server version metadata.
type AppProperties struct {
	Version     string
	BuildNumber string
	DisplayName string
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

// PRComment is the domain representation of a general comment on a pull request.
type PRComment struct {
	ID        int
	Author    User
	Text      string
	CreatedAt time.Time
}

// AddPRCommentInput carries the parameters for adding a comment to a PR.
type AddPRCommentInput struct {
	Text string
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
// the rest untouched.
type UpdateIssueInput struct {
	Title    string
	Content  string
	State    string
	Kind     string
	Priority string
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
