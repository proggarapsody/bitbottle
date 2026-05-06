package repo_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoFork_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := repo.NewCmdRepoFork(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoFork_RequiresInto(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFork(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--into")
}

func TestRepoFork_CallsBackendWithTargetWorkspace(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotIn backend.ForkRepoInput
	fake := &testhelpers.FakeClient{
		T: t,
		ForkRepoFn: func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
			gotNS, gotSlug, gotIn = ns, slug, in
			return backend.Repository{Slug: "my-service", Name: "my-service", Namespace: in.Workspace}, nil
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFork(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--into", "otherws"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "otherws", gotIn.Workspace)
	assert.Empty(t, gotIn.Name, "no --name flag means input.Name stays empty so Bitbucket reuses source name")
}

func TestRepoFork_NameFlagRenamesFork(t *testing.T) {
	t.Parallel()
	var gotIn backend.ForkRepoInput
	fake := &testhelpers.FakeClient{
		T: t,
		ForkRepoFn: func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
			gotIn = in
			return backend.Repository{Slug: "renamed-fork", Name: "renamed-fork", Namespace: in.Workspace}, nil
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFork(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--into", "otherws", "--name", "renamed-fork"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "renamed-fork", gotIn.Name)
}

func TestRepoFork_PrintsForkCoordinate(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ForkRepoFn: func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
			return backend.Repository{
				Slug:      "my-service",
				Name:      "my-service",
				Namespace: in.Workspace,
				WebURL:    "https://bitbucket.org/otherws/my-service",
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFork(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--into", "otherws"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "otherws/my-service")
}

func TestRepoFork_APIError_Propagates(t *testing.T) {
	t.Parallel()
	apiErr := errors.New("403 forbidden: workspace not authorized")
	fake := &testhelpers.FakeClient{
		T: t,
		ForkRepoFn: func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
			return backend.Repository{}, apiErr
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFork(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--into", "otherws"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
