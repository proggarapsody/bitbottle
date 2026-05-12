package delete_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/variable/delete"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdDelete_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing KEY
	require.Error(t, cmd.Execute())
}

func TestDelete_WithConfirmFlag_DeletesByKey(t *testing.T) {
	t.Parallel()
	var gotKey string
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineVariableFn: func(ns, slug, key string) error {
			gotKey = key
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "OBSOLETE", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "OBSOLETE", gotKey)
	assert.Contains(t, out.String(), "Deleted pipeline variable OBSOLETE")
}

func TestDelete_NoConfirm_NoTTY_RejectsForSafety(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "OBSOLETE"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm required")
}

func TestDelete_BackendNotFound_PropagatesError(t *testing.T) {
	t.Parallel()
	notFound := &backend.DomainError{Kind: backend.ErrNotFound, Message: `pipeline variable "X" not found`}
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineVariableFn: func(ns, slug, key string) error {
			return notFound
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "X", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, errors.Is(err, backend.ErrNotFound))
}

func TestDelete_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "X", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
