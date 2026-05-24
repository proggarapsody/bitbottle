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

func TestProjectView_Success(t *testing.T) {
	t.Parallel()
	var gotKey string
	fake := &testhelpers.FakeClient{
		T: t,
		GetServerProjectFn: func(key string) (backend.ServerProject, error) {
			gotKey = key
			return backend.ServerProject{
				Key:         key,
				Name:        "My Project",
				Description: "A great project",
				Public:      true,
				WebURL:      "https://git.example.com/projects/PRJ",
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"view", "PRJ", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "PRJ", gotKey)
	assert.Contains(t, out.String(), "PRJ")
	assert.Contains(t, out.String(), "My Project")
}

func TestProjectView_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetServerProjectFn: func(key string) (backend.ServerProject, error) {
			return backend.ServerProject{Key: key, Name: "My Project"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"view", "PRJ", "--hostname", "git.example.com", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, strings.Contains(out.String(), `"key"`))
}

func TestProjectView_RequiresKey(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"view", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
}
