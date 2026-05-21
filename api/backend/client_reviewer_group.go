package backend

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
	return requireFeature[ReviewerGroupClient](c, host, specFor(FeatureReviewerGroup))
}
