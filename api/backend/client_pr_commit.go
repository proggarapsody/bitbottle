package backend

// PRCommitClient is implemented by both Cloud and Server backends.
type PRCommitClient interface {
	ListPRCommits(ns, slug string, prID int) ([]Commit, error)
}

// FeaturePRCommits names the PR-commits capability for typed-error reporting.
const FeaturePRCommits Feature = "pr_commits"

// AsPRCommitClient returns the PRCommitClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the PRCommits capability.
func AsPRCommitClient(c Client, host string) (PRCommitClient, error) {
	return requireFeature[PRCommitClient](c, host, specFor(FeaturePRCommits))
}
