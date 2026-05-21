package backend

// DeploymentClient is implemented by Cloud only; Server returns ErrUnsupportedOnHost.
type DeploymentClient interface {
	ListDeployments(ns, slug string, limit int) ([]Deployment, error)
	GetDeployment(ns, slug, uuid string) (Deployment, error)

	ListEnvironments(ns, slug string) ([]Environment, error)
	CreateEnvironment(ns, slug string, in CreateEnvironmentInput) (Environment, error)
	DeleteEnvironment(ns, slug, uuid string) error

	ListEnvVariables(ns, slug, envUUID string) ([]EnvVariable, error)
	SetEnvVariable(ns, slug, envUUID string, in EnvVariableInput) (EnvVariable, error)
	DeleteEnvVariable(ns, slug, envUUID, varUUID string) error
}

// FeatureDeployments names the deployments capability for typed-error reporting.
const FeatureDeployments Feature = "deployments"

// AsDeploymentClient returns the DeploymentClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the Deployments capability.
func AsDeploymentClient(c Client, host string) (DeploymentClient, error) {
	return requireFeature[DeploymentClient](c, host, specFor(FeatureDeployments))
}
