package shared_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/variable/shared"
)

// --- Test fakes ----------------------------------------------------------

// fakeClient is a minimal backend.Client used for capability assertions.
// We embed the interface so we satisfy it without listing every method;
// only the variable subset is exercised here.
type fakeClient struct{ backend.Client }

// fakePipelineClient is a fakeClient that also implements PipelineClient.
type fakePipelineClient struct {
	backend.Client
	listFn   func(ns, slug string) ([]backend.PipelineVariable, error)
	setFn    func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error)
	deleteFn func(ns, slug, key string) error
}

func (f *fakePipelineClient) ListPipelines(string, string, int) ([]backend.Pipeline, error) {
	return nil, nil
}
func (f *fakePipelineClient) GetPipeline(string, string, string) (backend.Pipeline, error) {
	return backend.Pipeline{}, nil
}
func (f *fakePipelineClient) RunPipeline(string, string, backend.RunPipelineInput) (backend.Pipeline, error) {
	return backend.Pipeline{}, nil
}
func (f *fakePipelineClient) ListPipelineSteps(string, string, string) ([]backend.PipelineStep, error) {
	return nil, nil
}
func (f *fakePipelineClient) GetPipelineStepLog(string, string, string, string) (io.ReadCloser, error) {
	return nil, nil
}
func (f *fakePipelineClient) ListPipelineVariables(ns, slug string) ([]backend.PipelineVariable, error) {
	return f.listFn(ns, slug)
}
func (f *fakePipelineClient) SetPipelineVariable(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	return f.setFn(ns, slug, in)
}
func (f *fakePipelineClient) StopPipeline(string, string, string) error { return nil }
func (f *fakePipelineClient) RerunPipeline(string, string, string) (backend.Pipeline, error) {
	return backend.Pipeline{}, nil
}
func (f *fakePipelineClient) DeletePipelineVariable(ns, slug, key string) error {
	return f.deleteFn(ns, slug, key)
}

// fakeWorkspaceVarClient implements WorkspaceVariableClient.
type fakeWorkspaceVarClient struct {
	backend.Client
	listFn   func(ns string) ([]backend.PipelineVariable, error)
	setFn    func(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error)
	deleteFn func(ns, key string) error
}

func (f *fakeWorkspaceVarClient) ListWorkspaceVariables(ns string) ([]backend.PipelineVariable, error) {
	return f.listFn(ns)
}
func (f *fakeWorkspaceVarClient) SetWorkspaceVariable(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	return f.setFn(ns, in)
}
func (f *fakeWorkspaceVarClient) DeleteWorkspaceVariable(ns, key string) error {
	return f.deleteFn(ns, key)
}

// fakeDeploymentClient implements DeploymentClient.
type fakeDeploymentClient struct {
	backend.Client
	listFn   func(ns, slug, envUUID string) ([]backend.EnvVariable, error)
	setFn    func(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error)
	deleteFn func(ns, slug, envUUID, varUUID string) error
}

func (f *fakeDeploymentClient) ListDeployments(string, string, int) ([]backend.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentClient) GetDeployment(string, string, string) (backend.Deployment, error) {
	return backend.Deployment{}, nil
}
func (f *fakeDeploymentClient) ListEnvironments(string, string) ([]backend.Environment, error) {
	return nil, nil
}
func (f *fakeDeploymentClient) CreateEnvironment(string, string, backend.CreateEnvironmentInput) (backend.Environment, error) {
	return backend.Environment{}, nil
}
func (f *fakeDeploymentClient) DeleteEnvironment(string, string, string) error { return nil }
func (f *fakeDeploymentClient) ListEnvVariables(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
	return f.listFn(ns, slug, envUUID)
}
func (f *fakeDeploymentClient) SetEnvVariable(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
	return f.setFn(ns, slug, envUUID, in)
}
func (f *fakeDeploymentClient) DeleteEnvVariable(ns, slug, envUUID, varUUID string) error {
	return f.deleteFn(ns, slug, envUUID, varUUID)
}

// --- Tests ---------------------------------------------------------------

func TestResolveVariableOps_UnknownScope(t *testing.T) {
	t.Parallel()
	_, err := shared.ResolveVariableOps("not-a-scope", &fakeClient{}, "bitbucket.org", "ws", "repo", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scope")
}

func TestResolveVariableOps_DeploymentRequiresEnvUUID(t *testing.T) {
	t.Parallel()
	dc := &fakeDeploymentClient{}
	_, err := shared.ResolveVariableOps("deployment", dc, "bitbucket.org", "ws", "repo", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--env")
}

func TestResolveVariableOps_RepositoryScope_ListSetDelete(t *testing.T) {
	t.Parallel()
	pc := &fakePipelineClient{
		listFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "ws", ns)
			assert.Equal(t, "repo", slug)
			return []backend.PipelineVariable{
				{UUID: "u1", Key: "K1", Value: "v1", Secured: false},
				{UUID: "u2", Key: "K2", Value: "", Secured: true},
			}, nil
		},
		setFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			assert.Equal(t, "ws", ns)
			assert.Equal(t, "repo", slug)
			assert.Equal(t, "K3", in.Key)
			assert.Equal(t, "v3", in.Value)
			assert.True(t, in.Secured)
			return backend.PipelineVariable{UUID: "u3", Key: in.Key, Secured: in.Secured}, nil
		},
		deleteFn: func(ns, slug, key string) error {
			assert.Equal(t, "K1", key)
			return nil
		},
	}
	ops, err := shared.ResolveVariableOps("repository", pc, "bitbucket.org", "ws", "repo", "")
	require.NoError(t, err)

	items, err := ops.ListVariables()
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "K1", items[0].Key)
	assert.Equal(t, "v1", items[0].Value)
	assert.False(t, items[0].Secured)
	assert.Equal(t, "K2", items[1].Key)
	assert.True(t, items[1].Secured)

	res, err := ops.SetVariable("K3", "v3", true)
	require.NoError(t, err)
	assert.Equal(t, "K3", res.Key)
	assert.True(t, res.Secured)
	require.NoError(t, ops.DeleteVariableByKey("K1"))
}

func TestResolveVariableOps_WorkspaceScope_ListSetDelete(t *testing.T) {
	t.Parallel()
	wc := &fakeWorkspaceVarClient{
		listFn: func(ns string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "ws", ns)
			return []backend.PipelineVariable{
				{UUID: "u1", Key: "WK", Value: "wv"},
			}, nil
		},
		setFn: func(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			assert.Equal(t, "ws", ns)
			assert.Equal(t, "WK", in.Key)
			return backend.PipelineVariable{Key: in.Key}, nil
		},
		deleteFn: func(ns, key string) error {
			assert.Equal(t, "ws", ns)
			assert.Equal(t, "WK", key)
			return nil
		},
	}
	ops, err := shared.ResolveVariableOps("workspace", wc, "bitbucket.org", "ws", "repo", "")
	require.NoError(t, err)

	items, err := ops.ListVariables()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "WK", items[0].Key)

	_, err = ops.SetVariable("WK", "wv", false)
	require.NoError(t, err)
	require.NoError(t, ops.DeleteVariableByKey("WK"))
}

func TestResolveVariableOps_DeploymentScope_ListSetDelete(t *testing.T) {
	t.Parallel()
	dc := &fakeDeploymentClient{
		listFn: func(ns, slug, env string) ([]backend.EnvVariable, error) {
			assert.Equal(t, "ws", ns)
			assert.Equal(t, "repo", slug)
			assert.Equal(t, "env-uuid", env)
			return []backend.EnvVariable{
				{UUID: "var-uuid-1", Key: "EK1", Value: "ev1"},
				{UUID: "var-uuid-2", Key: "EK2", Secured: true},
			}, nil
		},
		setFn: func(ns, slug, env string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
			assert.Equal(t, "env-uuid", env)
			assert.Equal(t, "EK3", in.Key)
			return backend.EnvVariable{UUID: "var-uuid-3", Key: in.Key}, nil
		},
		deleteFn: func(ns, slug, env, varUUID string) error {
			assert.Equal(t, "env-uuid", env)
			assert.Equal(t, "var-uuid-1", varUUID)
			return nil
		},
	}
	ops, err := shared.ResolveVariableOps("deployment", dc, "bitbucket.org", "ws", "repo", "env-uuid")
	require.NoError(t, err)

	items, err := ops.ListVariables()
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "EK1", items[0].Key)
	assert.Equal(t, "var-uuid-1", items[0].UUID)
	assert.True(t, items[1].Secured)

	_, err = ops.SetVariable("EK3", "ev3", false)
	require.NoError(t, err)
	require.NoError(t, ops.DeleteVariableByKey("EK1"))
}

func TestResolveVariableOps_DeploymentScope_DeleteByKey_NotFound(t *testing.T) {
	t.Parallel()
	dc := &fakeDeploymentClient{
		listFn: func(ns, slug, env string) ([]backend.EnvVariable, error) {
			return []backend.EnvVariable{
				{UUID: "var-uuid-1", Key: "OTHER"},
			}, nil
		},
		deleteFn: func(ns, slug, env, varUUID string) error {
			t.Fatalf("DeleteEnvVariable should not be called for missing key")
			return nil
		},
	}
	ops, err := shared.ResolveVariableOps("deployment", dc, "bitbucket.org", "ws", "repo", "env-uuid")
	require.NoError(t, err)
	err = ops.DeleteVariableByKey("MISSING")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "MISSING")
}

func TestResolveVariableOps_DeploymentScope_DeleteByKey_ListError(t *testing.T) {
	t.Parallel()
	want := errors.New("list failed")
	dc := &fakeDeploymentClient{
		listFn: func(ns, slug, env string) ([]backend.EnvVariable, error) {
			return nil, want
		},
	}
	ops, err := shared.ResolveVariableOps("deployment", dc, "bitbucket.org", "ws", "repo", "env-uuid")
	require.NoError(t, err)
	err = ops.DeleteVariableByKey("ANY")
	require.Error(t, err)
	assert.ErrorIs(t, err, want)
}
