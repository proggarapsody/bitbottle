package backend

// RepoLabelClient manages labels on a repository. Implemented by both Cloud
// and Server/DC backends.
type RepoLabelClient interface {
	ListRepoLabels(ns, slug string) ([]RepoLabel, error)
	CreateRepoLabel(ns, slug string, in CreateRepoLabelInput) (RepoLabel, error)
	UpdateRepoLabel(ns, slug string, id int, in UpdateRepoLabelInput) (RepoLabel, error)
	DeleteRepoLabel(ns, slug string, id int) error
}

// FeatureRepoLabels names the repository labels capability.
const FeatureRepoLabels Feature = "repo-labels"

// AsRepoLabelClient returns the RepoLabelClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when the backend does not implement the capability.
func AsRepoLabelClient(c Client, host string) (RepoLabelClient, error) {
	return requireFeature[RepoLabelClient](c, host, specFor(FeatureRepoLabels))
}
