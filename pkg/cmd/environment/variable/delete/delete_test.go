package delete_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	vardelete "github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/delete"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdDelete_RequiresThreeArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := vardelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid"})
	require.Error(t, cmd.Execute())
}

func TestDelete_DeletesVariable(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteEnvVariableFn: func(ns, slug, envUUID, varUUID string) error {
			deleted = true
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "env-uuid-1", envUUID)
			assert.Equal(t, "var-uuid-1", varUUID)
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := vardelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid-1", "var-uuid-1"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, out.String(), "var-uuid-1")
}
