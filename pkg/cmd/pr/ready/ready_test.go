package ready_test

import (
	"errors"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/ready"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRReady_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ReadyPRFn: func(ns, slug string, id int) error {
			return nil
		},
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: id, WebURL: "https://example.com/pr/42"}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := ready.NewCmdReady(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Marked pull request #42 as ready for review")
	assert.Contains(t, out.String(), "https://example.com/pr/42")
}

func TestPRReady_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ReadyPRFn: func(ns, slug string, id int) error {
			return errors.New("422 unprocessable")
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := ready.NewCmdReady(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "422")
}
