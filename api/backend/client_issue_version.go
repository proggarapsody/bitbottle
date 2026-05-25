package backend

// IssueVersionClient is implemented only by Bitbucket Cloud clients.
// It provides CRUD access to issue versions on the Cloud issue tracker.
type IssueVersionClient interface {
	ListIssueVersions(ns, slug string, limit int) ([]IssueVersion, error)
	GetIssueVersion(ns, slug string, id int) (IssueVersion, error)
	CreateIssueVersion(ns, slug, name string) (IssueVersion, error)
	DeleteIssueVersion(ns, slug string, id int) error
}

// FeatureIssueVersions names the issue versions capability for typed-error reporting.
const FeatureIssueVersions Feature = "issue_versions"

// AsIssueVersionClient returns the IssueVersionClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsIssueVersionClient(c Client, host string) (IssueVersionClient, error) {
	return requireFeature[IssueVersionClient](c, host, specFor(FeatureIssueVersions))
}
