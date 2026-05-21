package setdefaultbranch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/test/testhelpers"

	setdefaultbranch "github.com/proggarapsody/bitbottle/pkg/cmd/repo/set-default-branch"
)

func TestRepoSetDefaultBranch_SetsViaExplicitRepo(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		SetRepoDefaultBranchFn: func(ns, slug, branch string) error {
			gotNS, gotSlug, gotBranch = ns, slug, branch
			return nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := setdefaultbranch.NewCmdSetDefaultBranch(f)
	cmd.SetArgs([]string{"main", "MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "main", gotBranch)
	assert.Contains(t, out.String(), "MYPROJ/my-service")
	assert.Contains(t, out.String(), "main")
}

func TestRepoSetDefaultBranch_InvalidRepo_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := newRepoFactory(t, nil)
	cmd := setdefaultbranch.NewCmdSetDefaultBranch(f)
	cmd.SetArgs([]string{"main", "not-a-valid-repo-ref"})
	err := cmd.Execute()
	require.Error(t, err)
}
