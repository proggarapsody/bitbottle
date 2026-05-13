package backend

import "fmt"

// AdminClient exposes Bitbucket Server / Data Center administration
// operations. Bitbucket Cloud does not expose these endpoints — calls against
// Cloud return ErrUnsupportedOnHost via AsAdminClient.
type AdminClient interface {
	RotateSecrets() error
	GetLoggingConfig() (LoggingConfig, error)
	SetLoggingConfig(in LoggingConfigInput) error
}

// FeatureAdmin names the admin capability.
const FeatureAdmin Feature = "admin"

// AsAdminClient returns the AdminClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend that
// doesn't implement admin operations (currently Bitbucket Cloud).
func AsAdminClient(c Client, host string) (AdminClient, error) {
	ac, ok := c.(AdminClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureAdmin),
			Message: fmt.Sprintf("admin operations are not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return ac, nil
}
