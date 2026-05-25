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

// ── list_workspace_audit_log ──────────────────────────────────────────────────

func TestListWorkspaceAuditLog_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			assert.Equal(t, "myws", workspace)
			return []backend.AuditEvent{
				{
					Actor:     backend.AuditActor{AccountID: "aid-1", DisplayName: "Alice", NickName: "alice"},
					Action:    "workspace.member.create",
					Object:    backend.AuditObject{Type: "team", Name: "acme"},
					CreatedAt: "2024-01-15T10:00:00.000000+00:00",
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspaceAuditLog(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "workspace.member.create", "Alice")
}

func TestListWorkspaceAuditLog_ForwardsFilters(t *testing.T) {
	t.Parallel()
	var gotOpts backend.AuditLogOpts
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			gotOpts = opts
			return nil, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspaceAuditLog(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"action":    "workspace.member.create",
		"from":      "2024-01-01T00:00:00Z",
		"limit":     float64(10),
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.Equal(t, "workspace.member.create", gotOpts.Action)
	assert.Equal(t, "2024-01-01T00:00:00Z", gotOpts.From)
	assert.Equal(t, 10, gotOpts.Limit)
}

func TestListWorkspaceAuditLog_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceAuditLog(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspaceAuditLog_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAuditLogFn: func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
			return nil, errors.New("403 Forbidden")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspaceAuditLog(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
