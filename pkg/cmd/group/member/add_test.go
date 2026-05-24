package member_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/group"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGroupMemberAdd_Success(t *testing.T) {
	t.Parallel()
	var gotGroup, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		AddGroupMemberFn: func(groupName, user string) error {
			gotGroup = groupName
			gotUser = user
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"member", "add", "developers", "alice", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "developers", gotGroup)
	assert.Equal(t, "alice", gotUser)
	assert.Contains(t, out.String(), "alice")
}

func TestGroupMemberAdd_RequiresBothArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"member", "add", "developers"})
	err := cmd.Execute()
	assert.Error(t, err)
}
