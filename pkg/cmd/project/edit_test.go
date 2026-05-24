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

func TestProjectEdit_Success(t *testing.T) {
	t.Parallel()
	var gotKey string
	var gotIn backend.UpdateServerProjectInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateServerProjectFn: func(key string, in backend.UpdateServerProjectInput) (backend.ServerProject, error) {
			gotKey = key
			gotIn = in
			return backend.ServerProject{Key: key, Name: *in.Name}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"edit", "PRJ", "--name", "New Name", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "PRJ", gotKey)
	require.NotNil(t, gotIn.Name)
	assert.Equal(t, "New Name", *gotIn.Name)
	assert.Contains(t, out.String(), "Updated project PRJ")
}

func TestProjectEdit_NoFlagsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"edit", "PRJ", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
}
