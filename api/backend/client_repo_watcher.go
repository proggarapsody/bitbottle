package backend

// RepoWatcherClient is implemented by both Cloud and Server backends.
type RepoWatcherClient interface {
	ListRepoWatchers(ns, slug string) ([]User, error)
}

// FeatureRepoWatchers names the repo-watchers capability for typed-error reporting.
const FeatureRepoWatchers Feature = "repo_watchers"

// AsRepoWatcherClient returns the RepoWatcherClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the RepoWatchers capability.
func AsRepoWatcherClient(c Client, host string) (RepoWatcherClient, error) {
	return requireFeature[RepoWatcherClient](c, host, specFor(FeatureRepoWatchers))
}
