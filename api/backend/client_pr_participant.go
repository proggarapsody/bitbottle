package backend

// PRParticipantClient is implemented by both Cloud and Server backends.
type PRParticipantClient interface {
	ListPRParticipants(ns, slug string, prID int) ([]PRParticipant, error)
}

// FeaturePRParticipants names the PR-participants capability for typed-error reporting.
const FeaturePRParticipants Feature = "pr_participants"

// AsPRParticipantClient returns the PRParticipantClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the PRParticipants capability.
func AsPRParticipantClient(c Client, host string) (PRParticipantClient, error) {
	return requireFeature[PRParticipantClient](c, host, specFor(FeaturePRParticipants))
}
