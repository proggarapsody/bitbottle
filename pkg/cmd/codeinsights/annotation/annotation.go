// Package annotation is the `code-insights annotation` subcommand group.
package annotation

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdAnnotation builds the `code-insights annotation` command tree.
func NewCmdAnnotation(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "annotation",
		Short: "Manage Code Insights annotations",
	}
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdAdd(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))
	return cmd
}
