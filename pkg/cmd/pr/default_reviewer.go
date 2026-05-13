package pr

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRDefaultReviewer builds the `pr default-reviewer` subcommand group.
func NewCmdPRDefaultReviewer(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default-reviewer",
		Short: "Manage PR default reviewers for a repository",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO (Server) or
WORKSPACE/REPO (Cloud). When omitted, the repository is inferred from
the "origin" git remote in the current directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdPRDefaultReviewerList(f))
	cmd.AddCommand(NewCmdPRDefaultReviewerAdd(f))
	cmd.AddCommand(NewCmdPRDefaultReviewerRemove(f))
	return cmd
}
