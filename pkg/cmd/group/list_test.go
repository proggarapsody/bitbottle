package group_test

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

func TestGroupList_Success(t *testing.T) {
	t.Parallel()
	var gotFilter string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupsFn: func(filter string, limit int) ([]backend.Group, error) {
			gotFilter = filter
			gotLimit = limit
			return []backend.Group{{Name: "developers"}, {Name: "admins"}}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"list", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "", gotFilter)
	assert.Equal(t, 100, gotLimit)
	assert.Contains(t, out.String(), "developers")
	assert.Contains(t, out.String(), "admins")
}

func TestGroupList_WithFilter(t *testing.T) {
	t.Parallel()
	var gotFilter string
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupsFn: func(filter string, limit int) ([]backend.Group, error) {
			gotFilter = filter
			return []backend.Group{{Name: "devs"}}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"list", "--hostname", "git.example.com", "--filter", "dev"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "dev", gotFilter)
}

func TestGroupList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupsFn: func(filter string, limit int) ([]backend.Group, error) {
			return []backend.Group{{Name: "qa"}}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := group.NewCmdGroup(f)
	cmd.SetArgs([]string{"list", "--hostname", "git.example.com", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, strings.Contains(out.String(), `"name"`))
}
