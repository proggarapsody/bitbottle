package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoRename(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "rename PROJECT/REPO NEW-NAME",
		Short: "Rename a repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}
			newName := args[1]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			updated, err := client.RenameRepo(ref.Project, ref.Slug, newName)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Renamed %s/%s to %s/%s\n",
				ref.Project, ref.Slug, updated.Namespace, updated.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
