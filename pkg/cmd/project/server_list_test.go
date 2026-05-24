package project_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/project"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestProjectServerList_Success(t *testing.T) {
	t.Parallel()
	var gotFilter string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListServerProjectsFn: func(filter string, limit int) ([]backend.ServerProject, error) {
			gotFilter = filter
			gotLimit = limit
			return []backend.ServerProject{
				{Key: "PRJ", Name: "My Project", Public: false},
				{Key: "DEV", Name: "Dev Project", Public: true},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"server-list", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "", gotFilter)
	assert.Equal(t, 30, gotLimit)
	assert.Contains(t, out.String(), "PRJ")
	assert.Contains(t, out.String(), "DEV")
}

func TestProjectServerList_WithFilter(t *testing.T) {
	t.Parallel()
	var gotFilter string
	fake := &testhelpers.FakeClient{
		T: t,
		ListServerProjectsFn: func(filter string, limit int) ([]backend.ServerProject, error) {
			gotFilter = filter
			return []backend.ServerProject{{Key: "PRJ", Name: "My Project"}}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"server-list", "--hostname", "git.example.com", "--filter", "My"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "My", gotFilter)
}

func TestProjectServerList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListServerProjectsFn: func(filter string, limit int) ([]backend.ServerProject, error) {
			return []backend.ServerProject{{Key: "PRJ", Name: "My Project"}}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"server-list", "--hostname", "git.example.com", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, strings.Contains(out.String(), `"key"`))
}
