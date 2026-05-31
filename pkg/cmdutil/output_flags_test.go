package cmdutil_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// cmdWithOutputFlags builds a leaf command under a root carrying the four
// output-format persistent flags, parses args, and returns the leaf so tests
// exercise the same inherited-flag shape production uses.
func cmdWithOutputFlags(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "root"}
	format.RegisterOutputFlags(root)
	leaf := &cobra.Command{Use: "leaf", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(leaf)
	root.SetArgs(append([]string{"leaf"}, args...))
	// Parse without executing RunE side effects beyond flag binding.
	require.NoError(t, root.ParseFlags(append([]string{"leaf"}, args...)))
	// ParseFlags binds to root; re-parse on the leaf so Changed/Get reflect the
	// inherited persistent flags exactly as cobra presents them in PreRunE.
	require.NoError(t, leaf.ParseFlags(args))
	return leaf
}

func TestValidateOutputFlags_JQRequiresJSON(t *testing.T) {
	t.Parallel()
	c := cmdWithOutputFlags(t, "--jq", ".title")
	err := cmdutil.ValidateOutputFlags(c, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--jq requires --json")
}

func TestValidateOutputFlags_JQWithJSON_OK(t *testing.T) {
	t.Parallel()
	c := cmdWithOutputFlags(t, "--json", "--jq", ".title")
	require.NoError(t, cmdutil.ValidateOutputFlags(c, nil))
}

func TestValidateOutputFlags_Mutex(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"json_yaml", []string{"--json", "--yaml"}, "--json and --yaml are mutually exclusive"},
		{"json_template", []string{"--json", "--template", "{{.}}"}, "--json and --template are mutually exclusive"},
		{"yaml_template", []string{"--yaml", "--template", "{{.}}"}, "--yaml and --template are mutually exclusive"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := cmdWithOutputFlags(t, tc.args...)
			err := cmdutil.ValidateOutputFlags(c, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestValidateOutputFlags_DisablesColorForStructured(t *testing.T) {
	t.Parallel()
	c := cmdWithOutputFlags(t, "--json")
	ios := iostreams.Test()
	ios.SetColorEnabled(true)
	require.NoError(t, cmdutil.ValidateOutputFlags(c, ios))
	assert.False(t, ios.ColorEnabled(), "structured output must disable color")
}

func TestValidateOutputFlags_NoFlags_OK(t *testing.T) {
	t.Parallel()
	// A bare command with no output flags registered must not panic.
	c := &cobra.Command{Use: "bare"}
	require.NoError(t, cmdutil.ValidateOutputFlags(c, nil))
}
