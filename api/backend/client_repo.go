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

// RepoVisibilitySetter sets a repository's public/private visibility.
type RepoVisibilitySetter interface {
	SetRepoVisibility(ns, slug string, isPrivate bool) error
}

// RepoDefaultBranchSetter sets the default branch of a repository.
type RepoDefaultBranchSetter interface {
	SetRepoDefaultBranch(ns, slug, branch string) error
}

// SourceReader reads file content and directory listings at a ref. Both
// backends implement it (Cloud via /src/{ref}/{path}, Server via
// /raw/{path}?at={ref} and /browse/{path}?at={ref}).
//
// GetFileContent returns the raw bytes of a file at ref. When path resolves
// to a directory the backend returns a 404 (Server) or directory listing
// (Cloud) — adapters normalise both to ErrNotFound so the cmd layer can
// suggest ListTree.
//
// ListTree returns immediate children of path at ref. path "" lists the
// repository root. Each entry has Type "file" or "dir" — backends with
// richer type vocabularies (Server's "FILE"/"DIRECTORY"/"SUBMODULE",
// Cloud's "commit_file"/"commit_directory") are normalised at the adapter
// boundary. Submodules are surfaced as Type "dir" so the renderer treats
// them as recursable; the Hash field carries the submodule pointer.
type SourceReader interface {
	GetFileContent(ns, slug, ref, path string) ([]byte, error)
	ListTree(ns, slug, ref, path string) ([]TreeEntry, error)
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
	return requireFeature[RepoForker](c, host, specFor(FeatureRepoFork))
}

// RepoForksLister lists forks of a repository. Both Cloud and Server support this.
type RepoForksLister interface {
	ListRepoForks(ns, slug string, limit int) ([]Repository, error)
}

// FeatureRepoForks names the repo-forks-list capability for typed-error reporting.
const FeatureRepoForks Feature = "repo-forks"

// AsRepoForksLister returns the RepoForksLister view of c, or a typed *DomainError.
func AsRepoForksLister(c Client, host string) (RepoForksLister, error) {
	return requireFeature[RepoForksLister](c, host, specFor(FeatureRepoForks))
}
