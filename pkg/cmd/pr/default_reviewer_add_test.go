package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDefaultReviewerAdd_CallsBackend(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		AddDefaultReviewerFn: func(ns, slug, userSlug string) error {
			gotNS = ns
			gotSlug = slug
			gotUser = userSlug
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pr.NewCmdPRDefaultReviewerAdd(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "alice"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "alice", gotUser)
}

func TestDefaultReviewerAdd_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoDefaultReviewerFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := pr.NewCmdPRDefaultReviewerAdd(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "alice"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default reviewer")
}
