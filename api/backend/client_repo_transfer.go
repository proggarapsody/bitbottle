package backend

// RepoTransferClient transfers a repository to another project (Server) or
// workspace (Cloud).
type RepoTransferClient interface {
	TransferRepo(ns, slug, target string) (Repository, error)
}

// FeatureRepoTransfer names the repo-transfer capability for typed-error reporting.
const FeatureRepoTransfer Feature = "repo_transfer"

// AsRepoTransferClient returns the RepoTransferClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the RepoTransfer capability.
func AsRepoTransferClient(c Client, host string) (RepoTransferClient, error) {
	return requireFeature[RepoTransferClient](c, host, specFor(FeatureRepoTransfer))
}
