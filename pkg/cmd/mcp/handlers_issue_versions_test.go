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

func TestListIssueVersions_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssueVersionsFn: func(ns, slug string, limit int) ([]backend.IssueVersion, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.IssueVersion{
				{ID: 1, Name: "1.0"},
				{ID: 2, Name: "2.0"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listIssueVersions(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "name", "1.0")
}

func TestListIssueVersions_MissingSlug(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listIssueVersions(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestListIssueVersions_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noVersionFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noV := &noVersionFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noV)
	h := newHandlers(f)
	result, err := h.listIssueVersions(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestViewIssueVersion_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetIssueVersionFn: func(ns, slug string, id int) (backend.IssueVersion, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 1, id)
			return backend.IssueVersion{ID: 1, Name: "1.0"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.viewIssueVersion(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
		"id":        float64(1),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "name", "1.0")
}

func TestViewIssueVersion_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewIssueVersion(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestCreateIssueVersion_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateIssueVersionFn: func(ns, slug, name string) (backend.IssueVersion, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "3.0", name)
			return backend.IssueVersion{ID: 3, Name: "3.0"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createIssueVersion(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
		"name":      "3.0",
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "Created")
}

func TestCreateIssueVersion_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createIssueVersion(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestDeleteIssueVersion_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteIssueVersionFn: func(ns, slug string, id int) error {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 1, id)
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteIssueVersion(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
		"id":        float64(1),
	}))
	require.NoError(t, err)
	text2 := extractText(t, result)
	assert.Contains(t, text2, "Deleted")
}

func TestDeleteIssueVersion_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteIssueVersion(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"slug":      "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}
