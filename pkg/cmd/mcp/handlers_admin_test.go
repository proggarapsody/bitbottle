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
