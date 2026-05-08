package backend

import "fmt"

// IssueClient is implemented only by Bitbucket Cloud clients. Bitbucket
// Server / Data Center has no built-in issue tracker, so the entire issue
// surface is gated behind AsIssueClient.
type IssueClient interface {
	ListIssues(ns, slug, state string, limit int) ([]Issue, error)
	GetIssue(ns, slug string, id int) (Issue, error)
	CreateIssue(ns, slug string, in CreateIssueInput) (Issue, error)
	UpdateIssue(ns, slug string, id int, in UpdateIssueInput) (Issue, error)
}

// FeatureIssues names the issues capability for typed-error reporting.
const FeatureIssues Feature = "issues"

// AsIssueClient returns the IssueClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsIssueClient(c Client, host string) (IssueClient, error) {
	ic, ok := c.(IssueClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureIssues),
			Message: fmt.Sprintf("issues are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return ic, nil
}
