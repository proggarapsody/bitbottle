package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoView(f *factory.Factory) *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO]",
		Short: "View a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}

			host, err := resolveHostname(f, "")
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			repo, err := client.GetRepo(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			// --web: open in browser and skip text output.
			if web {
				if repo.WebURL == "" {
					return fmt.Errorf("no web URL available for this repository")
				}
				return f.Browser.Browse(repo.WebURL)
			}

			fmt.Fprintf(f.IOStreams.Out, "Name:      %s\n", repo.Slug)
			fmt.Fprintf(f.IOStreams.Out, "Namespace: %s\n", repo.Namespace)
			fmt.Fprintf(f.IOStreams.Out, "SCM:       %s\n", repo.SCM)
			if repo.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "URL:       %s\n", repo.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	return cmd
}
