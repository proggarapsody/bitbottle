package backend

// IssueAttacher is implemented only by Bitbucket Cloud clients.
// It manages file attachments on issues.
type IssueAttacher interface {
	ListIssueAttachments(ns, slug string, id int) ([]IssueAttachment, error)
	DeleteIssueAttachment(ns, slug string, id int, filename string) error
}

// IssueVoter is implemented only by Bitbucket Cloud clients.
// It manages votes on issues.
type IssueVoter interface {
	VoteIssue(ns, slug string, id int) error
	UnvoteIssue(ns, slug string, id int) error
}

// IssueWatcher is implemented only by Bitbucket Cloud clients.
// It manages watches on issues.
type IssueWatcher interface {
	WatchIssue(ns, slug string, id int) error
	UnwatchIssue(ns, slug string, id int) error
}

// FeatureIssueAttachments names the issue attachments capability.
const FeatureIssueAttachments Feature = "issue-attachments"

// FeatureIssueVoting names the issue voting capability.
const FeatureIssueVoting Feature = "issue-voting"

// FeatureIssueWatching names the issue watching capability.
const FeatureIssueWatching Feature = "issue-watching"

// AsIssueAttacher returns the IssueAttacher view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsIssueAttacher(c Client, host string) (IssueAttacher, error) {
	return requireFeature[IssueAttacher](c, host, specFor(FeatureIssueAttachments))
}

// AsIssueVoter returns the IssueVoter view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsIssueVoter(c Client, host string) (IssueVoter, error) {
	return requireFeature[IssueVoter](c, host, specFor(FeatureIssueVoting))
}

// AsIssueWatcher returns the IssueWatcher view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsIssueWatcher(c Client, host string) (IssueWatcher, error) {
	return requireFeature[IssueWatcher](c, host, specFor(FeatureIssueWatching))
}
