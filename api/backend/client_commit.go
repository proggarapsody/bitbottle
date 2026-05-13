package backend

// CommitLister lists commits in a repository.
type CommitLister interface {
	ListCommits(ns, slug, branch string, limit int) ([]Commit, error)
}

// CommitReader reads a single commit by hash.
type CommitReader interface {
	GetCommit(ns, slug, hash string) (Commit, error)
}

// CommitStatusLister lists build/CI statuses reported against a commit hash.
type CommitStatusLister interface {
	ListCommitStatuses(ns, slug, hash string) ([]CommitStatus, error)
}

// CommitStatusReporter posts a build status against a commit hash.
type CommitStatusReporter interface {
	ReportCommitStatus(ns, slug, hash string, input CommitStatusInput) (CommitStatus, error)
}
