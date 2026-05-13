package pr

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRDefaultReviewerRemove builds the `pr default-reviewer remove` cobra command.
func NewCmdPRDefaultReviewerRemove(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "remove [PROJECT/REPO] USER",
		Short: "Remove a default reviewer from a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var repoArgs []string
			var userSlug string
			if len(args) == 2 {
				repoArgs = args[:1]
				userSlug = args[1]
			} else {
				userSlug = args[0]
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			dr, err := backend.AsDefaultReviewerClient(client, ref.Host)
			if err != nil {
				return err
			}
			return dr.RemoveDefaultReviewer(ref.Project, ref.Slug, userSlug)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
