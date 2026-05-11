package root_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdBrowse_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := root.NewCmdBrowse(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestBrowse_NoTarget_OpenRepoURL(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				WebURL: "https://bb.example.com/projects/MYPROJ/repos/my-service/browse",
			}, nil
		},
	}

	f, _, _ := newStatusFactory(t, fake)
	browser := &testhelpers.FakeBrowserLauncher{}
	f.Browser = browser

	cmd := root.NewCmdBrowse(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Equal(t, "https://bb.example.com/projects/MYPROJ/repos/my-service/browse", browser.URLs[0])
}

func TestBrowse_NumericTarget_OpensPRURL(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				WebURL: "https://bb.example.com/projects/MYPROJ/repos/my-service/browse",
			}, nil
		},
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{
				ID:     42,
				WebURL: "https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/42",
			}, nil
		},
	}

	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := newStatusFactory(t, fake)
	f.Browser = browser

	cmd := root.NewCmdBrowse(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "42"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Equal(t, "https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/42", browser.URLs[0])
}

func TestBrowse_HexTarget_OpensCommitURL(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				WebURL: "https://bb.example.com/projects/MYPROJ/repos/my-service/browse",
			}, nil
		},
	}

	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := newStatusFactory(t, fake)
	f.Browser = browser

	cmd := root.NewCmdBrowse(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc1234def"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Contains(t, browser.URLs[0], "/commits/abc1234def")
}

func TestBrowse_PathTarget_OpensSrcURL(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				WebURL: "https://bb.example.com/projects/MYPROJ/repos/my-service/browse",
			}, nil
		},
	}

	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := newStatusFactory(t, fake)
	f.Browser = browser

	cmd := root.NewCmdBrowse(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "README.md"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Contains(t, browser.URLs[0], "/src/")
	assert.Contains(t, browser.URLs[0], "README.md")
}
