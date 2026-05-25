package mcp

import (
	"context"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func fakeAdminHandlers(t *testing.T, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	return newHandlersWithFake(t, singleHostConfig, fake)
}

// ── rotate_secrets ────────────────────────────────────────────────────────────

func TestMCP_RotateSecrets_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		RotateSecretsFn: func() error {
			called = true
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.rotateSecrets(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "rotated", "")
	assert.True(t, called)
}

func TestMCP_RotateSecrets_PermissionError(t *testing.T) {
	t.Parallel()
	permErr := &backend.DomainError{Kind: backend.ErrPermission, Message: "forbidden"}
	fake := &testhelpers.FakeClient{T: t,
		RotateSecretsFn: func() error { return permErr },
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.rotateSecrets(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}

// ── get_logging_config ────────────────────────────────────────────────────────

func TestMCP_GetLoggingConfig_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetLoggingConfigFn: func() (backend.LoggingConfig, error) {
			return backend.LoggingConfig{Level: "WARN", Async: true}, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getLoggingConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "WARN", "")
	assertJSONContains(t, result, `"async":true`, "")
}

// ── set_logging_config ────────────────────────────────────────────────────────

func TestMCP_SetLoggingConfig_Level_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.LoggingConfigInput
	fake := &testhelpers.FakeClient{T: t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error {
			gotIn = in
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setLoggingConfig(context.Background(), makeReq(map[string]any{
		"level": "DEBUG",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "updated", "")
	assert.Equal(t, "DEBUG", gotIn.Level)
}

func TestMCP_SetLoggingConfig_Async_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.LoggingConfigInput
	fake := &testhelpers.FakeClient{T: t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error {
			gotIn = in
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setLoggingConfig(context.Background(), makeReq(map[string]any{
		"async": true,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "updated", "")
	assert.True(t, gotIn.Async)
}

func TestMCP_SetLoggingConfig_Persistent_NotePersistent(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error { return nil },
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setLoggingConfig(context.Background(), makeReq(map[string]any{
		"level": "ERROR", "persistent": true,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "persistent", "")
}

func TestMCP_SetLoggingConfig_NoParams_ReturnsError(t *testing.T) {
	t.Parallel()
	h := fakeAdminHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.setLoggingConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "at least one of")
}

func TestMCP_SetLoggingConfig_InvalidLevel_ReturnsError(t *testing.T) {
	t.Parallel()
	h := fakeAdminHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.setLoggingConfig(context.Background(), makeReq(map[string]any{
		"level": "verbose",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "case-sensitive")
}

// ── list_admin_users ──────────────────────────────────────────────────────────

func TestMCP_ListAdminUsers_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListAdminUsersFn: func(filter string, limit int) ([]backend.AdminUser, error) {
			return []backend.AdminUser{
				{Slug: "alice", DisplayName: "Alice A", Email: "alice@example.com", Active: true, Type: "NORMAL"},
			}, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.listAdminUsers(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "")
}

func TestMCP_ListAdminUsers_WithFilter(t *testing.T) {
	t.Parallel()
	var gotFilter string
	fake := &testhelpers.FakeClient{T: t,
		ListAdminUsersFn: func(filter string, limit int) ([]backend.AdminUser, error) {
			gotFilter = filter
			return nil, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.listAdminUsers(context.Background(), makeReq(map[string]any{"filter": "ali"}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.Equal(t, "ali", gotFilter)
}

// ── activate_user ─────────────────────────────────────────────────────────────

func TestMCP_ActivateUser_OK(t *testing.T) {
	t.Parallel()
	var gotSlug string
	fake := &testhelpers.FakeClient{T: t,
		ActivateUserFn: func(slug string) error {
			gotSlug = slug
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.activateUser(context.Background(), makeReq(map[string]any{"slug": "alice"}))
	require.NoError(t, err)
	assertJSONContains(t, result, "activated", "")
	assert.Equal(t, "alice", gotSlug)
}

func TestMCP_ActivateUser_MissingSlug(t *testing.T) {
	t.Parallel()
	h := fakeAdminHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.activateUser(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

// ── deactivate_user ───────────────────────────────────────────────────────────

func TestMCP_DeactivateUser_OK(t *testing.T) {
	t.Parallel()
	var gotSlug string
	fake := &testhelpers.FakeClient{T: t,
		DeactivateUserFn: func(slug string) error {
			gotSlug = slug
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.deactivateUser(context.Background(), makeReq(map[string]any{"slug": "bob"}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deactivated", "")
	assert.Equal(t, "bob", gotSlug)
}

// ── rename_user ───────────────────────────────────────────────────────────────

func TestMCP_RenameUser_OK(t *testing.T) {
	t.Parallel()
	var gotSlug, gotNewSlug string
	fake := &testhelpers.FakeClient{T: t,
		RenameUserFn: func(slug, newSlug string) error {
			gotSlug = slug
			gotNewSlug = newSlug
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.renameUser(context.Background(), makeReq(map[string]any{
		"slug": "old-alice", "new_slug": "alice",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "renamed", "")
	assert.Equal(t, "old-alice", gotSlug)
	assert.Equal(t, "alice", gotNewSlug)
}

func TestMCP_RenameUser_MissingNewSlug(t *testing.T) {
	t.Parallel()
	h := fakeAdminHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.renameUser(context.Background(), makeReq(map[string]any{"slug": "alice"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "new_slug")
}

// ── get_admin_license ─────────────────────────────────────────────────────────

func TestMCP_GetAdminLicense_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetLicenseFn: func() (backend.AdminLicense, error) {
			return backend.AdminLicense{
				Tier:     "ENTERPRISE",
				Users:    500,
				ServerId: "srv-abc123",
			}, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getAdminLicense(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "ENTERPRISE", "")
	assertJSONContains(t, result, "srv-abc123", "")
}

// ── get_cluster_nodes ─────────────────────────────────────────────────────────

func TestMCP_GetClusterNodes_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetClusterNodesFn: func() ([]backend.ClusterNode, error) {
			return []backend.ClusterNode{
				{NodeId: "node-1", Name: "Primary", Address: "10.0.0.1", State: "ACTIVE", Local: true},
			}, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getClusterNodes(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "node-1", "")
	assertJSONContains(t, result, "Primary", "")
}

// ── get_mail_server_config ────────────────────────────────────────────────────

func TestMCP_GetMailServerConfig_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetMailServerConfigFn: func() (backend.MailServerConfig, error) {
			return backend.MailServerConfig{
				Hostname:      "smtp.example.com",
				Port:          25,
				Protocol:      "smtp",
				UseStartTLS:   false,
				Username:      "mailer",
				SenderAddress: "no-reply@example.com",
			}, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getMailServerConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "smtp.example.com", "")
	assertJSONContains(t, result, "no-reply@example.com", "")
	// Password must never be in output
	text := extractText(t, result)
	assert.NotContains(t, text, "Password")
}

func TestMCP_GetMailServerConfig_Error(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetMailServerConfigFn: func() (backend.MailServerConfig, error) {
			return backend.MailServerConfig{}, &backend.DomainError{
				Kind:    backend.ErrPermission,
				Message: "access denied",
			}
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getMailServerConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}

// ── set_mail_server_config ────────────────────────────────────────────────────

func TestMCP_SetMailServerConfig_OK(t *testing.T) {
	t.Parallel()
	var gotCfg backend.MailServerConfig
	fake := &testhelpers.FakeClient{T: t,
		SetMailServerConfigFn: func(in backend.MailServerConfig) error {
			gotCfg = in
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setMailServerConfig(context.Background(), makeReq(map[string]any{
		"mail_hostname":  "smtp.example.com",
		"port":           float64(465),
		"protocol":       "smtps",
		"sender_address": "bot@example.com",
		"username":       "mailer",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "updated", "")
	assertJSONContains(t, result, "smtp.example.com", "")
	assert.Equal(t, "smtp.example.com", gotCfg.Hostname)
	assert.Equal(t, 465, gotCfg.Port)
	assert.Equal(t, "smtps", gotCfg.Protocol)
	assert.Equal(t, "bot@example.com", gotCfg.SenderAddress)
	assert.Equal(t, "mailer", gotCfg.Username)
}

func TestMCP_SetMailServerConfig_MissingMailHostname(t *testing.T) {
	t.Parallel()
	h := fakeAdminHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.setMailServerConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "mail_hostname")
}

// ── noAdminClient returns unsupported ────────────────────────────────────────

// noAdminClientWrapper wraps backend.Client without satisfying AdminClient,
// simulating a Cloud backend invocation.
type noAdminClientWrapper struct{ backend.Client }

func TestMCP_RotateSecrets_Unsupported(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	wrapper := noAdminClientWrapper{Client: &testhelpers.FakeClient{T: t}}
	factorytest.UseBackend(f, wrapper)
	h := newHandlers(f)

	result, err := h.rotateSecrets(context.Background(), makeReq(nil))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError, "expected error result for unsupported backend")
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "host.unsupported")
	assert.Contains(t, text.Text, "admin")
}
