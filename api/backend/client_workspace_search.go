package backend

// WorkspaceSearcher is Cloud-only; Server/DC returns host.unsupported.
type WorkspaceSearcher interface {
	SearchWorkspaces(opts WorkspaceSearchOpts) ([]Workspace, error)
}

// WorkspaceSearchOpts carries filter options for SearchWorkspaces.
type WorkspaceSearchOpts struct {
	Query string // slug/name prefix match
	Role  string // owner | collaborator | member | "" (all)
	Limit int
}

// FeatureWorkspaceSearch names the workspace-search capability for typed-error reporting.
const FeatureWorkspaceSearch Feature = "workspace_search"

// AsWorkspaceSearcher returns the WorkspaceSearcher view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the WorkspaceSearch capability.
func AsWorkspaceSearcher(c Client, host string) (WorkspaceSearcher, error) {
	return requireFeature[WorkspaceSearcher](c, host, specFor(FeatureWorkspaceSearch))
}
