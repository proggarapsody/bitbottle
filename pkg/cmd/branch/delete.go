package branch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

func NewCmdBranchDelete(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] BRANCH",
		Short: "Delete a branch",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			branchName := rest[0]

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			if err := client.DeleteBranch(ref.Project, ref.Slug, branchName); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Deleted branch %s\n", branchName)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
