package backend

// WorkspacePermsClient is implemented only by Bitbucket Cloud clients.
// Bitbucket Server / Data Center does not have workspace permission APIs,
// so this surface is gated behind AsWorkspacePermsClient.
type WorkspacePermsClient interface {
	ListWorkspaceMemberPerms(ws string, limit int) ([]WorkspaceMemberPerm, error)
	ListWorkspaceRepoPerms(ws string, limit int) ([]WorkspaceRepoPerm, error)
	GrantWorkspacePerm(ws, user, permission string) error
	RevokeWorkspacePerm(ws, user string) error
}

// FeatureWorkspacePerms names the workspace-permissions capability for
// typed-error reporting.
const FeatureWorkspacePerms Feature = "workspace_perms"

// AsWorkspacePermsClient returns the WorkspacePermsClient view of c, or a
// typed *DomainError (Kind=ErrUnsupportedOnHost) when called against a
// Server/DC backend.
func AsWorkspacePermsClient(c Client, host string) (WorkspacePermsClient, error) {
	return requireFeature[WorkspacePermsClient](c, host, specFor(FeatureWorkspacePerms))
}
