package backend

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
	BranchDeleter
	UserGetter
}

// ServerCapabilities is implemented only by Bitbucket Data Center clients.
type ServerCapabilities interface {
	GetApplicationProperties() (AppProperties, error)
}
