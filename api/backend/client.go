package backend

import "fmt"

// RepoLister can list repositories.
type RepoLister interface {
	ListRepos(limit int) ([]Repository, error)
}

// RepoReader can fetch a single repository.
type RepoReader interface {
	GetRepo(ns, slug string) (Repository, error)
}

// RepoWriter can create repositories.
type RepoWriter interface {
	CreateRepo(ns string, in CreateRepoInput) (Repository, error)
}

// RepoDeleter can delete repositories.
type RepoDeleter interface {
	DeleteRepo(ns, slug string) error
}

// PRLister can list pull requests.
type PRLister interface {
	ListPRs(ns, slug, state string, limit int) ([]PullRequest, error)
}

// PRReader can fetch a single pull request.
type PRReader interface {
	GetPR(ns, slug string, id int) (PullRequest, error)
}

// PRCreator can create pull requests.
type PRCreator interface {
	CreatePR(ns, slug string, in CreatePRInput) (PullRequest, error)
}

// PRMerger can merge pull requests.
type PRMerger interface {
	MergePR(ns, slug string, id int, in MergePRInput) (PullRequest, error)
}

// PRApprover can approve pull requests.
type PRApprover interface {
	ApprovePR(ns, slug string, id int) error
}

// PRDiffer can fetch pull request diffs.
type PRDiffer interface {
	GetPRDiff(ns, slug string, id int) (string, error)
}

// BranchLister can list branches.
type BranchLister interface {
	ListBranches(ns, slug string, limit int) ([]Branch, error)
}

// BranchDeleter can delete branches.
type BranchDeleter interface {
	DeleteBranch(ns, slug, branch string) error
}

// UserGetter can retrieve the currently authenticated user.
type UserGetter interface {
	GetCurrentUser() (User, error)
}

// Client is the full Bitbucket backend interface.
type Client interface {
	RepoLister
	RepoReader
	RepoWriter
	RepoDeleter
	PRLister
	PRReader
	PRCreator
	PRMerger
	PRApprover
	PRDiffer
	BranchLister
	BranchDeleter
	UserGetter
}

// ServerCapabilities is implemented only by Bitbucket Data Center clients.
type ServerCapabilities interface {
	GetApplicationProperties() (AppProperties, error)
}

// PipelineClient is implemented only by Bitbucket Cloud clients.
type PipelineClient interface {
	ListPipelines(ns, slug string, limit int) ([]Pipeline, error)
	GetPipeline(ns, slug, uuid string) (Pipeline, error)
	RunPipeline(ns, slug string, in RunPipelineInput) (Pipeline, error)
}

// AsPipelineClient returns the PipelineClient view of c, or an error if the
// backend does not support pipelines (i.e. is not Bitbucket Cloud).
func AsPipelineClient(c Client) (PipelineClient, error) {
	pc, ok := c.(PipelineClient)
	if !ok {
		return nil, fmt.Errorf("pipelines are only supported on Bitbucket Cloud")
	}
	return pc, nil
}
