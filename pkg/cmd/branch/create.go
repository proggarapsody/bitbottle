package branch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdBranchCreate(f *factory.Factory) *cobra.Command {
	var startAt string
	var hostname string

	cmd := &cobra.Command{
		Use:   "create PROJECT/REPO NAME",
		Short: "Create a new branch",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if startAt == "" {
				return fmt.Errorf("required flag \"start-at\" not set")
			}

			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}
			name := args[1]

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			_, err = client.CreateBranch(ref.Project, ref.Slug, backend.CreateBranchInput{
				Name:    name,
				StartAt: startAt,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Created branch %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&startAt, "start-at", "", "Branch name or commit hash to start from (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
