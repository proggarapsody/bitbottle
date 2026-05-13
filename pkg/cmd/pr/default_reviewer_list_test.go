package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRDefaultReviewerList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := pr.NewCmdPRDefaultReviewerList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestDefaultReviewerList_PrintsReviewers(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListDefaultReviewersFn: func(ns, slug string) ([]backend.DefaultReviewer, error) {
			return []backend.DefaultReviewer{
				{UserSlug: "alice", DisplayName: "Alice Smith", EmailAddress: "alice@co.com"},
				{UserSlug: "bob", DisplayName: "Bob Jones"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pr.NewCmdPRDefaultReviewerList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "Alice Smith")
	assert.Contains(t, got, "bob")
}

func TestDefaultReviewerList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListDefaultReviewersFn: func(ns, slug string) ([]backend.DefaultReviewer, error) {
			return []backend.DefaultReviewer{
				{UserSlug: "alice", DisplayName: "Alice Smith", EmailAddress: "alice@co.com"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pr.NewCmdPRDefaultReviewerList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"userSlug":"alice"`)
}

func TestDefaultReviewerList_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoDefaultReviewerFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pr.NewCmdPRDefaultReviewerList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default reviewer")
}
