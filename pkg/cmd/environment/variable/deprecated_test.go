package variable_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/delete"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/list"
	cmdSet "github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestDeprecatedFields guards against the Deprecated field being silently
// removed from the environment variable command tree.
func TestDeprecatedFields(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())

	t.Run("variable group", func(t *testing.T) {
		t.Parallel()
		cmd := variable.NewCmdVariable(f)
		require.NotEmpty(t, cmd.Deprecated)
		assert.True(t, strings.Contains(cmd.Deprecated, "--scope deployment"),
			"Deprecated message should mention --scope deployment, got: %q", cmd.Deprecated)
	})

	t.Run("list subcommand", func(t *testing.T) {
		t.Parallel()
		cmd := cmdList.NewCmdList(f, nil) //nolint:staticcheck
		require.NotEmpty(t, cmd.Deprecated)
		assert.True(t, strings.Contains(cmd.Deprecated, "--scope deployment"),
			"Deprecated message should mention --scope deployment, got: %q", cmd.Deprecated)
	})

	t.Run("set subcommand", func(t *testing.T) {
		t.Parallel()
		cmd := cmdSet.NewCmdSet(f, nil) //nolint:staticcheck
		require.NotEmpty(t, cmd.Deprecated)
		assert.True(t, strings.Contains(cmd.Deprecated, "--scope deployment"),
			"Deprecated message should mention --scope deployment, got: %q", cmd.Deprecated)
	})

	t.Run("delete subcommand", func(t *testing.T) {
		t.Parallel()
		cmd := cmdDelete.NewCmdDelete(f, nil) //nolint:staticcheck
		require.NotEmpty(t, cmd.Deprecated)
		assert.True(t, strings.Contains(cmd.Deprecated, "--scope deployment"),
			"Deprecated message should mention --scope deployment, got: %q", cmd.Deprecated)
	})
}
