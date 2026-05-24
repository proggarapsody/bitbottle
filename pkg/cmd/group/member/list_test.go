package member_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/group"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGroupMemberList_Success(t *testing.T) {
	t.Parallel()
	var gotGroup string
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupMembersFn: func(groupName string, limit int) ([]backend.GroupMember, error) {
			gotGroup = groupName
			return []backend.GroupMember{
				{Name: "alice", DisplayName: "Alice", EmailAddress: "alice@example.com"},
				{Name: "bob", DisplayName: "Bob", EmailAddress: "bob@example.com"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"member", "list", "developers", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "developers", gotGroup)
	assert.Contains(t, out.String(), "alice")
	assert.Contains(t, out.String(), "bob")
}

func TestGroupMemberList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupMembersFn: func(groupName string, limit int) ([]backend.GroupMember, error) {
			return []backend.GroupMember{
				{Name: "alice", DisplayName: "Alice", EmailAddress: "alice@example.com"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"member", "list", "developers", "--hostname", "git.example.com", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, strings.Contains(out.String(), `"name"`))
}
