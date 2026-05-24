package pat_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth/pat"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPATRevoke_ConfirmFlag(t *testing.T) {
	t.Parallel()
	var gotUserSlug, gotTokenID string
	fake := &testhelpers.FakeClient{
		T: t,
		RevokePATFn: func(userSlug, tokenID string) error {
			gotUserSlug = userSlug
			gotTokenID = tokenID
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"revoke", "42", "--hostname", "git.example.com", "--confirm"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "alice", gotUserSlug)
	assert.Equal(t, "42", gotTokenID)
	assert.Contains(t, out.String(), "Revoked PAT 42")
}

func TestPATRevoke_NonTTYRequiresConfirm(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"revoke", "42", "--hostname", "git.example.com"})
	err := cmd.Execute()
	// Non-TTY without --confirm must error.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}
