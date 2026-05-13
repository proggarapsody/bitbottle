package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdScheduleDelete_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := pipeline.NewCmdScheduleDelete(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestScheduleDelete_DeletesSchedule(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineScheduleFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "sched-abc", uuid)
			deleted = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pipeline.NewCmdScheduleDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "sched-abc"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, out.String(), "sched-abc")
}

func TestScheduleDelete_WithUUIDOnly(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineScheduleFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "sched-xyz", uuid)
			deleted = true
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pipeline.NewCmdScheduleDelete(f, nil)
	// Repo comes from the fake base repo; uuid is the only arg
	cmd.SetArgs([]string{"myws/my-repo", "sched-xyz"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
}

func TestScheduleDelete_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineScheduleFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pipeline.NewCmdScheduleDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "sched-abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline schedules")
}
