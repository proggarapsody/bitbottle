package backend

import "fmt"

// ReviewerGroupClient manages named reviewer-group conditions on a Bitbucket
// Server / Data Center repository. Cloud does not expose a compatible API, so
// AsReviewerGroupClient returns ErrUnsupportedOnHost for Cloud backends.
type ReviewerGroupClient interface {
	ListReviewerGroups(ns, slug string) ([]ReviewerGroup, error)
	CreateReviewerGroup(ns, slug string, in CreateReviewerGroupInput) (ReviewerGroup, error)
	DeleteReviewerGroup(ns, slug string, id int) error
}

// FeatureReviewerGroup names the reviewer-group capability for typed-error
// reporting.
const FeatureReviewerGroup Feature = "reviewer-group"

// AsReviewerGroupClient returns the ReviewerGroupClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend does not implement
// the capability (currently Cloud).
func AsReviewerGroupClient(c Client, host string) (ReviewerGroupClient, error) {
	rg, ok := c.(ReviewerGroupClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureReviewerGroup),
			Message: fmt.Sprintf("reviewer group management is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return rg, nil
}
