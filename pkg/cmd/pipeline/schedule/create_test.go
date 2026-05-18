package schedule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/schedule"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdScheduleCreate_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := schedule.NewCmdCreate(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("cron"))
	assert.NotNil(t, cmd.Flag("branch"))
	assert.NotNil(t, cmd.Flag("enabled"))
}

func TestScheduleCreate_PassesInputToAPI(t *testing.T) {
	t.Parallel()
	var gotInput backend.PipelineScheduleInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePipelineScheduleFn: func(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			gotInput = input
			return backend.PipelineSchedule{
				UUID:           "new-uuid",
				Enabled:        true,
				CronExpression: "0 0 * * *",
				Branch:         "main",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--cron", "0 0 * * *", "--branch", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "0 0 * * *", gotInput.CronExpression)
	assert.Equal(t, "main", gotInput.Branch)
	assert.True(t, gotInput.Enabled)
	assert.Contains(t, out.String(), "new-uuid")
}

func TestScheduleCreate_MissingCron_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestScheduleCreate_MissingBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--cron", "0 0 * * *"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestScheduleCreate_EnabledFalse(t *testing.T) {
	t.Parallel()
	var gotEnabled bool
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePipelineScheduleFn: func(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error) {
			gotEnabled = input.Enabled
			return backend.PipelineSchedule{UUID: "x", Branch: "main", CronExpression: "0 0 * * *"}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--cron", "0 0 * * *", "--branch", "main", "--enabled=false"})
	require.NoError(t, cmd.Execute())
	assert.False(t, gotEnabled)
}

func TestScheduleCreate_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineScheduleFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--cron", "0 0 * * *", "--branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline schedules")
}
