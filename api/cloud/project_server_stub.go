// Package cloud does not implement ServerProjectClient.
// Bitbucket Cloud has no equivalent server project management API.
// Calls to AsServerProjectClient against a Cloud backend return ErrUnsupportedOnHost.
package cloud
