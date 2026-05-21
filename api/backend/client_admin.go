package backend

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
	return requireFeature[AdminClient](c, host, specFor(FeatureAdmin))
}
