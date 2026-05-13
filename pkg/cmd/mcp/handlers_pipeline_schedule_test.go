package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListPipelineSchedules_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineSchedulesFn: func(ns, slug string) ([]backend.PipelineSchedule, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.PipelineSchedule{
				{UUID: "sched-1", Enabled: true, CronExpression: "0 0 * * *", Branch: "main"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listPipelineSchedules(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "sched-1", "0 0 * * *")
}

func TestListPipelineSchedules_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPipelineSchedules(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestCreatePipelineSchedule_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePipelineScheduleFn: func(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error) {
			assert.Equal(t, "0 12 * * 1", input.CronExpression)
			assert.Equal(t, "develop", input.Branch)
			assert.True(t, input.Enabled)
			return backend.PipelineSchedule{
				UUID:           "new-sched",
				Enabled:        true,
				CronExpression: "0 12 * * 1",
				Branch:         "develop",
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createPipelineSchedule(context.Background(), makeReq(map[string]any{
		"repo":    "myws/my-repo",
		"cron":    "0 12 * * 1",
		"branch":  "develop",
		"enabled": true,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "new-sched", "develop")
}

func TestCreatePipelineSchedule_MissingCron(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createPipelineSchedule(context.Background(), makeReq(map[string]any{
		"repo":   "myws/my-repo",
		"branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "cron")
}

func TestCreatePipelineSchedule_MissingBranch(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createPipelineSchedule(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"cron": "0 0 * * *",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "branch")
}

func TestDeletePipelineSchedule_Success(t *testing.T) {
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
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deletePipelineSchedule(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"uuid": "sched-xyz",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeletePipelineSchedule_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deletePipelineSchedule(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}
