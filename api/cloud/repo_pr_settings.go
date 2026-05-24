// Package cloud does not implement RepoPRSettingsClient. Bitbucket Cloud does
// not expose per-repo PR gate settings (required approvers, merge strategies,
// etc.) via the same dedicated API endpoint that Server/DC provides.
// Calls to AsRepoPRSettingsClient against a Cloud backend return
// ErrUnsupportedOnHost.
package cloud
