// Package cloud does not implement AdminClient. Admin operations (secret
// rotation, logging config) are Bitbucket Server / Data Center features only.
// Calls to AsAdminClient against a Cloud backend return ErrUnsupportedOnHost.
package cloud
