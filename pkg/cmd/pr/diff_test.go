package pr_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRDiff_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRDiff(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRDiff_StreamsDiffToStdout(t *testing.T) {
	t.Parallel()

	diffText := "diff --git a/foo b/foo\n+added\n"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRDiffFn: func(ns, slug string, id int) (string, error) {
			return diffText, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRDiff(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), diffText)
}

// TestPRDiff_TTY_StreamsThroughPager verifies that on a TTY the diff is
// piped through $PAGER. We use a pager command that transforms its input
// (`tr a-z A-Z`) so the assertion only passes if the bytes actually went
// through the subprocess. Without StartPager/StopPager wiring, the diff
// would land in the buffer unchanged in lowercase.
func TestPRDiff_TTY_StreamsThroughPager(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns a pager subprocess")
	}
	t.Setenv("PAGER", "tr a-z A-Z")

	diffText := "diff --git a/foo b/foo\n+added\n"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRDiffFn: func(ns, slug string, id int) (string, error) {
			return diffText, nil
		},
	}

	ios := iostreams.TestTTY()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		IOStreams:       ios,
		BackendOverride: fake,
		InitialConfig:   prConfig,
		GitRunner:       newPRRunner(),
	})
	cmd := pr.NewCmdPRDiff(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := ios.Out.(*bytes.Buffer).String()
	// Expect pager-transformed output — uppercase proves the diff passed
	// through `tr a-z A-Z` rather than going directly to the buffer.
	assert.Contains(t, got, "DIFF --GIT A/FOO B/FOO",
		"diff should be transformed by pager, got: %q", got)
}

func TestPRDiff_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("404 not found")
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRDiffFn: func(ns, slug string, id int) (string, error) {
			return "", apiErr
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRDiff(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
