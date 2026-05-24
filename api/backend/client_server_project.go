package backend

// ServerProjectClient is implemented only by Bitbucket Server/DC clients.
// Cloud workspaces have no equivalent project-management API.
type ServerProjectClient interface {
	ListServerProjects(filter string, limit int) ([]ServerProject, error)
	GetServerProject(key string) (ServerProject, error)
	CreateServerProject(in CreateServerProjectInput) (ServerProject, error)
	UpdateServerProject(key string, in UpdateServerProjectInput) (ServerProject, error)
	DeleteServerProject(key string) error
}

// ServerProject is the domain representation of a Bitbucket Server/DC project.
type ServerProject struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	WebURL      string `json:"webURL"`
}

// CreateServerProjectInput carries the fields for creating a new project.
type CreateServerProjectInput struct {
	Key         string
	Name        string
	Description string
	Public      bool
}

// UpdateServerProjectInput carries optional patch fields for updating a project.
// Nil pointer fields are left unchanged on the server.
type UpdateServerProjectInput struct {
	Name        *string
	Description *string
	Public      *bool
}

// FeatureServerProject names the server project management capability.
const FeatureServerProject Feature = "server-project"

// AsServerProjectClient returns the ServerProjectClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// support server project management.
func AsServerProjectClient(c Client, host string) (ServerProjectClient, error) {
	return requireFeature[ServerProjectClient](c, host, specFor(FeatureServerProject))
}
