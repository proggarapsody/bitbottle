package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/user"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestUserRename_Success(t *testing.T) {
	t.Parallel()
	var gotSlug, gotNewSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		RenameUserFn: func(slug, newSlug string) error {
			gotSlug = slug
			gotNewSlug = newSlug
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserRename(f, nil)
	cmd.SetArgs([]string{"old-alice", "alice"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "old-alice", gotSlug)
	assert.Equal(t, "alice", gotNewSlug)
	assert.Contains(t, out.String(), "User old-alice renamed to alice.")
}

func TestUserRename_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserRename(f, nil)
	cmd.SetArgs([]string{"only-one"})
	err := cmd.Execute()
	require.Error(t, err)
}
