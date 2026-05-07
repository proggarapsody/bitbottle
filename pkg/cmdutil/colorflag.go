package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// NoColorFlag is the persistent flag name that disables color output. Defined
// here so root registration and any consumer (e.g. PreRunE wiring,
// integration tests) reference the same string.
const NoColorFlag = "no-color"

// RegisterNoColorFlag adds the persistent --no-color flag to root. The default
// is false so the flag's presence alone is the disable signal.
func RegisterNoColorFlag(root *cobra.Command) {
	root.PersistentFlags().Bool(NoColorFlag, false,
		"Disable colored output (also honors NO_COLOR env)")
}

// ApplyNoColorFlag reads --no-color off the command's flag set and, if set,
// flips the IOStreams color decision off. Designed for use inside
// PersistentPreRunE so a single registration applies across every subcommand.
//
// Cobra walks PersistentFlags up the parent chain when resolving a flag, so
// passing --no-color on a leaf command (e.g. `bitbottle pr list --no-color`)
// still surfaces here. We deliberately do NOT call SetColorEnabled(true) when
// the flag is unset — leaving IOStreams as constructed lets NO_COLOR env and
// TTY detection keep their say.
func ApplyNoColorFlag(cmd *cobra.Command, ios *iostreams.IOStreams) {
	noColor, err := cmd.Flags().GetBool(NoColorFlag)
	if err != nil {
		// Flag not registered — nothing to apply. This path keeps the helper
		// safe to call from any command, even ones whose tree doesn't include
		// the flag (defensive; production wires it on root for everyone).
		return
	}
	if noColor {
		ios.SetColorEnabled(false)
	}
}
