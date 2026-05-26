// Package defaultreviewer implements the `bitbottle workspace project default-reviewer`
// command group. Workspace project default reviewers are a Bitbucket Cloud concept;
// the optional WorkspaceProjectDefaultReviewerClient interface gates these commands.
package defaultreviewer

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdWorkspaceProjectDefaultReviewer returns the `default-reviewer` sub-group
// under `workspace project`.
func NewCmdWorkspaceProjectDefaultReviewer(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default-reviewer",
		Short: "Manage workspace project default reviewers (Cloud only)",
	}
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdAdd(f, nil))
	cmd.AddCommand(NewCmdRemove(f, nil))
	return cmd
}
