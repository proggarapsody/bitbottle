package cache_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/cache"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cache.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestCacheList_PrintsCaches(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineCachesFn: func(ns, slug string) ([]backend.PipelineCache, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.PipelineCache{
				{UUID: "cache-1", Name: "node_modules", Path: "/app/node_modules", FileSizeBytes: 12345678, CreatedOn: "2024-01-01T00:00:00.000Z"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "cache-1")
	assert.Contains(t, out.String(), "node_modules")
	assert.Contains(t, out.String(), "/app/node_modules")
}

func TestCacheList_EmptyList(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineCachesFn: func(ns, slug string) ([]backend.PipelineCache, error) {
			return []backend.PipelineCache{}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "No pipeline caches found")
}

func TestCacheList_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineCacheFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline cache")
}

func TestCacheList_PartialResults(t *testing.T) {
	t.Parallel()
	listErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineCachesFn: func(ns, slug string) ([]backend.PipelineCache, error) {
			return []backend.PipelineCache{
				{UUID: "partial-cache", Name: "node_modules", Path: "/app/node_modules", FileSizeBytes: 1024, CreatedOn: "2024-01-01T00:00:00.000Z"},
			}, listErr
		},
	}
	f, out, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "partial-cache")
	assert.Contains(t, errOut.String(), "warning: partial results")
}

// noPipelineCacheFake wraps backend.Client without implementing
// backend.PipelineCacheClient — simulates a Bitbucket Server backend.
type noPipelineCacheFake struct {
	backend.Client
}
