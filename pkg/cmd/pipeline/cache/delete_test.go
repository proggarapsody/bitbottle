package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/cache"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdDelete_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cache.NewCmdDelete(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestCacheDelete_DeletesCache(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineCacheFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "cache-abc", uuid)
			deleted = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "cache-abc"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, out.String(), "cache-abc")
}

func TestCacheDelete_WithUUIDOnly(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineCacheFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "cache-xyz", uuid)
			deleted = true
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myws/my-repo", "cache-xyz"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
}

func TestCacheDelete_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineCacheFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cache.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "cache-abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline cache")
}
