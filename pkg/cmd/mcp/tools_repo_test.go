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

const repoToolsCfg = "git.example.com:\n  oauth_token: tok\n"

func TestGetFileContent_ReturnsBytesAsText(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetFileContentFn: func(ns, slug, ref, path string) ([]byte, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-svc", slug)
			assert.Equal(t, "main", ref)
			assert.Equal(t, "README.md", path)
			return []byte("hello world"), nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	factorytest.UseBackend(f, fake)
	h := newHandlers(f)
	result, err := h.getFileContent(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-svc",
		"ref":     "main",
		"path":    "README.md",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "hello world", extractText(t, result))
}

func TestGetFileContent_RequiresPath(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	h := newHandlers(f)
	result, err := h.getFileContent(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-svc",
		"ref":     "main",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError, "missing path must surface as a tool-result error")
}

func TestListTree_Root_OmitPathArg(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListTreeFn: func(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
			assert.Equal(t, "", path, "absent path arg means repo root")
			return []backend.TreeEntry{
				{Path: "README.md", Type: "file", Size: 1234},
				{Path: "cmd", Type: "dir"},
			}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	factorytest.UseBackend(f, fake)
	h := newHandlers(f)
	result, err := h.listTree(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-svc",
		"ref":     "main",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "README.md", "")
	assertJSONContains(t, result, "cmd", "")
}

func TestListTree_NotFound_ReturnsTypedEnvelope(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListTreeFn: func(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
			return nil, &backend.DomainError{
				Kind:    backend.ErrNotFound,
				Code:    "repo.not_found",
				Message: "no such path",
			}
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	factorytest.UseBackend(f, fake)
	h := newHandlers(f)
	result, err := h.listTree(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-svc",
		"ref":     "main",
		"path":    "missing",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo.not_found")
}
