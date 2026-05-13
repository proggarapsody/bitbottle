package sshkey

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdDelete builds the `ssh-key delete` cobra command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete an SSH key for the current user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil || id <= 0 {
				return fmt.Errorf("ID must be a positive integer, got %q", args[0])
			}
			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			sk, err := backend.AsSSHKeyClient(client, host)
			if err != nil {
				return err
			}
			if err := sk.DeleteSSHKey(id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "SSH key %d deleted\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
