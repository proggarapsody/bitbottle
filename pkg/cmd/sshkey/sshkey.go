// Package sshkey implements the `bitbottle ssh-key` command group.
// User SSH keys grant SSH access to Bitbucket Cloud. This is a Cloud-only feature.
package sshkey

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmd)
}

// NewCmd builds the root `ssh-key` command.
func NewCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage user SSH keys",
	}
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdAdd(f))
	cmd.AddCommand(NewCmdDelete(f))
	return cmd
}
