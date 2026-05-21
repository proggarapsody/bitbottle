package backend

// DeployKeyClient is implemented by both Cloud and Server backends.
type DeployKeyClient interface {
	ListDeployKeys(ns, slug string) ([]DeployKey, error)
	AddDeployKey(ns, slug string, input DeployKeyInput) (DeployKey, error)
	DeleteDeployKey(ns, slug string, id int) error
}

// FeatureDeployKeys names the deploy-keys capability for typed-error reporting.
const FeatureDeployKeys Feature = "deploy_keys"

// AsDeployKeyClient returns the DeployKeyClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the DeployKeys capability.
func AsDeployKeyClient(c Client, host string) (DeployKeyClient, error) {
	return requireFeature[DeployKeyClient](c, host, specFor(FeatureDeployKeys))
}
