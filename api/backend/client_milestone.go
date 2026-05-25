package backend

// MilestoneClient is implemented only by Bitbucket Cloud clients.
// It provides read access to issue milestones.
type MilestoneClient interface {
	ListMilestones(ns, slug string, limit int) ([]Milestone, error)
	GetMilestone(ns, slug string, id int) (Milestone, error)
}

// FeatureMilestones names the milestones capability for typed-error reporting.
const FeatureMilestones Feature = "milestones"

// AsMilestoneClient returns the MilestoneClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsMilestoneClient(c Client, host string) (MilestoneClient, error) {
	return requireFeature[MilestoneClient](c, host, specFor(FeatureMilestones))
}
