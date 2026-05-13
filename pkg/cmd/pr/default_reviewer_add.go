package pr

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRDefaultReviewerAdd builds the `pr default-reviewer add` cobra command.
func NewCmdPRDefaultReviewerAdd(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO] USER",
		Short: "Add a default reviewer to a repository",
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
			return dr.AddDefaultReviewer(ref.Project, ref.Slug, userSlug)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
