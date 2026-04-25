package backend

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
