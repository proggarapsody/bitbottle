package backend

// CloudProjectClient is implemented only by Bitbucket Cloud clients.
// It provides CRUD access to workspace projects.
type CloudProjectClient interface {
	CreateWorkspaceProject(ws string, input CreateWorkspaceProjectInput) (WorkspaceProject, error)
	GetWorkspaceProject(ws, key string) (WorkspaceProject, error)
	UpdateWorkspaceProject(ws, key string, input UpdateWorkspaceProjectInput) (WorkspaceProject, error)
	DeleteWorkspaceProject(ws, key string) error
}

// FeatureCloudProjects names the Cloud workspace projects capability for typed-error reporting.
const FeatureCloudProjects Feature = "cloud_projects"

// AsCloudProjectClient returns the CloudProjectClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsCloudProjectClient(c Client, host string) (CloudProjectClient, error) {
	return requireFeature[CloudProjectClient](c, host, specFor(FeatureCloudProjects))
}
