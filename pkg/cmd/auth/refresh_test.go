package auth_test

import (
	"errors"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestAuthRefresh_ValidToken_PrintsConfirmation(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	factorytest.UseBackend(f, fake)
	cmd := auth.NewCmdAuthRefresh(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "alice")
}

func TestAuthRefresh_APIError_PrintsToStderr(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, errors.New("unauthorized")
		},
	}
	f, _, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	factorytest.UseBackend(f, fake)
	cmd := auth.NewCmdAuthRefresh(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, errOut.String(), "token validation failed")
}
