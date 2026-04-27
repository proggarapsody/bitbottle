package root_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// TestRootHelp_PRList_ShowsArgumentsSection verifies that `bitbottle pr list
// --help` includes an ARGUMENTS section sourced from the pr parent's
// Annotations["help:arguments"]. This is the user-visible surface change.
func TestRootHelp_PRList_ShowsArgumentsSection(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := root.NewCmdRoot(f)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"pr", "list", "--help"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "ARGUMENTS",
		"pr list --help should include an ARGUMENTS section")
	assert.Contains(t, got, "PROJECT/REPO",
		"ARGUMENTS section should describe the PROJECT/REPO positional argument")
}

// TestRootHelp_RepoView_ShowsArgumentsSection verifies the same wiring for
// the repo command tree.
func TestRootHelp_RepoView_ShowsArgumentsSection(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := root.NewCmdRoot(f)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"repo", "view", "--help"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "ARGUMENTS",
		"repo view --help should include an ARGUMENTS section")
}
