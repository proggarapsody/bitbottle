package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var hostname string
	var key, name, description string
	var isPrivate bool

	cmd := &cobra.Command{
		Use:   "create WORKSPACE",
		Short: "Create a workspace project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := args[0]
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
			_, err = pc.CreateWorkspaceProject(ws, backend.CreateWorkspaceProjectInput{
				Key:         key,
				Name:        name,
				Description: description,
				IsPrivate:   isPrivate,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Created project %s in %s.\n", key, ws)
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Project key (e.g. MYPROJ)")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringVar(&name, "name", "", "Project name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	cmd.Flags().BoolVar(&isPrivate, "private", false, "Make project private")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
