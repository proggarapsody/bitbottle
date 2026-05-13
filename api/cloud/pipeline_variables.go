package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const variablesPath = "/repositories/%s/%s/pipelines_config/variables/"
const variablePath = "/repositories/%s/%s/pipelines_config/variables/%s"

type wireCloudPipelineVariable struct {
	UUID    string `json:"uuid"`
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Secured bool   `json:"secured"`
}

func (w wireCloudPipelineVariable) toDomain() backend.PipelineVariable {
	return backend.PipelineVariable{
		UUID:    stripBraces(w.UUID),
		Key:     w.Key,
		Value:   w.Value,
		Secured: w.Secured,
	}
}

// ListPipelineVariables returns repository-level pipeline variables.
// The Bitbucket Cloud API never includes the value of secured variables in the
// response, so PipelineVariable.Value is empty when Secured is true.
func (c *Client) ListPipelineVariables(ns, slug string) ([]backend.PipelineVariable, error) {
	return paging.Collect(c.http, fmt.Sprintf(variablesPath, ns, slug), func(body []byte) ([]backend.PipelineVariable, error) {
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

// SetPipelineVariable upserts a pipeline variable by Key. If a variable with
// the same Key already exists, it is updated via PUT; otherwise a new one is
// created via POST. Hides the UUID-vs-key wrinkle from the caller.
func (c *Client) SetPipelineVariable(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	existing, err := c.findPipelineVariable(ns, slug, in.Key)
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
		// PUT to {variable_uuid}; key field is ignored on update but kept for symmetry.
		path := fmt.Sprintf(variablePath, ns, slug, braceUUID(existing.UUID))
		if err := c.http.PutJSON(path, body, &w); err != nil {
			return backend.PipelineVariable{}, err
		}
	} else {
		path := fmt.Sprintf(variablesPath, ns, slug)
		if err := c.http.PostJSON(path, body, &w); err != nil {
			return backend.PipelineVariable{}, err
		}
	}
	return w.toDomain(), nil
}

// DeletePipelineVariable looks up the variable by Key and DELETEs it. Returns
// a typed ErrNotFound DomainError when no variable matches the key.
func (c *Client) DeletePipelineVariable(ns, slug, key string) error {
	existing, err := c.findPipelineVariable(ns, slug, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return &backend.DomainError{
			Kind:     backend.ErrNotFound,
			Resource: "pipeline-variable",
			ID:       key,
			Message:  fmt.Sprintf("pipeline variable %q not found", key),
		}
	}
	path := fmt.Sprintf(variablePath, ns, slug, braceUUID(existing.UUID))
	return c.http.DeleteJSON(path, nil)
}

// findPipelineVariable returns the variable matching key, or nil if none.
func (c *Client) findPipelineVariable(ns, slug, key string) (*backend.PipelineVariable, error) {
	vars, err := c.ListPipelineVariables(ns, slug)
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
