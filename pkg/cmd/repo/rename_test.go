package repo_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoRename_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := repo.NewCmdRepoRename(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoRename_CallsBackendWithNewName(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotNew string
	fake := &testhelpers.FakeClient{
		T: t,
		RenameRepoFn: func(ns, slug, newName string) (backend.Repository, error) {
			gotNS, gotSlug, gotNew = ns, slug, newName
			return backend.Repository{Slug: newName, Name: newName, Namespace: ns}, nil
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoRename(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "renamed", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "renamed", gotNew)
}

func TestRepoRename_PrintsNewCoordinate(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RenameRepoFn: func(ns, slug, newName string) (backend.Repository, error) {
			return backend.Repository{Slug: newName, Name: newName, Namespace: ns}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoRename(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "renamed", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "MYPROJ/renamed")
}

func TestRepoRename_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RenameRepoFn: func(ns, slug, newName string) (backend.Repository, error) {
			return backend.Repository{Slug: newName, Name: newName, Namespace: ns}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoRename(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "renamed", "--json", "--confirm"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"slug":"renamed"`)
	assert.Contains(t, got, `"namespace":"MYPROJ"`)
}

func TestRepoRename_WithoutConfirm_NonTTY_Errors(t *testing.T) {
	t.Parallel()
	// Default IOStreams is non-TTY; renaming changes the slug on Cloud and
	// breaks every existing clone's origin URL — must require --confirm to
	// match the safety bar set by `repo delete` for destructive ops.
	f, _, _ := newRepoFactory(t, nil)
	cmd := repo.NewCmdRepoRename(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "renamed"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirm")
}

func TestRepoRename_APIError_Propagates(t *testing.T) {
	t.Parallel()
	apiErr := errors.New("403 forbidden")
	fake := &testhelpers.FakeClient{
		T: t,
		RenameRepoFn: func(ns, slug, newName string) (backend.Repository, error) {
			return backend.Repository{}, apiErr
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoRename(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "renamed", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
