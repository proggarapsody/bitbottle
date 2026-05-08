// Package report is the `code-insights report` subcommand group.
package report

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdReport builds the `code-insights report` command tree.
func NewCmdReport(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Manage Code Insights reports",
	}
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdView(f, nil))
	cmd.AddCommand(NewCmdSet(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))
	return cmd
}
