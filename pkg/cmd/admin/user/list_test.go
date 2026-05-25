package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/user"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestUserList_TextOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAdminUsersFn: func(filter string, limit int) ([]backend.AdminUser, error) {
			return []backend.AdminUser{
				{Slug: "alice", DisplayName: "Alice A", Email: "alice@example.com", Active: true, Type: "NORMAL"},
				{Slug: "svc-bot", DisplayName: "Bot", Email: "bot@example.com", Active: false, Type: "SERVICE"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserList(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "alice")
	assert.Contains(t, out.String(), "Alice A")
	assert.Contains(t, out.String(), "svc-bot")
}

func TestUserList_WithFilter(t *testing.T) {
	t.Parallel()
	var gotFilter string
	fake := &testhelpers.FakeClient{
		T: t,
		ListAdminUsersFn: func(filter string, limit int) ([]backend.AdminUser, error) {
			gotFilter = filter
			return nil, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserList(f, nil)
	cmd.SetArgs([]string{"--filter", "ali"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "ali", gotFilter)
}

func TestUserList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListAdminUsersFn: func(filter string, limit int) ([]backend.AdminUser, error) {
			return []backend.AdminUser{
				{Slug: "alice", DisplayName: "Alice A", Email: "alice@example.com", Active: true, Type: "NORMAL"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := user.NewCmdUserList(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"Slug":"alice"`)
}

func TestUserList_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := user.NewCmdUserList(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
