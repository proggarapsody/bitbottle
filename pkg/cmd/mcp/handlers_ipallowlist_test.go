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

// ── list_workspace_ipallowlists ───────────────────────────────────────────────

func TestListWorkspaceIPAllowlists_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIPAllowlistsFn: func(workspace string) ([]backend.IPAllowlist, error) {
			assert.Equal(t, "myws", workspace)
			return []backend.IPAllowlist{
				{UUID: "entry-uuid-1", CIDR: "10.0.0.0/8", Description: "Corp VPN", Enabled: true},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspaceIPAllowlists(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "entry-uuid-1", "10.0.0.0/8")
}

func TestListWorkspaceIPAllowlists_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceIPAllowlists(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspaceIPAllowlists_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIPAllowlistsFn: func(workspace string) ([]backend.IPAllowlist, error) {
			return nil, errors.New("403 Forbidden")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspaceIPAllowlists(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── create_workspace_ipallowlist ─────────────────────────────────────────────

func TestCreateWorkspaceIPAllowlist_Success(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	var gotInput backend.CreateIPAllowlistInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateIPAllowlistFn: func(workspace string, in backend.CreateIPAllowlistInput) (backend.IPAllowlist, error) {
			gotWorkspace = workspace
			gotInput = in
			return backend.IPAllowlist{
				UUID:        "new-entry-uuid",
				CIDR:        in.CIDR,
				Description: in.Description,
				Enabled:     in.Enabled,
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"workspace":   "myws",
		"cidr":        "192.168.0.0/16",
		"description": "Office network",
		"enabled":     "true",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myws", gotWorkspace)
	assert.Equal(t, "192.168.0.0/16", gotInput.CIDR)
	assert.Equal(t, "Office network", gotInput.Description)
	assert.True(t, gotInput.Enabled)
	assertJSONContains(t, result, "new-entry-uuid", "192.168.0.0/16")
}

func TestCreateWorkspaceIPAllowlist_DefaultEnabled(t *testing.T) {
	t.Parallel()
	var gotEnabled bool
	fake := &testhelpers.FakeClient{
		T: t,
		CreateIPAllowlistFn: func(workspace string, in backend.CreateIPAllowlistInput) (backend.IPAllowlist, error) {
			gotEnabled = in.Enabled
			return backend.IPAllowlist{UUID: "uuid-x", CIDR: in.CIDR, Enabled: in.Enabled}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"cidr":      "10.0.0.0/8",
		// no "enabled" param — should default to true
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.True(t, gotEnabled)
}

func TestCreateWorkspaceIPAllowlist_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"cidr": "10.0.0.0/8",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestCreateWorkspaceIPAllowlist_MissingCIDR(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "cidr")
}

// ── delete_workspace_ipallowlist ─────────────────────────────────────────────

func TestDeleteWorkspaceIPAllowlist_Success(t *testing.T) {
	t.Parallel()
	var gotWorkspace, gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteIPAllowlistFn: func(workspace, uuid string) error {
			gotWorkspace = workspace
			gotUUID = uuid
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"uuid":      "entry-uuid-1",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myws", gotWorkspace)
	assert.Equal(t, "entry-uuid-1", gotUUID)
	assertJSONContains(t, result, "deleted", "entry-uuid-1")
}

func TestDeleteWorkspaceIPAllowlist_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"uuid": "entry-uuid-1",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestDeleteWorkspaceIPAllowlist_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

func TestDeleteWorkspaceIPAllowlist_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteIPAllowlistFn: func(workspace, uuid string) error {
			return errors.New("404 Not Found")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteWorkspaceIPAllowlist(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"uuid":      "ghost-uuid",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
