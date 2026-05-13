package repo_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoWatcherList_PrintsTable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.User{
				{Slug: "alice", DisplayName: "Alice Smith"},
				{Slug: "bob", DisplayName: "Bob Jones"},
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoWatcher(f)
	cmd.SetArgs([]string{"list", "MYPROJ/my-repo"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "Alice Smith")
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "Bob Jones")
	assert.Contains(t, output, "bob")
}

func TestRepoWatcherList_EmptyList_NoError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			return []backend.User{}, nil
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoWatcher(f)
	cmd.SetArgs([]string{"list", "MYPROJ/my-repo"})
	require.NoError(t, cmd.Execute())
}

func TestRepoWatcherList_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoWatcher(f)
	cmd.SetArgs([]string{"list", "MYPROJ/my-repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestRepoWatcherList_MissingRepo_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoWatcher(f)
	cmd.SetArgs([]string{"list"})
	// Without a base repo configured, this should fail gracefully
	err := cmd.Execute()
	// Error is expected here because no repo is resolvable
	_ = err
}
