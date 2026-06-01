package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func fakeRepoHookHandlers(t *testing.T, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	return newHandlersWithFake(t, singleHostConfig, fake)
}

// ── TestHandleListRepoHooks ───────────────────────────────────────────────────

func TestHandleListRepoHooks(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListRepoHooksFn: func(project, slug string) ([]backend.RepoHook, error) {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "myrepo", slug)
			return []backend.RepoHook{
				{Key: "com.example.hook", Name: "My Hook", Enabled: true},
			}, nil
		},
	}
	h := fakeRepoHookHandlers(t, fake)
	result, err := h.listRepoHooks(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "myrepo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "com.example.hook", "My Hook")
}

// ── TestHandleViewRepoHook ────────────────────────────────────────────────────

func TestHandleViewRepoHook(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetRepoHookFn: func(project, slug, hookKey string) (backend.RepoHook, error) {
			assert.Equal(t, "hook-key-1", hookKey)
			return backend.RepoHook{Key: hookKey, Name: "Hook One", Enabled: false, Configured: true}, nil
		},
	}
	h := fakeRepoHookHandlers(t, fake)
	result, err := h.viewRepoHook(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hook_key": "hook-key-1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "hook-key-1", "Hook One")
}

// ── TestHandleEnableRepoHook ──────────────────────────────────────────────────

func TestHandleEnableRepoHook(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		EnableRepoHookFn: func(project, slug, hookKey string) error {
			called = true
			assert.Equal(t, "my-hook", hookKey)
			return nil
		},
	}
	h := fakeRepoHookHandlers(t, fake)
	result, err := h.enableRepoHook(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hook_key": "my-hook",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "enabled", "my-hook")
}

// ── TestHandleDisableRepoHook ─────────────────────────────────────────────────

func TestHandleDisableRepoHook(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		DisableRepoHookFn: func(project, slug, hookKey string) error {
			called = true
			assert.Equal(t, "my-hook", hookKey)
			return nil
		},
	}
	h := fakeRepoHookHandlers(t, fake)
	result, err := h.disableRepoHook(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hook_key": "my-hook",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "disabled", "my-hook")
}

// ── TestHandleGetRepoHookSettings ─────────────────────────────────────────────

func TestHandleGetRepoHookSettings(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"timeout":30,"retries":3}`)
	fake := &testhelpers.FakeClient{T: t,
		GetRepoHookSettingsFn: func(project, slug, hookKey string) (json.RawMessage, error) {
			assert.Equal(t, "cfg-hook", hookKey)
			return raw, nil
		},
	}
	h := fakeRepoHookHandlers(t, fake)
	result, err := h.getRepoHookSettings(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hook_key": "cfg-hook",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "timeout", "retries")
}

// ── TestHandleSetRepoHookSettings ─────────────────────────────────────────────

func TestHandleSetRepoHookSettings(t *testing.T) {
	t.Parallel()
	var gotCfg json.RawMessage
	fake := &testhelpers.FakeClient{T: t,
		SetRepoHookSettingsFn: func(project, slug, hookKey string, cfg json.RawMessage) error {
			gotCfg = cfg
			return nil
		},
	}
	h := fakeRepoHookHandlers(t, fake)
	result, err := h.setRepoHookSettings(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hook_key": "cfg-hook",
		"config": `{"timeout":60}`,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "ok", "cfg-hook")
	assert.NotNil(t, gotCfg)
	assert.Contains(t, string(gotCfg), "timeout")
}
