package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// singleCloudConfig is a Cloud-host-only config for deployment tests.
const singleCloudConfig = "bitbucket.org:\n  oauth_token: tok\n"

func TestSplitRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		wantNS   string
		wantSlug string
		wantErr  bool
	}{
		{"myws/my-repo", "myws", "my-repo", false},
		{"myws/", "", "", true},
		{"/my-repo", "", "", true},
		{"noslash", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			ns, slug, err := splitRepo(tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantNS, ns)
				assert.Equal(t, tc.wantSlug, slug)
			}
		})
	}
}

func TestListDeployments_Success(t *testing.T) {
	t.Parallel()
	d := backend.Deployment{UUID: "dep-1", State: "COMPLETED"}
	d.Environment.Name = "Production"
	d.Release.Name = "v1.0"
	fake := &testhelpers.FakeClient{
		T: t,
		ListDeploymentsFn: func(ns, slug string, limit int) ([]backend.Deployment, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.Deployment{d}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listDeployments(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "dep-1", "COMPLETED")
}

func TestListDeployments_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listDeployments(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestGetDeployment_Success(t *testing.T) {
	t.Parallel()
	dep := backend.Deployment{UUID: "dep-abc", State: "FAILED"}
	fake := &testhelpers.FakeClient{
		T: t,
		GetDeploymentFn: func(ns, slug, uuid string) (backend.Deployment, error) {
			return dep, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getDeployment(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"uuid": "dep-abc",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "dep-abc", "FAILED")
}

func TestListEnvironments_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvironmentsFn: func(ns, slug string) ([]backend.Environment, error) {
			return []backend.Environment{
				{UUID: "env-1", Name: "Production", Type: "Production", Rank: 1},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listEnvironments(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "Production", "")
}

func TestCreateEnvironment_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateEnvironmentFn: func(ns, slug string, in backend.CreateEnvironmentInput) (backend.Environment, error) {
			return backend.Environment{UUID: "new-env", Name: in.Name, Type: in.Type}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createEnvironment(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"name": "QA",
		"type": "Test",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "new-env", "QA")
}

func TestDeleteEnvironment_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteEnvironmentFn: func(ns, slug, uuid string) error {
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteEnvironment(context.Background(), makeReq(map[string]any{
		"repo":     "myws/my-repo",
		"env_uuid": "env-123",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}
