package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── SSH key pair ──────────────────────────────────────────────────────────────

func TestViewPipelineSSHKeyPair_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineSSHKeyPairFn: func(ns, slug string) (backend.PipelineSSHKeyPair, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return backend.PipelineSSHKeyPair{
				PublicKey:    "ssh-rsa AAAA...",
				KeyTypeLabel: "RSA",
				Created:      time.Now(),
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.viewPipelineSSHKeyPair(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "public_key", "RSA")
}

func TestViewPipelineSSHKeyPair_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewPipelineSSHKeyPair(context.Background(), makeReq(map[string]any{
		"slug": "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestViewPipelineSSHKeyPair_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noSSHFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noSSH := &noSSHFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noSSH)
	h := newHandlers(f)
	result, err := h.viewPipelineSSHKeyPair(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestRegeneratePipelineSSHKeyPair_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RegeneratePipelineSSHKeyPairFn: func(ns, slug string, bits int) (backend.PipelineSSHKeyPair, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 4096, bits)
			return backend.PipelineSSHKeyPair{
				PublicKey:    "ssh-rsa BBBB...",
				KeyTypeLabel: "RSA",
				Created:      time.Now(),
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.regeneratePipelineSSHKeyPair(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"bits":    4096,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "public_key", "RSA")
}

// ── Known hosts ───────────────────────────────────────────────────────────────

func TestListPipelineKnownHosts_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineKnownHostsFn: func(ns, slug string) ([]backend.PipelineKnownHost, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.PipelineKnownHost{
				{UUID: "uuid-1", Hostname: "github.com", PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listPipelineKnownHosts(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "github.com", "uuid-1")
}

func TestListPipelineKnownHosts_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPipelineKnownHosts(context.Background(), makeReq(map[string]any{
		"slug": "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListPipelineKnownHosts_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noKHFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noKH := &noKHFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noKH)
	h := newHandlers(f)
	result, err := h.listPipelineKnownHosts(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestViewPipelineKnownHost_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineKnownHostFn: func(ns, slug, uuid string) (backend.PipelineKnownHost, error) {
			assert.Equal(t, "uuid-1", uuid)
			return backend.PipelineKnownHost{
				UUID:      "uuid-1",
				Hostname:  "github.com",
				PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA", MD5: "aa:bb"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.viewPipelineKnownHost(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"uuid":    "uuid-1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "github.com", "uuid-1")
}

func TestViewPipelineKnownHost_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewPipelineKnownHost(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

func TestAddPipelineKnownHost_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddPipelineKnownHostFn: func(ns, slug string, in backend.PipelineKnownHostInput) (backend.PipelineKnownHost, error) {
			assert.Equal(t, "github.com", in.Hostname)
			assert.Equal(t, "RSA", in.PublicKey.KeyType)
			return backend.PipelineKnownHost{
				UUID:      "new-uuid",
				Hostname:  "github.com",
				PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addPipelineKnownHost(context.Background(), makeReq(map[string]any{
		"project":      "myws",
		"slug":         "my-repo",
		"hostname_arg": "github.com",
		"key":          "AAAA...",
		"key_type":     "RSA",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "new-uuid", "github.com")
}

func TestAddPipelineKnownHost_MissingHostnameArg(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addPipelineKnownHost(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "hostname_arg")
}

func TestDeletePipelineKnownHost_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineKnownHostFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "uuid-1", uuid)
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deletePipelineKnownHost(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"uuid":    "uuid-1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "uuid-1")
}

func TestDeletePipelineKnownHost_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deletePipelineKnownHost(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}
