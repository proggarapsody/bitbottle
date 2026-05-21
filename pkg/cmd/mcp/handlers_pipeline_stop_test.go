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

func TestStopPipeline_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		StopPipelineFn: func(ws, slug, uuid string) error {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc-123", uuid)
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.stopPipeline(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"uuid": "abc-123",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "abc-123", "")
}

func TestStopPipeline_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.stopPipeline(context.Background(), makeReq(map[string]any{
		"uuid": "abc-123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestStopPipeline_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.stopPipeline(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

func TestStopPipeline_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noPipelineFake struct {
		backend.Client
	}
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noStop := &noPipelineFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noStop)
	h := newHandlers(f)
	result, err := h.stopPipeline(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"uuid": "abc-123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}
