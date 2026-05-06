package delete_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/delete"
	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdDelete_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing ID
	require.Error(t, cmd.Execute())
}

func TestDelete_WithConfirmFlag_DeletesByID(t *testing.T) {
	t.Parallel()
	var gotID string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWebhookFn: func(ns, slug, id string) error {
			gotID = id
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc-1", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "abc-1", gotID)
	assert.Contains(t, out.String(), "Deleted webhook abc-1")
}

func TestDelete_NoConfirm_NoTTY_RejectsForSafety(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc-1"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm required")
}

func TestDelete_BackendNotFound_PropagatesError(t *testing.T) {
	t.Parallel()
	notFound := &backend.DomainError{Kind: backend.ErrNotFound, Message: `webhook "x" not found`}
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWebhookFn: func(ns, slug, id string) error {
			return notFound
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "x", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, errors.Is(err, backend.ErrNotFound))
}
