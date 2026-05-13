package backend

import "fmt"

// PRFileClient is implemented by both Cloud and Server backends.
type PRFileClient interface {
	ListPRFiles(ns, slug string, prID int) ([]DiffStatEntry, error)
}

// FeaturePRFiles names the PR-files capability for typed-error reporting.
const FeaturePRFiles Feature = "pr_files"

// AsPRFileClient returns the PRFileClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the PRFiles capability.
func AsPRFileClient(c Client, host string) (PRFileClient, error) {
	pf, ok := c.(PRFileClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePRFiles),
			Message: fmt.Sprintf("PR files are not supported on %s", host),
		}
	}
	return pf, nil
}
