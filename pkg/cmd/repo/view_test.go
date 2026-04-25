package repo_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoView_HasWebFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoView(f)
	assert.NotNil(t, cmd.Flag("web"))
}

func TestNewCmdRepoView_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoView(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoView_PrintsRepoDetails(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(
				testhelpers.BackendRepoWithSlug("my-service"),
				testhelpers.BackendRepoWithWebURL("https://bb.example.com/projects/MYPROJ/repos/my-service/browse"),
			), nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoView(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "my-service")
	assert.Contains(t, got, "browse")
}

func TestRepoView_WebFlag_OpensBrowser(t *testing.T) {
	t.Parallel()

	url := "https://bb.example.com/projects/MYPROJ/repos/my-service/browse"
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(
				testhelpers.BackendRepoWithSlug("my-service"),
				testhelpers.BackendRepoWithWebURL(url),
			), nil
		},
	}
	browser := &testhelpers.FakeBrowserLauncher{}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   repoConfig,
		BackendOverride: fake,
		Browser:         browser,
	})
	cmd := repo.NewCmdRepoView(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--web"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1, "expected browser to be invoked once")
	assert.Equal(t, url, browser.URLs[0])
}

func TestRepoView_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("404 not found")
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{}, apiErr
		},
	}

	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoView(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
