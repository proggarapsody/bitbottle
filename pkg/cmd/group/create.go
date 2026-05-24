package group

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCreate builds `group create NAME [--hostname HOST]`.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create an admin group on Bitbucket Server/DC",
		Long: `Create a Bitbucket Server/DC admin group.

Examples:
  bitbottle group create developers
  bitbottle group create qa-team --hostname git.example.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			gc, err := backend.AsGroupClient(client, host)
			if err != nil {
				return err
			}

			g, err := gc.CreateGroup(name)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Created group %s\n", g.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
