package backend

import "fmt"

// DefaultReviewerClient is implemented by both Cloud and Server backends.
type DefaultReviewerClient interface {
	ListDefaultReviewers(ns, slug string) ([]DefaultReviewer, error)
	AddDefaultReviewer(ns, slug, userSlug string) error
	RemoveDefaultReviewer(ns, slug, userSlug string) error
}

// FeatureDefaultReviewerClient names the default-reviewer-client capability for
// typed-error reporting.
const FeatureDefaultReviewerClient Feature = "default_reviewer_client"

// AsDefaultReviewerClient returns the DefaultReviewerClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the DefaultReviewerClient capability.
func AsDefaultReviewerClient(c Client, host string) (DefaultReviewerClient, error) {
	dr, ok := c.(DefaultReviewerClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureDefaultReviewerClient),
			Message: fmt.Sprintf("default reviewer management is not supported on %s", host),
		}
	}
	return dr, nil
}
