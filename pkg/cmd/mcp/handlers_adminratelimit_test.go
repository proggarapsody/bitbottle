package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── get_rate_limit_config ─────────────────────────────────────────────────────

func TestMCP_GetRateLimitConfig_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{Enabled: true, RequestsPerHour: 3600, ThrottleWaitMS: 500}, nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getRateLimitConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, `"enabled":true`, "")
	assertJSONContains(t, result, `"requests_per_hour":3600`, "")
	assertJSONContains(t, result, `"throttle_wait_ms":500`, "")
}

func TestMCP_GetRateLimitConfig_Error(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{}, &backend.DomainError{
				Kind:    backend.ErrPermission,
				Message: "access denied",
			}
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.getRateLimitConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}

// ── set_rate_limit_config ─────────────────────────────────────────────────────

func TestMCP_SetRateLimitConfig_Enabled_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{Enabled: false, RequestsPerHour: 1000, ThrottleWaitMS: 100}, nil
		},
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setRateLimitConfig(context.Background(), makeReq(map[string]any{
		"enabled": true,
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.True(t, gotIn.Enabled)
	assert.Equal(t, 1000, gotIn.RequestsPerHour) // unchanged
}

func TestMCP_SetRateLimitConfig_RequestsPerHour_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{Enabled: true, RequestsPerHour: 500, ThrottleWaitMS: 0}, nil
		},
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setRateLimitConfig(context.Background(), makeReq(map[string]any{
		"requests_per_hour": float64(7200),
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.Equal(t, 7200, gotIn.RequestsPerHour)
	assert.True(t, gotIn.Enabled) // unchanged
}

func TestMCP_SetRateLimitConfig_NoParams_ReturnsError(t *testing.T) {
	t.Parallel()
	h := fakeAdminHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.setRateLimitConfig(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "at least one of")
}

func TestMCP_SetRateLimitConfig_PermissionError(t *testing.T) {
	t.Parallel()
	permErr := &backend.DomainError{Kind: backend.ErrPermission, Message: "forbidden"}
	fake := &testhelpers.FakeClient{T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{}, nil
		},
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error { return permErr },
	}
	h := fakeAdminHandlers(t, fake)
	result, err := h.setRateLimitConfig(context.Background(), makeReq(map[string]any{
		"enabled": true,
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}
