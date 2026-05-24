package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/group"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGroupDelete_WithConfirm(t *testing.T) {
	t.Parallel()
	var gotName string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteGroupFn: func(name string) error {
			gotName = name
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"delete", "oldgroup", "--confirm", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "oldgroup", gotName)
	assert.Contains(t, out.String(), "Deleted group")
}

func TestGroupDelete_NonTTY_RequiresConfirm(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "git.example.com:\n  oauth_token: tok\n",
	})
	// factorytest.New sets IsStdoutTTY to return false by default,
	// so without --confirm this must error.
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"delete", "oldgroup", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "confirm")
}
