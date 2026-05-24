package member

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMemberRemove builds `group member remove NAME USER [--confirm] [--hostname HOST]`.
func NewCmdMemberRemove(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "remove NAME USER",
		Short: "Remove a user from a Bitbucket Server/DC admin group",
		Long: `Remove a user from a Bitbucket Server/DC admin group.

Non-interactive use requires --confirm.

Examples:
  bitbottle group member remove developers alice --confirm
  bitbottle group member remove qa-team bob`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]
			user := args[1]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			if !confirm {
				proceed, err := confirmMemberRemove(f, groupName, user)
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(f.IOStreams.Out, "Removal aborted.")
					return nil
				}
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			gmc, err := backend.AsGroupMemberClient(client, host)
			if err != nil {
				return err
			}

			if err := gmc.RemoveGroupMember(groupName, user); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Removed %s from group %s\n", user, groupName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func confirmMemberRemove(f *factory.Factory, groupName, user string) (bool, error) {
	if !f.IOStreams.IsStdoutTTY() {
		return false, fmt.Errorf("requires --confirm to remove a user from a group")
	}
	fmt.Fprintf(f.IOStreams.ErrOut, "Are you sure you want to remove %s from group %s? [y/N] ",
		user, groupName)

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
