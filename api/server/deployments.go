package server

import "github.com/proggarapsody/bitbottle/api/backend"

func (c *Client) ListDeployments(ns, slug string, limit int) ([]backend.Deployment, error) {
	return nil, backend.ErrUnsupportedOnHost
}

func (c *Client) GetDeployment(ns, slug, uuid string) (backend.Deployment, error) {
	return backend.Deployment{}, backend.ErrUnsupportedOnHost
}

func (c *Client) ListEnvironments(ns, slug string) ([]backend.Environment, error) {
	return nil, backend.ErrUnsupportedOnHost
}

func (c *Client) CreateEnvironment(ns, slug string, in backend.CreateEnvironmentInput) (backend.Environment, error) {
	return backend.Environment{}, backend.ErrUnsupportedOnHost
}

func (c *Client) DeleteEnvironment(ns, slug, uuid string) error {
	return backend.ErrUnsupportedOnHost
}

func (c *Client) ListEnvVariables(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
	return nil, backend.ErrUnsupportedOnHost
}

func (c *Client) SetEnvVariable(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
	return backend.EnvVariable{}, backend.ErrUnsupportedOnHost
}

func (c *Client) DeleteEnvVariable(ns, slug, envUUID, varUUID string) error {
	return backend.ErrUnsupportedOnHost
}
