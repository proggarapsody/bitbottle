package cmdutil_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// newCmdWithOutputFlags builds a minimal cobra.Command with output flags registered.
func newCmdWithOutputFlags() *cobra.Command {
	cmd := &cobra.Command{Use: "test", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	format.RegisterOutputFlags(cmd)
	return cmd
}

func TestJSONFieldsFromCmd_NoFlag_ReturnsNil(t *testing.T) {
	t.Parallel()
	cmd := newCmdWithOutputFlags()
	// No --json flag passed at all
	require.NoError(t, cmd.Execute())
	fields := cmdutil.JSONFieldsFromCmd(cmd)
	assert.Nil(t, fields, "expected nil when --json not passed")
}

func TestJSONFieldsFromCmd_FlagWithNoValue_ReturnsNil(t *testing.T) {
	t.Parallel()
	cmd := newCmdWithOutputFlags()
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	fields := cmdutil.JSONFieldsFromCmd(cmd)
	assert.Nil(t, fields, "expected nil when --json passed without field list")
}

func TestJSONFieldsFromCmd_FlagWithFields_ReturnsSlice(t *testing.T) {
	t.Parallel()
	// Use --json=field1,field2 (equals form) since NoOptDefVal is set;
	// the space form (--json field1,field2) would treat field1,field2 as a
	// positional arg and use the NoOptDefVal sentinel instead.
	cmd := newCmdWithOutputFlags()
	cmd.SetArgs([]string{"--json=field1,field2"})
	require.NoError(t, cmd.Execute())
	fields := cmdutil.JSONFieldsFromCmd(cmd)
	assert.Equal(t, []string{"field1", "field2"}, fields)
}
