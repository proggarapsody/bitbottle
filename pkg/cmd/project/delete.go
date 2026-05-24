package project

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdDelete builds `project delete KEY [--confirm] [--hostname HOST]`.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete KEY",
		Short: "Delete a project on Bitbucket Server",
		Long: `Delete a project on Bitbucket Server / Data Center.

Non-interactive use requires --confirm.
Bitbucket Cloud returns an unsupported error.

Examples:
  bitbottle project delete PRJ --confirm --hostname git.example.com
  bitbottle project delete PRJ`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			if !confirm {
				proceed, err := confirmProjectDelete(f, key)
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(f.IOStreams.Out, "Deletion aborted.")
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

			pc, err := backend.AsServerProjectClient(client, host)
			if err != nil {
				return err
			}

			if err := pc.DeleteServerProject(key); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Deleted project %s\n", key)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func confirmProjectDelete(f *factory.Factory, key string) (bool, error) {
	if !f.IOStreams.IsStdoutTTY() {
		return false, fmt.Errorf("requires --confirm to delete a project")
	}
	fmt.Fprintf(f.IOStreams.ErrOut, "Delete project %s? [y/N] ", key)

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
