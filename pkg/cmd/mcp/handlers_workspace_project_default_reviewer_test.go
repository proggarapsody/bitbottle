package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListProjectDefaultReviewers_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListProjectDefaultReviewersFn: func(workspace, projectKey string, limit int) ([]backend.ProjectDefaultReviewer, error) {
			assert.Equal(t, "myworkspace", workspace)
			assert.Equal(t, "PROJ", projectKey)
			return []backend.ProjectDefaultReviewer{
				{AccountID: "abc123", DisplayName: "Alice", Nickname: "alice"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listProjectDefaultReviewers(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "abc123", "Alice")
}

func TestListProjectDefaultReviewers_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listProjectDefaultReviewers(context.Background(), makeReq(map[string]any{
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListProjectDefaultReviewers_MissingProjectKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listProjectDefaultReviewers(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project_key")
}

func TestListProjectDefaultReviewers_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	type noDefaultReviewerFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, noDefaultReviewerFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listProjectDefaultReviewers(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestAddProjectDefaultReviewer_Success(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		AddProjectDefaultReviewerFn: func(workspace, projectKey, accountID string) error {
			gotWS, gotKey, gotUser = workspace, projectKey, accountID
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.addProjectDefaultReviewer(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
		"user":        "abc123",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "abc123", gotUser)
	assertJSONContains(t, result, "added", "abc123")
}

func TestAddProjectDefaultReviewer_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addProjectDefaultReviewer(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}

func TestRemoveProjectDefaultReviewer_Success(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		RemoveProjectDefaultReviewerFn: func(workspace, projectKey, accountID string) error {
			gotWS, gotKey, gotUser = workspace, projectKey, accountID
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.removeProjectDefaultReviewer(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
		"user":        "abc123",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "abc123", gotUser)
	assertJSONContains(t, result, "removed", "abc123")
}

func TestRemoveProjectDefaultReviewer_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeProjectDefaultReviewer(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}
