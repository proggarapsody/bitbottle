package branch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdBranchDelete(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO BRANCH",
		Short: "Delete a branch",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := resolveRef(f, args[0], hostname)
			if err != nil {
				return err
			}
			branchName := args[1]

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
