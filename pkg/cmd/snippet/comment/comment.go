// Package comment implements the `bitbottle snippet comment` subcommand group.
package comment

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdSnippetComment builds the `snippet comment` parent command.
func NewCmdSnippetComment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage comments on a snippet",
	}
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdAdd(f))
	cmd.AddCommand(NewCmdDelete(f))
	return cmd
}
