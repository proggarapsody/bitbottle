package backend

import "fmt"

// WorkspaceMemberClient lists members of a Bitbucket Cloud workspace.
// Cloud-only — Server returns ErrUnsupportedOnHost.
type WorkspaceMemberClient interface {
	ListWorkspaceMembers(workspace string, limit int) ([]WorkspaceMember, error)
}

const FeatureWorkspaceMembers Feature = "workspace_members"

// AsWorkspaceMemberClient returns the WorkspaceMemberClient view of c, or a
// typed *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does
// not support workspace members.
func AsWorkspaceMemberClient(c Client, host string) (WorkspaceMemberClient, error) {
	wmc, ok := c.(WorkspaceMemberClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureWorkspaceMembers),
			Message: fmt.Sprintf("workspace members are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return wmc, nil
}
