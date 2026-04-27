package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestPRList_RepoOverrideFlag_PassesProjectAndSlugToBackend verifies that
// `pr list -R MYPROJ/myrepo` causes the backend's ListPRs to be called with
// the override's project/slug, even when no git remote is present.
func TestPRList_RepoOverrideFlag_PassesProjectAndSlugToBackend(t *testing.T) {
	t.Parallel()

	var gotNS, gotSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
			gotNS, gotSlug = ns, slug
			return nil, nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   "bb.example.com:\n  oauth_token: tok\n",
		BackendOverride: fake,
		GitRunner:       testhelpers.NewFakeRunner(),
	})

	root := pr.NewCmdPR(f)
	root.SetArgs([]string{"list", "-R", "MYPROJ/myrepo"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "myrepo", gotSlug)
}
