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
type Repository struct {
	Slug      string
	Name      string
	Namespace string
	SCM       string
	WebURL    string
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

// CreatePRInput carries the parameters for creating a pull request.
type CreatePRInput struct {
	Title       string
	Description string
	Draft       bool
	FromBranch  string
	ToBranch    string
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
