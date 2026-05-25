package backend

// WorkspaceProjectPermsClient is Cloud-only; Server/DC returns host.unsupported.
type WorkspaceProjectPermsClient interface {
	ListWorkspaceProjectPerms(workspace, projectKey string) ([]WorkspaceProjectPerm, error)
	GrantWorkspaceProjectPerm(workspace, projectKey string, in WorkspaceProjectPermInput) error
	RevokeWorkspaceProjectPerm(workspace, projectKey, subjectSlug string, isGroup bool) error
}

// WorkspaceProjectPerm is the domain representation of one permission entry for
// a Cloud workspace project. Exactly one of User/Group is non-nil.
type WorkspaceProjectPerm struct {
	Permission string
	User       *User             // nil if this is a group permission
	Group      *WorkspaceGroup   // nil if this is a user permission
}

// WorkspaceGroup is the Cloud representation of a workspace group as used in
// project and workspace permission APIs. It is distinct from the Server-side
// Group type (which only carries Name).
type WorkspaceGroup struct {
	Slug string
	Name string
}

// WorkspaceProjectPermInput carries the subject and permission level for
// GrantWorkspaceProjectPerm. Set exactly one of UserSlug/GroupSlug.
type WorkspaceProjectPermInput struct {
	Permission string
	UserSlug   string // set when granting a user permission
	GroupSlug  string // set when granting a group permission
}

// FeatureWorkspaceProjectPerms names the workspace-project-permissions
// capability for typed-error reporting.
const FeatureWorkspaceProjectPerms Feature = "workspace_project_perms"

// AsWorkspaceProjectPermsClient returns the WorkspaceProjectPermsClient view
// of c, or a typed *DomainError (Kind=ErrUnsupportedOnHost) when called
// against a Server/DC backend.
func AsWorkspaceProjectPermsClient(c Client, host string) (WorkspaceProjectPermsClient, error) {
	return requireFeature[WorkspaceProjectPermsClient](c, host, specFor(FeatureWorkspaceProjectPerms))
}
