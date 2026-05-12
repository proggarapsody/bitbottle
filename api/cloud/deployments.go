package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ── wire types ───────────────────────────────────────────────────────────────

type wireCloudEnvironment struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	EnvironmentType struct {
		Name string `json:"name"`
	} `json:"environment_type"`
	Rank int `json:"rank"`
}

func (w wireCloudEnvironment) toDomain() backend.Environment {
	return backend.Environment{
		UUID: stripBraces(w.UUID),
		Name: w.Name,
		Type: w.EnvironmentType.Name,
		Rank: w.Rank,
	}
}

type wireCloudDeployment struct {
	UUID  string `json:"uuid"`
	State struct {
		Name string `json:"name"`
	} `json:"state"`
	Environment wireCloudEnvironment `json:"environment"`
	Release     struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Commit struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"release"`
}

func (w wireCloudDeployment) toDomain() backend.Deployment {
	d := backend.Deployment{
		UUID:        stripBraces(w.UUID),
		State:       w.State.Name,
		Environment: w.Environment.toDomain(),
	}
	d.Release.Name = w.Release.Name
	d.Release.URL = w.Release.URL
	d.Release.CommitHash = w.Release.Commit.Hash
	return d
}

type wireCloudEnvVariable struct {
	UUID    string `json:"uuid"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Secured bool   `json:"secured"`
}

func (w wireCloudEnvVariable) toDomain() backend.EnvVariable {
	return backend.EnvVariable{
		UUID:    stripBraces(w.UUID),
		Key:     w.Key,
		Value:   w.Value,
		Secured: w.Secured,
	}
}

// ── write bodies ─────────────────────────────────────────────────────────────

type createEnvironmentBody struct {
	Name            string                    `json:"name"`
	EnvironmentType createEnvironmentTypeBody `json:"environment_type"`
	Rank            int                       `json:"rank,omitempty"`
}

type createEnvironmentTypeBody struct {
	Name string `json:"name"`
}

type setEnvVariableBody struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Secured bool   `json:"secured"`
}

// ── DeploymentClient methods ─────────────────────────────────────────────────

// ListDeployments returns up to limit deployments for the repository, most
// recent first.
func (c *Client) ListDeployments(ns, slug string, limit int) ([]backend.Deployment, error) {
	pagelen := limit
	if pagelen <= 0 || pagelen > 100 {
		pagelen = 100
	}
	path := fmt.Sprintf("/repositories/%s/%s/deployments?sort=-created_on&pagelen=%d",
		url.PathEscape(ns), url.PathEscape(slug), pagelen)

	return paging.Collect(c.http, path, func(body []byte) ([]backend.Deployment, error) {
		var page cloudPagedResponse[wireCloudDeployment]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Deployment, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

// GetDeployment fetches a single deployment by UUID.
func (c *Client) GetDeployment(ns, slug, uuid string) (backend.Deployment, error) {
	var w wireCloudDeployment
	path := fmt.Sprintf("/repositories/%s/%s/deployments/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(braceUUID(uuid)))
	if err := c.getJSON(path, &w); err != nil {
		return backend.Deployment{}, err
	}
	return w.toDomain(), nil
}

// ListEnvironments returns all deployment environments for the repository.
func (c *Client) ListEnvironments(ns, slug string) ([]backend.Environment, error) {
	path := fmt.Sprintf("/repositories/%s/%s/environments",
		url.PathEscape(ns), url.PathEscape(slug))

	return paging.Collect(c.http, path, func(body []byte) ([]backend.Environment, error) {
		var page cloudPagedResponse[wireCloudEnvironment]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Environment, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// CreateEnvironment creates a new deployment environment.
func (c *Client) CreateEnvironment(ns, slug string, in backend.CreateEnvironmentInput) (backend.Environment, error) {
	body := createEnvironmentBody{
		Name:            in.Name,
		EnvironmentType: createEnvironmentTypeBody{Name: in.Type},
		Rank:            in.Rank,
	}
	var w wireCloudEnvironment
	path := fmt.Sprintf("/repositories/%s/%s/environments",
		url.PathEscape(ns), url.PathEscape(slug))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Environment{}, err
	}
	return w.toDomain(), nil
}

// DeleteEnvironment deletes a deployment environment by UUID.
func (c *Client) DeleteEnvironment(ns, slug, uuid string) error {
	path := fmt.Sprintf("/repositories/%s/%s/environments/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(braceUUID(uuid)))
	return c.delete(path)
}

// ListEnvVariables returns all variables for a deployment environment.
func (c *Client) ListEnvVariables(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(braceUUID(envUUID)))

	return paging.Collect(c.http, path, func(body []byte) ([]backend.EnvVariable, error) {
		var page cloudPagedResponse[wireCloudEnvVariable]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.EnvVariable, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// SetEnvVariable upserts an environment variable by Key. If a variable with
// the same Key already exists it is updated via PUT; otherwise it is created
// via POST.
func (c *Client) SetEnvVariable(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
	existing, err := c.findEnvVariable(ns, slug, envUUID, in.Key)
	if err != nil {
		return backend.EnvVariable{}, err
	}
	body := setEnvVariableBody{
		Key:     in.Key,
		Value:   in.Value,
		Secured: in.Secured,
	}
	var w wireCloudEnvVariable
	if existing != nil {
		path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/%s",
			url.PathEscape(ns), url.PathEscape(slug),
			url.PathEscape(braceUUID(envUUID)), url.PathEscape(braceUUID(existing.UUID)))
		if err := c.putJSON(path, body, &w); err != nil {
			return backend.EnvVariable{}, err
		}
	} else {
		path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables",
			url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(braceUUID(envUUID)))
		if err := c.postJSON(path, body, &w); err != nil {
			return backend.EnvVariable{}, err
		}
	}
	return w.toDomain(), nil
}

// findEnvVariable returns the variable matching key, or nil if none.
func (c *Client) findEnvVariable(ns, slug, envUUID, key string) (*backend.EnvVariable, error) {
	vars, err := c.ListEnvVariables(ns, slug, envUUID)
	if err != nil {
		return nil, err
	}
	for i := range vars {
		if vars[i].Key == key {
			return &vars[i], nil
		}
	}
	return nil, nil
}

// DeleteEnvVariable deletes an environment variable by its UUID.
func (c *Client) DeleteEnvVariable(ns, slug, envUUID, varUUID string) error {
	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/%s",
		url.PathEscape(ns), url.PathEscape(slug),
		url.PathEscape(braceUUID(envUUID)), url.PathEscape(braceUUID(varUUID)))
	return c.delete(path)
}
