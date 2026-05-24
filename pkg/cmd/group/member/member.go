// Package member implements the `group member` subcommand group.
package member

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMember builds the `group member` subcommand group.
func NewCmdMember(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage members of a Bitbucket Server/DC admin group",
	}
	cmd.AddCommand(NewCmdMemberList(f))
	cmd.AddCommand(NewCmdMemberAdd(f))
	cmd.AddCommand(NewCmdMemberRemove(f))
	return cmd
}

// resolveHostname returns the hostname for group member commands.
func resolveHostname(f *factory.Factory, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return hosts[0], nil
	default:
		return "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
	}
}
