package trigger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/trigger"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdTrigger_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := trigger.NewCmdTrigger(f, nil)
	assert.NotNil(t, cmd.Flag("branch"))
	assert.NotNil(t, cmd.Flag("variable"))
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestTrigger_PrintsUUID(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TriggerPipelineFn: func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "main", input.Branch)
			return backend.PipelineTriggerResult{
				UUID:  "abc-123",
				State: "PENDING",
				Link:  "https://api.bitbucket.org/2.0/repositories/myworkspace/my-service/pipelines/%7Babc-123%7D",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := trigger.NewCmdTrigger(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "abc-123")
	assert.Contains(t, out.String(), "PENDING")
}

func TestTrigger_PassesBranchToAPI(t *testing.T) {
	t.Parallel()
	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		TriggerPipelineFn: func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
			gotBranch = input.Branch
			return backend.PipelineTriggerResult{UUID: "x", State: "PENDING"}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := trigger.NewCmdTrigger(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "feature/login"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "feature/login", gotBranch)
}

func TestTrigger_PassesVariablesToAPI(t *testing.T) {
	t.Parallel()
	var gotVars []backend.PipelineVariable
	fake := &testhelpers.FakeClient{
		T: t,
		TriggerPipelineFn: func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
			gotVars = input.Variables
			return backend.PipelineTriggerResult{UUID: "x", State: "PENDING"}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := trigger.NewCmdTrigger(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main", "--variable", "FOO=bar", "--variable", "BAZ=qux"})
	require.NoError(t, cmd.Execute())
	require.Len(t, gotVars, 2)
	assert.Equal(t, "FOO", gotVars[0].Key)
	assert.Equal(t, "bar", gotVars[0].Value)
	assert.Equal(t, "BAZ", gotVars[1].Key)
	assert.Equal(t, "qux", gotVars[1].Value)
}

func TestTrigger_InvalidVariable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	var gotErr error
	cmd := trigger.NewCmdTrigger(f, func(opts *trigger.Options) error {
		return nil
	})
	// Override the run function to test parseVariables indirectly via Execute
	cmd2 := trigger.NewCmdTrigger(f, nil)
	cmd2.SetArgs([]string{"myworkspace/my-service", "--branch", "main", "--variable", "INVALID"})
	gotErr = cmd2.Execute()
	_ = cmd
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "INVALID")
}

func TestTrigger_ClientNotPipelineTriggerCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineTriggerFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := trigger.NewCmdTrigger(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline trigger")
}

// noPipelineTriggerFake wraps backend.Client without implementing
// backend.PipelineTriggerClient — simulates a Bitbucket Server backend.
type noPipelineTriggerFake struct {
	backend.Client
}
