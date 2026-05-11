package commit_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitCommentList_PrintsTable(t *testing.T) {
	t.Parallel()

	now := time.Now()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitCommentsFn: func(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
			return []backend.CommitComment{
				{
					ID:        101,
					Author:    backend.User{Slug: "alice"},
					Body:      "Looks good to me",
					CreatedAt: now.Add(-1 * time.Hour),
				},
				{
					ID:        102,
					Author:    backend.User{Slug: "bob"},
					Body:      "Minor nit here",
					CreatedAt: now.Add(-2 * time.Hour),
				},
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentList(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "101")
	assert.Contains(t, got, "102")
	assert.Contains(t, got, "Looks good")
	assert.Contains(t, got, "Minor nit")
}

func TestCommitCommentList_Empty(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitCommentsFn: func(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
			return []backend.CommitComment{}, nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentList(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef"})
	require.NoError(t, cmd.Execute())
}

func TestCommitCommentList_BackendError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitCommentsFn: func(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
			return nil, errors.New("upstream connection failed")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentList(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upstream connection failed")
}
