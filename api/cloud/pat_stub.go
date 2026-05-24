// Package cloud does not implement PATClient.
// Bitbucket Cloud API token management is not available via the REST API.
// Calls to AsPATClient against a Cloud backend return ErrUnsupportedOnHost.
package cloud
