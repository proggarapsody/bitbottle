package group_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/group"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGroupCreate_Success(t *testing.T) {
	t.Parallel()
	var gotName string
	fake := &testhelpers.FakeClient{
		T: t,
		CreateGroupFn: func(name string) (backend.Group, error) {
			gotName = name
			return backend.Group{Name: name}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"create", "newgroup", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "newgroup", gotName)
	assert.Contains(t, out.String(), "newgroup")
}

func TestGroupCreate_RequiresName(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	assert.Error(t, err)
}
