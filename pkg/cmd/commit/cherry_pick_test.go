package commit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCherryPick_PrintsNewHash(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CherryPickCommitFn: func(ns, slug string, in backend.CherryPickInput) (backend.Commit, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc123", in.SourceHash)
			assert.Equal(t, "main", in.TargetBranch)
			assert.Empty(t, in.Message)
			return backend.Commit{Hash: "newdef456", Message: "Fix thing"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCherryPick(f)
	cmd.SetArgs([]string{"abc123", "main", "myworkspace/my-repo"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "newdef456")
}

func TestCherryPick_WithMessageFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CherryPickCommitFn: func(ns, slug string, in backend.CherryPickInput) (backend.Commit, error) {
			assert.Equal(t, "cherry-pick: fix", in.Message)
			return backend.Commit{Hash: "aabbcc"}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCherryPick(f)
	cmd.SetArgs([]string{"abc123", "main", "myws/repo", "--message", "cherry-pick: fix"})
	require.NoError(t, cmd.Execute())
}

func TestCherryPick_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CherryPickCommitFn: func(ns, slug string, in backend.CherryPickInput) (backend.Commit, error) {
			return backend.Commit{}, assert.AnError
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCherryPick(f)
	cmd.SetArgs([]string{"abc123", "main", "myws/repo"})
	err := cmd.Execute()
	require.Error(t, err)
}
