package pipelinevar_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/pipelinevar"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudCfg = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

var fakeVars = []backend.PipelineVariable{
	{UUID: "uuid-1", Key: "CI_TOKEN", Value: "tok123", Secured: false},
	{UUID: "uuid-2", Key: "DEPLOY_SECRET", Value: "", Secured: true},
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestList_AcceptsOptionalWorkspaceArg(t *testing.T) {
	t.Parallel()
	var gotWS string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg})
	cmd := pipelinevar.NewCmdList(f, func(opts *pipelinevar.ListOptions) error {
		gotWS = opts.Workspace
		return nil
	})
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWS)
}

func TestList_PrintsVariables(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "myws", workspace)
			return fakeVars, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "CI_TOKEN")
	assert.Contains(t, got, "DEPLOY_SECRET")
}

func TestList_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return nil, errors.New("api error")
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws"})
	require.Error(t, cmd.Execute())
}

// ── get ───────────────────────────────────────────────────────────────────────

func TestGet_AcceptsWorkspaceAndKey(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg})
	cmd := pipelinevar.NewCmdGet(f, func(opts *pipelinevar.GetOptions) error {
		gotWS = opts.Workspace
		gotKey = opts.Key
		return nil
	})
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws", "CI_TOKEN"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWS)
	assert.Equal(t, "CI_TOKEN", gotKey)
}

func TestGet_ResolvesKeyToUUID(t *testing.T) {
	t.Parallel()
	var gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return fakeVars, nil
		},
		GetWorkspacePipelineVariableFn: func(workspace, uuid string) (backend.PipelineVariable, error) {
			gotUUID = uuid
			return fakeVars[0], nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdGet(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws", "CI_TOKEN"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "uuid-1", gotUUID)
}

func TestGet_ReturnsNotFoundWhenKeyMissing(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return fakeVars, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdGet(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws", "NONEXISTENT"})
	require.Error(t, cmd.Execute())
}

// ── set ───────────────────────────────────────────────────────────────────────

func TestSet_AcceptsWorkspaceKeyValue(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey, gotVal string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg})
	cmd := pipelinevar.NewCmdSet(f, func(opts *pipelinevar.SetOptions) error {
		gotWS = opts.Workspace
		gotKey = opts.Key
		gotVal = opts.Value
		return nil
	})
	cmd.SetArgs([]string{"myws", "MY_KEY", "my_val"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWS)
	assert.Equal(t, "MY_KEY", gotKey)
	assert.Equal(t, "my_val", gotVal)
}

func TestSet_CallsSetWorkspacePipelineVariable(t *testing.T) {
	t.Parallel()
	var gotInput backend.PipelineVariableInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetWorkspacePipelineVariableFn: func(workspace string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			gotInput = in
			return backend.PipelineVariable{UUID: "new-uuid", Key: in.Key, Value: in.Value}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myws", "CI_KEY", "ci_value", "--secured"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "CI_KEY", gotInput.Key)
	assert.Equal(t, "ci_value", gotInput.Value)
	assert.True(t, gotInput.Secured)
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestDelete_RequiresConfirmInNonTTY(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myws", "CI_TOKEN"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestDelete_ResolvesKeyToUUIDAndDeletes(t *testing.T) {
	t.Parallel()
	var deletedUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return fakeVars, nil
		},
		DeleteWorkspacePipelineVariableFn: func(workspace, uuid string) error {
			deletedUUID = uuid
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myws", "CI_TOKEN", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "uuid-1", deletedUUID)
}

func TestDelete_ReturnsNotFoundWhenKeyMissing(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return fakeVars, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudCfg, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := pipelinevar.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myws", "NONEXISTENT", "--confirm"})
	require.Error(t, cmd.Execute())
}
