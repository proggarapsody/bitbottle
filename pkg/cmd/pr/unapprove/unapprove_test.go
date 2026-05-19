package unapprove_test

import (
	"errors"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/unapprove"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRUnapprove_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UnapprovePRFn: func(ns, slug string, id int) error {
			return nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := unapprove.NewCmdUnapprove(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Removed approval from pull request #42")
}

func TestPRUnapprove_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UnapprovePRFn: func(ns, slug string, id int) error {
			return errors.New("403 forbidden")
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := unapprove.NewCmdUnapprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
