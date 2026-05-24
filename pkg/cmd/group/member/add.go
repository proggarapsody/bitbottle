package member

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMemberAdd builds `group member add NAME USER [--hostname HOST]`.
func NewCmdMemberAdd(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "add NAME USER",
		Short: "Add a user to a Bitbucket Server/DC admin group",
		Long: `Add a user to a Bitbucket Server/DC admin group.

Examples:
  bitbottle group member add developers alice
  bitbottle group member add qa-team bob --hostname git.example.com`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]
			user := args[1]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			gmc, err := backend.AsGroupMemberClient(client, host)
			if err != nil {
				return err
			}

			if err := gmc.AddGroupMember(groupName, user); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Added %s to group %s\n", user, groupName)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
