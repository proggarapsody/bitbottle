package project_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/project"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestProjectCreate_Success(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreateServerProjectInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateServerProjectFn: func(in backend.CreateServerProjectInput) (backend.ServerProject, error) {
			gotInput = in
			return backend.ServerProject{Key: in.Key, Name: in.Name}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"create", "PRJ", "--name", "My Project", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "PRJ", gotInput.Key)
	assert.Equal(t, "My Project", gotInput.Name)
	assert.Contains(t, out.String(), "PRJ")
	assert.Contains(t, out.String(), "My Project")
}

func TestProjectCreate_RequiresName(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"create", "PRJ", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestProjectCreate_RequiresKey(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"create", "--name", "My Project", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
}
