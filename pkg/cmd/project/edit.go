package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdEdit builds `project edit KEY [--name NAME] [--description TEXT] [--public=BOOL]`.
func NewCmdEdit(f *factory.Factory) *cobra.Command {
	var hostname string
	var name string
	var description string
	// Use string pointer flags so we can detect "was set"
	var nameSet, descriptionSet, publicSet bool
	var public bool

	cmd := &cobra.Command{
		Use:   "edit KEY",
		Short: "Edit a project on Bitbucket Server",
		Long: `Edit properties of a project on Bitbucket Server / Data Center.

At least one of --name, --description, or --public must be provided.
Bitbucket Cloud returns an unsupported error.

Examples:
  bitbottle project edit PRJ --name "New Name" --hostname git.example.com
  bitbottle project edit PRJ --description "Updated desc" --public=true`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			nameSet = cmd.Flags().Changed("name")
			descriptionSet = cmd.Flags().Changed("description")
			publicSet = cmd.Flags().Changed("public")

			if !nameSet && !descriptionSet && !publicSet {
				return &backend.DomainError{
					Kind:    backend.ErrInvalidRequest,
					Code:    backend.CodeInvalidRequest,
					Message: "at least one of --name, --description, or --public must be provided",
				}
			}

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

			in := backend.UpdateServerProjectInput{}
			if nameSet {
				in.Name = &name
			}
			if descriptionSet {
				in.Description = &description
			}
			if publicSet {
				in.Public = &public
			}

			_, err = pc.UpdateServerProject(key, in)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Updated project %s\n", key)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&name, "name", "", "New project display name")
	cmd.Flags().StringVar(&description, "description", "", "New project description")
	cmd.Flags().BoolVar(&public, "public", false, "Set project public visibility")
	return cmd
}
