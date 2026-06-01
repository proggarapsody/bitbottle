package branch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

func NewCmdBranchCreate(f *factory.Factory) *cobra.Command {
	var startAt string
	var hostname string

	cmd := &cobra.Command{
		Use:   "create [PROJECT/REPO] NAME [START_AT]",
		Short: "Create a new branch",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			name := rest[0]

			// positional START_AT wins only when --start-at flag was not set
			if startAt == "" && len(rest) > 1 {
				startAt = rest[1]
			}
			if startAt == "" {
				return fmt.Errorf("start-at is required (pass it as the third positional or --start-at flag)")
			}

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
	cmd.Flags().StringVar(&startAt, "start-at", "", "Branch name or commit hash to start from")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
