package commit

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCommitComment returns the parent "commit comment" command group.
func NewCmdCommitComment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage comments on a commit",
	}
	cmd.AddCommand(NewCmdCommitCommentList(f))
	cmd.AddCommand(NewCmdCommitCommentAdd(f))
	cmd.AddCommand(NewCmdCommitCommentEdit(f))
	cmd.AddCommand(NewCmdCommitCommentDelete(f))
	return cmd
}
