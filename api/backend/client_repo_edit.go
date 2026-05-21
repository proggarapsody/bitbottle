package backend

import "fmt"

// RepoEditor updates mutable repository metadata fields.
type RepoEditor interface {
	EditRepo(ns, slug string, in EditRepoInput) (Repository, error)
}

// EditRepoInput holds the fields to update. Nil pointer = leave unchanged.
type EditRepoInput struct {
	Description *string
	Website     *string
	Language    *string
	ForkPolicy  *string
	HasIssues   *bool
	HasWiki     *bool
}

// FeatureRepoEdit names the repo-edit capability for typed-error reporting.
const FeatureRepoEdit Feature = "repo_edit"

// AsRepoEditor returns the RepoEditor view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) if the backend at host does not implement
// the RepoEdit capability.
func AsRepoEditor(c Client, host string) (RepoEditor, error) {
	re, ok := c.(RepoEditor)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureRepoEdit),
			Message: fmt.Sprintf("repo edit is not supported on %s", host),
		}
	}
	return re, nil
}
