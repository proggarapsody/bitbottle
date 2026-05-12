package repo_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoView_HasWebFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("web"))
}

func TestNewCmdRepoView_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
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

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, fake)
	f.Browser = browser
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestRepoView_ShowsDescription(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			r := testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("my-service"))
			r.Description = "bitbottle manual test"
			return r, nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "bitbottle manual test")
}

func TestRepoView_NoDescriptionLine_WhenEmpty(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("my-service")), nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	assert.NotContains(t, out.String(), "Description:")
}

func TestNewCmdRepoView_HasJSONAndJQFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestRepoView_JSON_EmitsObject(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(
				testhelpers.BackendRepoWithSlug("my-service"),
				testhelpers.BackendRepoWithWebURL("https://bb.example.com/browse"),
			), nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	got := strings.TrimSpace(out.String())
	assert.True(t, strings.HasPrefix(got, "{"), "expected JSON object, got: %s", got)
	assert.Contains(t, got, `"slug":"my-service"`)
	assert.Contains(t, got, `"webURL"`)
	// OUT2 ships all fields; field selection is deferred.
	assert.Contains(t, got, `"namespace"`)
}

func TestRepoView_JQ_FilterObject(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("my-service")), nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoView(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json", "--jq", ".slug"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, `"my-service"`, strings.TrimSpace(out.String()))
}
