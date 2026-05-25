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

// ── put_file ──────────────────────────────────────────────────────────────────

func TestPutFile_OK(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotPath string
	var gotIn backend.PutFileInput
	fake := &testhelpers.FakeClient{
		PutFileFn: func(ns, slug, path string, in backend.PutFileInput) error {
			gotNS, gotSlug, gotPath = ns, slug, path
			gotIn = in
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	factorytest.UseBackend(f, fake)
	h := newHandlers(f)
	result, err := h.putFile(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-svc",
		"path":    "README.md",
		"branch":  "main",
		"message": "Update README",
		"content": "# Hello",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "committed", "")
	assertJSONContains(t, result, "README.md", "")
	assert.Equal(t, "myws", gotNS)
	assert.Equal(t, "my-svc", gotSlug)
	assert.Equal(t, "README.md", gotPath)
	assert.Equal(t, "main", gotIn.Branch)
	assert.Equal(t, "Update README", gotIn.Message)
	assert.Equal(t, "# Hello", gotIn.Content)
}

func TestPutFile_WithSourceCommit(t *testing.T) {
	t.Parallel()
	var gotSourceCommit string
	fake := &testhelpers.FakeClient{
		PutFileFn: func(ns, slug, path string, in backend.PutFileInput) error {
			gotSourceCommit = in.SourceCommit
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	factorytest.UseBackend(f, fake)
	h := newHandlers(f)
	result, err := h.putFile(context.Background(), makeReq(map[string]any{
		"project":       "myws",
		"slug":          "my-svc",
		"path":          "README.md",
		"branch":        "main",
		"message":       "Update",
		"content":       "x",
		"source_commit": "deadbeef",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "deadbeef", gotSourceCommit)
}

func TestPutFile_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	h := newHandlers(f)
	result, err := h.putFile(context.Background(), makeReq(map[string]any{
		"slug": "my-svc", "path": "README.md", "branch": "main", "message": "x", "content": "x",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestPutFile_MissingBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	h := newHandlers(f)
	result, err := h.putFile(context.Background(), makeReq(map[string]any{
		"project": "myws", "slug": "my-svc", "path": "README.md", "message": "x", "content": "x",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "branch")
}

func TestPutFile_UnsupportedOnBackend_ReturnsError(t *testing.T) {
	t.Parallel()
	// Plain Client without SourceWriter
	type noWriteFake struct{ backend.Client }
	wrapped := noWriteFake{Client: &testhelpers.FakeClient{}}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoToolsCfg})
	factorytest.UseBackend(f, wrapped)
	h := newHandlers(f)
	result, err := h.putFile(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-svc",
		"path":    "README.md",
		"branch":  "main",
		"message": "Update",
		"content": "x",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}
