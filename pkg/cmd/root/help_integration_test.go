package root_test

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// TestRootHelp_PRList_ShowsArgumentsSection verifies that `bitbottle pr list
// --help` includes an ARGUMENTS section sourced from the pr parent's
// Annotations["help:arguments"]. This is the user-visible surface change.
func TestRootHelp_PRList_ShowsArgumentsSection(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{})
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

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := root.NewCmdRoot(f)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"repo", "view", "--help"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "ARGUMENTS",
		"repo view --help should include an ARGUMENTS section")
}

// TestRootHelp_Context_IsRegistered verifies that `bitbottle context --help`
// is reachable through the root command — the regression that would catch a
// missing AddCommand call when wiring CTX into root.
func TestRootHelp_Context_IsRegistered(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := root.NewCmdRoot(f)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"context", "--help"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "bitbottle context",
		"`bitbottle context --help` should print the context command's help")
	assert.Contains(t, got, "default branch",
		"context help should describe the orientation shape")
}
