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

func TestListSSHKeys_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSSHKeysFn: func() ([]backend.SSHKey, error) {
			return []backend.SSHKey{
				{ID: 1, Label: "Laptop key", Key: "ssh-rsa AAAA1"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listSSHKeys(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "Laptop key", "ssh-rsa")
}

func TestListSSHKeys_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	// Use a client wrapper that does NOT implement SSHKeyClient.
	type noSSHFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleCloudConfig})
	factorytest.UseBackend(f, noSSHFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listSSHKeys(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestAddSSHKey_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddSSHKeyFn: func(input backend.SSHKeyInput) (backend.SSHKey, error) {
			return backend.SSHKey{ID: 42, Label: input.Label, Key: input.Key}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addSSHKey(context.Background(), makeReq(map[string]any{
		"key":   "ssh-rsa AAAA",
		"label": "CI",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "42", "CI")
}

func TestAddSSHKey_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addSSHKey(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

func TestDeleteSSHKey_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSSHKeyFn: func(id int) error {
			assert.Equal(t, 5, id)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteSSHKey(context.Background(), makeReq(map[string]any{
		"id": float64(5),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeleteSSHKey_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteSSHKey(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}
