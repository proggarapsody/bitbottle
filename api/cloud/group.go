// Package cloud does not implement GroupClient or GroupMemberClient.
// Bitbucket Cloud uses a different workspace-permissions model and does not
// expose admin/groups endpoints. Calls to AsGroupClient or AsGroupMemberClient
// against a Cloud backend return ErrUnsupportedOnHost.
package cloud
