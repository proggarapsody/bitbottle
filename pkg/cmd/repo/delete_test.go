package repo_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoDelete_HasConfirmFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoDelete(f)
	assert.NotNil(t, cmd.Flag("confirm"))
}

func TestNewCmdRepoDelete_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoDelete(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoDelete_WithConfirmFlag_DeletesRepo(t *testing.T) {
	t.Parallel()

	var calledNS, calledSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRepoFn: func(ns, slug string) error {
			calledNS = ns
			calledSlug = slug
			return nil
		},
	}

	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoDelete(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", calledNS)
	assert.Equal(t, "my-service", calledSlug)
}

func TestRepoDelete_WithoutConfirm_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := newRepoFactory(t, nil)
	cmd := repo.NewCmdRepoDelete(f)
	// Default IOStreams is non-TTY; without --confirm, must error.
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confirm")
}

func TestRepoDelete_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("404 not found")
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRepoFn: func(ns, slug string) error {
			return apiErr
		},
	}

	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoDelete(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
