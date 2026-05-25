package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const workspacePipelineVarsPath = "/workspaces/%s/pipelines-config/variables/"
const workspacePipelineVarPath = "/workspaces/%s/pipelines-config/variables/%s"

// ListWorkspacePipelineVariables returns workspace-level pipeline variables.
// Secured variable values are not returned by the API.
func (c *Client) ListWorkspacePipelineVariables(workspace string) ([]backend.PipelineVariable, error) {
	path := fmt.Sprintf(workspacePipelineVarsPath, workspace)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PipelineVariable, error) {
		var page cloudPagedResponse[cloudgen.CloudPipelineVariable]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PipelineVariable, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toPipelineVariableDomain(w))
		}
		return out, nil
	}, 0)
}

// GetWorkspacePipelineVariable retrieves a single workspace pipeline variable
// by UUID.
func (c *Client) GetWorkspacePipelineVariable(workspace, uuid string) (backend.PipelineVariable, error) {
	path := fmt.Sprintf(workspacePipelineVarPath, workspace, url.PathEscape(braceUUID(uuid)))
	var w cloudgen.CloudPipelineVariable
	if err := c.http.GetJSON(path, &w); err != nil {
		return backend.PipelineVariable{}, err
	}
	return toPipelineVariableDomain(w), nil
}

// SetWorkspacePipelineVariable upserts a workspace-level pipeline variable.
// If in.UUID is non-empty it updates via PUT; otherwise it searches by Key
// and PUTs if found, POSTs if not.
func (c *Client) SetWorkspacePipelineVariable(workspace string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	body := cloudgen.CloudPipelineVariable{
		Key:     in.Key,
		Value:   in.Value,
		Secured: in.Secured,
	}

	// Find existing by key to upsert.
	existing, err := c.findWorkspacePipelineVariable(workspace, in.Key)
	if err != nil {
		return backend.PipelineVariable{}, err
	}

	var w cloudgen.CloudPipelineVariable
	if existing != nil {
		path := fmt.Sprintf(workspacePipelineVarPath, workspace, url.PathEscape(braceUUID(existing.UUID)))
		if err := c.http.PutJSON(path, body, &w); err != nil {
			return backend.PipelineVariable{}, err
		}
	} else {
		path := fmt.Sprintf(workspacePipelineVarsPath, workspace)
		if err := c.http.PostJSON(path, body, &w); err != nil {
			return backend.PipelineVariable{}, err
		}
	}
	return toPipelineVariableDomain(w), nil
}

// DeleteWorkspacePipelineVariable deletes a workspace pipeline variable by
// UUID. Returns a typed ErrNotFound DomainError when the UUID is not found.
func (c *Client) DeleteWorkspacePipelineVariable(workspace, uuid string) error {
	path := fmt.Sprintf(workspacePipelineVarPath, workspace, url.PathEscape(braceUUID(uuid)))
	return c.http.DeleteJSON(path, nil)
}

// findWorkspacePipelineVariable returns the variable matching key, or nil.
func (c *Client) findWorkspacePipelineVariable(workspace, key string) (*backend.PipelineVariable, error) {
	vars, err := c.ListWorkspacePipelineVariables(workspace)
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
