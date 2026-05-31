package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// ValidateOutputFlags enforces the output-format contract shared by every
// command: exactly one of --json / --yaml / --template may be active, and
// --jq is only meaningful alongside --json. It also disables color on the
// IOStreams when any structured mode is selected so consumers get raw values.
//
// This logic lives here (not inline in the root command's PersistentPreRunE)
// because cobra runs only the single deepest PersistentPreRunE in a command
// chain. Command groups that install their own PersistentPreRunE (e.g. via
// factory.EnableRepoOverride for the -R flag) would otherwise shadow the
// root's hook and silently skip these checks — the FMT-CONTRACT bug where
// `pr view --jq .x` was accepted without --json while `status --jq .x`
// correctly errored. Both call sites now invoke this one function.
//
// The flag lookups tolerate missing flags (returning zero values) so callers
// that construct a subcommand in isolation — without the root's persistent
// flags — do not panic.
func ValidateOutputFlags(c *cobra.Command, ios *iostreams.IOStreams) error {
	jsonMode := flagChanged(c, "json")
	yamlMode := flagBool(c, "yaml")
	tmpl := flagString(c, "template")
	jqExpr := flagString(c, "jq")

	if jsonMode && yamlMode {
		return fmt.Errorf("--json and --yaml are mutually exclusive")
	}
	if jsonMode && tmpl != "" {
		return fmt.Errorf("--json and --template are mutually exclusive")
	}
	if yamlMode && tmpl != "" {
		return fmt.Errorf("--yaml and --template are mutually exclusive")
	}
	if jqExpr != "" && !jsonMode {
		return fmt.Errorf("--jq requires --json")
	}

	if ios != nil && (jsonMode || yamlMode || tmpl != "") {
		ios.SetColorEnabled(false)
	}
	return nil
}

func flagChanged(c *cobra.Command, name string) bool {
	if c.Flags().Lookup(name) == nil {
		return false
	}
	return c.Flags().Changed(name)
}

func flagBool(c *cobra.Command, name string) bool {
	if c.Flags().Lookup(name) == nil {
		return false
	}
	v, _ := c.Flags().GetBool(name)
	return v
}

func flagString(c *cobra.Command, name string) string {
	if c.Flags().Lookup(name) == nil {
		return ""
	}
	v, _ := c.Flags().GetString(name)
	return v
}
