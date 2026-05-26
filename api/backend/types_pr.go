package backend

import "time"

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
	// Version is the optimistic-concurrency token used by Bitbucket Server / DC.
	// It is zero for Cloud PRs (Cloud does not expose a version field).
	// Use omitempty so Cloud JSON output does not include a spurious "version":0.
	Version int `json:"version,omitempty"`
}

// AutoMergeState records the auto-merge configuration for a pull request.
// Strategy is one of "merge", "squash", or "rebase" (the CLI vocabulary).
type AutoMergeState struct {
	Enabled  bool
	Strategy string // "merge" | "squash" | "rebase"
}

// MyPREntry is a cross-repo PR summary for the status dashboard.
type MyPREntry struct {
	PullRequest        // embed — carries ID, Title, State, Author, WebURL, HeadCommitHash
	Repo        string // "PROJECT/REPO" (Server) or "workspace/slug" (Cloud)
	Role        string // "AUTHOR" | "REVIEWER"
}

// PRParticipant is a user involved in a pull request.
type PRParticipant struct {
	User     User
	Role     string // AUTHOR | REVIEWER | PARTICIPANT
	Approved bool
	State    string // APPROVED | CHANGES_REQUESTED | "" (empty = unapproved)
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

// UpdatePRInput carries the parameters for editing a pull request.
type UpdatePRInput struct {
	Title       string // empty = no change
	Description string // empty = no change
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

// DefaultReviewer is the domain representation of a repository default reviewer.
type DefaultReviewer struct {
	UserSlug     string `json:"userSlug"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// ReviewerGroup is the domain representation of a Bitbucket Server / Data
// Center default-reviewers condition. Each condition defines a set of
// reviewers that are automatically required for PRs matching a source/target
// ref pattern.
type ReviewerGroup struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	RequiredApprovals int    `json:"requiredApprovals"`
	Reviewers         []User `json:"reviewers"`
}

// CreateReviewerGroupInput carries the parameters for creating a reviewer group
// condition.
type CreateReviewerGroupInput struct {
	Name              string   // used as sourceMatcher displayId / id
	UserSlugs         []string // reviewer user slugs
	RequiredApprovals int      // default 1
}

// MergeDryRunResult is the result of a dry-run merge check on a pull request.
// CanMerge is true when the PR can be merged without conflicts.
// Message is a human-readable summary from the backend (may be empty).
// ConflictedFiles lists file paths that have merge conflicts (Cloud only;
// empty on Server because the Server dry-run endpoint does not enumerate
// conflicted files — vetoes carry the conflict description instead).
// Vetoes lists blocking conditions returned by Bitbucket Server / Data Center.
type MergeDryRunResult struct {
	CanMerge        bool         `json:"can_merge"`
	Message         string       `json:"message,omitempty"`
	ConflictedFiles []string     `json:"conflicted_files,omitempty"`
	Vetoes          []MergeVeto  `json:"vetoes,omitempty"`
}

// MergeVeto is a single blocking condition returned by the Bitbucket Server
// dry-run merge endpoint. SummaryMessage is a short label; DetailMessage is
// the full explanation.
type MergeVeto struct {
	SummaryMessage string `json:"summary_message"`
	DetailMessage  string `json:"detail_message,omitempty"`
}
