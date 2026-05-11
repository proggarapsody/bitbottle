package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRStatus_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRStatus(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestPRStatus_ShowsAuthorAndReviewerSections(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListMyPRsFn: func(ns, slug string) ([]backend.MyPREntry, error) {
			return []backend.MyPREntry{
				{
					PullRequest: backend.PullRequest{ID: 1, Title: "My authored PR", State: "OPEN"},
					Repo:        "workspace/repo",
					Role:        "AUTHOR",
				},
				{
					PullRequest: backend.PullRequest{ID: 2, Title: "Assigned to me", State: "OPEN"},
					Repo:        "workspace/repo",
					Role:        "REVIEWER",
				},
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRStatus(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Pull requests assigned to you for review")
	assert.Contains(t, got, "Pull requests created by you")
	assert.Contains(t, got, "My authored PR")
	assert.Contains(t, got, "Assigned to me")
}

func TestPRStatus_EmptySections_ShowsNone(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListMyPRsFn: func(ns, slug string) ([]backend.MyPREntry, error) {
			return nil, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRStatus(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "(none)")
}
