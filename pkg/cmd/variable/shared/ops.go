package shared

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// VariableItem is a scope-agnostic view of a pipeline/workspace/deployment
// variable for use by callers that don't care which scope it came from.
// Value is empty whenever Secured is true.
type VariableItem struct {
	UUID    string
	Key     string
	Value   string
	Secured bool
}

// VariableOps is the scope-agnostic interface for variable CRUD. The concrete
// implementation is selected by ResolveVariableOps based on scope.
type VariableOps interface {
	// ListVariables returns all variables in the bound scope.
	ListVariables() ([]VariableItem, error)
	// GetVariableByKey returns the single variable matching key. Returns a
	// typed ErrNotFound DomainError when no variable with that key exists.
	GetVariableByKey(key string) (VariableItem, error)
	// SetVariable upserts the variable by key and returns the resulting item.
	SetVariable(key, value string, secured bool) (VariableItem, error)
	// DeleteVariableByKey removes the variable identified by key. For scopes
	// whose underlying API only accepts a UUID, the implementation performs
	// a list-then-find lookup and returns an error if the key is absent.
	DeleteVariableByKey(key string) error
}

// ResolveVariableOps maps scope + backend client to the right VariableOps
// implementation. scope must be one of "repository", "workspace", or
// "deployment". For "deployment", envUUID must be non-empty.
//
// The host argument is forwarded to the As* capability assertions so that
// any unsupported-on-host error carries the correct host label.
func ResolveVariableOps(scope string, client backend.Client, host, project, slug, envUUID string) (VariableOps, error) {
	switch scope {
	case "repository":
		pc, err := backend.AsPipelineClient(client, host)
		if err != nil {
			return nil, err
		}
		return &repoOps{pc: pc, ns: project, slug: slug}, nil

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, host)
		if err != nil {
			return nil, err
		}
		return &workspaceOps{wc: wc, ns: project}, nil

	case "deployment":
		if envUUID == "" {
			return nil, &backend.DomainError{
				Kind:    backend.ErrInvalidRequest,
				Message: "--env ENV-UUID is required for --scope deployment",
			}
		}
		dc, err := backend.AsDeploymentClient(client, host)
		if err != nil {
			return nil, err
		}
		return &deploymentOps{dc: dc, ns: project, slug: slug, envUUID: envUUID}, nil

	default:
		return nil, &backend.DomainError{
			Kind:    backend.ErrInvalidRequest,
			Message: fmt.Sprintf("unknown scope %q; valid: repository, workspace, deployment", scope),
		}
	}
}

// --- repository scope ----------------------------------------------------

type repoOps struct {
	pc       backend.PipelineClient
	ns, slug string
}

func (o *repoOps) ListVariables() ([]VariableItem, error) {
	vars, err := o.pc.ListPipelineVariables(o.ns, o.slug)
	if err != nil {
		return nil, err
	}
	out := make([]VariableItem, 0, len(vars))
	for _, v := range vars {
		out = append(out, VariableItem{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured})
	}
	return out, nil
}

func (o *repoOps) GetVariableByKey(key string) (VariableItem, error) {
	items, err := o.ListVariables()
	if err != nil {
		return VariableItem{}, err
	}
	for _, v := range items {
		if v.Key == key {
			return v, nil
		}
	}
	return VariableItem{}, &backend.DomainError{
		Kind:     backend.ErrNotFound,
		Resource: "pipeline-variable",
		ID:       key,
		Message:  fmt.Sprintf("pipeline variable %q not found", key),
	}
}

func (o *repoOps) SetVariable(key, value string, secured bool) (VariableItem, error) {
	v, err := o.pc.SetPipelineVariable(o.ns, o.slug, backend.PipelineVariableInput{
		Key:     key,
		Value:   value,
		Secured: secured,
	})
	if err != nil {
		return VariableItem{}, err
	}
	return VariableItem{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured}, nil
}

func (o *repoOps) DeleteVariableByKey(key string) error {
	return o.pc.DeletePipelineVariable(o.ns, o.slug, key)
}

// --- workspace scope -----------------------------------------------------

type workspaceOps struct {
	wc backend.WorkspaceVariableClient
	ns string
}

func (o *workspaceOps) ListVariables() ([]VariableItem, error) {
	vars, err := o.wc.ListWorkspaceVariables(o.ns)
	if err != nil {
		return nil, err
	}
	out := make([]VariableItem, 0, len(vars))
	for _, v := range vars {
		out = append(out, VariableItem{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured})
	}
	return out, nil
}

func (o *workspaceOps) GetVariableByKey(key string) (VariableItem, error) {
	items, err := o.ListVariables()
	if err != nil {
		return VariableItem{}, err
	}
	for _, v := range items {
		if v.Key == key {
			return v, nil
		}
	}
	return VariableItem{}, &backend.DomainError{
		Kind:     backend.ErrNotFound,
		Resource: "pipeline-variable",
		ID:       key,
		Message:  fmt.Sprintf("pipeline variable %q not found", key),
	}
}

func (o *workspaceOps) SetVariable(key, value string, secured bool) (VariableItem, error) {
	v, err := o.wc.SetWorkspaceVariable(o.ns, backend.PipelineVariableInput{
		Key:     key,
		Value:   value,
		Secured: secured,
	})
	if err != nil {
		return VariableItem{}, err
	}
	return VariableItem{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured}, nil
}

func (o *workspaceOps) DeleteVariableByKey(key string) error {
	return o.wc.DeleteWorkspaceVariable(o.ns, key)
}

// --- deployment scope ----------------------------------------------------

type deploymentOps struct {
	dc                backend.DeploymentClient
	ns, slug, envUUID string
}

func (o *deploymentOps) ListVariables() ([]VariableItem, error) {
	vars, err := o.dc.ListEnvVariables(o.ns, o.slug, o.envUUID)
	if err != nil {
		return nil, err
	}
	out := make([]VariableItem, 0, len(vars))
	for _, v := range vars {
		out = append(out, VariableItem{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured})
	}
	return out, nil
}

func (o *deploymentOps) GetVariableByKey(key string) (VariableItem, error) {
	items, err := o.ListVariables()
	if err != nil {
		return VariableItem{}, err
	}
	for _, v := range items {
		if v.Key == key {
			return v, nil
		}
	}
	return VariableItem{}, &backend.DomainError{
		Kind:     backend.ErrNotFound,
		Resource: "pipeline-variable",
		ID:       key,
		Message:  fmt.Sprintf("pipeline variable %q not found in environment %s", key, o.envUUID),
	}
}

func (o *deploymentOps) SetVariable(key, value string, secured bool) (VariableItem, error) {
	v, err := o.dc.SetEnvVariable(o.ns, o.slug, o.envUUID, backend.EnvVariableInput{
		Key:     key,
		Value:   value,
		Secured: secured,
	})
	if err != nil {
		return VariableItem{}, err
	}
	return VariableItem{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured}, nil
}

func (o *deploymentOps) DeleteVariableByKey(key string) error {
	vars, err := o.dc.ListEnvVariables(o.ns, o.slug, o.envUUID)
	if err != nil {
		return err
	}
	var varUUID string
	for _, v := range vars {
		if v.Key == key {
			varUUID = v.UUID
			break
		}
	}
	if varUUID == "" {
		return &backend.DomainError{
			Kind:     backend.ErrNotFound,
			Resource: "pipeline-variable",
			ID:       key,
			Message:  fmt.Sprintf("variable %q not found in environment %s", key, o.envUUID),
		}
	}
	return o.dc.DeleteEnvVariable(o.ns, o.slug, o.envUUID, varUUID)
}
