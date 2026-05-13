package backend

import "fmt"

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
	rt, ok := c.(RepoTransferClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureRepoTransfer),
			Message: fmt.Sprintf("repo transfer is not supported on %s", host),
		}
	}
	return rt, nil
}
