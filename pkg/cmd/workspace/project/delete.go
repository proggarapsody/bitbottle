package project

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete WORKSPACE KEY",
		Short: "Delete a workspace project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, key := args[0], args[1]
			if !confirm {
				if !f.IOStreams.IsStdoutTTY() {
					return fmt.Errorf("--confirm required when not running interactively")
				}
				fmt.Fprintf(f.IOStreams.Out, "Delete project %s? [y/N]: ", key)
				reader := bufio.NewReader(f.IOStreams.In)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(f.IOStreams.Out, "Aborted.")
					return nil
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
			pc, err := backend.AsCloudProjectClient(client, host)
			if err != nil {
				return err
			}
			if err := pc.DeleteWorkspaceProject(ws, key); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deleted project %s.\n", key)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
