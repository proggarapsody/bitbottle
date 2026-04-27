package factory_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestEnableRepoOverride_RepoFlag_OverridesBaseRepo verifies that --repo
// HOST/PROJ/REPO replaces what f.BaseRepo() returns.
func TestEnableRepoOverride_RepoFlag_OverridesBaseRepo(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb1.example.com:\n  oauth_token: tok\n",
		GitRunner:     testhelpers.NewFakeRunner(testhelpers.RunResponse{Stdout: "ssh://git@bb1.example.com/AUTO/auto.git\n"}),
	})

	root := &cobra.Command{Use: "pr"}
	factory.EnableRepoOverride(root, f)

	leaf := &cobra.Command{
		Use:  "list",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(leaf)

	root.SetArgs([]string{"list", "--repo", "bb2.example.com/MYPROJ/myrepo"})
	require.NoError(t, root.Execute())

	ref, err := f.BaseRepo()
	require.NoError(t, err)
	assert.Equal(t, "bb2.example.com", ref.Host)
	assert.Equal(t, "MYPROJ", ref.Project)
	assert.Equal(t, "myrepo", ref.Slug)
}

// TestEnableRepoOverride_BareProjectRepo_UsesSingleConfiguredHost verifies
// that --repo PROJ/REPO (no host component) uses the single configured host.
func TestEnableRepoOverride_BareProjectRepo_UsesSingleConfiguredHost(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n",
		GitRunner:     testhelpers.NewFakeRunner(),
	})

	root := &cobra.Command{Use: "pr"}
	factory.EnableRepoOverride(root, f)

	leaf := &cobra.Command{
		Use:  "list",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(leaf)

	root.SetArgs([]string{"list", "--repo", "MYPROJ/myrepo"})
	require.NoError(t, root.Execute())

	ref, err := f.BaseRepo()
	require.NoError(t, err)
	assert.Equal(t, "bb.example.com", ref.Host)
	assert.Equal(t, "MYPROJ", ref.Project)
	assert.Equal(t, "myrepo", ref.Slug)
}

// TestEnableRepoOverride_RegistersPersistentFlag verifies the -R/--repo
// flag is registered as a persistent flag with the expected description.
func TestEnableRepoOverride_RegistersPersistentFlag(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	root := &cobra.Command{Use: "pr"}
	factory.EnableRepoOverride(root, f)

	flag := root.PersistentFlags().Lookup("repo")
	require.NotNil(t, flag, "--repo flag must be registered persistently")
	assert.Equal(t, "R", flag.Shorthand, "should have -R short form")
}
