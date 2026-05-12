package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoCreate(f *factory.Factory) *cobra.Command {
	var hostname, project, description string
	var private bool

	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return fmt.Errorf("--project is required")
			}

			name := args[0]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			r, err := client.CreateRepo(project, backend.CreateRepoInput{
				Name:        name,
				SCM:         "git",
				Public:      !private,
				Description: description,
			})
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := repoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(r)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Created repository %s/%s\n", r.Namespace, r.Slug)
			if r.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", r.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&project, "project", "", "Project key")
	cmd.Flags().StringVar(&description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&private, "private", true, "Make repository private")
	return cmd
}
