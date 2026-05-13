package backend

import (
	"context"
	"fmt"
)

// PermissionsClient exposes Bitbucket Server / Data Center permission
// management for projects and repositories. Bitbucket Cloud has a different
// permissions model — AsPermissionsClient returns ErrUnsupportedOnHost when
// called against a Cloud backend.
type PermissionsClient interface {
	ListProjectPermissions(ctx context.Context, project string) ([]PermissionGrant, error)
	GrantProjectPermission(ctx context.Context, project string, subject PermissionSubject, perm string) error
	RevokeProjectPermission(ctx context.Context, project string, subject PermissionSubject) error

	ListRepoPermissions(ctx context.Context, project, slug string) ([]PermissionGrant, error)
	GrantRepoPermission(ctx context.Context, project, slug string, subject PermissionSubject, perm string) error
	RevokeRepoPermission(ctx context.Context, project, slug string, subject PermissionSubject) error
}

// FeaturePermissions names the permissions management capability.
const FeaturePermissions Feature = "permissions"

// AsPermissionsClient returns the PermissionsClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend that
// doesn't implement permissions management (currently Bitbucket Cloud).
func AsPermissionsClient(c Client, host string) (PermissionsClient, error) {
	pc, ok := c.(PermissionsClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePermissions),
			Message: fmt.Sprintf("permission management is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return pc, nil
}
