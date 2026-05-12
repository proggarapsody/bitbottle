package root_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// TestRootPersistentOutputFlags_Mutex verifies that the four output-format
// flags registered on the root in OUT2 reject incompatible combinations in
// PersistentPreRunE before any subcommand runs.
func TestRootPersistentOutputFlags_Mutex(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "json_and_yaml",
			args: []string{"status", "--json", "--yaml"},
			want: "--json and --yaml are mutually exclusive",
		},
		{
			name: "json_and_template",
			args: []string{"status", "--json", "--template", "{{.}}"},
			want: "--json and --template are mutually exclusive",
		},
		{
			name: "yaml_and_template",
			args: []string{"status", "--yaml", "--template", "{{.}}"},
			want: "--yaml and --template are mutually exclusive",
		},
		{
			name: "jq_without_json",
			args: []string{"status", "--jq", ".foo"},
			want: "--jq requires --json",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
			cmd := root.NewCmdRoot(f)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestRootPersistentOutputFlags_Registered(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
	cmd := root.NewCmdRoot(f)
	for _, name := range []string{"json", "yaml", "jq", "template"} {
		assert.NotNil(t, cmd.PersistentFlags().Lookup(name), "expected persistent flag %q to be registered", name)
	}
}
