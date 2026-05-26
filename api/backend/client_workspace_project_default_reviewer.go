package backend

// WorkspaceProjectDefaultReviewerClient is Cloud-only; Server/DC returns host.unsupported.
type WorkspaceProjectDefaultReviewerClient interface {
	ListProjectDefaultReviewers(workspace, projectKey string, limit int) ([]ProjectDefaultReviewer, error)
	AddProjectDefaultReviewer(workspace, projectKey, accountID string) error
	RemoveProjectDefaultReviewer(workspace, projectKey, accountID string) error
}

// ProjectDefaultReviewer is the domain representation of one default reviewer
// on a Cloud workspace project.
type ProjectDefaultReviewer struct {
	AccountID   string
	DisplayName string
	Nickname    string
	AvatarURL   string
}

// FeatureWorkspaceProjectDefaultReviewers names the workspace-project-default-reviewers
// capability for typed-error reporting.
const FeatureWorkspaceProjectDefaultReviewers Feature = "workspace_project_default_reviewers"

// AsWorkspaceProjectDefaultReviewerClient returns the WorkspaceProjectDefaultReviewerClient
// view of c, or a typed *DomainError (Kind=ErrUnsupportedOnHost) when called
// against a Server/DC backend.
func AsWorkspaceProjectDefaultReviewerClient(c Client, host string) (WorkspaceProjectDefaultReviewerClient, error) {
	return requireFeature[WorkspaceProjectDefaultReviewerClient](c, host, specFor(FeatureWorkspaceProjectDefaultReviewers))
}
