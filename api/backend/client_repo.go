package backend

// RepoLister lists repositories. ns is the workspace (Bitbucket Cloud) or
// project key (Bitbucket Server); pass "" for Server to list all repos.
type RepoLister interface {
	ListRepos(ns string, limit int) ([]Repository, error)
}

// RepoReader reads a single repository.
type RepoReader interface {
	GetRepo(ns, slug string) (Repository, error)
}

// RepoWriter creates a repository.
type RepoWriter interface {
	CreateRepo(ns string, in CreateRepoInput) (Repository, error)
}

// RepoDeleter deletes a repository.
type RepoDeleter interface {
	DeleteRepo(ns, slug string) error
}

// RepoRenamer renames a repository.
type RepoRenamer interface {
	RenameRepo(ns, slug, newName string) (Repository, error)
}

// RepoForker is implemented only by Bitbucket Cloud clients — Bitbucket
// Server / Data Center has no fork primitive in its REST API. Access via
// AsRepoForker so the Cloud-only constraint surfaces as a typed
// ErrUnsupportedOnHost error rather than a panic.
type RepoForker interface {
	ForkRepo(ns, slug string, in ForkRepoInput) (Repository, error)
}

// FeatureRepoFork names the repo-fork capability for typed-error reporting.
const FeatureRepoFork Feature = "repo-fork"

// AsRepoForker returns the RepoForker view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) if the backend at host has no fork primitive
// (Bitbucket Server / Data Center).
func AsRepoForker(c Client, host string) (RepoForker, error) {
	rf, ok := c.(RepoForker)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureRepoFork),
			Message: "repo fork is not supported on " + host + " (Bitbucket Cloud only)",
		}
	}
	return rf, nil
}
