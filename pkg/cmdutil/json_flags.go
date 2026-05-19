// Package cmdutil provides CLI command utilities including JSON output helpers.
package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
)

// JSONFieldsFromCmd returns the parsed JSON field list from the --json flag.
// Returns nil when --json was not passed or passed without a field list (= all fields).
// Returns a non-nil slice when --json field1,field2 was used.
func JSONFieldsFromCmd(cmd *cobra.Command) []string {
	return format.ConfigFromCmd(cmd).JSONFields
}
