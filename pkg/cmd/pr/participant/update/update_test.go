package update_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/participant/update"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRParticipantUpdate_RequiresUser(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := update.NewCmdPRParticipantUpdate(f)
	cmd.SetArgs([]string{"10", "--approve"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user")
}

func TestPRParticipantUpdate_RequiresMutuallyExclusiveFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := update.NewCmdPRParticipantUpdate(f)
	cmd.SetArgs([]string{"10", "--user", "acc123"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--approve")
}

func TestPRParticipantUpdate_Approve_CallsBackend(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePRParticipantFn: func(ns, slug string, prID int, accountID, state string) (backend.PRParticipant, error) {
			assert.Equal(t, 10, prID)
			assert.Equal(t, "acc123", accountID)
			gotState = state
			return backend.PRParticipant{
				User:     backend.User{DisplayName: "Alice", Slug: "alice"},
				State:    "APPROVED",
				Approved: true,
			}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := update.NewCmdPRParticipantUpdate(f)
	cmd.SetArgs([]string{"10", "--user", "acc123", "--approve"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "approved", gotState)
	assert.Contains(t, out.String(), "APPROVED")
}

func TestPRParticipantUpdate_Unapprove_CallsBackend(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePRParticipantFn: func(ns, slug string, prID int, accountID, state string) (backend.PRParticipant, error) {
			gotState = state
			return backend.PRParticipant{
				User:  backend.User{DisplayName: "Bob", Slug: "bob"},
				State: "",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := update.NewCmdPRParticipantUpdate(f)
	cmd.SetArgs([]string{"10", "--user", "acc456", "--unapprove"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "", gotState)
	assert.Contains(t, out.String(), "unapproved")
}

func TestPRParticipantUpdate_RequestChanges_CallsBackend(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePRParticipantFn: func(ns, slug string, prID int, accountID, state string) (backend.PRParticipant, error) {
			gotState = state
			return backend.PRParticipant{
				User:  backend.User{DisplayName: "Carol", Slug: "carol"},
				State: "CHANGES_REQUESTED",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := update.NewCmdPRParticipantUpdate(f)
	cmd.SetArgs([]string{"10", "--user", "acc789", "--request-changes"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "changes_requested", gotState)
	assert.Contains(t, out.String(), "CHANGES_REQUESTED")
}
