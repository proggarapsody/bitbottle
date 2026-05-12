package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const workspaceVariablesPath = "/workspaces/%s/pipelines-config/variables/"
const workspaceVariablePath = "/workspaces/%s/pipelines-config/variables/%s"

// ListWorkspaceVariables returns workspace-level pipeline variables.
// Secured variable values are not returned by the API.
func (c *Client) ListWorkspaceVariables(ns string) ([]backend.PipelineVariable, error) {
	path := fmt.Sprintf(workspaceVariablesPath, ns)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PipelineVariable, error) {
		var page cloudPagedResponse[wireCloudPipelineVariable]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PipelineVariable, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// SetWorkspaceVariable upserts a workspace-level pipeline variable by Key.
// If a variable with the same Key already exists it is updated via PUT;
// otherwise a new one is created via POST.
func (c *Client) SetWorkspaceVariable(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	existing, err := c.findWorkspaceVariable(ns, in.Key)
	if err != nil {
		return backend.PipelineVariable{}, err
	}
	body := wireCloudPipelineVariable{
		Key:     in.Key,
		Value:   in.Value,
		Secured: in.Secured,
	}
	var w wireCloudPipelineVariable
	if existing != nil {
		path := fmt.Sprintf(workspaceVariablePath, ns, braceUUID(existing.UUID))
		if err := c.http.PutJSON(path, body, &w); err != nil {
			return backend.PipelineVariable{}, err
		}
	} else {
		path := fmt.Sprintf(workspaceVariablesPath, ns)
		if err := c.http.PostJSON(path, body, &w); err != nil {
			return backend.PipelineVariable{}, err
		}
	}
	return w.toDomain(), nil
}

// DeleteWorkspaceVariable looks up a variable by Key and DELETEs it. Returns
// a typed ErrNotFound DomainError when no variable matches the key.
func (c *Client) DeleteWorkspaceVariable(ns, key string) error {
	existing, err := c.findWorkspaceVariable(ns, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return &backend.DomainError{
			Kind:     backend.ErrNotFound,
			Resource: "workspace-variable",
			ID:       key,
			Message:  fmt.Sprintf("workspace variable %q not found", key),
		}
	}
	path := fmt.Sprintf(workspaceVariablePath, ns, braceUUID(existing.UUID))
	return c.http.DeleteJSON(path, nil)
}

// findWorkspaceVariable returns the variable matching key, or nil if none.
func (c *Client) findWorkspaceVariable(ns, key string) (*backend.PipelineVariable, error) {
	vars, err := c.ListWorkspaceVariables(ns)
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
