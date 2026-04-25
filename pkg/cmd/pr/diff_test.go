package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
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
