package backend

// IssueActivityClient is Cloud-only; Server returns host.unsupported.
type IssueActivityClient interface {
	ListIssueActivity(ns, slug string, issueID int, limit int) ([]IssueChange, error)
}

// FeatureIssueActivity names the issue-activity capability for typed-error reporting.
const FeatureIssueActivity Feature = "issue_activity"

// AsIssueActivityClient returns the IssueActivityClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the IssueActivity capability.
func AsIssueActivityClient(c Client, host string) (IssueActivityClient, error) {
	return requireFeature[IssueActivityClient](c, host, specFor(FeatureIssueActivity))
}
