package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRepoSetDefaultBranch builds the `repo set-default-branch` command.
func NewCmdRepoSetDefaultBranch(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "set-default-branch BRANCH [PROJECT/REPO]",
		Short: "Set the default branch of a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]

			var ns, slug string
			if len(args) == 2 {
				ref, err := bbrepo.Parse(args[1])
				if err != nil {
					return err
				}
				ns, slug = ref.Project, ref.Slug
			} else {
				base, err := f.BaseRepo()
				if err != nil {
					return err
				}
				ns, slug = base.Project, base.Slug
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			if err := client.SetRepoDefaultBranch(ns, slug, branch); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Default branch of %s/%s set to %q.\n", ns, slug, branch)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
