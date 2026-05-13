package pr_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRCommits_ListsCommitsTable(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommitsFn: func(ns, slug string, prID int) ([]backend.Commit, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, 42, prID)
			return []backend.Commit{
				{
					Hash:      "abc1234def456abc1234def456abc1234def456ab",
					Message:   "Fix null pointer in auth",
					Author:    backend.User{Slug: "alice", DisplayName: "Alice Smith"},
					Timestamp: ts,
				},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommits(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "abc1234")
	assert.Contains(t, output, "Fix null pointer in auth")
	assert.Contains(t, output, "Alice Smith")
}

func TestPRCommits_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommitsFn: func(ns, slug string, prID int) ([]backend.Commit, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommits(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestPRCommits_InvalidPRID_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommits(f)
	cmd.SetArgs([]string{"not-a-number"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PR ID")
}

func TestPRCommits_EmptyList_NoError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommitsFn: func(ns, slug string, prID int) ([]backend.Commit, error) {
			return []backend.Commit{}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommits(f)
	cmd.SetArgs([]string{"7"})
	require.NoError(t, cmd.Execute())
}
