package branchrule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/branchrule"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDelete_DeletesRule(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteBranchRuleFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "7"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 7, gotID)
	assert.Contains(t, out.String(), "7")
}

func TestDelete_InvalidID(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "notanumber"})
	require.Error(t, cmd.Execute())
}

func TestDelete_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdDelete(f)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}
