package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdEdit(f *factory.Factory) *cobra.Command {
	var hostname, name, description string
	var isPrivate bool

	cmd := &cobra.Command{
		Use:   "edit WORKSPACE KEY",
		Short: "Edit a workspace project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, key := args[0], args[1]
			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			pc, err := backend.AsCloudProjectClient(client, host)
			if err != nil {
				return err
			}
			input := backend.UpdateWorkspaceProjectInput{}
			if cmd.Flags().Changed("name") {
				input.Name = &name
			}
			if cmd.Flags().Changed("description") {
				input.Description = &description
			}
			if cmd.Flags().Changed("private") {
				input.IsPrivate = &isPrivate
			}
			if _, err := pc.UpdateWorkspaceProject(ws, key, input); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Updated project %s.\n", key)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New project name")
	cmd.Flags().StringVar(&description, "description", "", "New project description")
	cmd.Flags().BoolVar(&isPrivate, "private", false, "Set project private flag")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
