package pipeline_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPipelineList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdPipelineList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	assert.Equal(t, "20", cmd.Flag("limit").DefValue)
}

func TestNewCmdPipelineList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPipelineList_PrintsBuildNumbers(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 42, State: "SUCCESSFUL", RefName: "main"},
				{BuildNumber: 43, State: "IN_PROGRESS", RefName: "feature/x"},
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "42")
	assert.Contains(t, got, "43")
}

func TestPipelineList_PrintsState(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 1, State: "SUCCESSFUL", RefName: "main"},
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "SUCCESSFUL")
}

func TestPipelineList_PrintsRefName(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 5, State: "SUCCESSFUL", RefName: "main"},
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "main")
}

func TestPipelineList_JSON_FieldsOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 42, State: "SUCCESSFUL", RefName: "main"},
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json", "buildNumber,state"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"buildNumber":42`)
	assert.Contains(t, got, `"state":"SUCCESSFUL"`)
}

func TestPipelineList_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()

	// FakeClient with NO pipeline Fn fields set — simulates a Server client
	// that doesn't implement PipelineClient. But FakeClient now always implements
	// PipelineClient (it has the methods). We test using a stripped-down fake.
	fake := &noPipelineFake{Client: &testhelpers.FakeClient{T: t}}

	f, _, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}

// noPipelineFake wraps backend.Client but does NOT implement backend.PipelineClient,
// simulating a Bitbucket Server backend. Embedding the interface rather than the
// concrete FakeClient prevents pipeline method promotion.
type noPipelineFake struct {
	backend.Client
}

func TestPipelineList_JQ_FilterOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 10, State: "SUCCESSFUL"},
				{BuildNumber: 20, State: "FAILED"},
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineList(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json", "buildNumber", "--jq", ".[] | .buildNumber"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{"10", "20"}, lines)
}
