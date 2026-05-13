package backend

import "fmt"

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
	pc, ok := c.(PRCommitClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePRCommits),
			Message: fmt.Sprintf("PR commits are not supported on %s", host),
		}
	}
	return pc, nil
}
