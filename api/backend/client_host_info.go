package backend

// HostInfoClient is implemented by both Cloud and Server/DC backends.
// It returns static metadata about the connected host: backend type,
// base URL, optional version (Server/DC only), and the list of Feature
// constants this backend supports.
type HostInfoClient interface {
	GetHostInfo() (HostInfo, error)
}

// FeatureHostInfo names the host-info capability for typed-error reporting.
const FeatureHostInfo Feature = "host_info"

// AsHostInfoClient returns the HostInfoClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when the backend does not
// implement this capability.
func AsHostInfoClient(c Client, host string) (HostInfoClient, error) {
	return requireFeature[HostInfoClient](c, host, specFor(FeatureHostInfo))
}
