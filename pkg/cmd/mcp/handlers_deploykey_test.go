package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListDeployKeys_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListDeployKeysFn: func(ns, slug string) ([]backend.DeployKey, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.DeployKey{
				{ID: 1, Label: "CI key", Key: "ssh-rsa AAAA1"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listDeployKeys(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "CI key", "ssh-rsa")
}

func TestListDeployKeys_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listDeployKeys(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestAddDeployKey_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			return backend.DeployKey{ID: 42, Label: input.Label, Key: input.Key}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addDeployKey(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "ssh-rsa AAAA",
		"label": "CI",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "42", "CI")
}

func TestAddDeployKey_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addDeployKey(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

func TestDeleteDeployKey_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteDeployKeyFn: func(ns, slug string, id int) error {
			assert.Equal(t, 5, id)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteDeployKey(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"id":   float64(5),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeleteDeployKey_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteDeployKey(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestAddDeployKey_DefaultPermissionIsRead(t *testing.T) {
	t.Parallel()
	var gotInput backend.DeployKeyInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			gotInput = input
			return backend.DeployKey{ID: 7, Label: input.Label, Key: input.Key}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addDeployKey(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"key":  "ssh-rsa BBBB",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "7", "")
	assert.Equal(t, "read", gotInput.Permission)
}

func TestAddDeployKey_InvalidPermission(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			t.Fatal("AddDeployKey should not be called on invalid permission")
			return backend.DeployKey{}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addDeployKey(context.Background(), makeReq(map[string]any{
		"repo":       "myws/my-repo",
		"key":        "ssh-rsa CCCC",
		"permission": "write",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}

func TestAddDeployKey_WithPermission(t *testing.T) {
	t.Parallel()
	var gotInput backend.DeployKeyInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			gotInput = input
			return backend.DeployKey{ID: 9, Label: input.Label, Key: input.Key}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addDeployKey(context.Background(), makeReq(map[string]any{
		"repo":       "myws/my-repo",
		"key":        "ssh-rsa ZZZZ",
		"permission": "read-write",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "9", "")
	assert.Equal(t, "read_write", gotInput.Permission)
}
