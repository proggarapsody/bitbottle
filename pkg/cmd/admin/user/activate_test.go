package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/user"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestUserActivate_Success(t *testing.T) {
	t.Parallel()
	var gotSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		ActivateUserFn: func(slug string) error {
			gotSlug = slug
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserActivate(f, nil)
	cmd.SetArgs([]string{"alice"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "alice", gotSlug)
	assert.Contains(t, out.String(), "User alice activated.")
}

func TestUserActivate_RequiresArg(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserActivate(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
}
