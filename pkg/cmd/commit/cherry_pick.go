package commit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCherryPick(f *factory.Factory) *cobra.Command {
	var hostname string
	var message string

	cmd := &cobra.Command{
		Use:   "cherry-pick HASH TARGET_BRANCH [PROJECT/REPO]",
		Short: "Cherry-pick a commit onto a branch",
		Long: `Cherry-pick a commit onto a branch.

Requires Bitbucket Server / Data Center with the branch-utils plugin.`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash := args[0]
			targetBranch := args[1]

			repoArgs := args[2:]
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			cp, err := backend.AsCommitCherryPicker(client, ref.Host)
			if err != nil {
				return err
			}

			result, err := cp.CherryPickCommit(ref.Project, ref.Slug, backend.CherryPickInput{
				SourceHash:   hash,
				TargetBranch: targetBranch,
				Message:      message,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "%s\n", result.Hash)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message override (default: original message)")
	return cmd
}
