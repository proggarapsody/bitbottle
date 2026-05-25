package mcp

import (
	"context"
	"encoding/base64"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListRepoDownloads_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoDownloadsFn: func(ns, slug string, limit int) ([]backend.RepoDownload, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.RepoDownload{
				{Name: "release.zip", Size: 1024, Downloads: 5},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listRepoDownloads(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "name", "release.zip")
}

func TestListRepoDownloads_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listRepoDownloads(context.Background(), makeReq(map[string]any{
		"slug": "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListRepoDownloads_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noDownloadFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noRD := &noDownloadFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noRD)
	h := newHandlers(f)
	result, err := h.listRepoDownloads(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestUploadRepoDownload_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UploadRepoDownloadFn: func(ns, slug, name string, _ io.Reader) (backend.RepoDownload, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "release.zip", name)
			return backend.RepoDownload{Name: name}, nil
		},
	}
	content := base64.StdEncoding.EncodeToString([]byte("fake file content"))
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.uploadRepoDownload(context.Background(), makeReq(map[string]any{
		"project":             "myws",
		"slug":                "my-repo",
		"name":                "release.zip",
		"file_content_base64": content,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "name", "release.zip")
}

func TestUploadRepoDownload_InvalidBase64(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.uploadRepoDownload(context.Background(), makeReq(map[string]any{
		"project":             "myws",
		"slug":                "my-repo",
		"name":                "release.zip",
		"file_content_base64": "!!!not-base64!!!",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "base64")
}

func TestDeleteRepoDownload_Success(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRepoDownloadFn: func(ns, slug, name string) error {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "release.zip", name)
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteRepoDownload(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"name":    "release.zip",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	text := extractText(t, result)
	assert.Contains(t, text, "Deleted")
}

func TestDeleteRepoDownload_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteRepoDownload(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}
