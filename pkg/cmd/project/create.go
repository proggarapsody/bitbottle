package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCreate builds `project create KEY --name NAME [--description TEXT] [--public]`.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var hostname string
	var name string
	var description string
	var public bool

	cmd := &cobra.Command{
		Use:   "create KEY",
		Short: "Create a new project on Bitbucket Server",
		Long: `Create a new project on a Bitbucket Server / Data Center host.

Bitbucket Cloud returns an unsupported error.

Examples:
  bitbottle project create PRJ --name "My Project" --hostname git.example.com
  bitbottle project create DEV --name "Dev" --description "Dev project" --public`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			pc, err := backend.AsServerProjectClient(client, host)
			if err != nil {
				return err
			}

			p, err := pc.CreateServerProject(backend.CreateServerProjectInput{
				Key:         key,
				Name:        name,
				Description: description,
				Public:      public,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Created project %s (%s)\n", p.Key, p.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&name, "name", "", "Project display name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	cmd.Flags().BoolVar(&public, "public", false, "Make project publicly accessible")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
