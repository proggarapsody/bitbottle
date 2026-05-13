// Package cloud does not implement PermissionsClient. Bitbucket Cloud uses a
// different permissions model that is not yet supported by this CLI.
// Calls to AsPermissionsClient against a Cloud backend return
// ErrUnsupportedOnHost.
package cloud
