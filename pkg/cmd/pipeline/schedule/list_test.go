package schedule_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/schedule"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdScheduleList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := schedule.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestScheduleList_PrintsSchedules(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineSchedulesFn: func(ns, slug string) ([]backend.PipelineSchedule, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.PipelineSchedule{
				{UUID: "sched-1", Enabled: true, CronExpression: "0 0 * * *", Branch: "main"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "sched-1")
	assert.Contains(t, out.String(), "0 0 * * *")
	assert.Contains(t, out.String(), "main")
}

func TestScheduleList_EmptyList(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineSchedulesFn: func(ns, slug string) ([]backend.PipelineSchedule, error) {
			return []backend.PipelineSchedule{}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "No pipeline schedules found")
}

func TestScheduleList_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineScheduleFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline schedules")
}

func TestScheduleList_PartialResults(t *testing.T) {
	t.Parallel()
	listErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineSchedulesFn: func(ns, slug string) ([]backend.PipelineSchedule, error) {
			return []backend.PipelineSchedule{
				{UUID: "partial-sched", Enabled: true, CronExpression: "0 0 * * *", Branch: "main"},
			}, listErr
		},
	}
	f, out, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := schedule.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "partial-sched")
	assert.Contains(t, errOut.String(), "warning: partial results")
}

// noPipelineScheduleFake wraps backend.Client without implementing
// backend.PipelineScheduleClient — simulates a Bitbucket Server backend.
type noPipelineScheduleFake struct {
	backend.Client
}
