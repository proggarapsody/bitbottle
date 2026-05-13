package backend

import "fmt"

// CommitFileClient is implemented by both Cloud and Server backends.
type CommitFileClient interface {
	ListCommitFiles(ns, slug, hash string) ([]DiffStatEntry, error)
}

// FeatureCommitFiles names the commit-files capability for typed-error reporting.
const FeatureCommitFiles Feature = "commit_files"

// AsCommitFileClient returns the CommitFileClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the CommitFiles capability.
func AsCommitFileClient(c Client, host string) (CommitFileClient, error) {
	cf, ok := c.(CommitFileClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureCommitFiles),
			Message: fmt.Sprintf("listing commit files is not supported on %s", host),
		}
	}
	return cf, nil
}
