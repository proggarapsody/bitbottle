package member_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/group"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGroupMemberRemove_WithConfirm(t *testing.T) {
	t.Parallel()
	var gotGroup, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		RemoveGroupMemberFn: func(groupName, user string) error {
			gotGroup = groupName
			gotUser = user
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"member", "remove", "developers", "alice", "--confirm", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "developers", gotGroup)
	assert.Equal(t, "alice", gotUser)
	assert.Contains(t, out.String(), "Removed alice")
}

func TestGroupMemberRemove_NonTTY_RequiresConfirm(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "git.example.com:\n  oauth_token: tok\n",
	})
	// factorytest.New sets IsStdoutTTY to return false by default.
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"member", "remove", "developers", "alice", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "confirm")
}
