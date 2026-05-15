package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── listWorkspaceHooks ────────────────────────────────────────────────────────

func TestListWorkspaceHooks_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceWebhooksFn: func(workspace string) ([]backend.Webhook, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.Webhook{
				{ID: "ws-uuid-1", URL: "https://example.com/hook", Active: true, Events: []string{"repo:push"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspaceHooks(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "ws-uuid-1", "")
	assertJSONContains(t, result, "https://example.com/hook", "")
}

func TestListWorkspaceHooks_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceHooks(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspaceHooks_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceWebhooksFn: func(workspace string) ([]backend.Webhook, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspaceHooks(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── createWorkspaceHook ───────────────────────────────────────────────────────

func TestCreateWorkspaceHook_Success(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	var gotInput backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWorkspaceWebhookFn: func(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			gotWorkspace = workspace
			gotInput = in
			return backend.Webhook{ID: "new-ws-uuid", URL: in.URL, Active: in.Active, Events: in.Events}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createWorkspaceHook(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"url":       "https://example.com/hook",
		"events":    "repo:push,pullrequest:created",
		"active":    true,
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "https://example.com/hook", gotInput.URL)
	assert.True(t, gotInput.Active)
	assert.Contains(t, gotInput.Events, "repo:push")
	assert.Contains(t, gotInput.Events, "pullrequest:created")
	assertJSONContains(t, result, "new-ws-uuid", "")
}

func TestCreateWorkspaceHook_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createWorkspaceHook(context.Background(), makeReq(map[string]any{
		"url":    "https://example.com/hook",
		"events": "repo:push",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestCreateWorkspaceHook_MissingURL(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createWorkspaceHook(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"events":    "repo:push",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "url")
}

func TestCreateWorkspaceHook_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWorkspaceWebhookFn: func(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			return backend.Webhook{}, errors.New("422 unprocessable entity")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createWorkspaceHook(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"url":       "https://example.com/hook",
		"events":    "repo:push",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── deleteWorkspaceHook ───────────────────────────────────────────────────────

func TestDeleteWorkspaceHook_Success(t *testing.T) {
	t.Parallel()
	var gotWorkspace, gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWorkspaceWebhookFn: func(workspace, uuid string) error {
			gotWorkspace = workspace
			gotUUID = uuid
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteWorkspaceHook(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"uuid":      "some-uuid-1",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "some-uuid-1", gotUUID)
	assertJSONContains(t, result, "some-uuid-1", "")
}

func TestDeleteWorkspaceHook_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteWorkspaceHook(context.Background(), makeReq(map[string]any{
		"uuid": "some-uuid-1",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestDeleteWorkspaceHook_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteWorkspaceHook(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

func TestDeleteWorkspaceHook_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWorkspaceWebhookFn: func(workspace, uuid string) error {
			return errors.New("404 not found")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteWorkspaceHook(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"uuid":      "some-uuid-1",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
