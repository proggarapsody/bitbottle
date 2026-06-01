package backend

// CommitSearcher searches commits in a repository by message, author, or date.
type CommitSearcher interface {
	SearchCommits(ns, slug string, opts CommitSearchOpts) ([]Commit, error)
}

// CommitSearchOpts controls which commits SearchCommits returns.
type CommitSearchOpts struct {
	Query  string // message keyword filter (substring match)
	Author string // author slug/account_id
	Since  string // ISO 8601 date or commit SHA (inclusive start)
	Until  string // ISO 8601 date or commit SHA (inclusive end)
	Limit  int
}

// FeatureCommitSearch names the commit-search capability.
const FeatureCommitSearch Feature = "commit-search"

// AsCommitSearcher returns the CommitSearcher view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when the backend doesn't implement
// commit search.
func AsCommitSearcher(c Client, host string) (CommitSearcher, error) {
	return requireFeature[CommitSearcher](c, host, specFor(FeatureCommitSearch))
}
