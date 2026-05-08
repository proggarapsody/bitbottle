package backend

import "fmt"

// WorkspaceClient is implemented only by Bitbucket Cloud clients. Bitbucket
// Server / Data Center has no workspace concept — its projects live directly
// under the instance — so the workspace and project list operations are
// Cloud-only and accessed via the AsWorkspaceClient type assertion.
type WorkspaceClient interface {
	ListWorkspaces(limit int) ([]Workspace, error)
	ListProjects(workspace string, limit int) ([]Project, error)
}

// FeatureWorkspaces names the workspace/project listing capability for
// typed-error reporting.
const FeatureWorkspaces Feature = "workspaces"

// AsWorkspaceClient returns the WorkspaceClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// support workspaces.
func AsWorkspaceClient(c Client, host string) (WorkspaceClient, error) {
	wc, ok := c.(WorkspaceClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureWorkspaces),
			Message: fmt.Sprintf("workspaces are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return wc, nil
}
