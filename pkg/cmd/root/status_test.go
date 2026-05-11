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

func TestNewCmdStatus_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := root.NewCmdStatus(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestStatus_ShowsBothSections(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListMyPRsFn: func(ns, slug string) ([]backend.MyPREntry, error) {
			return []backend.MyPREntry{
				{
					PullRequest: backend.PullRequest{ID: 5, Title: "Review this", State: "OPEN"},
					Repo:        "PROJ/repo",
					Role:        "REVIEWER",
				},
				{
					PullRequest: backend.PullRequest{ID: 6, Title: "My PR", State: "OPEN"},
					Repo:        "PROJ/repo",
					Role:        "AUTHOR",
				},
			}, nil
		},
	}

	f, out, _ := newStatusFactory(t, fake)
	cmd := root.NewCmdStatus(f)
	cmd.SetArgs([]string{"PROJ/repo"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Pull requests assigned to you for review")
	assert.Contains(t, got, "Pull requests created by you")
	assert.Contains(t, got, "Review this")
	assert.Contains(t, got, "My PR")
}

func TestStatus_NoRepoArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := root.NewCmdStatus(f)
	err := cmd.Execute()
	require.Error(t, err)
}
