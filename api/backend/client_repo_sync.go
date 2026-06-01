package backend

// RepoSyncer synchronises a fork branch with its upstream. Implemented only
// by the Bitbucket Cloud adapter; Bitbucket Server / Data Center has no
// fork-upstream concept and returns ErrUnsupportedOnHost via AsRepoSyncer.
type RepoSyncer interface {
	SyncRepo(ns, slug, branch string) (SyncResult, error)
}

// SyncResult is returned by SyncRepo.
type SyncResult struct {
	Behind        int `json:"behind"`
	CommitsMerged int `json:"commits_merged"`
}

// FeatureRepoSync names the repo-sync capability for typed-error reporting.
const FeatureRepoSync Feature = "repo-sync"

// AsRepoSyncer returns the RepoSyncer view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsRepoSyncer(c Client, host string) (RepoSyncer, error) {
	return requireFeature[RepoSyncer](c, host, specFor(FeatureRepoSync))
}
