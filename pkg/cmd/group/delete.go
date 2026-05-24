package group

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdDelete builds `group delete NAME [--confirm] [--hostname HOST]`.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete an admin group on Bitbucket Server/DC",
		Long: `Delete a Bitbucket Server/DC admin group.

Non-interactive use requires --confirm.

Examples:
  bitbottle group delete oldgroup --confirm
  bitbottle group delete oldgroup`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			if !confirm {
				proceed, err := confirmGroupDelete(f, name)
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(f.IOStreams.Out, "Deletion aborted.")
					return nil
				}
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			gc, err := backend.AsGroupClient(client, host)
			if err != nil {
				return err
			}

			if err := gc.DeleteGroup(name); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Deleted group %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func confirmGroupDelete(f *factory.Factory, name string) (bool, error) {
	if !f.IOStreams.IsStdoutTTY() {
		return false, fmt.Errorf("requires --confirm to delete a group")
	}
	fmt.Fprintf(f.IOStreams.ErrOut, "Are you sure you want to delete group %s? [y/N] ", name)

	scanner := bufio.NewScanner(f.IOStreams.In)
	var answer string
	if scanner.Scan() {
		answer = strings.TrimSpace(scanner.Text())
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
