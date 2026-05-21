package backend

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
	return requireFeature[DefaultReviewerClient](c, host, specFor(FeatureDefaultReviewerClient))
}
